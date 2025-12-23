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
	height          int
	help            help.Model
	urlPrompt       *urlPrompt
	topbar          topbar
	downloader      *downloader
	downloaderModel *downloaderModel
	datatable       *datatable
	errorStyle      lipgloss.Style
	err             error
	footerMsg       string
}

func newModel(d *downloader) appModel {
	return appModel{
		keymap:          newKeymap(),
		help:            help.New(),
		urlPrompt:       newURLPrompt(),
		topbar:          newTopbar(),
		downloader:      d,
		downloaderModel: newDownloaderModel(),
		datatable:       newDatatable(),
		errorStyle:      newErrorStyle(),
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
			return m, m.downloader.quit
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
			cmds = append(cmds, m.downloader.downloadCmd(filmUrl.String()))

			m.urlPrompt.Reset()
		}
	case tea.WindowSizeMsg:
		m.height = msg.Height
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
	m.downloaderModel, cmd = m.downloaderModel.Update(msg)
	cmds = append(cmds, cmd)
	m.datatable, cmd = m.datatable.Update(msg)
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

	topbar := m.topbar.View()
	urlPrompt := m.urlPrompt.View()
	downloaderView := m.downloaderModel.View()
	helpView := m.help.View(m.keymap)

	h := lipgloss.Height

	topbarHeight := h(topbar)
	urlPromptHeight := h(urlPrompt)
	downloaderViewHeight := h(downloaderView)
	helpViewHeight := h(helpView)
	footerHeight := h(footer)

	heightAdjusted := m.height - topbarHeight
	heightAdjusted -= urlPromptHeight + downloaderViewHeight + helpViewHeight + footerHeight
	heightAdjusted -= m.datatable.style.GetVerticalBorderSize()

	m.datatable.style = m.datatable.style.Height(heightAdjusted)
	dataTable := m.datatable.View()

	return lipgloss.JoinVertical(
		lipgloss.Center,
		topbar,
		urlPrompt,
		dataTable,
		downloaderView,
		helpView,
		footer,
	)
}
