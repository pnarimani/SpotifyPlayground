package config

import (
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
