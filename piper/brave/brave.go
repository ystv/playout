// Package brave is an implementation of Piper in brave.
// TODO: Remove piper as dependency and make piper adapt to brave in it's own package
package brave

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Brave piper instance
type Brave struct {
	c        http.Client
	endpoint string
	State    *State
}

type (
	// State represents an adaptation of the Brave state
	State struct {
		MainMixID int
		Inputs    []Input   `json:"inputs"`
		Overlays  []Overlay `json:"overlays"`
		Outputs   []Output  `json:"outputs"`
		Mixers    []Mixer   `json:"mixer"`
	}
	// Input represents a Brave input object.
	Input struct {
		ID              int     `json:"id"`
		UID             string  `json:"uid"`
		URI             string  `json:"uri"`
		Type            string  `json:"type"`
		HasAudio        bool    `json:"has_audio"`
		HasVideo        bool    `json:"has_video"`
		Volume          float64 `json:"volume"`
		Position        int     `json:"position"`
		State           string  `json:"state"`
		ConnectionSpeed int     `json:"connection_speed"`
		BufferSize      int     `json:"buffer_size"`
		BufferDuration  int     `json:"buffer_duration"`
		Width           int     `json:"width"`
		Height          int     `json:"height"`
	}
	// Overlay represents a Brave overlay object.
	Overlay struct {
		ID      int    `json:"id"`
		UID     string `json:"uid"`
		Type    string `json:"type"`
		Visible bool   `json:"visible"`
		Source  string `json:"source"`
	}
	// Output represents a Brave output object
	Output struct {
		ID       int    `json:"id"`
		UID      string `json:"uid"`
		Source   string `json:"source"`
		URI      string `json:"uri"`
		Type     string `json:"type"`
		HasAudio bool   `json:"has_audio"`
		HasVideo bool   `json:"has_video"`
		Width    int    `json:"width"`
		Height   int    `json:"height"`
		State    string `json:"string"`
	}
	// Mixer represents a Brave mixer object
	Mixer struct {
		HasAudio bool        `json:"has_audio"`
		HasVideo bool        `json:"has_video"`
		UID      string      `json:"uid"`
		Width    int         `json:"width"`
		Height   int         `json:"height"`
		Pattern  int         `json:"pattern"`
		State    string      `json:"state"`
		ID       int         `json:"id"`
		Type     string      `json:"type"`
		Sources  []MixSource `json:"sources"`
	}
	// MixSource represents a Brave mixer source's object
	MixSource struct {
		UID       string `json:"uid"`
		ID        int    `json:"id"`
		BlockType string `json:"block_type"`
		InMix     bool   `json:"in_mix"`
	}
)

// New creates a new brave object
//
// The URL is the endpoint of the brave instance
func New(ctx context.Context, endpoint string, width, height int) (*Brave, error) {
	b := &Brave{
		c:        http.Client{},
		endpoint: endpoint,
	}
	res, err := b.c.Get(b.endpoint + "/api/all")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to brave: %w", err)
	}
	defer res.Body.Close()

	reqBody := Mixer{
		Height: height,
		Width:  width,
	}
	reqJSON, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal new pipe json: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, "/mixers", bytes.NewReader(reqJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	_, err = b.c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to do request: %w", err)
	}
	return b, nil
}

// Restart will restart the Brave instance
//
// Currently will re-use the existing configuration.
// TODO: Might go to default state, and rebuild.
func (b *Brave) Restart() error {
	reqBody := struct {
		Config string `json:"config"`
	}{
		Config: "current",
	}
	reqJSON, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marhsal restart json: %w", err)
	}
	_, err = b.c.Post("/api/restart", "application/json", bytes.NewReader(reqJSON))
	if err != nil {
		return fmt.Errorf("failed to restart Brave: %w", err)
	}
	return nil
}

// GetState updates the internal state with what Brave is currently
func (b *Brave) GetState() (*State, error) {
	res, err := b.c.Get(b.endpoint + "/api/all")
	if err != nil {
		return nil, fmt.Errorf("failed to request state: %w", err)
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	err = json.Unmarshal(body, &b.State)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	return b.State, nil
}
