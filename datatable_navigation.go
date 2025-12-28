package main

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
	d.nameTruncateLeft = 0
	d.scrollUp()
}

func (d *datatable) lineDown(n int) {
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
