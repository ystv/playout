package brave

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ystv/playout/piper"
)

var _ piper.Piper = &Brave{}

type Brave struct {
	c        http.Client
	endpoint string
	state    State
}

type (
	// State represents an adaptation of the Brave state
	State struct {
		Inputs   []Input   `json:"inputs"`
		Overlays []Overlay `json:"overlays"`
		Outputs  []Output  `json:"outputs"`
		Mixers   []Mixer   `json:"mixer"`
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
		BufferDuration  int     `json:"buffer_duration`
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
		ID       int    `json:id"`
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
func New(ctx context.Context, p piper.New) (*Brave, error) {
	b := &Brave{
		c:        http.Client{},
		endpoint: p.Endpoint,
	}
	res, err := b.c.Get(b.endpoint + "/api/all")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to brave: %w", err)
	}
	defer res.Body.Close()

	reqBody := Mixer{
		Height: p.Height,
		Width:  p.Width,
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

func (b *Brave) GetState() (*piper.State, error) {
	p := piper.State{}
	for _, bInput := range b.state.Inputs {
		p.Inputs = append(p.Inputs, piper.Input{
			URL:    bInput.URI, // TODO: Look into this
			State:  bInput.State,
			Type:   bInput.Type,
			Width:  bInput.Width,
			Height: bInput.Height,
		})
	}

	for _, bOutput := range b.state.Outputs {
		p.Outputs = append(p.Outputs, piper.Output{
			URL: bOutput.URI,
		})
	}
	return nil, nil
}
