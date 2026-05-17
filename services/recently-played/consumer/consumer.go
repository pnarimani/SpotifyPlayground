package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"recently_played/config"
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
		GroupID: "recently-played",
		Topic:   "play-events",
		Brokers: config.GetKafkaBrokers(),
	})
	defer reader.Close()

	for {
		slog.DebugContext(ctx, "waiting for message")

		msg, err := reader.ReadMessage(ctx)
		if err != nil {
			if err == io.EOF || err == context.Canceled || err == context.DeadlineExceeded {
				return nil
			}
			return fmt.Errorf("failed to read kafka message, err: %w", err)
		}

		data := messageData{}
		if err := json.Unmarshal(msg.Value, &data); err != nil {
			slog.ErrorContext(ctx, "failed to parse json", "json", string(msg.Value), "key", string(msg.Key))
			continue
		}

		slog.DebugContext(ctx, "kafka message received", "data", data)

		if err := db.Write(ctx, db.Entry{
			UserId:   data.UserId,
			PlayedAt: data.PlayedAt,
			TrackId:  data.TrackId,
		}); err != nil {
			return fmt.Errorf("failed to write to cassandra, err: %w", err)
		}
	}
}
