package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/linnovs/ytqueue/database"
)

var activeBorderColor = lipgloss.Color("99") // nolint: gochecknoglobals

type (
	footerMsg      struct{ msg string }
	clearFooterMsg struct{ msg string }
)

func footerMsgCmd(msg string, clearAfter time.Duration) tea.Cmd {
	if clearAfter == 0 {
		clearAfter = time.Second * 3
	}

	return tea.Sequence(func() tea.Msg {
		return footerMsg{msg: msg}
	}, tea.Tick(clearAfter, func(t time.Time) tea.Msg { return clearFooterMsg{msg} }))
}

type quitMsg struct{}

type appModel struct {
	section       sectionType
	cancelFn      context.CancelFunc
	keymap        keymap
	width, height int
	help          help.Model
	urlPrompt     *urlPrompt
	topbar        topbar
	downloader    *downloader
	status        *status
	datatable     *datatable
	logging       *logging
	errorStyle    lipgloss.Style
	err           error
	footerMsg     string
}

type contextFn func() context.Context

func newModel(
	downloader *downloader,
	player *player,
	logger io.Reader,
	ctx context.Context,
	cancelFn context.CancelFunc,
	queries *database.Queries,
	cfg *config,
) appModel {
	getContext := func() context.Context {
		return ctx
	}

	return appModel{
		cancelFn:   cancelFn,
		keymap:     newKeymap(),
		help:       help.New(),
		urlPrompt:  newURLPrompt(),
		topbar:     newTopbar(),
		downloader: downloader,
		status:     newStatus(cfg.DownloadPath),
		datatable:  newDatatable(player, queries, getContext),
		logging:    newLogging(logger),
		errorStyle: newErrorStyle(),
	}
}

func (m appModel) Init() tea.Cmd {
	return tea.Batch(
		m.urlPrompt.Init(),
		m.topbar.Init(),
		m.status.Init(),
		m.datatable.Init(),
		m.logging.Init(),
		sectionChangedCmd(sectionDatatable),
	)
}

func (m appModel) exitCmd() tea.Cmd {
	return tea.Sequence(func() tea.Msg {
		return quitMsg{}
	}, func() tea.Msg {
		m.cancelFn()
		m.downloader.stop()

		return nil
	}, tea.Quit)
}

func (m appModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keymap.help):
			m.help.ShowAll = !m.help.ShowAll
		case key.Matches(msg, m.keymap.quit):
			return m, m.exitCmd()
		case key.Matches(msg, m.keymap.prev):
			cmds = append(cmds, sectionChangedCmd(m.section.prev()))
		case key.Matches(msg, m.keymap.next):
			cmds = append(cmds, sectionChangedCmd(m.section.next()))
		case key.Matches(msg, m.keymap.toggleLog):
			m.logging.visible = !m.logging.visible
			return m, nil
		}
	case sectionChangedMsg:
		m.section = msg.section
	case submitURLMsg:
		cmds = append(cmds, enqueueURLCmd(m.downloader, msg.url))
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case footerMsg:
		m.footerMsg = msg.msg
	case clearFooterMsg:
		if msg.msg == m.footerMsg {
			m.footerMsg = ""
		}
	case errorMsg:
		m.err = msg.err
		cmds = append(cmds, resetErrorMsgCmd(m.err))
	case resetErrorMsg:
		if errors.Is(msg.err, m.err) {
			m.err = nil
		}
	}

	var cmd tea.Cmd
	m.urlPrompt, cmd = m.urlPrompt.Update(msg)
	cmds = append(cmds, cmd)
	m.topbar, cmd = m.topbar.Update(msg)
	cmds = append(cmds, cmd)
	m.status, cmd = m.status.Update(msg)
	cmds = append(cmds, cmd)
	m.logging, cmd = m.logging.Update(msg)
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
