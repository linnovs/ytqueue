package main

import (
	"fmt"
	"log/slog"
	"slices"
	"strconv"

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
		d.viewport.SetYOffset(d.clampCursor(d.cursor - d.viewport.Height + 1))
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
	d.viewport.GotoTop()
	d.cursor = 0
}

func (d *datatable) gotoBottom() {
	d.rowMu.RLock()
	defer d.rowMu.RUnlock()

	d.viewport.GotoBottom()
	d.cursor = len(d.rows) - 1
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
	d.viewport.SetYOffset(d.clampCursorInViewport(d.cursor - (d.viewport.Height / 2)))
}

func (d *datatable) updateRowOrderCmd(id string) tea.Cmd {
	return func() tea.Msg {
		d.rowMu.Lock()
		defer d.rowMu.Unlock()

		slog.Debug("updating row order", slog.String("id", id))

		idx := slices.IndexFunc(d.rows, func(r row) bool { return r[colID] == id })
		upperIdx := clamp(idx-1, 0, len(d.rows)-1)
		lowerIdx := clamp(idx+1, 0, len(d.rows)-1)

		upperOrderUnix, err := strconv.Atoi(d.rows[upperIdx][colOrder])
		if err != nil {
			return errorMsg{fmt.Errorf("failed to parse upper order index: %w", err)}
		}

		lowerOrderUnix, err := strconv.Atoi(d.rows[lowerIdx][colOrder])
		if err != nil {
			return errorMsg{fmt.Errorf("failed to parse lower order index: %w", err)}
		}

		const inHalf = 2
		newOrderUnix := int64((upperOrderUnix + lowerOrderUnix) / inHalf)

		if err := d.datastore.updateVideoOrder(d.getCtx(), id, newOrderUnix); err != nil {
			return errorMsg{fmt.Errorf("failed to update video order: %w", err)}
		}

		slog.Debug(
			"updated row order",
			slog.String("id", id),
			slog.Int64("new_order_unix", newOrderUnix),
		)

		return updateRowOrderMsg{}
	}
}

func (d *datatable) moveRow(n int) tea.Cmd {
	d.rowMu.Lock()
	defer d.rowMu.Unlock()

	nextIdx := d.clampCursor(d.cursor + n)
	d.rows[d.cursor], d.rows[nextIdx] = d.rows[nextIdx], d.rows[d.cursor]
	id := d.rows[nextIdx][colID]
	d.cursor = nextIdx

	if n > 0 {
		d.scrollDown()
	} else {
		d.scrollUp()
	}

	return d.updateRowOrderCmd(id)
}

func (d *datatable) moveUp() tea.Cmd {
	return d.moveRow(-1)
}

func (d *datatable) moveDown() tea.Cmd {
	return d.moveRow(1)
}
