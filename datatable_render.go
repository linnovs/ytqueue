package main

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

func (d *datatable) renderHeader() string {
	var s strings.Builder

	for _, col := range d.columns {
		style := lipgloss.NewStyle().Width(d.widths[col]).Padding(0, dtCellPadding)
		s.WriteString(style.Render(string(col)))
	}

	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, true).
		Render(s.String())
}

// this function assumes the caller holds the rowMu Rlock.
func (d *datatable) renderRow(r int) string {
	d.cursorMu.RLock()
	defer d.cursorMu.RUnlock()

	var s strings.Builder

	if d.deleteConfirm && r == d.cursor {
		return lipgloss.NewStyle().
			Width(d.width).
			Align(lipgloss.Center).
			Foreground(lipgloss.Color("0")).
			Background(lipgloss.Color("9")).
			Render("Delete this row? (press 'x' again to confirm, 'esc' to cancel)")
	}

	isPlaying := d.player.getCurrentlyPlayingId() == d.rows[r][colID]
	rowStyle := lipgloss.NewStyle()

	if r == d.cursor || d.isSelected(r) {
		rowStyle = d.selectedRowStyle

		if d.isFocused {
			rowStyle = rowStyle.Background(d.focusedBGColor)
		}
	}

	for _, colKey := range d.columns {
		style := lipgloss.NewStyle().Width(d.widths[colKey]).Padding(0, dtCellPadding)
		cellWidth := d.widths[colKey] - style.GetHorizontalFrameSize()
		colValue := d.rows[r][colKey]

		switch colKey {
		case colName:
			if d.nameTruncateLeft > 0 && r == d.cursor {
				colValue = runewidth.TruncateLeft(colValue, d.nameTruncateLeft, "…")
			}
		case colWatched:
			style = style.AlignHorizontal(lipgloss.Center)

			if isPlaying {
				rowStyle = rowStyle.Bold(true).Background(lipgloss.Color("34"))
			}
		case colLocation:
			colValue = shortenPath(colValue)
		}

		colValue = runewidth.Truncate(colValue, cellWidth, "…")
		s.WriteString(rowStyle.Render(style.Render(colValue)))
	}

	return s.String()
}

func (d *datatable) View() string {
	playingRow := ""
	header := d.renderHeader()
	d.viewport.Height = d.styles.GetHeight() - lipgloss.Height(header)
	content := lipgloss.JoinVertical(lipgloss.Top, header, d.viewport.View())

	if playingRow != "" {
		content = lipgloss.JoinVertical(lipgloss.Top, content, playingRow)
	}

	return d.styles.Width(d.width).Render(content)
}
