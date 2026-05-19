package v1

const PlayEventTopic = "play-events"

type ClientMessage struct {
	RequestSequenceID int32         `json:"request_sequence_id"`
	TrackId           string        `json:"track_id"`
	PlaybackSessionID string        `json:"playback_session_id"`
	EventType         PlayEventType `json:"event_type"`
	PositionMs        int32         `json:"position_ms"`
	ClientTimestamp   int64         `json:"client_timestamp"`
}

type EventMessage struct {
	ClientMessage
	UserID           string `json:"user_id"`
	EventID          string `json:"event_id"`
	ServerReceivedAt int64  `json:"server_received_at"`
	Country          string `json:"country"`
}

type PlayEventType string

const (
	Start    PlayEventType = "start"
	Progress PlayEventType = "progress"
	Pause    PlayEventType = "pause"
	Resume   PlayEventType = "resume"
	Stop     PlayEventType = "stop"
)
