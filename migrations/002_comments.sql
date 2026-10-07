CREATE TABLE IF NOT EXISTS comments (
    id UUID PRIMARY KEY,
    topic_id UUID NOT NULL REFERENCES topics(id) ON DELETE CASCADE,
    author_id UUID NOT NULL REFERENCES authors(id),
    language TEXT NOT NULL,
    body TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS comments_topic_created_idx ON comments (topic_id, created_at ASC);
CREATE INDEX IF NOT EXISTS comments_author_idx ON comments (author_id);
