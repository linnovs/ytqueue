package main

import (
	"time"

	"github.com/charmbracelet/lipgloss"
)

func formatPlaytime(d time.Duration) string {
	var start time.Time

	return start.Add(d).Format("15:04:05")
}

func renderPlaytime(d time.Duration) string {
	return lipgloss.NewStyle().Padding(0, 1).Render(formatPlaytime(d))
}

func renderPlaytimeRemaining(d time.Duration) string {
	return lipgloss.NewStyle().Padding(0, 1).SetString("-").Render(formatPlaytime(d))
}
