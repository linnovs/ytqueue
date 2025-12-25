package main

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type urlPrompt struct {
	prompt textinput.Model
	style  lipgloss.Style
}

func newURLPrompt() *urlPrompt {
	i := textinput.New()
	i.Placeholder = "Enter URL here..."
	i.Focus()

	style := lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder())

	return &urlPrompt{i, style}
}

func (p *urlPrompt) Init() tea.Cmd {
	return textinput.Blink
}

func (p *urlPrompt) Update(msg tea.Msg) (*urlPrompt, tea.Cmd) {
	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		p.prompt.Width = msg.Width - p.style.GetHorizontalFrameSize()
		p.prompt.Width -= lipgloss.Width(p.prompt.Prompt) + 1
	}

	var cmd tea.Cmd
	p.prompt, cmd = p.prompt.Update(msg)

	return p, cmd
}

func (p *urlPrompt) View() string {
	return p.style.Render(p.prompt.View())
}

func (p *urlPrompt) Focus() {
	p.prompt.Focus()
}

func (p *urlPrompt) Reset() {
	p.prompt.Reset()
}
