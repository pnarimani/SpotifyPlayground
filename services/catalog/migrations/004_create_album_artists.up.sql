CREATE TABLE IF NOT EXISTS album_artists (
    album_id  UUID NOT NULL REFERENCES albums(id),
    artist_id UUID NOT NULL REFERENCES artists(id),
    role      TEXT NOT NULL DEFAULT 'primary',
    position  INT NOT NULL DEFAULT 0,
    PRIMARY KEY (album_id, artist_id)
);

CREATE INDEX idx_album_artists_artist_id ON album_artists(artist_id);
CREATE INDEX idx_album_artists_role ON album_artists(role) WHERE role = 'primary';
