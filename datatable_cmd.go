package main

import (
	"errors"
	"fmt"
	"os"
	"path"

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

func (d *datatable) playStopRowCmd(cursor int) tea.Cmd {
	return func() tea.Msg {
		if d.playingRow == cursor && d.player.playing {
			if err := d.player.stop(); err != nil {
				return errorMsg{err: fmt.Errorf("failed to stop player: %w", err)}
			}

			return finishPlayMsg{}
		}

		row := d.rows[cursor]
		filepath := path.Join(row[colLocation], row[colName])
		filepath = path.Clean(filepath)

		_, err := os.Stat(filepath) // #nosec G104
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return errorMsg{errors.New("file does not exist")}
			}

			return errorMsg{err: fmt.Errorf("failed to access file: %w", err)}
		}

		if d.playingRow == cursor {
			d.playingRow = -1

			return nil
		}

		d.playingRow = cursor

		if err := d.player.play(filepath); err != nil {
			return errorMsg{err: fmt.Errorf("failed to play file: %w", err)}
		}

		return finishPlayMsg{}
	}
}

func (d *datatable) setVideoWatchedCmd(cursor int) tea.Cmd {
	return func() tea.Msg {
		video, err := d.datastore.setWatched(d.getCtx(), d.rows[cursor][colID])
		if err != nil {
			return errorMsg{err: fmt.Errorf("failed to set video as watched: %w", err)}
		}

		videoRow := videoToRow(*video)

		rows := append([]row{}, d.rows...)
		rows[cursor] = videoRow
		d.setRows(rows)

		return nil
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
