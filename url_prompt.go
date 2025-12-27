package main

import (
	"net/url"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type enqueueURLMsg struct {
	url string
}

type urlPrompt struct {
	prompt textinput.Model
	style  lipgloss.Style
	keymap promptKeymap
}

func newURLPrompt() *urlPrompt {
	i := textinput.New()
	i.Placeholder = "Enter URL here..."
	i.Focus()

	style := lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder())

	return &urlPrompt{i, style, newPromptKeymap()}
}

func (p *urlPrompt) Init() tea.Cmd {
	return textinput.Blink
}

func (p *urlPrompt) enqueueURLCmd(url string) tea.Cmd {
	return func() tea.Msg {
		return enqueueURLMsg{url}
	}
}

func (p *urlPrompt) Update(msg tea.Msg) (*urlPrompt, tea.Cmd) {
	var cmds []tea.Cmd

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
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, p.keymap.clear):
			p.prompt.Reset()
		case key.Matches(msg, p.keymap.submit):
			filmUrl, err := url.Parse(p.prompt.Value())
			if err != nil {
				return p, errorCmd(err)
			}

			cmds = append(cmds, p.enqueueURLCmd(filmUrl.String()))

			p.prompt.Reset()
		}
	}

	var cmd tea.Cmd
	p.prompt, cmd = p.prompt.Update(msg)
	cmds = append(cmds, cmd)

	return p, tea.Batch(cmds...)
}

func (p *urlPrompt) View() string {
	return p.style.Render(p.prompt.View())
}
