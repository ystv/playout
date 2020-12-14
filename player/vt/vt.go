package vt

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ystv/playout/player"
)

// Player encapsulates VT's dependencies
type Player struct {
	endpoint string
	c        http.Client
}

var _ player.Player = &Player{}

// EncodeArgs are the FFmpeg arguements on the playout encode
type EncodeArgs struct {
	Args    string `json:"args"`    // Global arguments
	DstArgs string `json:"dstArgs"` // Output file options
	DstURL  string `json:"dstURL"`  // Destination
}

// Play will create a VT Task to play the programme
func (p *Player) Play(ctx context.Context, c player.Config) error {
	reqBody := struct {
		EncodeArgs EncodeArgs
		Videos     []string
	}{
		EncodeArgs: EncodeArgs{
			Args:    "-re",
			DstArgs: "-c:v libx264 -bitrate 10M -f flv",
			DstURL:  c.DstURL,
		},
		Videos: c.VideoURLs,
	}
	reqJSON, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal VT task: %w", err)
	}
	res, err := p.c.Post(p.endpoint+"/task/play", "application/json", bytes.NewReader(reqJSON))
	if err != nil {
		return fmt.Errorf("failed to submit VT task: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusCreated {
		return fmt.Errorf("VT failed to create task: %v", res.Body)
	}
	return nil
}

// New creates a new VT-based player
func New(endpoint string) (*Player, error) {
	p := Player{
		endpoint: endpoint,
		c:        http.Client{},
	}
	res, err := p.c.Get(endpoint + "/ok")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to VT: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("VT not OK: %v", res.Body)
	}
	return &p, nil
}
