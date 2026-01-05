package main

import (
	"slices"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func (d *datatable) clampCursor(cursor int) int {
	return clamp(cursor, 0, len(d.rows)-1)
}

func (d *datatable) clampCursorInViewport(cursor int) int {
	return clamp(cursor, 0, len(d.rows)-d.viewport.Height)
}

func (d *datatable) scrollUp() {
	if d.cursor < d.viewport.YOffset {
		d.viewport.SetYOffset(d.cursor)
	}
}

func (d *datatable) scrollDown() {
	if d.cursor >= d.viewport.YOffset+d.viewport.Height {
		d.viewport.SetYOffset(d.clampCursorInViewport(d.cursor - d.viewport.Height + 1))
	}
}

func (d *datatable) scrollToTop() {
	if d.viewport.YOffset-d.viewport.Height != d.cursor {
		d.viewport.SetYOffset(d.cursor)
	}
}

func (d *datatable) scrollToBottom() {
	if d.viewport.YOffset+d.viewport.Height-1 != d.cursor {
		d.viewport.SetYOffset(d.clampCursorInViewport(d.cursor - d.viewport.Height + 1))
	}
}

func (d *datatable) moveCursor(n int) {
	d.rowMu.RLock()
	defer d.rowMu.RUnlock()

	d.deleteConfirm = false
	d.cursor = d.clampCursor(d.cursor + n)
	d.nameTruncateLeft = 0

	if n < 0 {
		d.scrollUp()
	} else {
		d.scrollDown()
	}
}

func (d *datatable) lineUp(n int) {
	d.moveCursor(-n)
}

func (d *datatable) lineDown(n int) {
	d.moveCursor(n)
}

func (d *datatable) nameScrollLeft(n int) {
	d.nameTruncateLeft = clamp(d.nameTruncateLeft-n, 0, d.widths[colName])
}

func (d *datatable) nameScrollRight(n int) {
	d.nameTruncateLeft = clamp(d.nameTruncateLeft+n, 0, d.widths[colName])
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
	d.cursor = 0
	d.viewport.GotoTop()
}

func (d *datatable) gotoBottom() {
	d.rowMu.RLock()
	defer d.rowMu.RUnlock()

	d.cursor = len(d.rows) - 1
	d.viewport.GotoBottom()
}

func (d *datatable) gotoPlaying() {
	d.rowMu.RLock()
	defer d.rowMu.RUnlock()

	if !d.player.isPlaying() {
		return
	}

	idx := slices.IndexFunc(d.rows, playingIDIndexFunc(d.player.getCurrentlyPlayingId()))
	d.cursor = d.clampCursor(idx)
	d.cursor2middle()
}

func (d *datatable) cursor2middle() {
	d.viewport.YOffset = (d.clampCursorInViewport(d.cursor - (d.viewport.Height / 2)))
}

func (d *datatable) moveRow(n int) tea.Cmd {
	const moveRowDebounceDuration = time.Millisecond * 300

	d.rowMu.Lock()
	defer d.rowMu.Unlock()

	d.updateRowOrderTagMu.Lock()
	d.updateRowOrderTag++
	tag := d.updateRowOrderTag
	d.updateRowOrderTagMu.Unlock()

	nextIdx := d.clampCursor(d.cursor + n)
	d.rows[d.cursor], d.rows[nextIdx] = d.rows[nextIdx], d.rows[d.cursor]
	id := d.rows[nextIdx][colID]
	d.cursor = nextIdx

	if n > 0 {
		d.scrollDown()
	} else {
		d.scrollUp()
	}

	return tea.Tick(moveRowDebounceDuration, func(_ time.Time) tea.Msg {
		return updateRowOrderMsg{
			id:  id,
			tag: tag,
		}
	})
}

func (d *datatable) moveUp() tea.Cmd {
	return d.moveRow(-1)
}

func (d *datatable) moveDown() tea.Cmd {
	return d.moveRow(1)
}
