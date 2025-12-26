package main

import (
	"context"
	"database/sql"

	"github.com/linnovs/ytqueue/database"
)

func boolToYesNo(b bool) string {
	if b {
		return "✅"
	}

	return "❌"
}

func videosToRows(videos []database.Video) []row {
	rows := make([]row, 0, len(videos))

	for _, v := range videos {
		rows = append(rows, row{
			colName:     v.Name,
			colURL:      v.Url,
			colLocation: v.Location,
			colWatched:  boolToYesNo(*v.IsWatched),
		})
	}

	return rows
}

type datastore struct{ queries *database.Queries }

func newDatastore(db *sql.DB) *datastore {
	queries := database.New(db)

	return &datastore{queries: queries}
}

func (s *datastore) getVideos(ctx context.Context) ([]database.Video, error) {
	videos, err := s.queries.GetVideos(ctx)
	if err != nil {
		return nil, err
	}

	return videos, nil
}

func (s *datastore) Close() error {
	return s.queries.Close()
}
