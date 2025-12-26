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
	width, height   int
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

func newModel(d *downloader, t *datatable, cfg *config) appModel {
	return appModel{
		keymap:          newKeymap(),
		help:            help.New(),
		urlPrompt:       newURLPrompt(),
		topbar:          newTopbar(),
		downloader:      d,
		downloaderModel: newDownloaderModel(cfg.DownloadPath),
		datatable:       t,
		errorStyle:      newErrorStyle(),
	}
}

func (m appModel) Init() tea.Cmd {
	return tea.Batch(
		m.urlPrompt.Init(),
		m.topbar.Init(),
		m.downloaderModel.Init(),
		m.datatable.Init(),
	)
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
		case key.Matches(msg, m.keymap.prev):
			m.keymap.section = m.keymap.section.prev()
			cmds = append(cmds, sectionChangedCmd(m.keymap.section))
		case key.Matches(msg, m.keymap.next):
			m.keymap.section = m.keymap.section.next()
			cmds = append(cmds, sectionChangedCmd(m.keymap.section))
		case key.Matches(msg, m.keymap.prompt.clear):
			m.urlPrompt.prompt.Reset()

			cmds = append(cmds, errorCmd(nil))
		case key.Matches(msg, m.keymap.prompt.submit):
			filmUrl, err := url.Parse(m.urlPrompt.prompt.Value())
			if err != nil {
				return m, errorCmd(err)
			}

			cmds = append(cmds, errorCmd(nil)) // clear previous error
			cmds = append(cmds, m.downloader.downloadCmd(filmUrl.String()))

			m.urlPrompt.prompt.Reset()
		case key.Matches(msg, m.keymap.datatable.moveUp):
			m.datatable.MoveUp()
		case key.Matches(msg, m.keymap.datatable.moveDown):
			m.datatable.MoveDown()
		}
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
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
	helpView := m.help.View(m.keymap)

	var footerItem []string

	if m.err != nil {
		footerItem = append(footerItem, m.errorStyle.Render(fmt.Sprintf("Error: %s", m.err)))
	}

	if m.footerMsg != "" {
		footerItem = append(footerItem, m.errorStyle.Render(m.footerMsg))
	}

	if helpView != "" {
		footerItem = append(footerItem, helpView)
	}

	if len(footerItem) != 0 {
		return lipgloss.JoinVertical(lipgloss.Center, footerItem...)
	}

	return ""
}

func (m appModel) View() string {
	if m.width < minWidth || m.height < minHeight {
		return m.terminalTooSmall()
	}

	topbar := m.topbar.View()
	urlPrompt := m.urlPrompt.View()
	downloaderView := m.downloaderModel.View()
	footerView := m.footerView()

	h := lipgloss.Height

	topbarHeight := h(topbar)
	urlPromptHeight := h(urlPrompt)
	downloaderViewHeight := h(downloaderView)
	footerHeight := h(footerView)

	if footerView == "" {
		footerHeight = 0
	}

	heightAdjusted := m.height - topbarHeight
	heightAdjusted -= urlPromptHeight + downloaderViewHeight + footerHeight

	m.datatable.setHeight(heightAdjusted)
	dataTable := m.datatable.View()

	verticalItems := []string{topbar, urlPrompt, dataTable, downloaderView}
	if footerView != "" {
		verticalItems = append(verticalItems, footerView)
	}

	return lipgloss.JoinVertical(lipgloss.Center, verticalItems...)
}
