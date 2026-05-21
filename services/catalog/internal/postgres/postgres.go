package postgres

import (
	"catalog/internal/albums"
	"catalog/internal/config"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
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

func (s *Store) CreateAlbum(ctx context.Context, params albums.CreateAlbumParams) (*albums.Album, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("generate uuid v7: %w", err)
	}

	row := s.pool.QueryRow(ctx, `
		INSERT INTO albums (id, title, release_date, cover_url, label, total_tracks)
		VALUES ($1, $2, $3, $4, $5, 0)
		RETURNING id, title, release_date, cover_url, label, total_tracks, created_at, updated_at
	`, id, params.Title, params.ReleaseDate, params.CoverURL, params.Label)

	var a albumRow
	if err := row.Scan(
		&a.ID, &a.Title, &a.ReleaseDate, &a.CoverURL, &a.Label,
		&a.TotalTracks, &a.CreatedAt, &a.UpdatedAt,
	); err != nil {
		return nil, fmt.Errorf("insert album: %w", err)
	}

	return albumRowToDomain(&a), nil
}

func (s *Store) UpdateAlbum(ctx context.Context, id string, params albums.UpdateAlbumParams) (*albums.Album, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid album id: %w", err)
	}

	setClauses := []string{}
	args := []any{}
	argIdx := 1

	if params.Title != nil {
		setClauses = append(setClauses, fmt.Sprintf("title = $%d", argIdx))
		args = append(args, *params.Title)
		argIdx++
	}
	if params.ReleaseDate != nil {
		setClauses = append(setClauses, fmt.Sprintf("release_date = $%d", argIdx))
		args = append(args, params.ReleaseDate)
		argIdx++
	}
	if params.CoverURL != nil {
		setClauses = append(setClauses, fmt.Sprintf("cover_url = $%d", argIdx))
		args = append(args, *params.CoverURL)
		argIdx++
	}
	if params.Label != nil {
		setClauses = append(setClauses, fmt.Sprintf("label = $%d", argIdx))
		args = append(args, *params.Label)
		argIdx++
	}

	if len(setClauses) == 0 {
		return s.GetAlbum(ctx, id)
	}

	setClauses = append(setClauses, fmt.Sprintf("updated_at = $%d", argIdx))
	args = append(args, time.Now())
	argIdx++

	args = append(args, uid)

	query := fmt.Sprintf(`
		UPDATE albums SET %s
		WHERE id = $%d AND deleted_at IS NULL
		RETURNING id, title, release_date, cover_url, label, total_tracks, created_at, updated_at
	`, strings.Join(setClauses, ", "), argIdx)

	row := s.pool.QueryRow(ctx, query, args...)

	var a albumRow
	if err := row.Scan(
		&a.ID, &a.Title, &a.ReleaseDate, &a.CoverURL, &a.Label,
		&a.TotalTracks, &a.CreatedAt, &a.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, albums.ErrNotFound
		}
		return nil, fmt.Errorf("update album: %w", err)
	}

	return albumRowToDomain(&a), nil
}

