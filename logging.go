package main

import (
	"bufio"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

type logMsg string

type logging struct {
	visible  bool
	style    lipgloss.Style
	header   string
	result   *strings.Builder
	logger   io.Reader
	viewport viewport.Model
	ch       chan string
}

func newLogging(logger io.Reader) *logging {
	var result strings.Builder

	style := lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder())
	header := lipgloss.NewStyle().
		Background(lipgloss.Color("240")).
		Foreground(lipgloss.Color("15")).
		Bold(true).
		Padding(0, 1).
		Render("Logs")

	const logHeight = 5

	return &logging{
		logger:   logger,
		ch:       make(chan string),
		viewport: viewport.New(0, logHeight),
		result:   &result,
		style:    style,
		header:   header,
	}
}

func (l *logging) fetchLogsCmd() tea.Cmd {
	l.appendLog("Waiting for logs...")

	return func() tea.Msg {
		scanner := bufio.NewScanner(l.logger)

		for {
			for scanner.Scan() {
				l.ch <- scanner.Text()
			}

			if scanner.Err() != nil {
				l.appendLog("Error reading logs: " + scanner.Err().Error())
			}
		}
	}
}

func (l *logging) waitForLogMsg() tea.Cmd {
	return func() tea.Msg {
		return logMsg(<-l.ch)
	}
}

func (l *logging) Init() tea.Cmd {
	return tea.Batch(l.fetchLogsCmd(), l.waitForLogMsg())
}

func (l *logging) appendLog(msg string) {
	const maxLogSize = 10 * 1024 // 10 KB

	if l.result.Len() > maxLogSize {
		l.result.Reset()
	}

	l.result.WriteString("\n" + msg)
	resultStr := ansi.Hardwrap(l.result.String(), l.viewport.Width, false)
	result := lipgloss.NewStyle().Width(l.viewport.Width).Render(resultStr)

	l.viewport.SetContent(result)
	l.viewport.GotoBottom()
}

func (l *logging) Update(msg tea.Msg) (*logging, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		l.viewport.Width = msg.Width - l.style.GetHorizontalFrameSize()
		l.style = l.style.Width(msg.Width - l.style.GetHorizontalFrameSize())

		return l, nil
	case logMsg:
		l.appendLog(string(msg))
		cmds = append(cmds, l.waitForLogMsg())
	}

	return l, tea.Batch(cmds...)
}

func (l *logging) View() string {
	if l.visible {
		return l.style.Render(lipgloss.JoinVertical(lipgloss.Top, l.header, l.viewport.View()))
	}

	return ""
}
