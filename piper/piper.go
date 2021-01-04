package piper

import (
	"context"
)

type (
	// Piper is a buffer which can swap video inputs without
	// dropping it's output feed with fallible sources
	Piper interface {
		Restart() error
		GetState() (*State, error)
	}
	State struct {
		Inputs      []Input
		Composition Composition
		Outputs     []Output
	}
	New struct {
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
		URL    string `json:"url"`
		State  string `json:"state"` // NULL / READY / PAUSED / PLAYING
		Type   string `json:"type"`  // LIVE / VT / TEST
		Width  int    `json:"width"`
		Height int    `json:"hieght"`
	}
	// NewInput is used to create a new input
	NewInput struct {
		Input
	}
)

type (
	// Composition is the video that will be outputted
	Composition interface {
		SetSource(sourceID int) error
	}
	// Output handles providing a video output to a defined URL
	IOutput interface {
		NewOutput(url string) error
		DeleteOutput(outputID string) error
	}
	Output struct {
		URL     string `json:"url"`
		State   string `json:"state"`
		Type    string `json:"type"`
		Width   int    `json:"width"`
		Height  int    `json:"height"`
		Bitrate int    `json:"bitrate"`
		Codec   string `json:"codec"`
	}
)
