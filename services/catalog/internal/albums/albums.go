package albums

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var ErrNotFound = errors.New("not found")

type Album struct {
	ID          string
	Title       string
	ReleaseDate *time.Time
	CoverURL    *string
	Label       *string
	TotalTracks int
	Tracks      []Track
}

type Track struct {
	ID          string
	AlbumID     string
	Name        string
	TrackNumber int
	DiscNumber  int
	DurationMs  int
	Explicit    bool
}

type CreateAlbumParams struct {
	Title       string
	ReleaseDate *time.Time
	CoverURL    *string
	Label       *string
}

type UpdateAlbumParams struct {
	Title       *string
	ReleaseDate *time.Time
	CoverURL    *string
	Label       *string
}

type CreateTrackParams struct {
	AlbumID     string
	Name        string
	TrackNumber int
	DiscNumber  int
	DurationMs  int
	Explicit    bool
}

type UpdateTrackParams struct {
	Name        *string
	TrackNumber *int
	DiscNumber  *int
	DurationMs  *int
	Explicit    *bool
}

type Cursor struct {
	CreatedAt time.Time
	ID        string
}

type Repository interface {
	GetAlbum(ctx context.Context, id string) (*Album, error)
	ListAlbums(ctx context.Context, cursor *Cursor, limit int) ([]Album, error)
	ListAlbumTracks(ctx context.Context, albumID string) ([]Track, error)
	CreateAlbum(ctx context.Context, params CreateAlbumParams) (*Album, error)
	UpdateAlbum(ctx context.Context, id string, params UpdateAlbumParams) (*Album, error)
	DeleteAlbum(ctx context.Context, id string) error
	GetTrack(ctx context.Context, id string) (*Track, error)
	CreateTrack(ctx context.Context, params CreateTrackParams) (*Track, error)
	UpdateTrack(ctx context.Context, id string, params UpdateTrackParams) (*Track, error)
	DeleteTrack(ctx context.Context, id string) error
}

type Service struct {
	repo Repository
}

func New(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetAlbum(ctx context.Context, id string) (*Album, error) {
	album, err := s.repo.GetAlbum(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get album: %w", err)
	}

	tracks, err := s.repo.ListAlbumTracks(ctx, album.ID)
	if err != nil {
		return nil, fmt.Errorf("get album tracks: %w", err)
	}

	album.Tracks = tracks
	return album, nil
}

func (s *Service) ListAlbums(ctx context.Context, cursor *Cursor, limit int) ([]Album, error) {
	albums, err := s.repo.ListAlbums(ctx, cursor, limit)
	if err != nil {
		return nil, fmt.Errorf("list albums: %w", err)
	}
	return albums, nil
}

func (s *Service) CreateAlbum(ctx context.Context, params CreateAlbumParams) (*Album, error) {
	if params.Title == "" {
		return nil, errors.New("title is required")
	}

	album, err := s.repo.CreateAlbum(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("create album: %w", err)
	}
	return album, nil
}

func (s *Service) UpdateAlbum(ctx context.Context, id string, params UpdateAlbumParams) (*Album, error) {
	if id == "" {
		return nil, errors.New("album id is required")
	}

	album, err := s.repo.UpdateAlbum(ctx, id, params)
	if err != nil {
		return nil, fmt.Errorf("update album: %w", err)
	}
	return album, nil
}

func (s *Service) DeleteAlbum(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("album id is required")
	}

	if err := s.repo.DeleteAlbum(ctx, id); err != nil {
		return fmt.Errorf("delete album: %w", err)
	}
	return nil
}

func (s *Service) GetTrack(ctx context.Context, id string) (*Track, error) {
	track, err := s.repo.GetTrack(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get track: %w", err)
	}
	return track, nil
}

func (s *Service) CreateTrack(ctx context.Context, params CreateTrackParams) (*Track, error) {
	if params.AlbumID == "" {
		return nil, errors.New("album id is required")
	}
	if params.Name == "" {
		return nil, errors.New("track name is required")
	}
	if params.DurationMs <= 0 {
		return nil, errors.New("duration must be positive")
	}

	track, err := s.repo.CreateTrack(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("create track: %w", err)
	}
	return track, nil
}

func (s *Service) UpdateTrack(ctx context.Context, id string, params UpdateTrackParams) (*Track, error) {
	if id == "" {
		return nil, errors.New("track id is required")
	}

	track, err := s.repo.UpdateTrack(ctx, id, params)
	if err != nil {
		return nil, fmt.Errorf("update track: %w", err)
	}
	return track, nil
}

func (s *Service) DeleteTrack(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("track id is required")
	}

	if err := s.repo.DeleteTrack(ctx, id); err != nil {
		return fmt.Errorf("delete track: %w", err)
	}
	return nil
}
