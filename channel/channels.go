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
func (mcr *MCR) GetChannel(ctx context.Context, channelID string) (*Channel, error) {
	ch, ok := mcr.channels[channelID]
	if !ok {
		return nil, errors.New("channel doesn't exist")
	}
	return ch, nil
}

// NewChannel creates a new channel to playout
func (mcr *MCR) NewChannel(ctx context.Context, newCh NewChannelStruct) (*Channel, error) {
	channel := &Channel{
		Name:        newCh.Name,
		Description: newCh.Description,
		ChannelType: newCh.ChannelType,
		IngestURL:   newCh.IngestURL,
		IngestType:  newCh.IngestType,
		SlateURL:    newCh.SlateURL,
		Outputs:     newCh.Outputs,
		Archive:     newCh.Archive,
		DVR:         newCh.DVR,
	}
	channel.Status = "pending"

	for {
		channel.ID = randString()
		_, exists := mcr.channels[channel.ID]
		if !exists {
			break
		}
	}

	mcr.channels[channel.ID] = channel

	return channel, nil
}

// DeleteChannel removes a channel from playout
func (mcr *MCR) DeleteChannel(channelID string) error {
	if _, ok := mcr.channels[channelID]; ok {
		err := mcr.channels[channelID].Stop()
		if err != nil {
			return fmt.Errorf("failed to delete channel: %w", err)
		}
		delete(mcr.channels, channelID)
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
