package main

import (
	"time"

	"github.com/charmbracelet/lipgloss"
)

func formatPlaytime(d time.Duration) string {
	var start time.Time

	return start.Add(d).Format("15:04:05")
}

func renderPlayingStatus(status playingStatus) string {
	var color string

	switch status {
	case playingStatusPlaying:
		color = "120" // Green
	case playingStatusPaused:
		color = "214" // Yellow
	case playingStatusStopped:
		color = "160" // Red
	}

	return lipgloss.NewStyle().
		Width(playingStatusLength).
		Padding(0, 1).
		AlignHorizontal(lipgloss.Center).
		Bold(true).
		Foreground(lipgloss.Color("231")).
		Background(lipgloss.Color(color)).
		Render(status.String())
}

func renderPlaytime(d time.Duration) string {
	return lipgloss.NewStyle().Padding(0, 1).Render(formatPlaytime(d))
}

func renderPlaytimeRemaining(d time.Duration) string {
	return lipgloss.NewStyle().Padding(0, 1).SetString("-").Render(formatPlaytime(d))
}
