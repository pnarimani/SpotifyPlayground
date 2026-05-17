package reader

import (
	"context"
	"recently_played/config"
	"recently_played/db"
)

type Entry struct {
	UserId   string `json:"user_id"`
	PlayedAt int64  `json:"played_at"`
	TrackId  string `json:"track_id"`
}

type Service struct {
	config config.Config
	db     db.Service
}

func New(config config.Config, db db.Service) Service {
	return Service{config: config, db: db}
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
