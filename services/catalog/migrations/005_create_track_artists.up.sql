CREATE TABLE IF NOT EXISTS track_artists (
    track_id  UUID NOT NULL REFERENCES tracks(id),
    artist_id UUID NOT NULL REFERENCES artists(id),
    role      TEXT NOT NULL DEFAULT 'primary',
    position  INT NOT NULL DEFAULT 0,
    PRIMARY KEY (track_id, artist_id)
);

CREATE INDEX idx_track_artists_artist_id ON track_artists(artist_id);
CREATE INDEX idx_track_artists_role ON track_artists(role) WHERE role = 'primary';
