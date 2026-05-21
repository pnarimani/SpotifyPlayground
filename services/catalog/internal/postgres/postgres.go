package postgres

import (
	"catalog/internal/config"
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewID() uuid.UUID {
	id, err := uuid.NewV7()
	if err != nil {
		panic(fmt.Sprintf("failed to generate UUIDv7: %v", err))
	}
	return id
}

type Track struct {
	ID          uuid.UUID
	AlbumID     uuid.UUID
	Name        string
	TrackNumber int
	DiscNumber  int
	DurationMs  int
	Explicit    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Album struct {
	ID          uuid.UUID
	Title       string
	ReleaseDate *time.Time
	CoverURL    *string
	Label       *string
	TotalTracks int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Artist struct {
	ID        uuid.UUID
	Name      string
	ImageURL  *string
	Bio       *string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Cursor struct {
	CreatedAt time.Time
	ID        uuid.UUID
}

type store struct {
	log  *slog.Logger
	pool *pgxpool.Pool
}

type Store interface {
	ListAlbums(ctx context.Context, cursor *Cursor, limit int) ([]Album, error)
	ListAlbumTracks(ctx context.Context, albumID uuid.UUID) ([]Track, error)
	Close() error
}

func New(ctx context.Context, cfg config.Config, logger *slog.Logger) (Store, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	poolCfg, err := pgxpool.ParseConfig(cfg.PostgresDSN)
	if err != nil {
		return nil, fmt.Errorf("parse pg config: %w", err)
	}

	poolCfg.MaxConns = 20
	poolCfg.MinConns = 4
	poolCfg.MaxConnIdleTime = 5 * time.Minute
	poolCfg.MaxConnLifetime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("create pg pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return &store{logger, pool}, nil
}

func (s *store) ListAlbums(ctx context.Context, cursor *Cursor, limit int) ([]Album, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	var rows pgx.Rows
	var err error

	if cursor == nil {
		// First page — no cursor, just grab the newest.
		rows, err = s.pool.Query(ctx, `
			SELECT id, title, release_date, cover_url, label, total_tracks, created_at, updated_at
			FROM albums
			WHERE deleted_at IS NULL
			ORDER BY created_at DESC, id DESC
			LIMIT $1
		`, limit)
	} else {
		// Keyset pagination: fetch rows older than the cursor position.
		rows, err = s.pool.Query(ctx, `
			SELECT id, title, release_date, cover_url, label, total_tracks, created_at, updated_at
			FROM albums
			WHERE deleted_at IS NULL
			  AND (created_at, id) < ($1, $2)
			ORDER BY created_at DESC, id DESC
			LIMIT $3
		`, cursor.CreatedAt, cursor.ID, limit)
	}
	if err != nil {
		return nil, fmt.Errorf("query albums: %w", err)
	}
	defer rows.Close()

	var albums []Album
	for rows.Next() {
		var a Album
		if err := rows.Scan(
			&a.ID,
			&a.Title,
			&a.ReleaseDate,
			&a.CoverURL,
			&a.Label,
			&a.TotalTracks,
			&a.CreatedAt,
			&a.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan album: %w", err)
		}
		albums = append(albums, a)
	}

	return albums, rows.Err()
}

func (s *store) ListAlbumTracks(ctx context.Context, albumID uuid.UUID) ([]Track, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, album_id, name, track_number, disc_number, duration_ms, explicit, created_at, updated_at
		FROM tracks
		WHERE album_id = $1 AND deleted_at IS NULL
		ORDER BY disc_number, track_number
	`, albumID)
	if err != nil {
		return nil, fmt.Errorf("query tracks: %w", err)
	}
	defer rows.Close()

	var tracks []Track
	for rows.Next() {
		var t Track
		if err := rows.Scan(
			&t.ID,
			&t.AlbumID,
			&t.Name,
			&t.TrackNumber,
			&t.DiscNumber,
			&t.DurationMs,
			&t.Explicit,
			&t.CreatedAt,
			&t.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan track: %w", err)
		}

		tracks = append(tracks, t)
	}

	return tracks, rows.Err()
}

func (s *store) Close() error {
	s.pool.Close()
	return nil
}
