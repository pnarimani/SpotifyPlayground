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

type Cursor struct {
	CreatedAt time.Time
	ID        string
}

type Repository interface {
	GetAlbum(ctx context.Context, id string) (*Album, error)
	ListAlbums(ctx context.Context, cursor *Cursor, limit int) ([]Album, error)
	ListAlbumTracks(ctx context.Context, albumID string) ([]Track, error)
	GetTrack(ctx context.Context, id string) (*Track, error)
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

func (s *Service) GetTrack(ctx context.Context, id string) (*Track, error) {
	track, err := s.repo.GetTrack(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get track: %w", err)
	}
	return track, nil
}
