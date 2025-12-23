package main

import "github.com/charmbracelet/bubbles/key"

type keymap struct {
	help, quit, next, prev, clear, submit key.Binding
}

func newKeymap() keymap {
	return keymap{
		help: key.NewBinding(key.WithKeys("f1"), key.WithHelp("F1", "show help")),
		quit: key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q/ctrl+c", "quit")),
		next: key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next section")),
		prev: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "previous section"),
		),
		clear:  key.NewBinding(key.WithKeys("ctrl+l"), key.WithHelp("ctrl+l", "clear URL")),
		submit: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "submit URL")),
	}
}

func (m keymap) ShortHelp() []key.Binding {
	return []key.Binding{}
}

func (m keymap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{m.next, m.prev, m.quit},
		{m.clear, m.submit},
	}
}
