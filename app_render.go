package main

import (
	"fmt"
	"slices"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/lipgloss"
)

const (
	minWidth  = 100
	minHeight = 35
)

func (m appModel) terminalTooSmall() string {
	message := lipgloss.NewStyle().Bold(true).Render("Terminal size too small")
	currentSize := lipgloss.NewStyle().
		Foreground(lipgloss.Color("9")).
		Render(fmt.Sprintf("%dx%d", m.width, m.height))
	minimumSize := lipgloss.NewStyle().
		Foreground(lipgloss.Color("10")).
		Render(fmt.Sprintf("%dx%d", minWidth, minHeight))

	return lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center, lipgloss.Center).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Center, message,
				"Current size: "+currentSize,
				"Minimum size: "+minimumSize,
			),
		)
}

func (m appModel) footerView() string {
	var keymap help.KeyMap

	switch m.section {
	case sectionURLPrompt:
		keymap = m.keymap.prompt
	case sectionDatatable:
		keymap = m.keymap.datatable
	}

	m.help.Width = m.width
	helpView := m.help.View(keymap)

	var footerItem []string

	if m.err != nil {
		footerItem = append(
			footerItem,
			m.errorStyle.Width(m.width).
				AlignHorizontal(lipgloss.Center).
				Render(fmt.Sprintf("Error: %s", m.err)),
		)
	}

	if m.footerMsg != "" {
		footerItem = append(
			footerItem,
			m.errorStyle.Width(m.width).
				AlignHorizontal(lipgloss.Center).
				UnsetForeground().
				Faint(true).
				Render(m.footerMsg),
		)
	}

	if helpView != "" {
		footerItem = append(footerItem, helpView)
	}

	if len(footerItem) != 0 {
		return lipgloss.JoinVertical(lipgloss.Center, footerItem...) + "\n"
	}

	return ""
}

func (m appModel) calculateHeights(components []string) int {
	var totalHeight int

	h := lipgloss.Height

	for _, component := range components {
		if component == "" {
			continue
		}

		totalHeight += h(component)
	}

	return totalHeight
}

func (m appModel) View() string {
	if m.width < minWidth || m.height < minHeight {
		return m.terminalTooSmall()
	}

	const datatableIdx = 2
	const totalSections = 6

	sections := make([]string, 0, totalSections)
	sections = append(sections, m.topbar.View())
	sections = append(sections, m.urlPrompt.View())
	sections = append(sections, m.playingNow.View())
	sections = append(sections, m.status.View())
	sections = append(sections, m.logging.View())
	sections = append(sections, m.footerView())

	heightAdjusted := m.height - m.calculateHeights(sections)
	m.datatable.setHeight(heightAdjusted)
	sections = slices.Insert(sections, datatableIdx, m.datatable.View())

	return lipgloss.JoinVertical(lipgloss.Center, slices.DeleteFunc(sections, func(c string) bool {
		return c == ""
	})...)
}
