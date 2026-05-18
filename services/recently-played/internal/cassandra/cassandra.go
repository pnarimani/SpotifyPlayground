package cassandra

import (
	"context"
	"errors"
	"fmt"
	"recently_played/internal/config"
	"time"

	"github.com/gocql/gocql"
)

type Entry struct {
	UserId   string
	PlayedAt int64
	TrackId  string
}

type service struct {
	session *gocql.Session
}

type Service interface {
	Read(ctx context.Context, userId string, readCount int) ([]Entry, error)
	Write(ctx context.Context, entry Entry) error
	Close()
}

func New(config config.Config) (Service, error) {
	cluster := gocql.NewCluster(config.CassandraHosts...)
	cluster.Keyspace = "recently_played"
	cluster.Consistency = gocql.LocalQuorum

	s, err := cluster.CreateSession()
	if err != nil {
		return nil, err
	}

	return &service{session: s}, nil
}

func (s *service) Close() {
	if s.session != nil {
		s.session.Close()
	}
}

func (s *service) Write(ctx context.Context, entry Entry) error {
	if s.session == nil {
		return errors.New("session is nil")
	}

	playedAt := time.UnixMilli(entry.PlayedAt)
	bucketDate := getBucketDate(playedAt)
	playedAtUUID := gocql.UUIDFromTime(playedAt)

	query := `INSERT INTO plays_by_user (user_id, bucket_date, played_at, track_id, played_at_ts)
				VALUES (?, ?, ?, ?, ?)`

	return s.session.Query(query, entry.UserId, bucketDate, playedAtUUID, entry.TrackId, playedAt).
		WithContext(ctx).
		Exec()
}

func (s *service) Read(ctx context.Context, userId string, limit int) ([]Entry, error) {
	if s.session == nil {
		return nil, errors.New("call Init before attempting to Read")
	}

	query := `SELECT track_id, played_at_ts FROM plays_by_user
				WHERE user_id = ? AND bucket_date = ? LIMIT ?`

	var results []Entry
	now := time.Now()

	weekLimit := 14

	for i := 0; i < weekLimit && len(results) < limit; i++ {
		t := now.AddDate(0, 0, -7*i)
		bucketDate := getBucketDate(t)
		remaining := limit - len(results)

		iter := s.session.Query(query, userId, bucketDate, remaining).WithContext(ctx).Iter()

		var trackId string
		var playedAt time.Time
		for iter.Scan(&trackId, &playedAt) {
			results = append(results, Entry{
				UserId:   userId,
				PlayedAt: playedAt.UnixMilli(),
				TrackId:  trackId,
			})
		}

		if err := iter.Close(); err != nil {
			return results, err
		}
	}

	return results, nil
}

func getBucketDate(playedAt time.Time) string {
	year, week := playedAt.ISOWeek()
	bucketDate := fmt.Sprintf("%d-W%02d", year, week)
	return bucketDate
}
