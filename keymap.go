package main

import "github.com/charmbracelet/bubbles/key"

type baseKeymap struct {
	help, quit, next, prev key.Binding
}

func (b baseKeymap) Help() []key.Binding {
	return []key.Binding{b.help, b.quit, b.next, b.prev}
}

func newBaseKeymap() baseKeymap {
	return baseKeymap{
		help: key.NewBinding(key.WithKeys("f1"), key.WithHelp("F1", "show help")),
		quit: key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q/ctrl+c", "quit")),
		next: key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next section")),
		prev: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "previous section"),
		),
	}
}

type promptKeymap struct {
	baseKeymap
	clear, submit key.Binding
}

func (p promptKeymap) ShortHelp() []key.Binding {
	return []key.Binding{}
}

func (p promptKeymap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		p.Help(),
		{p.clear, p.submit},
	}
}

type datatableKeymap struct {
	baseKeymap
	lineUp, lineDown, moveUp, moveDown         key.Binding
	pageUp, pageDown, halfPageUp, halfPageDown key.Binding
	gotoTop, gotoBottom                        key.Binding
}

func (d datatableKeymap) ShortHelp() []key.Binding {
	return []key.Binding{}
}

func (d datatableKeymap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		d.Help(),
		{d.lineUp, d.lineDown, d.moveUp, d.moveDown},
		{d.pageUp, d.pageDown, d.halfPageUp, d.halfPageDown},
		{d.gotoTop, d.gotoBottom},
	}
}

type keymap struct {
	baseKeymap
	prompt    promptKeymap
	datatable datatableKeymap
}

func newKeymap() keymap {
	return keymap{
		baseKeymap: newBaseKeymap(),
		prompt:     newPromptKeymap(),
		datatable:  newDatatableKeymap(),
	}
}

func newPromptKeymap() promptKeymap {
	return promptKeymap{
		baseKeymap: newBaseKeymap(),
		clear:      key.NewBinding(key.WithKeys("ctrl+l"), key.WithHelp("ctrl+l", "clear URL")),
		submit:     key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "submit URL")),
	}
}

func newDatatableKeymap() datatableKeymap {
	return datatableKeymap{
		baseKeymap: newBaseKeymap(),
		lineUp: key.NewBinding(
			key.WithKeys("k", "up"),
			key.WithHelp("↑/k", "move cursor up"),
		),
		lineDown: key.NewBinding(
			key.WithKeys("j", "down"),
			key.WithHelp("↓/j", "move cursor down"),
		),
		moveUp: key.NewBinding(key.WithKeys("K"), key.WithHelp("shift+k", "move row up")),
		moveDown: key.NewBinding(
			key.WithKeys("J"),
			key.WithHelp("shift+j", "move row down"),
		),
		pageUp:   key.NewBinding(key.WithKeys("ctrl+f"), key.WithHelp("ctrl+f", "page up")),
		pageDown: key.NewBinding(key.WithKeys("ctrl+b"), key.WithHelp("ctrl+b", "page down")),
		halfPageUp: key.NewBinding(
			key.WithKeys("ctrl+u"),
			key.WithHelp("ctrl+u", "half page up"),
		),
		halfPageDown: key.NewBinding(
			key.WithKeys("ctrl+d"),
			key.WithHelp("ctrl+d", "half page down"),
		),
		gotoTop:    key.NewBinding(key.WithKeys("g"), key.WithHelp("g", "go to top")),
		gotoBottom: key.NewBinding(key.WithKeys("G"), key.WithHelp("G", "go to bottom")),
	}
}
