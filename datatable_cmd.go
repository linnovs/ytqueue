package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type deletedRowMsg struct {
	filename string
	notFound bool
}

type deletedMultipleRowsMsg struct {
	filenames []string
	notFounds []string
	dbErrors  []error
}

type updateRowOrderMsg struct {
	id  string
	tag int
}

type updatedRowOrderMsg struct{}

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
	}, func() tea.Msg {
		if d.isInitialRefresh {
			d.isInitialRefresh = false

			return nil
		}

		return footerMsgCmd("Refreshed video list", 0)()
	})
}

func (d *datatable) newVideoCmd(name, url, location string) tea.Cmd {
	return func() tea.Msg {
		video, err := d.datastore.addVideo(d.getCtx(), name, url, location)
		if err != nil {
			return errorMsg{err}
		}

		oldId := d.getCursorID()
		rows := append([]row{videoToRow(*video)}, d.getCopyOfRows()...)

		d.setRows(rows)

		d.cursorMu.Lock()
		idx := slices.IndexFunc(rows, playingIDIndexFunc(oldId))
		d.cursor = clamp(idx, 0, len(rows)-1)
		d.cursorMu.Unlock()

		return nil
	}
}

func (d *datatable) playStopRowCmd(id string) tea.Cmd {
	return func() tea.Msg {
		d.rowMu.RLock()
		defer d.rowMu.RUnlock()

		slog.Debug("playStopRowCmd", slog.String("requestedId", id))

		if d.player.isRunning() && d.player.getCurrentlyPlayingId() == id && d.player.isPlaying() {
			if err := d.player.stop(); err != nil {
				return errorMsg{err: fmt.Errorf("failed to stop player: %w", err)}
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
				slog.Debug(
					"playing next unwatched video",
					slog.String("id", row[colID]),
					slog.String("name", row[colName]),
				)
				d.player.setPlaying(playingStatusStopped)

				return d.playStopRowCmd(row[colID])()
			}
		}

		return nil
	}
}

func (d *datatable) toggleSelectModeCmd() tea.Cmd {
	return func() tea.Msg {
		d.selectModeMu.Lock()
		defer d.selectModeMu.Unlock()

		d.cursorMu.RLock()
		defer d.cursorMu.RUnlock()

		d.selectMode = !d.selectMode
		d.selectModeStart = d.cursor

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

func (d *datatable) updateRowOrderCmd(id string) tea.Cmd {
	return func() tea.Msg {
		d.rowMu.Lock()
		defer d.rowMu.Unlock()

		slog.Debug("updating row order", slog.String("id", id))

		idx := slices.IndexFunc(d.rows, func(r row) bool { return r[colID] == id })
		upperIdx := clamp(idx-1, 0, len(d.rows)-1)
		lowerIdx := clamp(idx+1, 0, len(d.rows)-1)

		upperOrderUnix, err := strconv.Atoi(d.rows[upperIdx][colOrder])
		if err != nil {
			return errorMsg{fmt.Errorf("failed to parse upper order index: %w", err)}
		}

		lowerOrderUnix, err := strconv.Atoi(d.rows[lowerIdx][colOrder])
		if err != nil {
			return errorMsg{fmt.Errorf("failed to parse lower order index: %w", err)}
		}

		const inHalf = 2
		newOrderUnix := int64((upperOrderUnix + lowerOrderUnix) / inHalf)

		if err := d.datastore.updateVideoOrder(d.getCtx(), id, newOrderUnix); err != nil {
			return errorMsg{fmt.Errorf("failed to update video order: %w", err)}
		}

		slog.Debug(
			"updated row order",
			slog.String("id", id),
			slog.Int64("new_order_unix", newOrderUnix),
		)

		return updatedRowOrderMsg{}
	}
}

func (d *datatable) deleteRowCmd(cursor int) tea.Cmd {
	return func() tea.Msg {
		rows := d.getCopyOfRows()
		row := rows[cursor]
		rows = append(rows[:cursor], rows[cursor+1:]...)
		msg := deletedRowMsg{filename: row[colName]}

		fname := filepath.Clean(filepath.Join(row[colLocation], row[colName]))
		if err := os.Remove(fname); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				msg.notFound = true
			} else {
				return errorMsg{fmt.Errorf("failed to delete video file: %w", err)}
			}
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

		return msg
	}
}

func (d *datatable) deleteSelectedRowsCmd() tea.Cmd {
	return func() tea.Msg {
		d.selectModeMu.RLock()
		defer d.selectModeMu.RUnlock()

		rows := d.getCopyOfRows()
		deletedFilenames := make([]string, 0)
		deleteFailedFiles := make([]string, 0)
		remainingRows := make([]row, 0)
		dbErrors := make([]error, 0)

		for idx, row := range rows {
			if !d.isSelected(idx) {
				remainingRows = append(remainingRows, row)
				continue
			}

			deleted := true

			fname := filepath.Clean(filepath.Join(row[colLocation], row[colName]))
			if err := os.Remove(fname); err != nil {
				if !errors.Is(err, os.ErrNotExist) {
					deleteFailedFiles = append(deleteFailedFiles, row[colName])
					deleted = false
				}
			}

			if err := d.datastore.deleteVideo(d.getCtx(), row[colID]); err != nil {
				dbErrors = append(
					dbErrors,
					fmt.Errorf("failed to delete video %s from datastore: %w", row[colName], err),
				)
				deleted = false
			}

			if !deleted {
				remainingRows = append(remainingRows, row)
			} else {
				deletedFilenames = append(deletedFilenames, row[colName])
			}
		}

		d.setRows(remainingRows)

		d.cursorMu.Lock()
		if d.cursor >= len(remainingRows)-1 && d.cursor > 0 {
			d.cursor = len(remainingRows) - 1
		}
		d.cursorMu.Unlock()

		return deletedMultipleRowsMsg{
			filenames: deletedFilenames,
			notFounds: deleteFailedFiles,
			dbErrors:  dbErrors,
		}
	}
}
