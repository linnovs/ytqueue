package main

import "github.com/charmbracelet/bubbles/key"

type promptKeymap struct {
	clear, submit key.Binding
}

type datatableKeymap struct {
	moveUp, moveDown key.Binding
}

type keymap struct {
	help, quit, next, prev key.Binding
	prompt                 promptKeymap
	datatable              datatableKeymap
	section                sectionType
}

func newKeymap() keymap {
	return keymap{
		section: sectionURLPrompt,
		help:    key.NewBinding(key.WithKeys("f1"), key.WithHelp("F1", "show help")),
		quit:    key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q/ctrl+c", "quit")),
		next:    key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next section")),
		prev: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "previous section"),
		),
		prompt: promptKeymap{
			clear:  key.NewBinding(key.WithKeys("ctrl+l"), key.WithHelp("ctrl+l", "clear URL")),
			submit: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "submit URL")),
		},
		datatable: datatableKeymap{
			moveUp: key.NewBinding(key.WithKeys("K"), key.WithHelp("shift+k", "move row up")),
			moveDown: key.NewBinding(
				key.WithKeys("J"),
				key.WithHelp("shift+j", "move row down"),
			),
		},
	}
}

func (m keymap) ShortHelp() []key.Binding {
	return []key.Binding{}
}

func (m keymap) FullHelp() [][]key.Binding {
	base := []key.Binding{
		m.help, m.next, m.prev, m.quit,
	}

	switch m.section {
	case sectionURLPrompt:
		return [][]key.Binding{
			base,
			{m.prompt.clear, m.prompt.submit},
		}
	case sectionDatatable:
		return [][]key.Binding{
			base,
			{m.datatable.moveUp, m.datatable.moveDown},
		}
	default:
		return [][]key.Binding{}
	}
}
