package player

import (
	"context"
)

// Player will create a stream to playout a programme
type Player interface {
	Play(ctx context.Context, c Config) error
}

// Config is the required video information for processing
type Config struct {
	DstURL    string
	Width     int
	Height    int
	Bitrate   int
	VideoURLs []string
}
