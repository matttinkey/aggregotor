-- +goose Up
CREATE TABLE posts(
    id INTEGER PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    title TEXT NOT NULL,
    url TEXT NOT NULL UNIQUE,
    description TEXT,
    published_at TIMESTAMP NOT NULL,
    feed_id INTEGER,
    FOREIGN KEY (feed_id)
    REFERENCES feeds(id)
);

-- +goose Down
DROP TABLE posts;