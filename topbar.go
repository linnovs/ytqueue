package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type topbar struct {
	titleStyle    lipgloss.Style
	versionStyle  lipgloss.Style
	helpTextStyle lipgloss.Style
	topbarStyle   lipgloss.Style

	width int
}

func newTopbar() topbar {
	titleStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("91")).
		Padding(0, 1).
		SetString("Short Film Downloader")
	versionStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("177")).
		Padding(0, 1)
	topbarStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("255")).
		Background(lipgloss.Color("0"))
	helpTextStyle := topbarStyle.Padding(0, 1).SetString("Press F1 for help")

	return topbar{titleStyle, versionStyle, helpTextStyle, topbarStyle, 0}
}

func (t topbar) Init() tea.Cmd { return nil }

func (t topbar) Update(msg tea.Msg) (topbar, tea.Cmd) {
	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		t.width = msg.Width
	}

	return t, nil
}

func (t topbar) View() string {
	w := lipgloss.Width

	title := t.titleStyle.Render()
	version := t.versionStyle.Render(fmt.Sprintf("%s-(#%s)-%s", version, commit, buildDate))
	helpText := t.helpTextStyle.Render()
	spacing := t.topbarStyle.Width(t.width - w(title) - w(version) - w(helpText)).Render()

	return t.topbarStyle.Width(t.width).
		Render(lipgloss.JoinHorizontal(lipgloss.Top, title, version, spacing, helpText))
}
