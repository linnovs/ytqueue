package main

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type enqueueDownloadMsg struct {
	url string
}

type finishDownloadMsg struct{}

type downloadProgressMsg struct {
	Filename        string  `json:"filename"`
	DownloadedBytes int64   `json:"downloaded_bytes"`
	TotalBytes      int64   `json:"total_bytes"`
	Speed           float64 `json:"speed"`
	Elapsed         float64 `json:"elapsed"`
	Eta             float64 `json:"eta"`
}

type downloadStatus int

func (s downloadStatus) String() string {
	return [...]string{"IDLE", "DOWNLOADING", "FINISHED", "ERROR"}[s]
}

const (
	downloadStatusIdle downloadStatus = iota
	downloadStatusDownloading
	downloadStatusFinished
	downloadStatusError
)

type downloaderModel struct {
	queued         int
	titleStyle     lipgloss.Style
	titleBarStyle  lipgloss.Style
	inqueueStyle   lipgloss.Style
	queueSizeStyle lipgloss.Style
	filenameStyle  lipgloss.Style
	statusStyle    lipgloss.Style
	style          lipgloss.Style
	progress       progress.Model
	width          int
	status         downloadStatus
	filename       string
	speed          float64
	elapsed        float64
	eta            float64
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
	downloadFilenameStyle := lipgloss.NewStyle().
		Bold(true).
		Padding(0, 1).
		Foreground(lipgloss.Color("39"))
	downloadingStyle := lipgloss.NewStyle().AlignHorizontal(lipgloss.Center).Padding(0, 1)
	style := lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder())

	return &downloaderModel{
		queued:         0,
		width:          0,
		titleStyle:     titleStyle,
		titleBarStyle:  titleBarStyle,
		inqueueStyle:   inqueueStyle,
		queueSizeStyle: queueSizeStyle,
		filenameStyle:  downloadFilenameStyle,
		statusStyle:    downloadingStyle,
		style:          style,
		progress:       progress.New(progress.WithDefaultGradient()),
		status:         downloadStatusIdle,
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
		d.status, d.filename, d.speed, d.elapsed, d.eta = downloadStatusIdle, "", 0, 0, 0

		return d, d.progress.SetPercent(0)
	case downloadProgressMsg:
		if msg.DownloadedBytes < msg.TotalBytes {
			d.status = downloadStatusDownloading
		} else {
			d.status = downloadStatusFinished
		}

		msg.TotalBytes = max(msg.TotalBytes, 1)
		msg.DownloadedBytes = min(msg.DownloadedBytes, msg.TotalBytes)
		d.filename = msg.Filename
		d.speed = msg.Speed
		d.elapsed = msg.Elapsed
		d.eta = msg.Eta
		percent := (float64(msg.DownloadedBytes) / float64(msg.TotalBytes))

		return d, d.progress.SetPercent(percent)
	case progress.FrameMsg:
		progressModel, cmd := d.progress.Update(msg)
		d.progress = progressModel.(progress.Model)

		return d, cmd
	}

	return d, nil
}

const (
	kibps = 1024
	mibps = 1024 * 1024
	gibps = 1024 * 1024 * 1024
	tibps = 1024 * 1024 * 1024 * 1024
)

func formatSpeed(speed float64) string {
	switch {
	case speed >= tibps:
		return fmt.Sprintf("%.2f TiB/s", speed/tibps)
	case speed >= gibps:
		return fmt.Sprintf("%.2f GiB/s", speed/gibps)
	case speed >= mibps:
		return fmt.Sprintf("%.2f MiB/s", speed/mibps)
	case speed >= kibps:
		fallthrough
	default:
		return fmt.Sprintf("%.2f KiB/s", speed/kibps)
	}
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

	status := d.statusStyle.Render(d.status.String())
	filename := d.filenameStyle.Render(d.filename)
	speed := d.statusStyle.Render(formatSpeed(d.speed))
	elapsed := d.statusStyle.Render(time.Duration(d.elapsed).String())
	eta := d.statusStyle.Render(fmt.Sprintf("ETA %s", time.Duration(d.eta).String()))
	d.progress.Width = d.width - w(status) - w(filename) - w(speed) - w(elapsed) - w(eta)
	downloadBar := lipgloss.JoinHorizontal(
		lipgloss.Top,
		status,
		filename,
		d.progress.View(),
		speed,
		elapsed,
		eta,
	)

	return d.style.Width(d.width).Render(lipgloss.JoinVertical(lipgloss.Top, titleBar, downloadBar))
}
