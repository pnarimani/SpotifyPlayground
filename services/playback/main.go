package main

import (
	"context"
	pb "contracts/events/playback/v1"
	"encoding/json"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/segmentio/kafka-go"
)

const sampleUserId = "default_user_id"

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	slog.SetDefault(logger)

	slog.Info("service starting")

	balancer := kafka.Hash{}

	brokers := strings.Split(os.Getenv("KAFKA_BROKERS"), ",")

	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  brokers,
		Topic:    pb.PlayEventTopic,
		Balancer: &balancer,
	})
	defer writer.Close()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	for {
		select {
		case <-ticker.C:
			now := time.Now()
			msg := pb.EventMessage{
				UserID: sampleUserId,
				ClientMessage: pb.ClientMessage{
					TrackId:         "sample_track_id",
					ClientTimestamp: now.UnixMilli(),
				},
			}

			slog.Debug("sending", "msg", msg)

			msgBytes, err := json.Marshal(msg)
			if err != nil {
				log.Fatal(err)
				return
			}

			slog.Debug("writing")
			if err := writer.WriteMessages(ctx, kafka.Message{
				Key:   []byte(sampleUserId),
				Value: msgBytes,
				Time:  now,
			}); err != nil {
				slog.Error("failed to write kafka message", "err", err)
			}
		case <-ctx.Done():
			return
		}
	}

}
