package main

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

func (d *datatable) keyMsgHandler(msg tea.KeyMsg) tea.Cmd {
	var cmd tea.Cmd

	if !d.isFocused {
		return cmd
	}

	switch {
	case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
		d.deleteConfirm = false
	case key.Matches(msg, d.keymap.lineUp):
		d.lineUp(1)
	case key.Matches(msg, d.keymap.lineDown):
		d.lineDown(1)
	case key.Matches(msg, d.keymap.pageUp):
		d.pageUp()
	case key.Matches(msg, d.keymap.pageDown):
		d.pageDown()
	case key.Matches(msg, d.keymap.halfPageUp):
		d.halfPageUp()
	case key.Matches(msg, d.keymap.halfPageDown):
		d.halfPageDown()
	case key.Matches(msg, d.keymap.gotoTop):
		d.gotoTop()
	case key.Matches(msg, d.keymap.gotoBottom):
		d.gotoBottom()
	case key.Matches(msg, d.keymap.moveUp):
		d.moveUp()
	case key.Matches(msg, d.keymap.moveDown):
		d.moveDown()
	case key.Matches(msg, d.keymap.playFromRow):
	case key.Matches(msg, d.keymap.stopPlaying):
	case key.Matches(msg, d.keymap.toggleWatched):
		cmd = d.toggleWatchedStatusCmd(d.cursor)
	case key.Matches(msg, d.keymap.deleteRow):
		if d.deleteConfirm {
			cmd = d.deleteCurrentRow(d.cursor)
		}

		d.deleteConfirm = !d.deleteConfirm
	}

	return cmd
}
