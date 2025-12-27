package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os/exec"
	"path"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type downloader struct {
	p           *tea.Program
	downloadDir string
	tempDir     string
	queue       chan string
	closeCh     chan struct{}
	wg          *sync.WaitGroup
}

func newDownloader(cfg *config, ds *datastore) *downloader {
	const queueSize = 100
	q := make(chan string, queueSize)
	closeCh := make(chan struct{})
	wg := &sync.WaitGroup{}

	return &downloader{nil, cfg.DownloadPath, cfg.tempDir, q, closeCh, wg}
}

func (d *downloader) setProgram(p *tea.Program) {
	d.p = p
}

const (
	titleFormat            = "%(title).50s [%(id)s].%(ext)s"
	progressUpdateInterval = time.Millisecond * 100
)

func (d *downloader) readStdout(stdoutPipe io.ReadCloser, url string) {
	scanner := bufio.NewScanner(stdoutPipe)
	ticker := time.NewTicker(progressUpdateInterval)

	for scanner.Scan() {
		var msg downloadProgressMsg
		if err := json.Unmarshal(scanner.Bytes(), &msg); err != nil {
			err := fmt.Errorf("[downloader] failed to unmarshal progress message: %w", err)
			d.p.Send(errorMsg{err})

			continue
		}

		switch msg.Status {
		case "downloading":
			select {
			case <-ticker.C:
				d.p.Send(msg)
			default:
			}
		case "finished":
			d.p.Send(finishDownloadMsg{
				filename:     path.Base(msg.Filename),
				downloadPath: d.downloadDir,
				url:          url,
			})
		case "error":
			slog.Error("download error", slog.String("stdout", scanner.Text()))
			d.p.Send(downloadErrorMsg{"[downloader] download error occurred"}) // TODO: update msg
		}
	}
}

func (d *downloader) readStderr(stderrPipe io.ReadCloser) {
	scanner := bufio.NewScanner(stderrPipe)

	for scanner.Scan() {
		err := fmt.Errorf("[downloader] %s", scanner.Text())
		d.p.Send(errorMsg{err})
	}
}

func (d *downloader) download(ctx context.Context, url string) {
	const concurrentFragments = "100"
	cmd := exec.CommandContext(
		ctx,
		"yt-dlp",
		"--concurrent-fragments",
		concurrentFragments,
		"--progress",
		"--progress-template",
		"%(progress)j",
		"--newline",
		"--quiet",
		"--no-warning",
		"--output",
		titleFormat,
		"--paths",
		fmt.Sprintf("home:%s", d.downloadDir), // #nosec G204
		"--paths",
		fmt.Sprintf("temp:%s", d.tempDir), // #nosec G204
		url,
	)

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		err = fmt.Errorf("[downloader] failed to get stdout pipe: %w", err)
		d.p.Send(errorMsg{err})

		return
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		err = fmt.Errorf("[downloader] failed to get stderr pipe: %w", err)
		d.p.Send(errorMsg{err})

		return
	}

	if err := cmd.Start(); err != nil {
		err = fmt.Errorf("[downloader] failed to start download command: %w", err)
		d.p.Send(errorMsg{err})

		return
	}

	go d.readStdout(stdoutPipe, url)
	go d.readStderr(stderrPipe)

	if err := cmd.Wait(); err != nil {
		return
	}
}

func (d *downloader) downloadCmd(url string) tea.Cmd {
	return func() tea.Msg {
		if url == "" {
			return nil
		}

		d.queue <- url

		return enqueueDownloadMsg{url}
	}
}

func (d *downloader) startDownload(ctx context.Context, url string) {
	d.wg.Add(1)
	defer d.wg.Done()

	slog.Info("starting download", slog.String("url", url))

	if d.p == nil {
		slog.Error("program pointer is nil, cannot send finish download message")
		return
	}

	d.download(ctx, url)
	d.p.Send(downloadCompletedMsg{})
	slog.Info("download completed", slog.String("url", url))
}

func (d *downloader) start() {
	ctx, cancel := context.WithCancel(context.Background())

	for {
		select {
		case <-d.closeCh:
			cancel()
			return
		case url := <-d.queue:
			d.startDownload(ctx, url)
		}
	}
}

func (d *downloader) quit() tea.Msg {
	close(d.closeCh)
	d.p.Send(waitingMsg{})
	d.wg.Wait()

	return tea.Quit()
}
