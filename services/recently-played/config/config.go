package config

import (
	"log/slog"
	"os"
	"strings"
)

// TODO: read from env or config

func GetRecentTracksCount() int {
	return 50
}

func GetDatabaseReadCount() int {
	return GetRecentTracksCount() * 4
}

func GetCassandraHosts() []string {
	return strings.Split(os.Getenv("CASSANDRA_HOSTS"), ",")
}

func GetKafkaBrokers() []string {
	return strings.Split(os.Getenv("KAFKA_BROKERS"), ",")
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
