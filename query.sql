-- name: GetVideos :many
SELECT id, name, url, location, is_watched, order_index, created_at FROM videos ORDER BY order_index DESC;
