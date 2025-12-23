package main

import (
	"fmt"
	"log/slog"
	"net/url"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type errorMsg struct {
	err error
}

func errorCmd(err error) tea.Cmd {
	return func() tea.Msg {
		return errorMsg{err}
	}
}

type appModel struct {
	height       int
	width        int
	keymap       keymap
	help         help.Model
	urlPrompt    *urlPrompt
	topbarStyle  lipgloss.Style
	headerStyle  lipgloss.Style
	versionStyle lipgloss.Style
	errorStyle   lipgloss.Style
	err          error
}

func newModel() appModel {
	prompt := newURLPrompt()
	prompt.Focus()

	topbarStyle := lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "250", Dark: "255"}).
		Background(lipgloss.AdaptiveColor{Light: "232", Dark: "0"})
	headerStyle := lipgloss.NewStyle().
		Background(lipgloss.AdaptiveColor{Light: "98", Dark: "91"}).
		Padding(0, 1)
	versionStyle := lipgloss.NewStyle().
		Background(lipgloss.AdaptiveColor{Light: "240", Dark: "238"}).
		Padding(0, 1)
	errorStyle := lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "196", Dark: "124"})

	return appModel{
		keymap:       newKeymap(),
		help:         help.New(),
		urlPrompt:    prompt,
		topbarStyle:  topbarStyle,
		headerStyle:  headerStyle,
		versionStyle: versionStyle,
		errorStyle:   errorStyle,
	}
}

// Init implements [tea.Model].
func (m appModel) Init() tea.Cmd {
	return nil
}

// Update implements [tea.Model].
func (m appModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keymap.help):
			m.help.ShowAll = !m.help.ShowAll
		case key.Matches(msg, m.keymap.quit):
			return m, tea.Quit
		case key.Matches(msg, m.keymap.next):
		case key.Matches(msg, m.keymap.prev):
		case key.Matches(msg, m.keymap.clear):
			m.urlPrompt.Reset()

			cmds = append(cmds, errorCmd(nil))
		case key.Matches(msg, m.keymap.submit):
			filmUrl, err := url.Parse(m.urlPrompt.prompt.Value())
			if err != nil {
				return m, errorCmd(err)
			}

			cmds = append(cmds, errorCmd(nil))

			slog.Info("current URL", slog.String("url", filmUrl.String()))
			m.urlPrompt.Reset()
		}
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width
		m.urlPrompt.setWidth(msg.Width)
	case errorMsg:
		m.err = msg.err
	}

	var cmd tea.Cmd
	m.urlPrompt, cmd = m.urlPrompt.Update(msg)

	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// View implements [tea.Model].
func (m appModel) View() string {
	header := m.headerStyle.Render("Short Film Downloader")
	helpText := m.topbarStyle.Render("Press F1 for help")
	version := m.versionStyle.Width(m.width - lipgloss.Width(header) - lipgloss.Width(helpText)).
		Render(version)
	topbar := m.topbarStyle.Width(m.width).
		Render(lipgloss.JoinHorizontal(lipgloss.Top, header, version, helpText))

	var errorSection string
	if m.err != nil {
		errorSection = m.errorStyle.Render(fmt.Sprintf("Error: %s", m.err))
	}

	return lipgloss.JoinVertical(
		lipgloss.Center,
		topbar,
		m.urlPrompt.View(),
		m.help.View(m.keymap),
		errorSection,
	)
}
