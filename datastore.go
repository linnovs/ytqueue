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

func videoToRow(v database.Video) row {
	return row{
		colName:     v.Name,
		colURL:      v.Url,
		colLocation: v.Location,
		colWatched:  boolToYesNo(*v.IsWatched),
	}
}

func videosToRows(videos []database.Video) []row {
	rows := make([]row, 0, len(videos))

	for _, v := range videos {
		rows = append(rows, videoToRow(v))
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

func (s *datastore) addVideo(
	ctx context.Context,
	name, url, location string,
) (*database.Video, error) {
	video, err := s.queries.AddVideo(ctx, database.AddVideoParams{
		Name:     name,
		Url:      url,
		Location: location,
	})
	if err != nil {
		return nil, err
	}

	return &video, nil
}

func (s *datastore) Close() error {
	return s.queries.Close()
}
