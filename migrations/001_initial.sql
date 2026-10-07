CREATE TABLE IF NOT EXISTS authors (
    id UUID PRIMARY KEY,
    nickname TEXT NOT NULL,
    email TEXT NOT NULL,
    bio TEXT NOT NULL DEFAULT '',
    preferred_language TEXT NOT NULL,
    status TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS authors_nickname_lower_idx ON authors (lower(nickname));
CREATE UNIQUE INDEX IF NOT EXISTS authors_email_lower_idx ON authors (lower(email));

CREATE TABLE IF NOT EXISTS topics (
    id UUID PRIMARY KEY,
    slug TEXT NOT NULL UNIQUE,
    original_language TEXT NOT NULL,
    status TEXT NOT NULL,
    created_by UUID NOT NULL REFERENCES authors(id),
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS topics_status_created_idx ON topics (status, created_at DESC);

CREATE TABLE IF NOT EXISTS topic_translations (
    topic_id UUID NOT NULL REFERENCES topics(id) ON DELETE CASCADE,
    language TEXT NOT NULL,
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    source TEXT NOT NULL,
    translated_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (topic_id, language)
);
CREATE INDEX IF NOT EXISTS topic_translations_language_idx ON topic_translations (language, topic_id);
