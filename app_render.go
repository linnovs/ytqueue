package main

import (
	"fmt"

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

func (m appModel) View() string {
	if m.width < minWidth || m.height < minHeight {
		return m.terminalTooSmall()
	}

	topbar := m.topbar.View()
	urlPrompt := m.urlPrompt.View()
	statusView := m.status.View()
	loggingView := m.logging.View()
	footerView := m.footerView()

	h := lipgloss.Height

	topbarHeight := h(topbar)
	urlPromptHeight := h(urlPrompt)
	statusViewHeight := h(statusView)
	loggingViewHeight := h(loggingView)
	footerHeight := h(footerView)

	if footerView == "" {
		footerHeight = 0
	}

	if loggingView == "" {
		loggingViewHeight = 0
	}

	heightAdjusted := m.height - topbarHeight
	heightAdjusted -= urlPromptHeight + statusViewHeight + loggingViewHeight + footerHeight

	m.datatable.setHeight(heightAdjusted)
	dataTable := m.datatable.View()
	verticalItems := []string{topbar, urlPrompt, dataTable, statusView}

	if loggingView != "" {
		verticalItems = append(verticalItems, loggingView)
	}

	if footerView != "" {
		verticalItems = append(verticalItems, footerView)
	}

	return lipgloss.JoinVertical(lipgloss.Center, verticalItems...)
}
