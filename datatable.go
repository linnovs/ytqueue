package main

import (
	"database/sql"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type datatable struct {
	style lipgloss.Style
	width int
}

func newDatatable(db *sql.DB) *datatable {
	style := lipgloss.NewStyle().Border(lipgloss.RoundedBorder())

	return &datatable{style: style}
}

func (d *datatable) Init() tea.Cmd {
	return nil
}

func (d *datatable) Update(msg tea.Msg) (*datatable, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		d.width = msg.Width
	}

	return d, nil
}

func (d *datatable) View() string {
	widthAdjusted := d.width - d.style.GetHorizontalBorderSize()

	return d.style.Width(widthAdjusted).Render("Datatable Placeholder")
}
