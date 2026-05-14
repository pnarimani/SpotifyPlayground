package main

import (
	"context"
	"time"

	"github.com/segmentio/kafka-go"
)

const sampleUserId = "default_user_id"

func main() {
	balancer := kafka.Hash{}
	const playTopic = "play-events"

	writer := kafka.NewWriter(kafka.WriterConfig{
		Topic:    playTopic,
		Balancer: &balancer,
	})

	writer.WriteMessages(context.Background(), kafka.Message{
		Topic:         playTopic,
		Partition:     0,
		Offset:        0,
		HighWaterMark: 0,
		Key:           []byte{},
		Value:         []byte{},
		Headers:       []kafka.Header{},
		WriterData:    writer,
		Time:          time.Time{},
	})
}
