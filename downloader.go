package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os/exec"
	"path/filepath"
	"slices"
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

func newDownloader(cfg *config) *downloader {
	const queueSize = 100
	q := make(chan string, queueSize)
	closeCh := make(chan struct{})
	wg := new(sync.WaitGroup)

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
		case "after_move":
			d.p.Send(finishDownloadMsg{
				filename:     filepath.Base(msg.Filename),
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

func (d *downloader) download(ctx context.Context, url string, secondTry ...bool) {
	const concurrentFragments = "100"
	const impersonateArgPos = 2

	args := make([]string, 0)
	args = append(args,
		"--concurrent-fragments",
		concurrentFragments,
		"--print",
		`after_move:{"status": "after_move", "filename": "%(filepath)s"}`,
		"--progress",
		"--progress-template",
		"%(progress)j",
		"--newline",
		"--quiet",
		"--no-warning",
		"--output",
		titleFormat,
		"--paths",
		fmt.Sprintf("home:%s", d.downloadDir),
		"--paths",
		fmt.Sprintf("temp:%s", d.tempDir),
		url,
	)

	if len(secondTry) > 0 && secondTry[0] {
		args = slices.Insert(args, impersonateArgPos, "--impersonate", "chrome")
	}

	cmd := exec.CommandContext(ctx, "yt-dlp", args...) // #nosec G204

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
		if len(secondTry) == 0 {
			d.download(ctx, url, true)
		}

		return
	}
}

func (d *downloader) enqueue(url string) {
	if url == "" {
		return
	}

	d.queue <- url
}

func (d *downloader) startDownload(ctx context.Context, url string) {
	d.wg.Add(1)
	defer d.wg.Done()

	slog.Info("starting download", slog.String("url", url))

	if d.p == nil {
		slog.Error("program pointer is nil, cannot send finish download message")
		return
	}

	d.p.Send(startDownloadMsg{url})
	d.download(ctx, url)
	d.p.Send(downloadCompletedMsg{url})
	slog.Info("download completed", slog.String("url", url))
}

func (d *downloader) start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case url := <-d.queue:
			d.startDownload(ctx, url)
		}
	}
}

func (d *downloader) stop() {
	d.p.Send(footerMsg{"Waiting for downloads to finish..."})
	d.wg.Wait()
}
