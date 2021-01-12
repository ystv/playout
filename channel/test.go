package channel

import (
	"context"
	"log"
)

func demo() {
	// make channel
	newCh := NewChannelStruct{
		Name:        "Cooking time",
		Description: "Very cool cooking show",
		ChannelType: "linear",
		OriginURL:   "rtmp://stream.ystv.co.uk/live/rhys",
		OriginType:  "rtmp",
		SlateURL:    "https://cdn.ystv.co.uk/ystv-holding.mp4",
		Renditions: []Rendition{{
			Width:   1920,
			Height:  1080,
			Bitrate: 8000,
			FPS:     25,
		}, {
			Width:   1280,
			Height:  720,
			Bitrate: 4000,
			FPS:     25,
		},
		},
		Archive:    true,
		DVR:        true,
		OriginOnly: true,
	}
	chService := New()
	ch, err := chService.New(context.Background(), newCh)
	if err != nil {
		log.Fatalf("%+v", err)
	}
	err = ch.Start()
	if err != nil {
		log.Fatalf("%+v", err)
	}
}
