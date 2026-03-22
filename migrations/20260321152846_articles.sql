-- +goose Up
CREATE TABLE IF NOT EXISTS articles (
    id INTEGER PRIMARY KEY,
    slug TEXT NOT NULL,
    host TEXT NOT NULL,
    iri TEXT NOT NULL,
    published INTEGER DEFAULT (unixepoch('now')),
    -- Whether the article can be edited by users from other servers.
    federated_edits BOOLEAN DEFAULT false,
    restricted_edits BOOLEAN DEFAULT false,

    UNIQUE (iri),
    UNIQUE (slug, host)
);

CREATE TABLE IF NOT EXISTS localized_articles (
    id INTEGER PRIMARY KEY,
    lang_code TEXT NOT NULL,
    --
    title TEXT,
    content TEXT NOT NULL,
    summary TEXT,
    url TEXT,
    article_id INTEGER NOT NULL,
    published INTEGER DEFAULT (unixepoch('now')),
    updated INTEGER,
    last_fetched INTEGER, -- Only used for foreign articles.

    FOREIGN KEY (article_id) REFERENCES articles (id)
    UNIQUE (article_id, lang_code),
    UNIQUE (url)
);

CREATE TABLE IF NOT EXISTS revisions (
    id INTEGER PRIMARY KEY,
    iri TEXT NOT NULL,
    code TEXT NOT NULL,
    diff TEXT NOT NULL,
    prev_revision_id INTEGER,
    summary TEXT,
    localized_article_id INTEGER NOT NULL,
    published INTEGER DEFAULT (unixepoch('now')) NOT NULL,
    actor_internal_id INTEGER,
    actor_iri TEXT,

    UNIQUE (iri),
    FOREIGN KEY (actor_internal_id) REFERENCES actors (id),
    FOREIGN KEY (localized_article_id) REFERENCES localized_articles (id),
    FOREIGN KEY (prev_revision_id) REFERENCES revisions (id)
);

CREATE TABLE IF NOT EXISTS categories (
    id INTEGER PRIMARY KEY,
    slug TEXT NOT NULL,
    host TEXT NOT NULL,
    iri TEXT  UNIQUE NOT NULL,

    UNIQUE (slug, host)
);

CREATE TABLE IF NOT EXISTS article_categories (
    article_id INTEGER NOT NULL,
    category_id INTEGER NOT NULL,

    FOREIGN KEY (article_id) REFERENCES articles (id),
    FOREIGN KEY (category_id) REFERENCES categories (id),
    PRIMARY KEY (article_id, category_id)
);

CREATE INDEX IF NOT EXISTS idx_local_articles_id ON localized_articles (article_id);
CREATE INDEX IF NOT EXISTS idx_revision_local_article ON revisions (localized_article_id);
CREATE INDEX IF NOT EXISTS idx_revision_local_article_timestamp ON revisions (localized_article_id, published);
CREATE INDEX IF NOT EXISTS idx_revision_actor_internal ON revisions (actor_internal_id, published);

-- +goose Down
DROP INDEX idx_revision_actor_internal;
DROP INDEX idx_revision_local_article_timestamp;
DROP INDEX idx_revision_local_article;
DROP INDEX idx_local_articles_id;
DROP TABLE article_categories;
DROP TABLE categories;
DROP TABLE revisions;
DROP TABLE localized_articles;
DROP TABLE articles;
