package main

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

func (d *datatable) scrollUp() {
	if d.cursor < d.viewport.YOffset {
		d.viewport.SetYOffset(d.cursor)
	}
}

func (d *datatable) scrollDown() {
	if d.cursor >= d.viewport.YOffset+d.viewport.Height {
		d.viewport.SetYOffset(d.cursor - d.viewport.Height + 1)
	}
}

func (d *datatable) lineUp(n int) {
	d.deleteConfirm = false
	d.cursor = clamp(d.cursor-n, 0, len(d.rows)-1)
	d.scrollUp()
}

func (d *datatable) lineDown(n int) {
	d.deleteConfirm = false
	d.cursor = clamp(d.cursor+n, 0, len(d.rows)-1)
	d.scrollDown()
}

func (d *datatable) pageUp() {
	d.lineUp(d.viewport.Height)
}

func (d *datatable) pageDown() {
	d.lineDown(d.viewport.Height)
}

const halfPageFactor = 2

func (d *datatable) halfPageUp() {
	d.lineUp(d.viewport.Height / halfPageFactor)
}

func (d *datatable) halfPageDown() {
	d.lineDown(d.viewport.Height / halfPageFactor)
}

func (d *datatable) gotoTop() {
	d.viewport.GotoTop()
	d.cursor = 0
}

func (d *datatable) gotoBottom() {
	d.viewport.GotoBottom()
	d.cursor = len(d.rows) - 1
}

func (d *datatable) moveUp() {
	upperIdx := clamp(d.cursor-1, 0, len(d.rows)-1)
	d.rows[d.cursor], d.rows[upperIdx] = d.rows[upperIdx], d.rows[d.cursor]
	d.cursor = upperIdx
	d.scrollUp()
	d.updateViewport()
}

func (d *datatable) moveDown() {
	lowerIdx := clamp(d.cursor+1, 0, len(d.rows)-1)
	d.rows[d.cursor], d.rows[lowerIdx] = d.rows[lowerIdx], d.rows[d.cursor]
	d.cursor = lowerIdx
	d.scrollDown()
	d.updateViewport()
}

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
	case key.Matches(msg, d.keymap.deleteRow):
		if d.deleteConfirm {
			cmd = d.deleteCurrentRow(d.cursor)
		}

		d.deleteConfirm = !d.deleteConfirm
	}

	return cmd
}
