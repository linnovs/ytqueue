package main

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/linnovs/ytqueue/database"
)

const (
	isWatchedYes = "✅"
	isWatchedNo  = "❌"
)

func boolToYesNo(b bool) string {
	if b {
		return isWatchedYes
	}

	return isWatchedNo
}

func videoToRow(v database.Video) row {
	return row{
		colID:       fmt.Sprint(v.ID),
		colName:     v.Name,
		colURL:      v.Url,
		colLocation: v.Location,
		colWatched:  boolToYesNo(*v.IsWatched),
		colOrder:    fmt.Sprint(v.OrderIndex.Unix()),
	}
}

func videosToRows(videos []database.Video) []row {
	rows := make([]row, 0, len(videos))

	for _, v := range videos {
		rows = append(rows, videoToRow(v))
	}

	return rows
}

func idStrToInt(idStr string) (int64, error) {
	return strconv.ParseInt(idStr, 10, 0)
}

type datastore struct{ queries *database.Queries }

func newDatastore(queries *database.Queries) *datastore {
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

func (s *datastore) updateVideoOrder(ctx context.Context, idStr string, orderUnix int64) error {
	id, err := idStrToInt(idStr)
	if err != nil {
		return err
	}

	idx := time.Unix(orderUnix, 0).UTC()

	if err := s.queries.UpdateVideoOrder(ctx, database.UpdateVideoOrderParams{
		ID:         id,
		OrderIndex: &idx,
	}); err != nil {
		return err
	}

	return nil
}

func (s *datastore) setWatched(ctx context.Context, idStr string) (*database.Video, error) {
	id, err := idStrToInt(idStr)
	if err != nil {
		return nil, err
	}

	video, err := s.queries.SetWatchedVideo(ctx, id)
	if err != nil {
		return nil, err
	}

	return &video, nil
}

func (s *datastore) toggleWatched(ctx context.Context, idStr string) (*database.Video, error) {
	id, err := idStrToInt(idStr)
	if err != nil {
		return nil, err
	}

	video, err := s.queries.ToggleWatchedStatus(ctx, id)
	if err != nil {
		return nil, err
	}

	return &video, nil
}

func (s *datastore) deleteVideo(ctx context.Context, idStr string) error {
	id, err := idStrToInt(idStr)
	if err != nil {
		return err
	}

	return s.queries.DeleteVideo(ctx, id)
}

func (s *datastore) Close() error {
	return s.queries.Close()
}
