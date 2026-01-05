package main

import (
	"sync"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/linnovs/ytqueue/database"
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
	isInitialRefresh    bool
	width               int
	nameTruncateLeft    int
	widths              map[column]int
	getCtx              contextFn
	datastore           *datastore
	viewport            viewport.Model
	styles              lipgloss.Style
	headerStyle         lipgloss.Style
	selectedRowStyle    lipgloss.Style
	focusedBGColor      lipgloss.Color
	keymap              datatableKeymap
	columns             []column
	rowMu               sync.RWMutex
	rows                []row
	updateRowOrderTagMu sync.RWMutex
	updateRowOrderTag   int
	cursorMu            sync.RWMutex
	cursor              int
	isFocused           bool
	player              *player
	deleteConfirm       bool
}

func newDatatable(player *player, queries *database.Queries, getCtx contextFn) *datatable {
	// minus topbar, urlPrompt, downloaderView, datatable's header (include borders)
	const defaultViewportHeight = minHeight - 1 - 3 - 4 - 4

	styles := lipgloss.NewStyle().Border(lipgloss.RoundedBorder())
	d := &datatable{
		isInitialRefresh: true,
		widths:           make(map[column]int),
		getCtx:           getCtx,
		datastore:        newDatastore(queries),
		viewport:         viewport.New(0, defaultViewportHeight),
		headerStyle:      lipgloss.NewStyle().Bold(true),
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
	case updateRowOrderMsg:
		d.updateRowOrderTagMu.RLock()
		if d.updateRowOrderTag == msg.tag {
			cmds = append(cmds, d.updateRowOrderCmd(msg.id))
		}
		d.updateRowOrderTagMu.RUnlock()
	case playbackChangedMsg, updatedRowOrderMsg:
		d.updateViewport()
	case sectionChangedMsg:
		d.isFocused = msg.section == sectionDatatable
		if msg.section == sectionDatatable {
			d.styles = d.styles.BorderForeground(activeBorderColor)
		} else {
			d.styles = d.styles.UnsetBorderForeground()
		}
	case deletedRowMsg:
		footerMsg := "Deleted video: " + msg.filename
		if msg.notFound {
			footerMsg = "Video not found, removed entry from database: " + msg.filename
		}

		cmds = append(cmds, footerMsgCmd(footerMsg, 0))
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
