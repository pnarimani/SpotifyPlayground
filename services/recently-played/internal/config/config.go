package config

import (
	"log/slog"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	RecentTracksCount int
	DatabaseReadCount int
	CassandraHosts    []string
	KafkaBrokers      []string
	LogLevel          slog.Level
}

func ReadFromEnv() Config {
	return Config{
		RecentTracksCount: readIntWithDefault("RECENT_TRACKS_COUNT", 50),
		DatabaseReadCount: readIntWithDefault("DATABASE_READ_COUNT", 200),
		CassandraHosts:    getCassandraHosts(),
		KafkaBrokers:      getKafkaBrokers(),
		LogLevel:          GetLogLevel(),
	}
}

func getCassandraHosts() []string {
	return strings.Split(os.Getenv("CASSANDRA_HOSTS"), ",")
}

func getKafkaBrokers() []string {
	return strings.Split(os.Getenv("KAFKA_BROKERS"), ",")
}

func readIntWithDefault(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	num, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}

	return num
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
