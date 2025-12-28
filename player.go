package main

import (
	"log/slog"
	"os"
	"path/filepath"
	"sync"

	"github.com/adrg/xdg"
	tea "github.com/charmbracelet/bubbletea"
)

type (
	finishPlayingMsg  struct{}
	stoppedPlayingMsg struct{ id string }
)

type player struct {
	program            *tea.Program
	playingMu          *sync.Mutex
	playing            bool
	currentlyPlayingId string
	processMu          *sync.Mutex
	process            *os.Process
	sockPath           string
	commandCh          chan []string
}

func newPlayer() *player {
	sockPath, err := xdg.RuntimeFile("ytqueue/mpv.sock")
	if err != nil {
		slog.Error("unable to get mpv socket path", slog.String("error", err.Error()))
		panic(err)
	}

	sockPath = filepath.Clean(sockPath)
	commandCh := make(chan []string)

	return &player{
		playingMu: new(sync.Mutex),
		processMu: new(sync.Mutex),
		sockPath:  sockPath,
		commandCh: commandCh,
	}
}

func (p *player) setProgram(program *tea.Program) {
	p.program = program
}

func (p *player) play(filePath, id string) error {
	if !p.isRunning() {
		if err := p.startPlayer(); err != nil {
			return err
		}
	}

	defer p.setPlaying(true, id)

	slog.Debug("sending loadfile command to mpv", slog.String("file", filePath))

	return p.sendMPVCommand("loadfile", filePath, "replace")
}

func (p *player) stop() error {
	if !p.isRunning() || !p.isPlaying() {
		return nil
	}

	defer p.setPlaying(false)

	return p.sendMPVCommand("quit")
}

func (p *player) quit() tea.Cmd {
	return func() tea.Msg {
		if !p.isRunning() {
			return nil
		}

		if err := p.sendMPVCommand("quit"); err != nil {
			return errorMsg{err}
		}

		return nil
	}
}
