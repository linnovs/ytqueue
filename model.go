package main

import (
	"fmt"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type appModel struct {
	height    int
	width     int
	keymap    keymap
	help      help.Model
	urlPrompt *urlPrompt
	section   int
}

func newModel() appModel {
	prompt := newURLPrompt()
	prompt.Focus()

	return appModel{
		keymap:    newKeymap(),
		help:      help.New(),
		urlPrompt: prompt,
		section:   0,
	}
}

// Init implements [tea.Model].
func (m appModel) Init() tea.Cmd {
	return nil
}

// Update implements [tea.Model].
func (m appModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keymap.quit):
			return m, tea.Quit
		case key.Matches(msg, m.keymap.next):
			m.section++
		case key.Matches(msg, m.keymap.prev):
			m.section--
		case key.Matches(msg, m.keymap.submit):
			m.urlPrompt.Reset()
		}
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width
		m.urlPrompt.setWidth(msg.Width)
	}

	newPrompt, cmd := m.urlPrompt.Update(msg)
	m.urlPrompt = newPrompt

	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// View implements [tea.Model].
func (m appModel) View() string {
	header := fmt.Sprintf("Width: %d, Height: %d\n", m.width, m.height)
	urlPromptView := fmt.Sprintf("%s\n", m.urlPrompt.View())
	sectionInfo := fmt.Sprintf("Section: %d\n", m.section)

	return lipgloss.JoinVertical(lipgloss.Center, header, urlPromptView, sectionInfo)
}
