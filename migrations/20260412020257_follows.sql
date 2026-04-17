-- +goose Up
CREATE TABLE IF NOT EXISTS follows (
    id INTEGER PRIMARY KEY,
    iri VARCHAR NOT NULL,
    follower_id INTEGER NOT NULL,
    followee_id INTEGER NOT NULL,
    accepted BOOLEAN,
    published INTEGER DEFAULT (unixepoch('now')) NOT NULL,

    UNIQUE (iri),
    UNIQUE (follower_id, followee_id),
    FOREIGN KEY (follower_id) REFERENCES actors (id),
    FOREIGN KEY (followee_id) REFERENCES actors (id)
);

CREATE INDEX IF NOT EXISTS idx_followee_accepted ON follows (followee_id, accepted);

-- +goose Down
DROP INDEX idx_followee_accepted;
DROP TABLE follows;
