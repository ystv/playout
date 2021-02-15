package channel

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
)

type (
	// MCR manages a group of channels
	MCR struct {
		conf     Config
		channels map[string]*Channel
	}
	// Config to specify available endpoints
	Config struct {
		VTEndpoint string
		Endpoints  []Endpoint
	}
	// Endpoint a usable output by playout
	Endpoint struct {
		Type string
		URL  string
	}
)

// NewMCR creates a new "Master Control Room"
// effictively manages a group of channels
func NewMCR() *MCR {
	mcr := &MCR{
		conf: Config{
			VTEndpoint: "http://localhost:7071",
			Endpoints: []Endpoint{
				{
					Type: "rtmp",
					URL:  "rtmp://stream.ystv.co.uk/internal/",
				},
				{
					Type: "hls",
					URL:  "https://video-cdn.ystv.co.uk/",
				},
			},
		},
		channels: make(map[string]*Channel),
	}
	return mcr
}

// GetChannel retrieves a channel from playout
func (mcr *MCR) GetChannel(ctx context.Context, shortName string) (*Channel, error) {
	ch, ok := mcr.channels[shortName]
	if !ok {
		return nil, errors.New("channel doesn't exist")
	}
	return ch, nil
}

// NewChannel creates a new channel to playout
func (mcr *MCR) NewChannel(ctx context.Context, newCh NewChannelStruct) (*Channel, error) {
	channel := &Channel{
		ShortName:   newCh.ShortName,
		Name:        newCh.Name,
		Description: newCh.Description,
		ChannelType: newCh.ChannelType,
		IngestURL:   newCh.IngestURL,
		IngestType:  newCh.IngestType,
		SlateURL:    newCh.SlateURL,
		Outputs:     newCh.Outputs,
		Archive:     newCh.Archive,
	}
	channel.Status = "pending"

	// Default values
	if channel.Name == "" {
		channel.Name = "A random livestream"
	}

	for {
		// Generate a random short-name if one wasn't provided
		if channel.ShortName == "" {
			channel.ShortName = randString()
		}
		_, exists := mcr.channels[channel.ShortName]
		if !exists {
			break
		}
	}

	mcr.channels[channel.ShortName] = channel
	// TODO: Reflect creation in DB

	return channel, nil
}

// DeleteChannel removes a channel from playout
func (mcr *MCR) DeleteChannel(shortName string) error {
	if _, ok := mcr.channels[shortName]; ok {
		err := mcr.channels[shortName].Stop()
		if err != nil {
			return fmt.Errorf("failed to delete channel: %w", err)
		}
		delete(mcr.channels, shortName)
	}
	return nil
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randString() string {
	b := make([]rune, 10)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
