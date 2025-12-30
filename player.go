package main

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/adrg/xdg"
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type (
	finishPlayingMsg   struct{}
	playbackChangedMsg struct{}
	updateProgressMsg  struct{ percent float64 }
)

type player struct {
	program            *tea.Program
	playingMu          *sync.Mutex
	playing            bool
	currentlyPlayingId string
	processMu          *sync.Mutex
	process            *os.Process
	sockPath           string
	commandCh          chan []any
	playtime           time.Duration
	playtimeRemaining  time.Duration
	progress           progress.Model
}

func newPlayer() *player {
	sockPath, err := xdg.RuntimeFile(fmt.Sprintf("ytqueue/mpv.%d.sock", os.Getpid()))
	if err != nil {
		slog.Error("unable to get mpv socket path", slog.String("error", err.Error()))
		panic(err)
	}

	sockPath = filepath.Clean(sockPath)
	commandCh := make(chan []any)

	return &player{
		playingMu: new(sync.Mutex),
		processMu: new(sync.Mutex),
		sockPath:  sockPath,
		commandCh: commandCh,
		progress:  progress.New(progress.WithDefaultGradient(), progress.WithoutPercentage()),
	}
}

func (p *player) setProgram(program *tea.Program) {
	p.program = program
}

func (p *player) renderPlayProgress(width int) string {
	w := lipgloss.Width
	playtime := renderPlaytime(p.playtime)
	remaining := renderPlaytimeRemaining(p.playtimeRemaining)
	p.progress.Width = width - w(playtime) - w(remaining)
	playProgress := p.progress.View()

	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		playtime,
		playProgress,
		remaining,
	)
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
