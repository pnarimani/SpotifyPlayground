package reader

import (
	"context"
	"recently_played/internal/cassandra"
	"recently_played/internal/config"
)

type Entry struct {
	UserId   string `json:"user_id"`
	PlayedAt int64  `json:"played_at"`
	TrackId  string `json:"track_id"`
}

type Service struct {
	config config.Config
	db     cassandra.Service
}

func New(config config.Config, db cassandra.Service) Service {
	return Service{config, db}
}

func (s *Service) GetLastPlayedTracks(ctx context.Context, userId string) ([]Entry, error) {
	count := s.config.RecentTracksCount

	dbResults, err := s.db.Read(ctx, userId, s.config.DatabaseReadCount)
	if err != nil {
		return nil, err
	}

	seenTracks := make(map[string]bool, count)

	var entries []Entry
	for _, row := range dbResults {
		if _, ok := seenTracks[row.TrackId]; ok {
			continue
		}

		entries = append(entries, Entry{
			UserId:   userId,
			PlayedAt: row.PlayedAt,
			TrackId:  row.TrackId,
		})

		seenTracks[row.TrackId] = true

		if len(entries) >= count {
			break
		}
	}

	return entries, nil
}
