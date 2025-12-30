ALTER TABLE videos RENAME TO old_videos;

CREATE TABLE videos (
    id INTEGER PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL,
    url VARCHAR UNIQUE NOT NULL,
    location VARCHAR NOT NULL,
    is_watched BOOLEAN DEFAULT FALSE,
    order_index TIMESTAMP UNIQUE DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO videos SELECT * FROM old_videos
ON CONFLICT(name,url) DO UPDATE
SET name=excluded.name || "-dup" || RANDOM, url=excluded.url || "-dup" || RANDOM;

DROP TABLE old_videos;
