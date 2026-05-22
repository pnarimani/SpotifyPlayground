package config

import (
	"errors"
	"log/slog"
	"os"
	"strings"
)

type Config struct {
	LogLevel    slog.Level
	PostgresDSN string
	RedisURL    string
}

func ReadFromEnv() (Config, error) {
	dsn := os.Getenv("POSTGRES_DSN")
	if dsn == "" {
		return Config{}, errors.New("POSTGRES_DSN environment variable is required")
	}

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		return Config{}, errors.New("REDIS_URL environment variable is required")
	}

	return Config{
		PostgresDSN: dsn,
		RedisURL:    redisURL,
		LogLevel:    getLogLevel(),
	}, nil
}

func getLogLevel() slog.Level {
	level := strings.ToLower(os.Getenv("LOG_LEVEL"))
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
