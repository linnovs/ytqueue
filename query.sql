-- name: GetVideos :many
SELECT id, name, url, location, is_watched, order_index, created_at FROM videos ORDER BY order_index DESC;

-- name: AddVideo :one
INSERT INTO videos (name, url, location) values (?, ?, ?) RETURNING *;

-- name: DeleteVideo :exec
DELETE FROM videos WHERE id = ?;
