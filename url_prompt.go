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
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		p.prompt.Width = msg.Width - p.style.GetHorizontalFrameSize()
		p.prompt.Width -= lipgloss.Width(p.prompt.Prompt) + 1
	case sectionChangedMsg:
		if msg.section == sectionURLPrompt {
			p.prompt.Focus()
		} else {
			p.prompt.Blur()
		}
	}

	var cmd tea.Cmd
	p.prompt, cmd = p.prompt.Update(msg)

	return p, cmd
}

func (p *urlPrompt) View() string {
	return p.style.Render(p.prompt.View())
}
