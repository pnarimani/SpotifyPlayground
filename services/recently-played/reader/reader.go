package reader

import "context"

type Entry struct {
	UserID   string
	PlayedAt int64
	TrackId  string
}

func GetLastPlayedTracks(ctx context.Context, userId string) ([]Entry, error) {

}
