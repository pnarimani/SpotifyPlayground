package main

import (
	"github.com/segmentio/kafka-go"
)

type messageData struct {
	UserId   string `json:"user_id"`
	TrackId  string `json:"track_id"`
	PlayedAt int64  `json:"played_at"`
}

func main() {
	kafka.NewReader(kafka.ReaderConfig{
		Brokers:                []string{},
		GroupID:                "",
		GroupTopics:            []string{},
		Topic:                  "",
		Partition:              0,
		Dialer:                 &kafka.Dialer{},
		QueueCapacity:          0,
		MinBytes:               0,
		MaxBytes:               0,
		MaxWait:                0,
		ReadBatchTimeout:       0,
		ReadLagInterval:        0,
		GroupBalancers:         []kafka.GroupBalancer{},
		HeartbeatInterval:      0,
		CommitInterval:         0,
		PartitionWatchInterval: 0,
		WatchPartitionChanges:  false,
		SessionTimeout:         0,
		RebalanceTimeout:       0,
		JoinGroupBackoff:       0,
		RetentionTime:          0,
		StartOffset:            0,
		ReadBackoffMin:         0,
		ReadBackoffMax:         0,
		Logger:                 nil,
		ErrorLogger:            nil,
		IsolationLevel:         0,
		MaxAttempts:            0,
		OffsetOutOfRangeError:  false,
	})
}
