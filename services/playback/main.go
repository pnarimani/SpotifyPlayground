package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/segmentio/kafka-go"
)

const sampleUserId = "default_user_id"

type MessageData struct {
	UserId   string `json:"user_id"`
	TrackId  string `json:"track_id"`
	PlayedAt int64  `json:"played_at"`
}

func main() {
	balancer := kafka.Hash{}
	const playTopic = "play-events"

	brokers := strings.Split(os.Getenv("KAFKA_BROKERS"), ",")

	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  brokers,
		Topic:    playTopic,
		Balancer: &balancer,
	})

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	for {
		select {
		case <-ticker.C:
			now := time.Now()
			msg := MessageData{
				UserId:   sampleUserId,
				TrackId:  "sample_track_id",
				PlayedAt: now.Unix(),
			}

			msgBytes, err := json.Marshal(msg)
			if err != nil {
				log.Fatal(err)
				return
			}

			writer.WriteMessages(ctx, kafka.Message{
				Topic:      playTopic,
				Key:        []byte(sampleUserId),
				Value:      msgBytes,
				WriterData: writer,
				Time:       now,
			})
		case <-ctx.Done():
			return
		}
	}

}
