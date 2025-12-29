package main

import (
	"strings"
	"sync"

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
	cursor           int
	isFocused        bool
	playingId        string
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

func (d *datatable) playingIDIndexFunc(r row) bool {
	return r[colID] == d.playingId
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
	case stoppedPlayingMsg:
		if d.playingId == msg.id {
			d.playingId = ""
		}
	case sectionChangedMsg:
		d.isFocused = msg.section == sectionDatatable
		if msg.section == sectionDatatable {
			d.styles = d.styles.BorderForeground(activeBorderColor)
		} else {
			d.styles = d.styles.UnsetBorderForeground()
		}
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

func (d *datatable) renderRow(r int) string {
	var s strings.Builder

	if d.deleteConfirm && r == d.cursor {
		return lipgloss.NewStyle().
			Width(d.width).
			Align(lipgloss.Center).
			Foreground(lipgloss.Color("0")).
			Background(lipgloss.Color("9")).
			Render("Delete this row? (press 'x' again to confirm, 'esc' to cancel)")
	}

	for _, colKey := range d.columns {
		colValue := d.rows[r][colKey]
		style := lipgloss.NewStyle().Width(d.widths[colKey]).Padding(0, dtCellPadding)

		switch colKey {
		case colWatched:
			style = style.Align(lipgloss.Center)

			if d.playingId == d.rows[r][colID] {
				colValue = "Playing"
				style = style.Foreground(lipgloss.Color("0")).
					Background(lipgloss.Color("10")).
					Bold(true)
			}
		case colLocation:
			colValue = shortenPath(colValue)
		}

		cellWidth := d.widths[colKey] - style.GetHorizontalFrameSize()

		if colKey == colName && d.nameTruncateLeft > 0 && r == d.cursor {
			colValue = runewidth.TruncateLeft(colValue, d.nameTruncateLeft, "…")
		}

		s.WriteString(style.Render(runewidth.Truncate(colValue, cellWidth, "…")))
	}

	if r == d.cursor {
		if d.isFocused {
			return d.selectedRowStyle.Background(d.focusedBGColor).Render(s.String())
		}

		return d.selectedRowStyle.Render(s.String())
	}

	return s.String()
}

func (d *datatable) View() string {
	header := d.renderHeader()
	d.viewport.Height = d.styles.GetHeight() - lipgloss.Height(header)
	content := lipgloss.JoinVertical(lipgloss.Top, header, d.viewport.View())

	return d.styles.Width(d.width).Render(content)
}
