package main

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type errorMsg struct {
	err error
}

const defaultErrorTimeout = time.Second * 3

func errorCmd(err error) tea.Cmd {
	return func() tea.Msg {
		return errorMsg{err}
	}
}

type resetErrorMsg struct {
	err error
}

func resetErrorMsgCmd(err error, timeoutArg ...time.Duration) tea.Cmd {
	timeout := defaultErrorTimeout

	if len(timeoutArg) >= 1 {
		timeout = timeoutArg[0]
	}

	return tea.Tick(timeout, func(t time.Time) tea.Msg {
		return resetErrorMsg{err}
	})
}

func newErrorStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "196", Dark: "124"})
}
