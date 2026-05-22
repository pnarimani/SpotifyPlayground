package cache

import (
	"catalog/internal/albums"
	"catalog/internal/config"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
)

const (
	albumKeyPrefix = "catalog:album:"
	trackKeyPrefix = "catalog:track:"
	defaultTTL     = 10 * time.Minute
)

type CachedRepository struct {
	repo  albums.Repository
	rdb   *redis.Client
	log   *slog.Logger
	group singleflight.Group
	ttl   time.Duration
}

func New(ctx context.Context, repo albums.Repository, logger *slog.Logger, cfg config.Config) (*CachedRepository, error) {
	rdb, err := newRedisClient(ctx, cfg.RedisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create redis client. err: %w", err)
	}

	return &CachedRepository{
		repo: repo,
		rdb:  rdb,
		log:  logger,
		ttl:  defaultTTL,
	}, nil
}

func (c *CachedRepository) GetAlbum(ctx context.Context, id string) (*albums.Album, error) {
	key := albumKeyPrefix + id

	v, err, _ := c.group.Do(key, func() (any, error) {
		data, err := c.rdb.Get(ctx, key).Bytes()
		if err == nil {
			var album albums.Album
			if err := json.Unmarshal(data, &album); err == nil {
				return &album, nil
			}
			c.log.WarnContext(ctx, "cache unmarshal failed, falling through to db", "key", key, "err", err)
		} else if !errors.Is(err, redis.Nil) {
			c.log.WarnContext(ctx, "cache get failed, falling through to db", "key", key, "err", err)
		}

		album, err := c.repo.GetAlbum(ctx, id)
		if err != nil {
			return nil, err
		}

		if data, err := json.Marshal(album); err == nil {
			if err := c.rdb.Set(ctx, key, data, c.ttl).Err(); err != nil {
				c.log.WarnContext(ctx, "cache set failed", "key", key, "err", err)
			}
		}

		return album, nil
	})
	if err != nil {
		return nil, err
	}

	return v.(*albums.Album), nil
}

func (c *CachedRepository) GetTrack(ctx context.Context, id string) (*albums.Track, error) {
	key := trackKeyPrefix + id

	v, err, _ := c.group.Do(key, func() (any, error) {
		data, err := c.rdb.Get(ctx, key).Bytes()
		if err == nil {
			var track albums.Track
			if err := json.Unmarshal(data, &track); err == nil {
				return &track, nil
			}
			c.log.WarnContext(ctx, "cache unmarshal failed, falling through to db", "key", key, "err", err)
		} else if !errors.Is(err, redis.Nil) {
			c.log.WarnContext(ctx, "cache get failed, falling through to db", "key", key, "err", err)
		}

		track, err := c.repo.GetTrack(ctx, id)
		if err != nil {
			return nil, err
		}

		if data, err := json.Marshal(track); err == nil {
			if err := c.rdb.Set(ctx, key, data, c.ttl).Err(); err != nil {
				c.log.WarnContext(ctx, "cache set failed", "key", key, "err", err)
			}
		}

		return track, nil
	})
	if err != nil {
		return nil, err
	}

	return v.(*albums.Track), nil
}

func (c *CachedRepository) CreateAlbum(ctx context.Context, params albums.CreateAlbumParams) (*albums.Album, error) {
	return c.repo.CreateAlbum(ctx, params)
}

func (c *CachedRepository) UpdateAlbum(ctx context.Context, id string, params albums.UpdateAlbumParams) (*albums.Album, error) {
	album, err := c.repo.UpdateAlbum(ctx, id, params)
	if err != nil {
		return nil, err
	}

	c.invalidateAlbum(ctx, id)
	return album, nil
}

func (c *CachedRepository) DeleteAlbum(ctx context.Context, id string) error {
	// Get tracks before deleting so we can invalidate them
	tracks, _ := c.repo.ListAlbumTracks(ctx, id)

	if err := c.repo.DeleteAlbum(ctx, id); err != nil {
		return err
	}

	c.invalidateAlbum(ctx, id)
	for _, t := range tracks {
		c.invalidateTrack(ctx, t.ID)
	}

	return nil
}

func (c *CachedRepository) CreateTrack(ctx context.Context, params albums.CreateTrackParams) (*albums.Track, error) {
	track, err := c.repo.CreateTrack(ctx, params)
	if err != nil {
		return nil, err
	}

	c.invalidateAlbum(ctx, params.AlbumID)
	return track, nil
}

func (c *CachedRepository) UpdateTrack(ctx context.Context, id string, params albums.UpdateTrackParams) (*albums.Track, error) {
	track, err := c.repo.UpdateTrack(ctx, id, params)
	if err != nil {
		return nil, err
	}

	c.invalidateTrack(ctx, id)
	c.invalidateAlbum(ctx, track.AlbumID)
	return track, nil
}

func (c *CachedRepository) DeleteTrack(ctx context.Context, id string) error {
	// Get the track first to know which album to invalidate
	track, err := c.repo.GetTrack(ctx, id)
	if err != nil {
		return c.repo.DeleteTrack(ctx, id)
	}

	if err := c.repo.DeleteTrack(ctx, id); err != nil {
		return err
	}

	c.invalidateTrack(ctx, id)
	c.invalidateAlbum(ctx, track.AlbumID)
	return nil
}

func (c *CachedRepository) ListAlbums(ctx context.Context, cursor *albums.Cursor, limit int) ([]albums.Album, error) {
	return c.repo.ListAlbums(ctx, cursor, limit)
}

func (c *CachedRepository) ListAlbumTracks(ctx context.Context, albumID string) ([]albums.Track, error) {
	return c.repo.ListAlbumTracks(ctx, albumID)
}

func (c *CachedRepository) invalidateAlbum(ctx context.Context, id string) {
	key := albumKeyPrefix + id
	if err := c.rdb.Del(ctx, key).Err(); err != nil {
		c.log.WarnContext(ctx, "cache invalidation failed", "key", key, "err", err)
	}
}

func (c *CachedRepository) invalidateTrack(ctx context.Context, id string) {
	key := trackKeyPrefix + id
	if err := c.rdb.Del(ctx, key).Err(); err != nil {
		c.log.WarnContext(ctx, "cache invalidation failed", "key", key, "err", err)
	}
}

func (c *CachedRepository) Close() error {
	return c.rdb.Close()
}

func newRedisClient(ctx context.Context, redisURL string) (*redis.Client, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}

	opts.PoolSize = 20
	opts.MinIdleConns = 4
	opts.ReadTimeout = 2 * time.Second
	opts.WriteTimeout = 2 * time.Second
	opts.DialTimeout = 3 * time.Second

	rdb := redis.NewClient(opts)

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return rdb, nil
}
