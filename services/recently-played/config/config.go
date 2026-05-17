package config

// TODO: read from env or config

func GetRecentTracksCount() int {
	return 50
}

func GetDatabaseReadCount() int {
	return GetRecentTracksCount() * 4
}
