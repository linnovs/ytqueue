package main

import "github.com/charmbracelet/bubbles/key"

type keymap struct {
	quit, next, prev, submit key.Binding
}

func newKeymap() keymap {
	return keymap{
		quit: key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q/ctrl+c", "quit")),
		next: key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next section")),
		prev: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "previous section"),
		),
		submit: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "submit URL")),
	}
}
