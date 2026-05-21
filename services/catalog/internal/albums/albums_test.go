package albums_test

import (
	"catalog/internal/albums"
	"context"
	"errors"
	"testing"
)

type mockRepo struct {
	album      *albums.Album
	albumsList []albums.Album
	tracks     []albums.Track
	err        error
}

func (m *mockRepo) GetAlbum(ctx context.Context, id string) (*albums.Album, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.album, nil
}

func (m *mockRepo) ListAlbums(ctx context.Context, cursor *albums.Cursor, limit int) ([]albums.Album, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.albumsList, nil
}

func (m *mockRepo) ListAlbumTracks(ctx context.Context, albumID string) ([]albums.Track, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.tracks, nil
}

func (m *mockRepo) CreateAlbum(ctx context.Context, params albums.CreateAlbumParams) (*albums.Album, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.album, nil
}

func (m *mockRepo) UpdateAlbum(ctx context.Context, id string, params albums.UpdateAlbumParams) (*albums.Album, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.album, nil
}

func (m *mockRepo) DeleteAlbum(ctx context.Context, id string) error {
	return m.err
}

func (m *mockRepo) GetTrack(ctx context.Context, id string) (*albums.Track, error) {
	if m.err != nil {
		return nil, m.err
	}
	if len(m.tracks) > 0 {
		return &m.tracks[0], nil
	}
	return nil, albums.ErrNotFound
}

func (m *mockRepo) CreateTrack(ctx context.Context, params albums.CreateTrackParams) (*albums.Track, error) {
	if m.err != nil {
		return nil, m.err
	}
	if len(m.tracks) > 0 {
		return &m.tracks[0], nil
	}
	return &albums.Track{}, nil
}

func (m *mockRepo) UpdateTrack(ctx context.Context, id string, params albums.UpdateTrackParams) (*albums.Track, error) {
	if m.err != nil {
		return nil, m.err
	}
	if len(m.tracks) > 0 {
		return &m.tracks[0], nil
	}
	return &albums.Track{}, nil
}

func (m *mockRepo) DeleteTrack(ctx context.Context, id string) error {
	return m.err
}

func TestGetAlbum_Success(t *testing.T) {
	repo := &mockRepo{
		album: &albums.Album{
			ID:    "00000000-0000-0000-0000-000000000001",
			Title: "Test Album",
		},
		tracks: []albums.Track{
			{ID: "t1", Name: "Track 1", TrackNumber: 1, DiscNumber: 1, DurationMs: 200000},
			{ID: "t2", Name: "Track 2", TrackNumber: 2, DiscNumber: 1, DurationMs: 180000},
		},
	}

	svc := albums.New(repo)
	album, err := svc.GetAlbum(context.Background(), "00000000-0000-0000-0000-000000000001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if album.Title != "Test Album" {
		t.Errorf("expected title 'Test Album', got %q", album.Title)
	}
	if len(album.Tracks) != 2 {
		t.Errorf("expected 2 tracks, got %d", len(album.Tracks))
	}
}

func TestGetAlbum_NotFound(t *testing.T) {
	repo := &mockRepo{
		err: albums.ErrNotFound,
	}

	svc := albums.New(repo)
	_, err := svc.GetAlbum(context.Background(), "00000000-0000-0000-0000-000000000001")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, albums.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestListAlbums_Success(t *testing.T) {
	repo := &mockRepo{
		albumsList: []albums.Album{
			{ID: "a1", Title: "Album 1"},
			{ID: "a2", Title: "Album 2"},
		},
	}

	svc := albums.New(repo)
	result, err := svc.ListAlbums(context.Background(), nil, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 albums, got %d", len(result))
	}
}
