package main

import (
	"errors"
	"fmt"
	"path"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type downloaderModel struct {
	queued         int
	titleStyle     lipgloss.Style
	titleBarStyle  lipgloss.Style
	inqueueStyle   lipgloss.Style
	queueSizeStyle lipgloss.Style
	statusStyle    lipgloss.Style
	etaStyle       lipgloss.Style
	style          lipgloss.Style
	downloadPath   string
	progress       progress.Model
	width          int
	status         downloadStatus
	filename       *runningTextModel
	speed          float64
	elapsed        float64
	eta            float64
}

const filenameWidth = 20

func newDownloaderModel(downloadDir string) *downloaderModel {
	titleBarStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("250")).
		Background(lipgloss.Color("0"))
	componentStyle := lipgloss.NewStyle().Bold(true).Padding(0, 1)

	titleStyle := componentStyle.Italic(true).
		Background(lipgloss.Color("57")).
		SetString("DOWNLOADER")
	inqueueStyle := componentStyle.Background(lipgloss.Color("30")).SetString("INQUEUE")
	queueSizeStyle := componentStyle.Background(lipgloss.Color("39"))
	filenameStyle := componentStyle.Foreground(lipgloss.Color("39"))

	const statusPadding = 2
	statusStyle := componentStyle.Width(len(downloadStatusDownloading.String()) + statusPadding).
		AlignHorizontal(lipgloss.Center).
		Foreground(lipgloss.Color("229"))
	etaStyle := componentStyle

	downloadPath := lipgloss.JoinHorizontal(
		lipgloss.Top,
		componentStyle.Background(lipgloss.Color("27")).Render("PATH"),
		componentStyle.Background(lipgloss.Color("39")).Render(downloadDir),
	)

	style := lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder())

	return &downloaderModel{
		queued:         0,
		width:          0,
		titleStyle:     titleStyle,
		titleBarStyle:  titleBarStyle,
		inqueueStyle:   inqueueStyle,
		queueSizeStyle: queueSizeStyle,
		filename:       newRunningTextModel(filenameWidth, filenameStyle),
		statusStyle:    statusStyle,
		etaStyle:       etaStyle,
		style:          style,
		downloadPath:   downloadPath,
		progress:       progress.New(progress.WithDefaultGradient()),
		status:         downloadStatusIdle,
	}
}

func (d *downloaderModel) Init() tea.Cmd {
	return tea.Batch(d.progress.Init(), d.filename.Init())
}

func (d *downloaderModel) Update(msg tea.Msg) (*downloaderModel, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		d.width = msg.Width - d.style.GetHorizontalFrameSize()
	case downloadQueuedMsg:
		d.queued++
	case startDownloadMsg:
		d.status = downloadStatusPreparing
	case downloadCompletedMsg:
		d.queued--
		d.status, d.speed, d.elapsed, d.eta = downloadStatusIdle, 0, 0, 0
		cmds = append(cmds, d.progress.SetPercent(0))
	case downloadErrorMsg:
		d.status = downloadStatusError
		err := errors.New(msg.msg)
		cmds = append(cmds, errorCmd(err))
	case downloadProgressMsg:
		total := max(msg.TotalBytes, msg.TotalBytesEst, 1)
		downloaded := min(msg.DownloadedBytes, total)

		d.status = downloadStatusDownloading
		d.speed = msg.Speed
		d.elapsed = msg.Elapsed * float64(time.Second)
		d.eta = msg.Eta * float64(time.Second)
		filename := path.Base(msg.Filename)
		percent := downloaded / total
		cmds = append(cmds, d.progress.SetPercent(percent), d.filename.updateText(filename))
	case finishDownloadMsg:
		d.status = downloadStatusFinished
	case quitMsg:
		d.status = downloadStatusQuitting
	case progress.FrameMsg:
		progressModel, cmd := d.progress.Update(msg)
		cmds = append(cmds, cmd)
		d.progress = progressModel.(progress.Model)
	}

	filename, cmd := d.filename.Update(msg)
	cmds = append(cmds, cmd)
	d.filename = filename.(*runningTextModel)

	return d, tea.Batch(cmds...)
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
	info := lipgloss.NewStyle().
		Width(d.width - w(title)).
		AlignHorizontal(lipgloss.Right).
		Render(lipgloss.JoinHorizontal(lipgloss.Top, d.downloadPath, inqueue, queueSize))
	titleBar := d.titleBarStyle.Render(lipgloss.JoinHorizontal(lipgloss.Top, title, info))

	var filename, speed, elapsed, eta string
	status := d.statusStyle.Render(d.status.String())
	d.progress.ShowPercentage = false

	if d.status == downloadStatusDownloading {
		filename = d.filename.View()
		speed = d.etaStyle.Render(formatSpeed(d.speed))
		elapsed = d.etaStyle.Render(time.Duration(d.elapsed).Round(time.Second).String())
		eta = d.etaStyle.Render(
			fmt.Sprintf("ETA %s", time.Duration(d.eta).Round(time.Second).String()),
		)
		d.progress.ShowPercentage = true
	}

	d.progress.Width = d.width - w(status) - w(filename) - w(speed) - w(elapsed) - w(eta)
	progress := d.progress.View()

	downloadBar := lipgloss.JoinHorizontal(
		lipgloss.Top,
		status,
		filename,
		progress,
		speed,
		elapsed,
		eta,
	)

	return d.style.Width(d.width).Render(lipgloss.JoinVertical(lipgloss.Top, titleBar, downloadBar))
}
