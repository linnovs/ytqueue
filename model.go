package main

import tea "github.com/charmbracelet/bubbletea"

type model struct{}

func newModel() model {
	return model{}
}

// Init implements [tea.Model].
func (m model) Init() tea.Cmd {
	panic("unimplemented")
}

// Update implements [tea.Model].
func (m model) Update(tea.Msg) (tea.Model, tea.Cmd) {
	panic("unimplemented")
}

// View implements [tea.Model].
func (m model) View() string {
	panic("unimplemented")
}
