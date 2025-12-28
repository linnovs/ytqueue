package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"

	tea "github.com/charmbracelet/bubbletea"
)

func (d *datatable) newVideoCmd(name, url, location string) tea.Cmd {
	return func() tea.Msg {
		video, err := d.datastore.addVideo(d.getCtx(), name, url, location)
		if err != nil {
			return errorMsg{err}
		}

		rows := append([]row{videoToRow(*video)}, d.rows...)
		d.setRows(rows)

		return nil
	}
}

func (d *datatable) playStopRowCmd(id string) tea.Cmd {
	return func() tea.Msg {
		slog.Debug(
			"playStopRowCmd",
			slog.String("currentPlayingId", d.playingId),
			slog.String("requestedId", id),
		)

		if d.playingId == id {
			d.playingId = ""

			if d.player.isPlaying() {
				if err := d.player.stop(); err != nil {
					return errorMsg{err: fmt.Errorf("failed to stop player: %w", err)}
				}
			}

			return nil
		}

		d.playingId = id

		idx := slices.IndexFunc(d.rows, d.playingIDIndexFunc)
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

		if err := d.player.play(file); err != nil {
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

	d.rowMu.Lock()
	idx := slices.IndexFunc(d.rows, d.playingIDIndexFunc)
	d.rows[idx] = videoToRow(*video)
	d.rowMu.Unlock()

	return idx, nil
}

func (d *datatable) playNextOrStopCmd() tea.Cmd {
	return func() tea.Msg {
		idx, err := d.setVideoWatched(d.playingId)
		if err != nil {
			return errorMsg{fmt.Errorf("failed to set video as watched: %w", err)}
		}

		if idx <= 0 {
			d.playingId = ""

			return nil
		}

		return d.playStopRowCmd(d.rows[idx-1][colID])()
	}
}

func (d *datatable) toggleWatchedStatusCmd(cursor int) tea.Cmd {
	return func() tea.Msg {
		rows := append([]row{}, d.rows...)

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
		row := d.rows[cursor]
		rows := append(d.rows[:cursor], d.rows[cursor+1:]...)

		if err := d.datastore.deleteVideo(d.getCtx(), row[colID]); err != nil {
			return errorMsg{err}
		}

		d.setRows(rows)

		return nil
	}
}
