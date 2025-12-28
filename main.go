package main

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/linnovs/ytqueue/database"
)

const fileperm = 0o600

func runApp() int {
	logFile, err := xdg.StateFile("ytqueue/ytqueue.log")
	if err != nil {
		slog.Error("unable to get log file path", slog.String("error", err.Error()))
		return 1
	}

	teaLogFile, err := xdg.StateFile("ytqueue/tea.log")
	if err != nil {
		slog.Error("unable to get tea log file path", slog.String("error", err.Error()))
		return 1
	}

	logFile = filepath.Clean(logFile)
	teaLogFile = filepath.Clean(teaLogFile)

	teaLog, err := tea.LogToFile(teaLogFile, "")
	if err != nil {
		slog.Error("unable to open tea log file", slog.String("error", err.Error()))
		return 1
	}

	defer teaLog.Close() // nolint:errcheck

	out, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, fileperm)
	if err != nil {
		slog.Error("unable to open log file", slog.String("error", err.Error()))
		return 1
	}

	cfg, err := loadConfig()
	if err != nil {
		slog.Error("unable to load config", slog.String("error", err.Error()))
		return 1
	}

	defer cleanupTempDir(cfg.tempDir)

	db, err := openAndMigrateDB()
	if err != nil {
		slog.Error("database migration failed", slog.String("error", err.Error()))
		return 1
	}

	ctx, cancel := context.WithCancel(context.Background())

	queries, err := database.Prepare(ctx, db)
	if err != nil {
		slog.Error("unable to prepare database queries", slog.String("error", err.Error()))
		return 1
	}

	handler := slog.NewTextHandler(out, &slog.HandlerOptions{Level: slog.LevelDebug})
	slog.SetDefault(slog.New(handler))


	d := newDownloader(cfg)
	p := tea.NewProgram(newModel(d, ctx, cancel, queries, cfg), tea.WithAltScreen())
	d.setProgram(p)

	go d.start(ctx)

	if _, err := p.Run(); err != nil {
		slog.Error("application crashed", slog.String("error", err.Error()))
		return 1
	}

	return 0
}

func main() {
	os.Exit(runApp())
}
