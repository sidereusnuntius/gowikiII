-- +goose Up
CREATE TABLE IF NOT EXISTS shared_inboxes (
       id INTEGER PRIMARY KEY,
       uri TEXT NOT NULL,
       UNIQUE (uri)
);

CREATE TABLE IF NOT EXISTS actors (
       id INTEGER PRIMARY KEY,
       uri TEXT NOT NULL,
       type TEXT NOT NULL,

       username VARCHAR NOT NULL,
       host VARCHAR NOT NULL,
       name VARCHAR,
       summary TEXT,

       inbox TEXT,
       outbox TEXT,
       followers TEXT,
       following TEXT,
       url TEXT,
       shared_inbox INTEGER,

       user_id INTEGER,

       published INTEGER DEFAULT (unixepoch('now')),
       updated INTEGER,

       UNIQUE (uri),
       UNIQUE (username, host),
       FOREIGN KEY (user_id) REFERENCES users (id),
       FOREIGN KEY (shared_inbox) REFERENCES shared_inboxes (id)
);

CREATE TABLE IF NOT EXISTS public_keys (
       id INTEGER PRIMARY KEY,
       iri TEXT NOT NULL,
       owner_id INTEGER NOT NULL,
       type INTEGER,
       key_pem BINARY NOT NULL,

       UNIQUE (iri),
       FOREIGN KEY (owner_id) REFERENCES actors (id)
);

CREATE TABLE IF NOT EXISTS private_keys (
       id INTEGER PRIMARY KEY,
       key_pem BINARY NOT NULL,
       owner_id INTEGER NOT NULL,
       type INTEGER,

       FOREIGN KEY (owner_id) REFERENCES actors (id)
);

CREATE INDEX IF NOT EXISTS idx_owner_pub_key ON public_keys (owner_id);
CREATE INDEX IF NOT EXISTS idx_owner_priv_key ON private_keys (owner_id);
CREATE INDEX IF NOT EXISTS idx_actor_user ON actors (user_id);

-- +goose Down
DROP INDEX idx_actor_user;
DROP INDEX idx_owner_priv_key;
DROP INDEX idx_owner_pub_key;
DROP TABLE actors;
DROP TABLE shared_inboxes;
