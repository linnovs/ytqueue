package main

import (
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/compat"
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
	return lipgloss.NewStyle().
		Foreground(compat.AdaptiveColor{Light: lipgloss.Color("196"), Dark: lipgloss.Color("124")})
}
