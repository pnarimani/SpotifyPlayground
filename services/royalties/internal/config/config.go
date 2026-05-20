package config

import (
	"log/slog"
	"os"
	"strings"
)

type Config struct {
	LogLevel     slog.Level
	KafkaBrokers []string
	PostgresDSN  string
}

func ReadFromEnv() Config {
	return Config{
		PostgresDSN:  getPostgresDSN(),
		KafkaBrokers: getKafkaBrokers(),
		LogLevel:     GetLogLevel(),
	}
}

func getKafkaBrokers() []string {
	return strings.Split(os.Getenv("KAFKA_BROKERS"), ",")
}

func getPostgresDSN() string {
	return os.Getenv("POSTGRES_DSN")
}

func GetLogLevel() slog.Level {
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
