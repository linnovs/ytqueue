package main

import tea "github.com/charmbracelet/bubbletea"

type downloadQueuedMsg struct{}

func enqueueURLCmd(d *downloader, url string) tea.Cmd {
	return func() tea.Msg {
		d.enqueue(url)

		return downloadQueuedMsg{}
	}
}

type finishDownloadMsg struct {
	filename     string
	downloadPath string
	url          string
}

type downloadErrorMsg struct {
	msg string
}

type downloadCompletedMsg struct{}

type downloadProgressMsg struct {
	Status          string  `json:"status"`
	Filename        string  `json:"filename"`
	DownloadedBytes float64 `json:"downloaded_bytes"`
	TotalBytes      float64 `json:"total_bytes"`
	TotalBytesEst   float64 `json:"total_bytes_estimate"`
	Speed           float64 `json:"speed"`
	Elapsed         float64 `json:"elapsed"`
	Eta             float64 `json:"eta"`
}

type downloadStatus int

func (s downloadStatus) String() string {
	return [...]string{"IDLE", "PREPARING", "DOWNLOADING", "FINISHED", "ERROR"}[s]
}

const (
	downloadStatusIdle downloadStatus = iota
	downloadStatusPreparing
	downloadStatusDownloading
	downloadStatusFinished
	downloadStatusError
)
