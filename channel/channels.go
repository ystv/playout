package channel

import (
	"context"
	"errors"
	"fmt"
	"math/rand"

	"github.com/jmoiron/sqlx"
	"github.com/ystv/playout/piper"
	"github.com/ystv/playout/scheduler"
)

type (
	// MCR manages a group of channels
	MCR struct {
		db       *sqlx.DB
		conf     *Config
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
		conf: &Config{
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
	ch := &Channel{
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
	ch.Status = "pending"

	// Default values
	if ch.Name == "" {
		ch.Name = "A random livestream"
	}

	for {
		// Generate a random short-name if one wasn't provided
		if ch.ShortName == "" {
			ch.ShortName = randString()
		}
		_, exists := mcr.channels[ch.ShortName]
		if !exists {
			break
		}
	}

	mcr.channels[ch.ShortName] = ch

	channelID := 0
	err := mcr.db.GetContext(ctx, &channelID, `
		INSERT INTO playout.channel(
			short_name,
			name,
			description,
			type,
			ingest_url,
			ingest_type,
			slate_url,
			visibility,
			archive,
			dvr)
		RETURNING channel_id;`,
		ch.ShortName, ch.Name, ch.Description, ch.ChannelType,
		ch.IngestURL, ch.IngestURL, ch.SlateURL, newCh.Visible,
		ch.Archive, newCh.DVR)
	if err != nil {
		return nil, fmt.Errorf("failed to insert channel to DB: %w", err)
	}

	if newCh.HasScheduler {
		sch, err := scheduler.New(mcr.db, channelID)
		if err != nil {
			return nil, fmt.Errorf("failed to start scheduler: %w", err)
		}
		ch.sch = sch
	}

	if newCh.HasPiper {
		piper, err := piper.New(ctx, piper.Config{
			Endpoint: "",
			Width:    1920,
			Height:   1080,
			FPS:      50,
		}, "brave")
		if err != nil {
			return nil, fmt.Errorf("failed to start piper: %w", err)
		}
		ch.piper = piper
	}

	return ch, nil
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
