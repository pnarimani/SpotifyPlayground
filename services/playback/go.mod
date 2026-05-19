module playback

go 1.26.3

require github.com/segmentio/kafka-go v0.4.51

require (
	contracts v0.0.0
	github.com/klauspost/compress v1.15.9 // indirect
	github.com/pierrec/lz4/v4 v4.1.15 // indirect
)

replace contracts => ../contracts
