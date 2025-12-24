-- name: GetVideos :many
SELECT id, name, url, file_path, is_watched, order_index, created_at FROM videos;
