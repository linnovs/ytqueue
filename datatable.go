package main

import (
	"strings"
	"sync"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/linnovs/ytqueue/database"
	"github.com/mattn/go-runewidth"
)

func clamp(n, low, high int) int {
	return min(max(n, low), high)
}

type column string

const (
	colID       column = "ID"
	colWatched  column = "Watched"
	colName     column = "Name"
	colURL      column = "URL"
	colLocation column = "Location"
	colOrder    column = "Order"
)

type row map[column]string

type datatable struct {
	width            int
	nameTruncateLeft int
	widths           map[column]int
	getCtx           contextFn
	datastore        *datastore
	viewport         viewport.Model
	styles           lipgloss.Style
	headerStyle      lipgloss.Style
	selectedRowStyle lipgloss.Style
	focusedBGColor   lipgloss.Color
	keymap           datatableKeymap
	columns          []column
	rowMu            sync.RWMutex
	rows             []row
	cursorMu         sync.RWMutex
	cursor           int
	isFocused        bool
	player           *player
	deleteConfirm    bool
}

func newDatatable(player *player, queries *database.Queries, getCtx contextFn) *datatable {
	// minus topbar, urlPrompt, downloaderView, datatable's header (include borders)
	const defaultViewportHeight = minHeight - 1 - 3 - 4 - 4

	styles := lipgloss.NewStyle().Border(lipgloss.RoundedBorder())
	d := &datatable{
		widths:      make(map[column]int),
		getCtx:      getCtx,
		datastore:   newDatastore(queries),
		viewport:    viewport.New(0, defaultViewportHeight),
		headerStyle: lipgloss.NewStyle().Bold(true),
		selectedRowStyle: lipgloss.NewStyle().
			Background(lipgloss.Color("244")).
			Foreground(lipgloss.Color("229")),
		focusedBGColor: lipgloss.Color("141"),
		styles:         styles,
		keymap:         newDatatableKeymap(),
		columns:        []column{colWatched, colName, colURL, colLocation},
		player:         player,
	}

	return d
}

const dtCellPadding = 1

func (d *datatable) calculateColWidth() {
	const colPadding = dtCellPadding * 2 // padding left and right
	const isWatchedColWidth = 7
	adjustedWidth := d.width - (colPadding * len(d.columns)) - isWatchedColWidth // exclude isWatched
	defaultWidth := (adjustedWidth) / (len(d.columns) - 1)                       // exclude isWatched
	d.widths[colName] = defaultWidth + colPadding
	d.widths[colURL] = defaultWidth + colPadding
	d.widths[colLocation] = defaultWidth + colPadding
	d.widths[colWatched] = isWatchedColWidth + colPadding
}

func playingIDIndexFunc(id string) func(r row) bool {
	return func(r row) bool {
		return r[colID] == id
	}
}

func (d *datatable) getIDAtIndex(idx int) string {
	d.rowMu.RLock()
	defer d.rowMu.RUnlock()

	return d.rows[idx][colID]
}

func (d *datatable) getCopyOfRows() []row {
	d.rowMu.RLock()
	defer d.rowMu.RUnlock()

	rows := make([]row, len(d.rows))
	copy(rows, d.rows)

	return rows
}

func (d *datatable) updateViewport() {
	d.rowMu.RLock()
	defer d.rowMu.RUnlock()

	renderedRows := make([]string, 0, len(d.rows))
	for i := range d.rows {
		renderedRows = append(renderedRows, d.renderRow(i))
	}

	d.viewport.SetContent(lipgloss.JoinVertical(lipgloss.Left, renderedRows...))
}

func (d *datatable) setRows(rows []row) {
	d.rowMu.Lock()
	d.rows = rows
	d.rowMu.Unlock()

	d.updateViewport()
}

func (d *datatable) Init() tea.Cmd {
	return d.refreshRowsCmd()
}

func (d *datatable) Update(msg tea.Msg) (*datatable, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		d.width = msg.Width - d.styles.GetHorizontalFrameSize()
		d.viewport.Width = d.width
		d.calculateColWidth()
	case finishDownloadMsg:
		cmds = append(cmds, d.newVideoCmd(msg.filename, msg.url, msg.downloadPath))
	case finishPlayingMsg:
		cmds = append(cmds, d.playNextOrStopCmd())
	case playbackChangedMsg:
		d.updateViewport()
	case sectionChangedMsg:
		d.isFocused = msg.section == sectionDatatable
		if msg.section == sectionDatatable {
			d.styles = d.styles.BorderForeground(activeBorderColor)
		} else {
			d.styles = d.styles.UnsetBorderForeground()
		}
	case updateProgressMsg:
		cmds = append(cmds, d.player.progress.SetPercent(msg.percent))
	case progress.FrameMsg:
		model, cmd := d.player.progress.Update(msg)
		cmds = append(cmds, cmd)
		d.player.progress = model.(progress.Model)
	case quitMsg:
		cmds = append(cmds, d.player.quit())
	case tea.KeyMsg:
		cmds = append(cmds, d.keyMsgHandler(msg))
	}

	return d, tea.Batch(cmds...)
}

func (d *datatable) setHeight(height int) {
	height -= d.styles.GetVerticalFrameSize()
	d.styles = d.styles.Height(height)
	d.updateViewport()
}

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

	if r == d.cursor {
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
			style = style.Align(lipgloss.Center)

			if isPlaying {
				colValue = "PLAYING"
				rowStyle = rowStyle.Bold(true).Background(lipgloss.Color("34"))
			}
		case colURL:
			if isPlaying {
				fullWidth := d.widths[colURL] + d.widths[colLocation]
				p := d.player.renderPlayProgress(cellWidth + d.widths[colLocation])
				s.WriteString(style.Width(fullWidth).Render(p))

				continue
			}
		case colLocation:
			if isPlaying {
				continue
			}

			colValue = shortenPath(colValue)
		}

		colValue = runewidth.Truncate(colValue, cellWidth, "…")
		s.WriteString(rowStyle.Render(style.Render(colValue)))
	}

	return s.String()
}

func (d *datatable) View() string {
	header := d.renderHeader()
	d.viewport.Height = d.styles.GetHeight() - lipgloss.Height(header)
	content := lipgloss.JoinVertical(lipgloss.Top, header, d.viewport.View())

	return d.styles.Width(d.width).Render(content)
}
