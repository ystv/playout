// Package piper is a safety-buffer and a mixer for a channel.
package piper

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/ystv/playout/piper/brave"
)

var (
	// ErrUnknownMixer is when an unsupported mixer is
	// attempted to be used
	ErrUnknownMixer = errors.New("unknown mixer")
)

type (
	// IPiper is a buffer which can swap video inputs without
	// dropping it's output feed with fallible sources
	IPiper interface {
		Restart() error
		GetState() (*Piper, error)
	}
	// State is the internal representation of a piper
	Piper struct {
		Inputs      []Input
		Composition Composition
		Outputs     []Output

		lock sync.RWMutex

		// States
		// When new mixers are offered their state would
		// be stored here.
		mixer string // i.e. brave, obs, liquidsoap
		brave *brave.Brave
	}
	// Config base requirements for a new Piper
	Config struct {
		Endpoint string
		Width    int
		Height   int
		FPS      int
	}
)

type (
	// InputStore handles ingesting a video source to our composition
	InputStore interface {
		New(ctx context.Context, i NewInput) error
		Delete(ctx context.Context, sourceID int) error
	}
	// InputObject methods providing individual control over source
	InputObject interface {
		Start() error
		Stop() error
		Delete() error
	}
	// Input is a video source that can be put in a composition
	Input struct {
		InputID string
		URL     string `json:"url"`
		State   string `json:"state"` // NULL / READY / PAUSED / PLAYING
		Type    string `json:"type"`  // LIVE / VT / TEST
		Width   int    `json:"width"`
		Height  int    `json:"hieght"`
	}
	// NewInput is used to create a new input
	NewInput struct {
		Input
	}

	// Composition is the video that will be outputted
	Composition struct {
		CompositionID string
		State         string `json:"state"`
		Width         int    `json:"width"`
		Height        int    `json:"height"`
		Sources       []int  `json:"sources"` // source IDs
	}

	// Compositioner handles changing the mix
	// TODO: Is this needed?
	Compositioner interface {
		SetSource(sourceID int) error
	}
	// Outputer handles providing a video output to a defined URL
	Outputer interface {
		NewOutput(url string) error
		DeleteOutput(outputID string) error
	}
	// Output is a piper output
	Output struct {
		OutputID string
		URL      string `json:"url"`
		State    string `json:"state"`
		Type     string `json:"type"`
		Width    int    `json:"width"`
		Height   int    `json:"height"`
		Bitrate  int    `json:"bitrate"`
		Codec    string `json:"codec"`
	}
)

// New creates a new Piper instance
func New(ctx context.Context, conf Config, mixer string) (*Piper, error) {
	s := &Piper{}
	var err error
	switch mixer {
	case "brave":
		s.brave, err = brave.New(ctx, conf.Endpoint, conf.Width, conf.Height)
		if err != nil {
			return nil, fmt.Errorf("failed to create new brave piper: %w", err)
		}

	default:
		return nil, ErrUnknownMixer
	}
	s.mixer = mixer
	err = s.UpdateState(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to update state: %w", err)
	}
	return s, nil
}

// UpdateState will update the state to reflect the chosen mixer
func (p *Piper) UpdateState(ctx context.Context) error {
	p.lock.Lock()
	defer p.lock.Unlock()
	switch p.mixer {
	case "brave":
		b, err := p.brave.GetState()
		if err != nil {
			return fmt.Errorf("failed to get brave state: %w", err)
		}
		// Brave inputs to Piper inputs
		p.Inputs = []Input{}
		for _, input := range b.Inputs {
			p.Inputs = append(p.Inputs, Input{
				InputID: string(rune(input.ID)),
				URL:     input.URI,
				State:   input.State,
				Type:    input.Type,
				Width:   input.Width,
				Height:  input.Height,
			})
		}
		// Brave Outputs to Piper outputs
		p.Outputs = []Output{}
		for _, output := range b.Outputs {
			p.Outputs = append(p.Outputs, Output{
				OutputID: string(rune(output.ID)),
				URL:      output.URI,
				State:    output.State,
				Type:     output.Type,
				Width:    output.Width,
				Height:   output.Height,
				Bitrate:  0, // TODO: Look into
				Codec:    "unknown",
			})
		}
		p.Composition = Composition{}
		for _, mixer := range b.Mixers {
			if mixer.ID == b.MainMixID {
				p.Composition = Composition{
					CompositionID: string(rune(mixer.ID)),
					State:         mixer.State,
					Width:         mixer.Height,
					Height:        mixer.Height,
				}
				for _, source := range mixer.Sources {
					p.Composition.Sources = append(p.Composition.Sources, source.ID)
				}
			}
		}
		// TODO: If no mixer, make one
	default:
		return ErrUnknownMixer
	}
	return nil
}
