package main

import (
	"log/slog"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	tea "github.com/charmbracelet/bubbletea"
)

const fileperm = 0o600

func main() {
	logFile, err := xdg.StateFile("ytqueue/ytqueue.log")
	if err != nil {
		slog.Error("unable to get log file path", slog.String("error", err.Error()))
		os.Exit(1)
	}

	logFile = filepath.Clean(logFile)

	out, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, fileperm)
	if err != nil {
		slog.Error("unable to open log file", slog.String("error", err.Error()))
		os.Exit(1)
	}

	handler := slog.NewTextHandler(out, &slog.HandlerOptions{Level: slog.LevelDebug})
	slog.SetDefault(slog.New(handler))

	p := tea.NewProgram(newModel(), tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		slog.Error("application crashed", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
