-- +goose Up
CREATE TABLE IF NOT EXISTS hosts (
    id INTEGER PRIMARY KEY,
    host VARCHAR NOT NULL,
    status INTEGER NOT NULL,
    is_wiki BOOLEAN NOT NULL,
    instance_actor_id INTEGER,

    UNIQUE (host),
    FOREIGN KEY (instance_actor_id) REFERENCES actors (id)
);

-- +goose Down
DROP TABLE hosts;
