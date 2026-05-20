package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"royalties/internal/config"
	"royalties/internal/postgres"

	playback "contracts/events/playback/v1"

	"github.com/segmentio/kafka-go"
)

type service struct {
	logger *slog.Logger
	pg     postgres.Store
	config config.Config
}

type Consumer interface {
	ConsumeEvents(ctx context.Context) error
}

func New(cfg config.Config, logger *slog.Logger, pg postgres.Store) Consumer {
	return &service{logger, pg, cfg}
}

func (s *service) ConsumeEvents(ctx context.Context) error {
	reader := kafka.NewReader(kafka.ReaderConfig{
		GroupID: "recently-played",
		Topic:   playback.PlayEventTopic,
		Brokers: s.config.KafkaBrokers,
	})
	defer reader.Close()

	for {
		s.logger.DebugContext(ctx, "waiting for message")

		msg, err := reader.FetchMessage(ctx)
		if err != nil {
			if err == io.EOF || err == context.Canceled || err == context.DeadlineExceeded {
				return nil
			}
			return fmt.Errorf("failed to read kafka message, err: %w", err)
		}

		data := playback.EventMessage{}
		if err := json.Unmarshal(msg.Value, &data); err != nil {
			slog.ErrorContext(ctx, "failed to parse json", "json", string(msg.Value), "key", string(msg.Key))
			continue
		}

		s.logger.DebugContext(ctx, "kafka message received", "data", data)

		if err := s.pg.WriteRawPlayEvent(ctx, postgres.RawEvent(data)); err != nil {
			return fmt.Errorf("failed to write raw events to postgres, err: %w", err)
		}

		if err := reader.CommitMessages(ctx, msg); err != nil {
			return fmt.Errorf("failed to commit kafka message, err: %w", err)
		}
	}
}
