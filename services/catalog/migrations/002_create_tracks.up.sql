CREATE TABLE IF NOT EXISTS tracks (
    id           UUID PRIMARY KEY,
    album_id     UUID NOT NULL REFERENCES albums(id),
    name         TEXT NOT NULL,
    track_number INT NOT NULL,
    disc_number  INT NOT NULL DEFAULT 1,
    duration_ms  INT NOT NULL,
    explicit     BOOLEAN NOT NULL DEFAULT false,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at   TIMESTAMPTZ
);

CREATE INDEX idx_tracks_album_order ON tracks(album_id, disc_number, track_number) WHERE deleted_at IS NULL;
CREATE INDEX idx_tracks_created_at ON tracks(created_at DESC, id DESC) WHERE deleted_at IS NULL;
