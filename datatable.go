package main

import (
	"context"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

func clamp(n, low, high int) int {
	return min(max(n, low), high)
}

type column string

const (
	colWatched  column = "Watched"
	colName     column = "Name"
	colURL      column = "URL"
	colLocation column = "Location"
)

type row map[column]string

type datatable struct {
	width              int
	widths             map[column]int
	datastore          *datastore
	viewport           viewport.Model
	styles             lipgloss.Style
	headerStyle        lipgloss.Style
	selectedRowStyle   lipgloss.Style
	focusedBGColor     lipgloss.Color
	columns            []column
	rows               []row
	cursor, start, end int
	isFocused          bool
}

func newDatatable(ds *datastore) *datatable {
	// minus topbar, urlPrompt, downloaderView, datatable's header (include borders)
	const defaultViewportHeight = minHeight - 1 - 3 - 4 - 4

	styles := lipgloss.NewStyle().Border(lipgloss.RoundedBorder())
	d := &datatable{
		widths:      make(map[column]int),
		datastore:   ds,
		viewport:    viewport.New(0, defaultViewportHeight),
		headerStyle: lipgloss.NewStyle().Bold(true),
		selectedRowStyle: lipgloss.NewStyle().
			Background(lipgloss.Color("244")).
			Foreground(lipgloss.Color("229")),
		focusedBGColor: lipgloss.Color("141"),
		styles:         styles,
		columns:        []column{colWatched, colName, colURL, colLocation},
		cursor:         0,
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

func (d *datatable) updateViewport() {
	renderedRows := make([]string, 0, len(d.rows))

	if d.cursor >= 0 {
		d.start = clamp(d.cursor-d.viewport.Height, 0, d.cursor)
	} else {
		d.start = 0
	}

	d.end = clamp(d.start+d.viewport.Height, d.cursor, len(d.rows))

	for i := d.start; i < d.end; i++ {
		renderedRows = append(renderedRows, d.renderRow(i))
	}

	d.viewport.SetContent(lipgloss.JoinVertical(lipgloss.Left, renderedRows...))
}

func (d *datatable) setRows(rows []row) {
	d.rows = rows
	d.updateViewport()
}

func (d *datatable) Init() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		videos, err := d.datastore.getVideos(ctx)
		if err != nil {
			return errorMsg{err: err}
		}

		d.setRows(videosToRows(videos))

		return nil
	}
}

func (d *datatable) Update(msg tea.Msg) (*datatable, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		d.width = msg.Width - d.styles.GetHorizontalFrameSize()
		d.viewport.Width = d.width
		d.calculateColWidth()
	case sectionChangedMsg:
		if msg.section == sectionDatatable {
			d.isFocused = true
		} else {
			d.isFocused = false
		}
	}

	return d, nil
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

	for _, colKey := range d.columns {
		colValue := d.rows[r][colKey]
		style := lipgloss.NewStyle().Width(d.widths[colKey]).Padding(0, dtCellPadding)

		if colKey == colWatched {
			style = style.Align(lipgloss.Center)
		}

		cellWidth := d.widths[colKey] - style.GetHorizontalFrameSize()
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