func (s *Store) DeleteAlbum(ctx context.Context, id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid album id: %w", err)
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// Soft-delete tracks belonging to this album
	_, err = tx.Exec(ctx, `
		UPDATE tracks SET deleted_at = now(), updated_at = now()
		WHERE album_id = $1 AND deleted_at IS NULL
	`, uid)
	if err != nil {
		return fmt.Errorf("delete album tracks: %w", err)
	}

	// Soft-delete the album
	tag, err := tx.Exec(ctx, `
		UPDATE albums SET deleted_at = now(), updated_at = now()
		WHERE id = $1 AND deleted_at IS NULL
	`, uid)
	if err != nil {
		return fmt.Errorf("delete album: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return albums.ErrNotFound
	}

	return tx.Commit(ctx)
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

func (s *Store) CreateTrack(ctx context.Context, params albums.CreateTrackParams) (*albums.Track, error) {
	albumUID, err := uuid.Parse(params.AlbumID)
	if err != nil {
		return nil, fmt.Errorf("invalid album id: %w", err)
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	id, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("generate uuid v7: %w", err)
	}
	row := tx.QueryRow(ctx, `
		INSERT INTO tracks (id, album_id, name, track_number, disc_number, duration_ms, explicit)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, album_id, name, track_number, disc_number, duration_ms, explicit, created_at, updated_at
	`, id, albumUID, params.Name, params.TrackNumber, params.DiscNumber, params.DurationMs, params.Explicit)

	var t trackRow
	if err := row.Scan(
		&t.ID, &t.AlbumID, &t.Name, &t.TrackNumber,
		&t.DiscNumber, &t.DurationMs, &t.Explicit, &t.CreatedAt, &t.UpdatedAt,
	); err != nil {
		return nil, fmt.Errorf("insert track: %w", err)
	}

	// Update album total_tracks count
	_, err = tx.Exec(ctx, `
		UPDATE albums SET total_tracks = (
			SELECT COUNT(*) FROM tracks WHERE album_id = $1 AND deleted_at IS NULL
		), updated_at = now()
		WHERE id = $1
	`, albumUID)
	if err != nil {
		return nil, fmt.Errorf("update album track count: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	track := trackRowToDomain(&t)
	return &track, nil
}

func (s *Store) UpdateTrack(ctx context.Context, id string, params albums.UpdateTrackParams) (*albums.Track, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid track id: %w", err)
	}

	setClauses := []string{}
	args := []any{}
	argIdx := 1

	if params.Name != nil {
		setClauses = append(setClauses, fmt.Sprintf("name = $%d", argIdx))
		args = append(args, *params.Name)
		argIdx++
	}
	if params.TrackNumber != nil {
		setClauses = append(setClauses, fmt.Sprintf("track_number = $%d", argIdx))
		args = append(args, *params.TrackNumber)
		argIdx++
	}
	if params.DiscNumber != nil {
		setClauses = append(setClauses, fmt.Sprintf("disc_number = $%d", argIdx))
		args = append(args, *params.DiscNumber)
		argIdx++
	}
	if params.DurationMs != nil {
		setClauses = append(setClauses, fmt.Sprintf("duration_ms = $%d", argIdx))
		args = append(args, *params.DurationMs)
		argIdx++
	}
	if params.Explicit != nil {
		setClauses = append(setClauses, fmt.Sprintf("explicit = $%d", argIdx))
		args = append(args, *params.Explicit)
		argIdx++
	}

	if len(setClauses) == 0 {
		return s.GetTrack(ctx, id)
	}

	setClauses = append(setClauses, fmt.Sprintf("updated_at = $%d", argIdx))
	args = append(args, time.Now())
	argIdx++

	args = append(args, uid)

	query := fmt.Sprintf(`
		UPDATE tracks SET %s
		WHERE id = $%d AND deleted_at IS NULL
		RETURNING id, album_id, name, track_number, disc_number, duration_ms, explicit, created_at, updated_at
	`, strings.Join(setClauses, ", "), argIdx)

	row := s.pool.QueryRow(ctx, query, args...)

	var t trackRow
	if err := row.Scan(
		&t.ID, &t.AlbumID, &t.Name, &t.TrackNumber,
		&t.DiscNumber, &t.DurationMs, &t.Explicit, &t.CreatedAt, &t.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, albums.ErrNotFound
		}
		return nil, fmt.Errorf("update track: %w", err)
	}

	track := trackRowToDomain(&t)
	return &track, nil
}

func (s *Store) DeleteTrack(ctx context.Context, id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid track id: %w", err)
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// Get album_id before deleting
	var albumID uuid.UUID
	err = tx.QueryRow(ctx, `
		SELECT album_id FROM tracks WHERE id = $1 AND deleted_at IS NULL
	`, uid).Scan(&albumID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return albums.ErrNotFound
		}
		return fmt.Errorf("get track album: %w", err)
	}

	// Soft-delete the track
	_, err = tx.Exec(ctx, `
		UPDATE tracks SET deleted_at = now(), updated_at = now()
		WHERE id = $1 AND deleted_at IS NULL
	`, uid)
	if err != nil {
		return fmt.Errorf("delete track: %w", err)
	}

	// Update album total_tracks count
	_, err = tx.Exec(ctx, `
		UPDATE albums SET total_tracks = (
			SELECT COUNT(*) FROM tracks WHERE album_id = $1 AND deleted_at IS NULL
		), updated_at = now()
		WHERE id = $1
	`, albumID)
	if err != nil {
		return fmt.Errorf("update album track count: %w", err)
	}

	return tx.Commit(ctx)
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
