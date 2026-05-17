package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"recently_played/db"

	"github.com/segmentio/kafka-go"
)

type messageData struct {
	UserId   string `json:"user_id"`
	TrackId  string `json:"track_id"`
	PlayedAt int64  `json:"played_at"`
}

func ConsumeEvents(ctx context.Context) error {
	reader := kafka.NewReader(kafka.ReaderConfig{
		GroupID:     "recently-played",
		Topic:       "play-events",
	})

	for {
		msg, err := reader.ReadMessage(ctx)
		if err != nil {
			return fmt.Errorf("failed to read kafka message, err: %w", err)
		}

		data := messageData{}
		if err := json.Unmarshal(msg.Value, &data); err != nil {
			log.Fatal(err) 
			continue
		}

		if err := db.Write(ctx, db.Entry{
			UserID: data.UserId,
			PlayedAt: data.PlayedAt,
			TrackId: data.TrackId,
		}); err != nil {
			return fmt.Errorf("failed to write to cassandra, err: %w", err)
		}
	}
}
