package main

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type runningTextTickMsg struct{}

func doRunningTextTickCmd() tea.Cmd {
	const interval = time.Millisecond * 100

	return tea.Tick(interval, func(time.Time) tea.Msg {
		return runningTextTickMsg{}
	})
}

type runningTextModel struct {
	style  lipgloss.Style
	text   string
	width  int
	offset int
}

func newRunningTextModel(text string, width int, style lipgloss.Style) *runningTextModel {
	return &runningTextModel{text: text, width: width, style: style}
}

func (r *runningTextModel) Init() tea.Cmd {
	return doRunningTextTickCmd()
}

func (r *runningTextModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	if _, ok := msg.(runningTextTickMsg); ok {
		r.offset = (r.offset + 1) % max(len(r.text), r.width)
		cmd = doRunningTextTickCmd()
	}

	return r, cmd
}

func (r *runningTextModel) View() string {
	fullText := []rune(r.text + "     " + r.text)
	display := string(fullText[r.offset : r.offset+r.width])

	return r.style.Render(display)
}
