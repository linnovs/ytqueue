package main

import (
	"strings"
	"time"
	"unicode/utf8"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

type runningTextTickMsg struct{}

func doRunningTextTickCmd() tea.Cmd {
	const interval = time.Millisecond * 100

	return tea.Tick(interval, func(time.Time) tea.Msg {
		return runningTextTickMsg{}
	})
}

type runningTextFullTextUpdateMsg struct {
	text string
}

type runningTextModel struct {
	style    lipgloss.Style
	fullText []rune
	textLen  int
	width    int
	offset   int
}

const emptyRunningText = "     "

func newRunningTextModel(width int, style lipgloss.Style) *runningTextModel {
	return &runningTextModel{
		fullText: []rune{' '},
		textLen:  1,
		width:    width,
		style:    style,
	}
}

func (r *runningTextModel) Init() tea.Cmd {
	return doRunningTextTickCmd()
}

func (r *runningTextModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case runningTextFullTextUpdateMsg:
		r.fullText = []rune(msg.text + emptyRunningText + msg.text)
		r.textLen = utf8.RuneCountInString(msg.text + emptyRunningText)
	case runningTextTickMsg:
		r.offset = (r.offset + 1) % r.textLen
		cmd = doRunningTextTickCmd()
	}

	return r, cmd
}

func (r *runningTextModel) updateText(text string) tea.Cmd {
	return func() tea.Msg {
		return runningTextFullTextUpdateMsg{text: text}
	}
}

func (r *runningTextModel) View() string {
	width := 0
	var display strings.Builder

	for _, ch := range r.fullText[r.offset:] {
		width += runewidth.RuneWidth(ch)
		if width > r.width {
			break
		}

		display.WriteRune(ch)
	}

	return r.style.Width(r.width + r.style.GetHorizontalPadding()).Render(display.String())
}
