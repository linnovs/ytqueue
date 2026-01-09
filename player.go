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

type playingStatus int

const (
	playingStatusStopped playingStatus = iota
	playingStatusPlaying
	playingStatusPaused
)

const playingStatusLength = 9

func (s playingStatus) String() string {
	return [...]string{"STOPPED", "PLAYING", "PAUSED"}[s]
}

type player struct {
	program                  *tea.Program
	playingMu                *sync.RWMutex
	playing                  playingStatus
	currentlyPlayingId       string
	currentlyPlayingFilename string
	processMu                *sync.RWMutex
	process                  *os.Process
	sockPath                 string
	commandCh                chan []any
	playtimeMu               *sync.RWMutex
	playtime                 time.Duration
	playtimeRemaining        time.Duration
	progress                 progress.Model
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
		playingMu:  new(sync.RWMutex),
		playtimeMu: new(sync.RWMutex),
		processMu:  new(sync.RWMutex),
		sockPath:   sockPath,
		commandCh:  commandCh,
		progress:   progress.New(progress.WithDefaultGradient(), progress.WithoutPercentage()),
	}
}

func (p *player) setProgram(program *tea.Program) {
	p.program = program
}

func (p *player) renderPlayProgress(width int) string {
	w := lipgloss.Width
	playtime := renderPlaytime(p.getPlaytime())
	playStatus := renderPlayingStatus(p.getPlaying())
	remaining := renderPlaytimeRemaining(p.getRemainingTime())
	p.progress.Width = width - w(playStatus) - w(playtime) - w(remaining)
	playProgress := p.progress.View()

	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		playStatus,
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

	defer p.setPlaying(playingStatusPlaying, id)

	slog.Debug("sending loadfile command to mpv", slog.String("file", filePath))

	if err := p.sendMPVCommand("loadfile", filePath, "replace"); err != nil {
		slog.Error("failed to send loadfile command to mpv", slog.String("error", err.Error()))

		return err
	}

	return p.sendMPVCommand("set_property", "pause", false)
}

func (p *player) stop() error {
	if !p.isRunning() || !p.isPlaying() {
		return nil
	}

	defer p.setPlaying(playingStatusStopped)

	return p.sendMPVCommand("stop")
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
