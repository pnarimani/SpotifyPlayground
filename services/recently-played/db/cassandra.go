package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gocql/gocql"
)

type Entry struct {
	UserId   string
	PlayedAt int64
	TrackId  string
}

var session *gocql.Session

func Init(hosts []string) error {
	cluster := gocql.NewCluster(hosts...)
	cluster.Keyspace = "recently_played"
	cluster.Consistency = gocql.LocalQuorum

	s, err := cluster.CreateSession()
	if err != nil {
		return err
	}

	session = s
	return nil
}

func Close() {
	if session != nil {
		session.Close()
	}
}

func Write(ctx context.Context, entry Entry) error {
	if session == nil {
		return errors.New("call Init before attempting to Write")
	}

	playedAt := time.UnixMilli(entry.PlayedAt)

	bucketDate := getBucketDate(playedAt)

	playedAtUUID := gocql.UUIDFromTime(playedAt)

	query := `INSERT INTO plays_by_user (user_id, bucket_date, played_at, track_id, played_at_ts) 
				VALUES (?, ?, ?, ?, ?)`

	return session.Query(query, entry.UserId, bucketDate, playedAtUUID, entry.TrackId, playedAt).
		WithContext(ctx).
		Exec()
}

func Read(ctx context.Context, userId string, limit int) ([]Entry, error) {
	if session == nil {
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

		
		iter := session.Query(query, userId, bucketDate, remaining).WithContext(ctx).Iter()

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
