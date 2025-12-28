-- name: GetVideos :many
SELECT id, name, url, location, is_watched, order_index, created_at FROM videos ORDER BY order_index DESC;

-- name: AddVideo :one
INSERT INTO videos (name, url, location) values (?, ?, ?) RETURNING *;

-- name: ToggleWatchedStatus :one
UPDATE videos SET is_watched = not is_watched WHERE id = ? RETURNING *;

-- name: SetWatchedVideo :one
UPDATE videos SET is_watched = true WHERE id = ? RETURNING *;

-- name: DeleteVideo :exec
DELETE FROM videos WHERE id = ?;
