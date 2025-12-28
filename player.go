package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/exec"

	"github.com/adrg/xdg"
)

type finishPlayMsg struct{}

var (
	errAlreadyPlaying   = errors.New("a file is already being played")
	errPlayerNotRunning = errors.New("player is not running")
)

type player struct {
	playing bool
	stopFn  func() error
}

func newPlayer() *player {
	return &player{}
}

func (p *player) play(filePath string) error {
	if p.playing {
		return errAlreadyPlaying
	}

	logFile, err := xdg.StateFile("ytqueue/mpv.log")
	if err != nil {
		return err
	}

	f, err := os.Create(logFile) // #nosec G304
	if err != nil {
		return err
	}

	cmd := exec.CommandContext(
		context.Background(),
		"mpv",
		"--keep-open=no",
		"--idle=no",
		filePath,
	) // #nosec G204
	cmd.Stdout = f
	p.stopFn = cmd.Cancel
	p.playing = true

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
	if !p.playing {
		return nil
	}

	if p.stopFn == nil {
		return errPlayerNotRunning
	}

	defer func() {
		p.playing = false
		p.stopFn = nil
	}()

	if err := p.stopFn(); err != nil {
		return err
	}

	return nil
}
