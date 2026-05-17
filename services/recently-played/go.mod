module recently_played

go 1.26

require (
	github.com/gocql/gocql v1.7.0
	github.com/segmentio/kafka-go v0.4.51
	golang.org/x/sync v0.20.0
	middleware v0.0.0
)

replace middleware => ../middleware

require (
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/golang/snappy v0.0.3 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/hailocab/go-hostpool v0.0.0-20160125115350-e80d13ce29ed // indirect
	github.com/klauspost/compress v1.15.9 // indirect
	github.com/pierrec/lz4/v4 v4.1.15 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
)
