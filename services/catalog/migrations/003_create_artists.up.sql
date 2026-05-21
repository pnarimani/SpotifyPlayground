CREATE TABLE IF NOT EXISTS artists (
    id         UUID PRIMARY KEY,
    name       TEXT NOT NULL,
    image_url  TEXT,
    bio        TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_artists_name ON artists(name) WHERE deleted_at IS NULL;
CREATE INDEX idx_artists_created_at ON artists(created_at DESC, id DESC) WHERE deleted_at IS NULL;
