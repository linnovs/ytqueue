package main

import (
	"context"
	"errors"
	"log/slog"
	"os/exec"
	"sync/atomic"

	"github.com/adrg/xdg"
	tea "github.com/charmbracelet/bubbletea"
)

type finishPlayMsg struct{}

var (
	errAlreadyPlaying   = errors.New("a file is already being played")
	errPlayerNotRunning = errors.New("player is not running")
)

type player struct {
	playing *atomic.Bool
	stopFn  *atomic.Value
}

func newPlayer() *player {
	return &player{playing: new(atomic.Bool), stopFn: new(atomic.Value)}
}

func (p *player) play(filePath string) error {
	if p.playing.Load() {
		return errAlreadyPlaying
	}

	logFile, err := xdg.StateFile("ytqueue/mpv.log")
	if err != nil {
		return err
	}

	cmd := exec.CommandContext(
		context.Background(),
		"mpv",
		"--log-file="+logFile,
		"--keep-open=no",
		"--idle=no",
		filePath,
	) // #nosec G204
	p.stopFn.Store(cmd.Cancel)
	p.playing.Store(true)

	defer func() {
		p.playing.Store(false)
		p.stopFn.Store(nil)
	}()

	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			slog.Error(
				"player exited with error",
				slog.String("stderr", string(exitErr.Stderr)),
				slog.Int("exitCode", exitErr.ExitCode()),
			)

			return nil
		}

		return err
	}

	return nil
}

func (p *player) stop() error {
	if !p.playing.Load() {
		return nil
	}

	stopFn := p.stopFn.Load().(func() error)

	if stopFn == nil {
		return errPlayerNotRunning
	}

	defer func() {
		p.playing.Store(false)
		p.stopFn.Store(nil)
	}()

	if err := stopFn(); err != nil {
		return err
	}

	return nil
}

func (p *player) quit() tea.Cmd {
	return func() tea.Msg {
		stopFn := p.stopFn.Load().(func() error)

		if stopFn == nil {
			return nil
		}

		p.playing.Store(false)

		if err := stopFn(); err != nil {
			slog.Error("failed to stop player on quit", slog.String("error", err.Error()))
		}

		return nil
	}
}
