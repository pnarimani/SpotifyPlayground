package postgres

import (
	"context"
	"fmt"
	"log/slog"
	"royalties/internal/config"
	"time"

	playback "contracts/events/playback/v1"

	"github.com/jackc/pgx/v5/pgxpool"
)

type RawEvent playback.EventMessage

type store struct {
	log  *slog.Logger
	pool *pgxpool.Pool
}

type Store interface {
	WriteRawPlayEvent(ctx context.Context, ev RawEvent) error
	Close() error
}

func New(ctx context.Context, cfg config.Config, logger *slog.Logger) (Store, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, cfg.PostgresDSN)
	if err != nil {
		return nil, fmt.Errorf("create pg pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return &store{logger, pool}, nil
}

func (s *store) WriteRawPlayEvent(ctx context.Context, ev RawEvent) error {
	const q = `
		INSERT INTO raw_play_events (
			event_id,
			user_id,
			track_id,
			playback_session_id,
			request_sequence_id,
			event_type,
			position_ms,
			client_timestamp,
			server_received_at,
			country
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (event_id) DO NOTHING;
		`

	_, err := s.pool.Exec(
		ctx, q,
		ev.EventID,
		ev.UserID,
		ev.TrackId,
		ev.PlaybackSessionID,
		ev.RequestSequenceID,
		string(ev.EventType),
		ev.PositionMs,
		ev.ClientTimestamp,
		ev.ServerReceivedAt,
		ev.Country,
	)
	if err != nil {
		return fmt.Errorf("failed to execute query. err: %w", err)
	}

	return nil
}

func (s *store) Close() error {
	s.pool.Close()
	return nil
}
