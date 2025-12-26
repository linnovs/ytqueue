package main

import (
	"database/sql"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/linnovs/ytqueue/database"
)

type datatable struct {
	style       lipgloss.Style
	table       table.Model
	tableStyle  table.Styles
	width       int
	queries     *database.Queries
	rows        []table.Row
	selectedRow int
}

func newDatatable(db *sql.DB) *datatable {
	style := lipgloss.NewStyle().Border(lipgloss.RoundedBorder())
	s := table.DefaultStyles()
	s.Selected = s.Selected.Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("236")).
		Bold(true)
	t := table.New(
		table.WithColumns([]table.Column{
			{Title: "Name", Width: 20},
			{Title: "URL", Width: 100},
			{Title: "Path", Width: 20},
			{Title: "Watched", Width: 20},
		}),
		table.WithFocused(false),
		table.WithStyles(s),
	)

	queries := database.New(db)

	// Sample data for demonstration
	rows := []table.Row{
		{"Video 1", "https://example.com/1", "/path/to/1.mp4", "No"},
		{"Video 2", "https://example.com/2", "/path/to/2.mp4", "Yes"},
		{"Video 3", "https://example.com/3", "/path/to/3.mp4", "No"},
	}
	t.SetRows(rows)

	return &datatable{
		style:       style,
		tableStyle:  s,
		table:       t,
		queries:     queries,
		rows:        rows,
		selectedRow: 0,
	}
}

func (d *datatable) Init() tea.Cmd {
	return nil
}

func (d *datatable) Update(msg tea.Msg) (*datatable, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		d.width = msg.Width - d.style.GetHorizontalFrameSize()
	case sectionChangedMsg:
		if msg.section == sectionDatatable {
			d.table.Focus()
			d.tableStyle.Selected = d.tableStyle.Selected.Background(lipgloss.Color("57"))
		} else {
			d.table.Blur()
			d.tableStyle.Selected = d.tableStyle.Selected.Background(lipgloss.Color("236"))
		}

		d.table.SetStyles(d.tableStyle)
	}

	var cmd tea.Cmd
	d.table, cmd = d.table.Update(msg)
	d.selectedRow = d.table.Cursor()

	return d, cmd
}

func (d *datatable) SetRows(rows []table.Row) {
	d.rows = rows
	d.table.SetRows(rows)
}

func (d *datatable) MoveUp() {
	if !d.table.Focused() {
		return
	}

	if d.selectedRow > 0 {
		d.rows[d.selectedRow], d.rows[d.selectedRow-1] = d.rows[d.selectedRow-1], d.rows[d.selectedRow]
		d.selectedRow--
		d.table.SetRows(d.rows)
		d.table.SetCursor(d.selectedRow)
	}
}

func (d *datatable) MoveDown() {
	if !d.table.Focused() {
		return
	}

	if d.selectedRow < len(d.rows)-1 {
		d.rows[d.selectedRow], d.rows[d.selectedRow+1] = d.rows[d.selectedRow+1], d.rows[d.selectedRow]
		d.selectedRow++
		d.table.SetRows(d.rows)
		d.table.SetCursor(d.selectedRow)
	}
}

func (d *datatable) View() string {
	return d.style.Width(d.width).Render(d.table.View())
}
