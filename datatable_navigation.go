package main

import (
	"fmt"
	"log/slog"
	"slices"
	"strconv"

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
	d.rowMu.RLock()
	defer d.rowMu.RUnlock()

	d.deleteConfirm = false
	d.cursor = clamp(d.cursor-n, 0, len(d.rows)-1)
	d.nameTruncateLeft = 0
	d.scrollUp()
}

func (d *datatable) lineDown(n int) {
	d.rowMu.RLock()
	defer d.rowMu.RUnlock()

	d.deleteConfirm = false
	d.cursor = clamp(d.cursor+n, 0, len(d.rows)-1)
	d.nameTruncateLeft = 0
	d.scrollDown()
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

func (d *datatable) cursor2middle() {
	d.viewport.SetYOffset(clamp(d.cursor-(d.viewport.Height/2), 0, len(d.rows)-d.viewport.Height))
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

		return nil
	}
}

func (d *datatable) moveUp() tea.Cmd {
	d.rowMu.RLock()
	upperIdx := clamp(d.cursor-1, 0, len(d.rows)-1)
	d.rowMu.RUnlock()

	d.rowMu.Lock()
	d.rows[d.cursor], d.rows[upperIdx] = d.rows[upperIdx], d.rows[d.cursor]
	d.rowMu.Unlock()

	d.cursor = upperIdx
	d.scrollUp()
	d.cursorMu.Unlock() // temporary unlock to avoid deadlock in updateViewport
	d.updateViewport()
	d.cursorMu.Lock()

	return d.updateRowOrderCmd(d.getIDAtIndex(d.cursor))
}

func (d *datatable) moveDown() tea.Cmd {
	d.rowMu.RLock()
	lowerIdx := clamp(d.cursor+1, 0, len(d.rows)-1)
	d.rowMu.RUnlock()

	d.rowMu.Lock()
	d.rows[d.cursor], d.rows[lowerIdx] = d.rows[lowerIdx], d.rows[d.cursor]
	d.rowMu.Unlock()

	d.cursor = lowerIdx
	d.scrollDown()
	d.cursorMu.Unlock() // temporary unlock to avoid deadlock in updateViewport
	d.updateViewport()
	d.cursorMu.Lock()

	return d.updateRowOrderCmd(d.getIDAtIndex(d.cursor))
}
