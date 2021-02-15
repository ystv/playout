package public

import (
	"context"
	"fmt"
	"time"
)

type (
	// Channel public representation
	//
	// Linear or event video stream
	Channel struct {
		ShortName   string    `json:"shortName"`
		Name        string    `json:"name"`
		Description string    `json:"description"`
		Thumbnail   string    `json:"thumbnail"`
		OutputURL   string    `json:"outputURL"`
		Schedule    []Playout `json:"schedule"`
	}
	// Playout public representation
	//
	// A video stream
	Playout struct {
		PlayoutID int       `json:"playoutID"`
		Start     time.Time `json:"start"`
		End       time.Time `json:"end"`
		// If Archived, this is the URL to it
		PlayoutVOD string    `json:"playoutVOD"`
		DVR        bool      `json:"dvr"`
		Archive    bool      `json:"archive"`
		Programme  Programme `json:"programme"`
	}
	// Programme public representation
	//
	// A video with some extra data
	Programme struct {
		ProgrammeID int
		Title       string
		Description string
		Thumbnail   string
		Type        string
		PrimaryVOD  string
		VODs        []string // Other playout VOD's with the same programme
	}
)

// GetAll retrieves all channels
func (p *Publicer) GetAll(ctx context.Context) ([]Channel, error) {
	return nil, nil
}

// GetChannel returns the public representation of a channel
func (p *Publicer) GetChannel(ctx context.Context, shortName string) (*Channel, error) {
	ch, err := p.mcr.GetChannel(ctx, shortName)
	if err != nil {
		return nil, fmt.Errorf("failed to get public channel: %w", err)
	}

	chPublic := &Channel{
		ShortName:   ch.ShortName,
		Name:        ch.Name,
		Description: ch.Description,
	}

	return chPublic, nil
}
