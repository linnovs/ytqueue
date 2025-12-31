package main

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

func (d *datatable) keyMsgHandler(msg tea.KeyMsg) tea.Cmd {
	d.cursorMu.Lock()
	defer d.cursorMu.Unlock()

	var cmd tea.Cmd

	if !d.isFocused {
		return cmd
	}

	const nameScrollAmount = 5

	switch {
	case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
		d.deleteConfirm = false
	case key.Matches(msg, d.keymap.lineUp):
		d.lineUp(1)
	case key.Matches(msg, d.keymap.lineDown):
		d.lineDown(1)
	case key.Matches(msg, d.keymap.nameScrollLeft):
		d.nameScrollLeft(nameScrollAmount)
	case key.Matches(msg, d.keymap.nameScrollRight):
		d.nameScrollRight(nameScrollAmount)
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
	case key.Matches(msg, d.keymap.cursor2middle):
		d.cursor2middle()
	case key.Matches(msg, d.keymap.moveUp):
		cmd = d.moveUp()
	case key.Matches(msg, d.keymap.moveDown):
		cmd = d.moveDown()
	case key.Matches(msg, d.keymap.playOrStop):
		cmd = d.playStopRowCmd(d.getIDAtIndex(d.cursor))
	case key.Matches(msg, d.keymap.toggleWatched):
		cmd = d.toggleWatchedStatusCmd(d.cursor)
	case key.Matches(msg, d.keymap.deleteRow):
		if d.deleteConfirm {
			cmd = d.deleteRowCmd(d.cursor)
		}

		d.deleteConfirm = !d.deleteConfirm
	case key.Matches(msg, d.keymap.refresh):
		cmd = d.refreshRowsCmd()
	case key.Matches(msg, d.keymap.copyURL):
		cmd = d.copyURLCmd(d.cursor)
	case key.Matches(msg, d.keymap.pasteURL):
		cmd = d.pasteURLCmd()
	}

	return cmd
}
