CREATE TABLE IF NOT EXISTS albums (
    id           UUID PRIMARY KEY,
    title        TEXT NOT NULL,
    release_date DATE,
    cover_url    TEXT,
    label        TEXT,
    total_tracks INT NOT NULL DEFAULT 0,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at   TIMESTAMPTZ
);

CREATE INDEX idx_albums_release_date ON albums(release_date DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_albums_cursor ON albums(created_at DESC, id DESC) WHERE deleted_at IS NULL;
