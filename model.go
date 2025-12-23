package main

import (
	"fmt"
	"net/url"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type waitingMsg struct{}

type appModel struct {
	keymap          keymap
	help            help.Model
	urlPrompt       *urlPrompt
	topbar          topbar
	downloader      *downloader
	downloaderModel *downloaderModel
	errorStyle      lipgloss.Style
	err             error
	footerMsg       string
}

func newModel() appModel {
	return appModel{
		keymap:     newKeymap(),
		help:       help.New(),
		urlPrompt:  newURLPrompt(),
		topbar:     newTopbar(),
		downloader: newDownloaderModel(),
		errorStyle: newErrorStyle(),
	}
}

func (m appModel) Init() tea.Cmd {
	return nil
}

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

			cmds = append(cmds, errorCmd(nil)) // clear previous error
			cmds = append(cmds, downloadCmd(filmUrl.String()))

			m.urlPrompt.Reset()
		}
	case waitingMsg:
		m.footerMsg = "Waiting for downloads to finish..."
	case errorMsg:
		m.err = msg.err
	}

	var cmd tea.Cmd
	m.urlPrompt, cmd = m.urlPrompt.Update(msg)
	cmds = append(cmds, cmd)
	m.topbar, cmd = m.topbar.Update(msg)
	cmds = append(cmds, cmd)
	m.downloader, cmd = m.downloader.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m appModel) View() string {
	var footer string
	if m.err != nil {
		footer = m.errorStyle.Render(fmt.Sprintf("Error: %s", m.err))
	} else if m.footerMsg != "" {
		footer = m.errorStyle.Render(m.footerMsg)
	}

	return lipgloss.JoinVertical(
		lipgloss.Center,
		m.topbar.View(),
		m.urlPrompt.View(),
		m.downloader.View(),
		m.help.View(m.keymap),
		footer,
	)
}
