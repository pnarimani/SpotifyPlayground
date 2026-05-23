package postgres

import (
	"catalog/internal/albums"
	"catalog/internal/config"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type albumRow struct {
	ID          uuid.UUID
	Title       string
	ReleaseDate *time.Time
	CoverURL    *string
	Label       *string
	TotalTracks int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type trackRow struct {
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

type Store struct {
	log  *slog.Logger
	pool *pgxpool.Pool
}

func New(ctx context.Context, cfg config.Config, logger *slog.Logger) (*Store, error) {
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

	return &Store{logger, pool}, nil
}

func (s *Store) GetAlbum(ctx context.Context, id string) (*albums.Album, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid album id: %w", err)
	}

	row := s.pool.QueryRow(ctx, `
		SELECT id, title, release_date, cover_url, label, total_tracks, created_at, updated_at
		FROM albums
		WHERE deleted_at IS NULL AND id = $1
	`, uid)

	var a albumRow
	if err := row.Scan(
		&a.ID,
		&a.Title,
		&a.ReleaseDate,
		&a.CoverURL,
		&a.Label,
		&a.TotalTracks,
		&a.CreatedAt,
		&a.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, albums.ErrNotFound
		}
		return nil, fmt.Errorf("scan album: %w", err)
	}

	return albumRowToDomain(&a), nil
}

func (s *Store) ListAlbums(ctx context.Context, cursor *albums.Cursor, limit int) ([]albums.Album, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	var rows pgx.Rows
	var err error

	if cursor == nil {
		rows, err = s.pool.Query(ctx, `
			SELECT id, title, release_date, cover_url, label, total_tracks, created_at, updated_at
			FROM albums
			WHERE deleted_at IS NULL
			ORDER BY created_at DESC, id DESC
			LIMIT $1
		`, limit)
	} else {
		cursorID, parseErr := uuid.Parse(cursor.ID)
		if parseErr != nil {
			return nil, fmt.Errorf("invalid cursor id: %w", parseErr)
		}
		rows, err = s.pool.Query(ctx, `
			SELECT id, title, release_date, cover_url, label, total_tracks, created_at, updated_at
			FROM albums
			WHERE deleted_at IS NULL
			  AND (created_at, id) < ($1, $2)
			ORDER BY created_at DESC, id DESC
			LIMIT $3
		`, cursor.CreatedAt, cursorID, limit)
	}
	if err != nil {
		return nil, fmt.Errorf("query albums: %w", err)
	}
	defer rows.Close()

	var result []albums.Album
	for rows.Next() {
		var a albumRow
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
		result = append(result, *albumRowToDomain(&a))
	}

	return result, rows.Err()
}

func (s *Store) ListAlbumTracks(ctx context.Context, albumID string) ([]albums.Track, error) {
	uid, err := uuid.Parse(albumID)
	if err != nil {
		return nil, fmt.Errorf("invalid album id: %w", err)
	}

	rows, err := s.pool.Query(ctx, `
		SELECT id, album_id, name, track_number, disc_number, duration_ms, explicit, created_at, updated_at
		FROM tracks
		WHERE album_id = $1 AND deleted_at IS NULL
		ORDER BY disc_number, track_number
	`, uid)
	if err != nil {
		return nil, fmt.Errorf("query tracks: %w", err)
	}
	defer rows.Close()

	var tracks []albums.Track
	for rows.Next() {
		var t trackRow
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
		tracks = append(tracks, trackRowToDomain(&t))
	}

	return tracks, rows.Err()
}

func (s *Store) GetTrack(ctx context.Context, id string) (*albums.Track, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid track id: %w", err)
	}

	row := s.pool.QueryRow(ctx, `
		SELECT id, album_id, name, track_number, disc_number, duration_ms, explicit, created_at, updated_at
		FROM tracks
		WHERE id = $1 AND deleted_at IS NULL
	`, uid)

	var t trackRow
	if err := row.Scan(
		&t.ID, &t.AlbumID, &t.Name, &t.TrackNumber,
		&t.DiscNumber, &t.DurationMs, &t.Explicit, &t.CreatedAt, &t.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, albums.ErrNotFound
		}
		return nil, fmt.Errorf("scan track: %w", err)
	}

	track := trackRowToDomain(&t)
	return &track, nil
}

func (s *Store) Close() error {
	s.pool.Close()
	return nil
}

func albumRowToDomain(a *albumRow) *albums.Album {
	return &albums.Album{
		ID:          a.ID.String(),
		Title:       a.Title,
		ReleaseDate: a.ReleaseDate,
		CoverURL:    a.CoverURL,
		Label:       a.Label,
		TotalTracks: a.TotalTracks,
	}
}

func trackRowToDomain(t *trackRow) albums.Track {
	return albums.Track{
		ID:          t.ID.String(),
		AlbumID:     t.AlbumID.String(),
		Name:        t.Name,
		TrackNumber: t.TrackNumber,
		DiscNumber:  t.DiscNumber,
		DurationMs:  t.DurationMs,
		Explicit:    t.Explicit,
	}
}
