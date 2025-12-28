package main

import (
	"errors"
	"fmt"
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
		if d.playingId == id && d.player.playing.Load() {
			if err := d.player.stop(); err != nil {
				return errorMsg{err: fmt.Errorf("failed to stop player: %w", err)}
			}

			return finishPlayMsg{}
		}

		d.playingId = id
		idx := slices.IndexFunc(d.rows, d.playingIDIndexFunc)
		row := d.rows[idx]
		fpath := filepath.Join(row[colLocation], row[colName])
		fpath = filepath.Clean(fpath)

		_, err := os.Stat(filepath) // #nosec G104
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return errorMsg{errors.New("file does not exist")}
			}

			return errorMsg{err: fmt.Errorf("failed to access file: %w", err)}
		}

		if err := d.player.play(filepath); err != nil {
			return errorMsg{err: fmt.Errorf("failed to play file: %w", err)}
		}

		return finishPlayMsg{}
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
