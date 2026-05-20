package main

import (
	"context"
	pb "contracts/events/playback/v1"
	"encoding/json"
	"log"
	"log/slog"
	"math/rand/v2"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	var sampleTrackIds []string
	for i := 0; i < 20; i += 1 {
		sampleTrackIds = append(sampleTrackIds, uuid.NewString())
	}

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
			userId := uuid.NewString()
			msg := pb.EventMessage{
				UserID:  userId,
				EventID: uuid.NewString(),
				ClientMessage: pb.ClientMessage{
					TrackId:           sampleTrackIds[rand.N(len(sampleTrackIds))],
					ClientTimestamp:   now.UnixMilli() - 50,
					RequestSequenceID: 0,
					PlaybackSessionID: uuid.NewString(),
					EventType:         pb.Start,
					PositionMs:        0,
				},
				ServerReceivedAt: now.UnixMilli(),
				Country:          "se",
			}

			slog.Debug("sending", "msg", msg)

			msgBytes, err := json.Marshal(msg)
			if err != nil {
				log.Fatal(err)
				return
			}

			slog.Debug("writing")
			if err := writer.WriteMessages(ctx, kafka.Message{
				Key:   []byte(userId),
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
