package reader

import (
	"context"
	"recently_played/config"
	"recently_played/db"
)

type Entry struct {
	UserId   string
	PlayedAt int64
	TrackId  string
}

func GetLastPlayedTracks(ctx context.Context, userId string) ([]Entry, error) {
	count := config.GetRecentTracksCount()

	dbResults, err := db.Read(ctx, userId, config.GetDatabaseReadCount())
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
