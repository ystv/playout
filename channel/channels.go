package channel

import (
	"context"
	"errors"
	"fmt"
	"log"
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
func NewMCR(db *sqlx.DB) (*MCR, error) {
	mcr := &MCR{
		db: db,
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
	err := mcr.Reload(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to reload mcr: %w", err)
	}
	return mcr, nil
}

func (mcr *MCR) Reload(ctx context.Context) error {
	chs := []Channel{}
	err := mcr.db.SelectContext(ctx, &chs,
		`SELECT short_name, name, description, type, ingest_url, ingest_type,
		slate_url, visibility, archive, dvr
		FROM playout.channel;`)
	if err != nil {
		return fmt.Errorf("failed to get channels from db: %w", err)
	}
	for _, ch := range chs {
		ch.Status = "running"
		err = mcr.newChannel(ctx, ch, false)
		if err != nil {
			return fmt.Errorf("failed to add channel: %w", err)
		}
	}
	log.Printf("loaded %d channels", len(chs))
	return nil
}

// GetChannel retrieves a channel from playout
func (mcr *MCR) GetChannel(ctx context.Context, shortName string) (*Channel, error) {
	ch, ok := mcr.channels[shortName]
	if !ok {
		return nil, errors.New("channel doesn't exist")
	}
	return ch, nil
}

// GetChannels retrieves all channels
func (mcr *MCR) GetChannels() (map[string]*Channel, error) {
	return mcr.channels, nil
}

// newChannel adds the channel to memory and adds the helper services
func (mcr *MCR) newChannel(ctx context.Context, ch Channel, updateDB bool) error {
	mcr.channels[ch.ShortName] = &ch

	if updateDB {
		// TODO handle existing
		err := mcr.addChannelToDB(ctx, ch)
		if err != nil {
			return fmt.Errorf("failed to add channel to DB: %w", err)
		}
	}

	if ch.hasScheduler {
		sch, err := scheduler.New(mcr.db, ch.ID)
		if err != nil {
			return fmt.Errorf("failed to start scheduler: %w", err)
		}
		ch.sch = sch
	}

	if ch.hasPiper {
		piper, err := piper.New(ctx, piper.Config{
			Endpoint: "",
			Width:    1920,
			Height:   1080,
			FPS:      50,
		}, "brave")
		if err != nil {
			return fmt.Errorf("failed to start piper: %w", err)
		}
		ch.piper = piper
	}
	return nil
}

// addChannelToDB will add a channel
//
// Will update channel ID to the new one
func (mcr *MCR) addChannelToDB(ctx context.Context, ch Channel) error {
	channelID := 0
	err := mcr.db.GetContext(ctx, &channelID, `
		INSERT INTO playout.channel(
			channel_id,
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
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING channel_id;`,
		ch.ID, ch.ShortName, ch.Name, ch.Description, ch.ChannelType,
		ch.IngestURL, ch.IngestURL, ch.SlateURL, ch.Visibilty,
		ch.Archive, ch.DVR)
	if err != nil {
		return fmt.Errorf("failed to insert channel to DB: %w", err)
	}
	ch.ID = channelID
	return nil
}

// NewChannel creates a new channel to playout
func (mcr *MCR) NewChannel(ctx context.Context, newCh NewChannelStruct) (*Channel, error) {
	ch := Channel{
		ShortName:   newCh.ShortName,
		Name:        newCh.Name,
		Description: newCh.Description,
		ChannelType: newCh.ChannelType,
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

	err := mcr.newChannel(ctx, ch, true)
	if err != nil {
		return nil, fmt.Errorf("failed to add channel to memory: %w", err)
	}

	return &ch, nil
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
