CREATE TABLE IF NOT EXISTS raw_play_events (
     event_id TEXT PRIMARY KEY,
     user_id TEXT NOT NULL,
     track_id TEXT NOT NULL,
     playback_session_id TEXT NOT NULL,
     request_sequence_id INT NOT NULL,
     event_type TEXT NOT NULL,
     position_ms INT NOT NULL,
     client_timestamp BIGINT NOT NULL,
     server_received_at BIGINT NOT NULL,
     country TEXT NOT NULL,
     created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
