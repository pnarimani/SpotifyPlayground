package cache

import (
	"catalog/internal/albums"
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

// mockRepo is a test double that tracks call counts
type mockRepo struct {
	getAlbumCalls  atomic.Int64
	getTrackCalls  atomic.Int64
	album          *albums.Album
	track          *albums.Track
}

func (m *mockRepo) GetAlbum(_ context.Context, id string) (*albums.Album, error) {
	m.getAlbumCalls.Add(1)
	if m.album == nil {
		return nil, albums.ErrNotFound
	}
	return m.album, nil
}

func (m *mockRepo) GetTrack(_ context.Context, id string) (*albums.Track, error) {
	m.getTrackCalls.Add(1)
	if m.track == nil {
		return nil, albums.ErrNotFound
	}
	return m.track, nil
}

func (m *mockRepo) ListAlbums(_ context.Context, _ *albums.Cursor, _ int) ([]albums.Album, error) {
	return nil, nil
}
func (m *mockRepo) ListAlbumTracks(_ context.Context, _ string) ([]albums.Track, error) {
	return nil, nil
}

func setupTestCache(t *testing.T, repo *mockRepo) *CachedRepository {
	t.Helper()

	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379", DB: 15})
	ctx := context.Background()

	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Skipf("redis not available: %v", err)
	}

	// Flush test DB
	rdb.FlushDB(ctx)

	t.Cleanup(func() {
		rdb.FlushDB(ctx)
		rdb.Close()
	})

	return &CachedRepository{
		repo: repo,
		rdb:  rdb,
		ttl:  defaultTTL,
	}
}

func TestGetAlbum_CacheHit(t *testing.T) {
	repo := &mockRepo{
		album: &albums.Album{
			ID:          "test-album-id",
			Title:       "Test Album",
			TotalTracks: 5,
		},
	}

	c := setupTestCache(t, repo)
	ctx := context.Background()

	// First call — cache miss, hits DB
	album, err := c.GetAlbum(ctx, "test-album-id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if album.Title != "Test Album" {
		t.Fatalf("expected title 'Test Album', got %q", album.Title)
	}
	if repo.getAlbumCalls.Load() != 1 {
		t.Fatalf("expected 1 DB call, got %d", repo.getAlbumCalls.Load())
	}

	// Second call — cache hit, no DB call
	album, err = c.GetAlbum(ctx, "test-album-id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if album.Title != "Test Album" {
		t.Fatalf("expected title 'Test Album', got %q", album.Title)
	}
	if repo.getAlbumCalls.Load() != 1 {
		t.Fatalf("expected 1 DB call (cache hit), got %d", repo.getAlbumCalls.Load())
	}
}

func TestGetAlbum_Singleflight(t *testing.T) {
	repo := &mockRepo{
		album: &albums.Album{
			ID:          "test-album-id",
			Title:       "Test Album",
			TotalTracks: 5,
		},
	}

	c := setupTestCache(t, repo)
	ctx := context.Background()

	// Launch many concurrent requests for the same key
	const numGoroutines = 50
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	errs := make([]error, numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			_, errs[idx] = c.GetAlbum(ctx, "test-album-id")
		}(i)
	}

	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Fatalf("goroutine %d got error: %v", i, err)
		}
	}

	// Singleflight should collapse all concurrent requests into 1 DB call
	calls := repo.getAlbumCalls.Load()
	if calls != 1 {
		t.Fatalf("expected 1 DB call due to singleflight, got %d", calls)
	}
}

func TestGetTrack_CacheHit(t *testing.T) {
	repo := &mockRepo{
		track: &albums.Track{
			ID:          "test-track-id",
			AlbumID:     "test-album-id",
			Name:        "Test Track",
			TrackNumber: 1,
			DiscNumber:  1,
			DurationMs:  240000,
		},
	}

	c := setupTestCache(t, repo)
	ctx := context.Background()

	// First call
	track, err := c.GetTrack(ctx, "test-track-id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if track.Name != "Test Track" {
		t.Fatalf("expected name 'Test Track', got %q", track.Name)
	}

	// Second call — cache hit
	_, err = c.GetTrack(ctx, "test-track-id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.getTrackCalls.Load() != 1 {
		t.Fatalf("expected 1 DB call (cache hit), got %d", repo.getTrackCalls.Load())
	}
}

// Suppress unused import warning
var _ = time.Second
