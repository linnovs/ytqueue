package main

import (
	"fmt"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type enqueueDownloadMsg struct {
	url string
}

type finishDownloadMsg struct {
	title    string
	filepath string
}

type downloaderModel struct {
	queued           int
	titleStyle       lipgloss.Style
	titleBarStyle    lipgloss.Style
	inqueueStyle     lipgloss.Style
	queueSizeStyle   lipgloss.Style
	downloadingStyle lipgloss.Style
	style            lipgloss.Style
	progress         progress.Model
	width            int
}

func newDownloaderModel() *downloaderModel {
	titleBarStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("250")).
		Background(lipgloss.Color("0"))
	titleStyle := titleBarStyle.
		Bold(true).
		Italic(true).
		Background(lipgloss.Color("57")).
		Padding(0, 1).
		SetString("DOWNLOADER")
	inqueueStyle := titleStyle.Italic(false).
		Background(lipgloss.Color("30")).
		Padding(0, 1).
		SetString("INQUEUE")
	queueSizeStyle := inqueueStyle.Background(lipgloss.Color("39")).
		AlignHorizontal(lipgloss.Right).
		SetString("")
	downloadingStyle := lipgloss.NewStyle().AlignHorizontal(lipgloss.Center).Padding(0, 1)
	style := lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder())

	return &downloaderModel{
		queued:           0,
		width:            0,
		titleStyle:       titleStyle,
		titleBarStyle:    titleBarStyle,
		inqueueStyle:     inqueueStyle,
		queueSizeStyle:   queueSizeStyle,
		downloadingStyle: downloadingStyle,
		style:            style,
		progress:         progress.New(progress.WithDefaultGradient()),
	}
}

func (d *downloaderModel) Init() tea.Cmd {
	return nil
}

func (d *downloaderModel) Update(msg tea.Msg) (*downloaderModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		d.width = msg.Width - d.style.GetHorizontalBorderSize()
	case enqueueDownloadMsg:
		d.queued++
	case finishDownloadMsg:
		d.queued--
	case progress.FrameMsg:
		progressModel, cmd := d.progress.Update(msg)
		d.progress = progressModel.(progress.Model)

		return d, cmd
	}

	return d, nil
}

func (d *downloaderModel) View() string {
	w := lipgloss.Width

	title := d.titleStyle.Render()
	inqueue := d.inqueueStyle.Render()
	queueSize := d.queueSizeStyle.Render(fmt.Sprint(d.queued))
	spacing := lipgloss.NewStyle().Width(d.width - w(title) - w(inqueue) - w(queueSize)).Render()
	titleBar := d.titleBarStyle.Render(lipgloss.JoinHorizontal(
		lipgloss.Top, title, spacing, inqueue, queueSize,
	))

	downloadingMsg := d.downloadingStyle.Render("IDLE")
	d.progress.Width = d.width - lipgloss.Width(downloadingMsg)
	downloadBar := lipgloss.JoinHorizontal(lipgloss.Top, downloadingMsg, d.progress.View())

	return d.style.Width(d.width).Render(lipgloss.JoinVertical(lipgloss.Top, titleBar, downloadBar))
}
