package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	wlCopyCmd  = "wl-copy"
	wlPasteCmd = "wl-paste"
)

func (d *datatable) copyURLCmd(cursor int) tea.Cmd {
	return func() tea.Msg {
		d.rowMu.RLock()
		defer d.rowMu.RUnlock()

		url := d.rows[cursor][colURL]

		// #nosec G204
		if err := exec.Command(wlCopyCmd, d.rows[cursor][colURL]).Run(); err != nil {
			return errorMsg{fmt.Errorf("failed to copy URL to clipboard: %w", err)}
		}

		const msgDelay = 2 * time.Second

		return footerMsgCmd(fmt.Sprintf("Copied URL (%s) to clipboard", url), msgDelay)()
	}
}

func (d *datatable) pasteURLCmd() tea.Cmd {
	return func() tea.Msg {
		d.rowMu.RLock()
		defer d.rowMu.RUnlock()

		data, err := exec.Command(wlPasteCmd).CombinedOutput()
		if err != nil {
			return errorMsg{fmt.Errorf("failed to paste URL from clipboard: %w", err)}
		}

		url := strings.TrimSpace(string(data))

		slog.Debug("pasted URL from clipboard", slog.String("url", url))

		return submitURLMsg{url}
	}
}

func (d *datatable) refreshRowsCmd() tea.Cmd {
	return tea.Sequence(func() tea.Msg {
		videos, err := d.datastore.getVideos(d.getCtx())
		if err != nil {
			return errorMsg{err: err}
		}

		d.setRows(videosToRows(videos))

		return nil
	}, footerMsgCmd("Refreshed video list", 0))
}

func (d *datatable) newVideoCmd(name, url, location string) tea.Cmd {
	return func() tea.Msg {
		video, err := d.datastore.addVideo(d.getCtx(), name, url, location)
		if err != nil {
			return errorMsg{err}
		}

		rows := append([]row{videoToRow(*video)}, d.getCopyOfRows()...)
		d.setRows(rows)

		return nil
	}
}

func (d *datatable) playStopRowCmd(id string) tea.Cmd {
	return func() tea.Msg {
		d.rowMu.RLock()
		defer d.rowMu.RUnlock()

		slog.Debug("playStopRowCmd", slog.String("requestedId", id))

		if d.player.getCurrentlyPlayingId() == id {
			if d.player.isPlaying() {
				if err := d.player.stop(); err != nil {
					return errorMsg{err: fmt.Errorf("failed to stop player: %w", err)}
				}
			}

			return nil
		}

		idx := slices.IndexFunc(d.rows, playingIDIndexFunc(id))
		if idx == -1 {
			return errorMsg{errors.New("playing video not found in datatable")}
		}

		slog.Debug(
			"playing video found",
			slog.Int("index", idx),
			slog.Int("totalRows", len(d.rows)),
		)

		row := d.rows[idx]
		file := filepath.Join(row[colLocation], row[colName])
		file = filepath.Clean(file)

		_, err := os.Stat(file)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return errorMsg{errors.New("file does not exist")}
			}

			return errorMsg{err: fmt.Errorf("failed to access file: %w", err)}
		}

		slog.Debug("starting to play video", slog.String("id", id), slog.String("file", file))

		if err := d.player.play(file, id); err != nil {
			return errorMsg{err: fmt.Errorf("failed to play file: %w", err)}
		}

		return nil
	}
}

func (d *datatable) setVideoWatched(id string) (int, error) {
	video, err := d.datastore.setWatched(d.getCtx(), id)
	if err != nil {
		return 0, err
	}

	rows := d.getCopyOfRows()
	idx := slices.IndexFunc(rows, playingIDIndexFunc(id))
	rows[idx] = videoToRow(*video)
	d.setRows(rows)

	return idx, nil
}

func (d *datatable) playNextOrStopCmd() tea.Cmd {
	return func() tea.Msg {
		idx, err := d.setVideoWatched(d.player.getCurrentlyPlayingId())
		if err != nil {
			return errorMsg{fmt.Errorf("failed to set video as watched: %w", err)}
		}

		if idx <= 0 {
			return nil
		}

		d.rowMu.RLock()
		defer d.rowMu.RUnlock()

		for i := idx - 1; i >= 0; i-- {
			row := d.rows[i]
			if row[colWatched] == isWatchedNo {
				return d.playStopRowCmd(row[colID])()
			}
		}

		return nil
	}
}

func (d *datatable) toggleWatchedStatusCmd(cursor int) tea.Cmd {
	return func() tea.Msg {
		rows := d.getCopyOfRows()

		video, err := d.datastore.toggleWatched(d.getCtx(), rows[cursor][colID])
		if err != nil {
			return errorMsg{err}
		}

		rows[cursor] = videoToRow(*video)
		d.setRows(rows)

		return nil
	}
}

func (d *datatable) deleteRowCmd(cursor int) tea.Cmd {
	return func() tea.Msg {
		rows := d.getCopyOfRows()
		row := rows[cursor]
		rows = append(rows[:cursor], rows[cursor+1:]...)

		fname := filepath.Clean(filepath.Join(row[colLocation], row[colName]))
		if err := os.Remove(fname); err != nil {
			return errorMsg{fmt.Errorf("failed to delete video file: %w", err)}
		}

		if err := d.datastore.deleteVideo(d.getCtx(), row[colID]); err != nil {
			return errorMsg{err}
		}

		d.setRows(rows)

		d.cursorMu.Lock()
		defer d.cursorMu.Unlock()

		if d.cursor >= len(rows)-1 && d.cursor > 0 {
			d.cursor = len(rows) - 1
		}

		return nil
	}
}
