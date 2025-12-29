package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/adrg/xdg"
)

func closeMPVConn(conn net.Conn) {
	if err := conn.Close(); err != nil {
		slog.Error("failed to close mpv socket connection", slog.String("error", err.Error()))
	}
}

type mpvCommand struct {
	Command []string `json:"command"`
}

type mpvEvent struct {
	Event     string `json:"event"`
	Error     string `json:"error,omitempty"`
	RequestID *int   `json:"request_id,omitempty"`
	Reason    string `json:"reason,omitempty"`
}

func (p *player) sendMPVCommand(command ...string) error {
	commandStr := strings.Join(command, " ")

	p.commandCh <- command

	slog.Debug("mpv command sent", slog.String("command", commandStr))

	return nil
}

func (p *player) readMPVEvents(conn net.Conn) {
	scanner := bufio.NewScanner(conn)

	for scanner.Scan() {
		var msg mpvEvent

		if err := json.Unmarshal(scanner.Bytes(), &msg); err != nil {
			slog.Error(
				"failed to unmarshal mpv event",
				slog.String("error", err.Error()),
				slog.String("data", scanner.Text()),
			)

			continue
		}

		switch {
		case msg.RequestID != nil:
			slog.Debug(
				"mpv command response received",
				slog.Int("request_id", *msg.RequestID),
				slog.String("data", scanner.Text()),
			)
		case msg.Error != "":
			slog.Error(
				"mpv event error",
				slog.String("event", msg.Event),
				slog.String("error", msg.Error),
				slog.String("data", scanner.Text()),
			)
		case msg.Event != "":
			slog.Debug("mpv event received", slog.String("event", msg.Event))

			switch msg.Event {
			case "file-loaded":
				slog.Debug("mpv playback started", slog.String("id", p.getCurrentlyPlayingId()))
				p.setPlaying(true, p.getCurrentlyPlayingId())
			case "end-file":
				slog.Debug(
					"mpv playback ended",
					slog.String("reason", msg.Reason),
					slog.String("id", p.getCurrentlyPlayingId()),
				)

				if msg.Reason != "eof" {
					p.program.Send(stoppedPlayingMsg{p.getCurrentlyPlayingId()})
				} else {
					p.program.Send(finishPlayingMsg{})
				}

				p.setPlaying(false)
			}
		}
	}

	if scanner.Err() != nil {
		slog.Error("error reading from mpv socket", slog.String("error", scanner.Err().Error()))
	}
}

func (p *player) writeMPVCommands(conn net.Conn) {
	for cmd := range p.commandCh {
		msg := mpvCommand{Command: cmd}

		data, err := json.Marshal(msg)
		if err != nil {
			slog.Error(
				"failed to marshal mpv command",
				slog.String("error", err.Error()),
				slog.Any("command", msg),
			)

			continue
		}

		if _, err := conn.Write(append(data, '\n')); err != nil {
			if errors.Is(err, net.ErrClosed) {
				go func() {
					p.commandCh <- cmd
				}()

				return
			}

			slog.Error(
				"failed to write mpv command to socket",
				slog.String("error", err.Error()),
				slog.Any("command", msg),
			)
		}
	}
}

func (p *player) createMPVConn(ctx context.Context, wg *sync.WaitGroup) {
	var conn net.Conn
	var err error

	for {
		conn, err = net.Dial("unix", p.sockPath)
		if err != nil {
			continue
		}

		const delay = 100 * time.Millisecond
		<-time.After(delay)

		break
	}

	defer closeMPVConn(conn)

	wg.Done()

	go p.readMPVEvents(conn)
	go p.writeMPVCommands(conn)

	<-ctx.Done()
}

func (p *player) monitorProcess(cmd *exec.Cmd, wg *sync.WaitGroup) {
	slog.Debug("mpv player started", slog.Int("pid", cmd.Process.Pid))

	p.processMu.Lock()
	p.process = cmd.Process
	p.processMu.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go p.createMPVConn(ctx, wg)

	if err := cmd.Wait(); err != nil {
		slog.Error("mpv player exited with error", slog.String("error", err.Error()))
	}

	slog.Debug("mpv player exited", slog.Int("pid", cmd.Process.Pid))
}

func (p *player) startPlayer() error {
	logPath, err := xdg.StateFile("ytqueue/mpv.log")
	if err != nil {
		return err
	}

	f, err := os.Create(filepath.Clean(logPath))
	if err != nil {
		return err
	}

	cmd := exec.Command(
		"mpv",
		"--save-position-on-quit",
		"--keep-open=yes",
		"--idle=yes",
		"--input-ipc-server="+p.sockPath, // #nosec G204
	)
	cmd.Stdout = f
	cmd.Stderr = f

	if err := cmd.Start(); err != nil {
		return err
	}

	var wg sync.WaitGroup
	wg.Add(1)

	go p.monitorProcess(cmd, &wg)
	wg.Wait()

	return nil
}
