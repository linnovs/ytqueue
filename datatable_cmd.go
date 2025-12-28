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
