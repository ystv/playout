package channel

import (
	"context"
	"log"
)

func Demo() {
	// make channel
	newCh := NewChannelStruct{
		Name:        "Cooking time",
		Description: "Very cool cooking show",
		ChannelType: "linear",
		IngestURL:   "rtmp://stream.ystv.co.uk/internal/rhys",
		IngestType:  "rtmp",
		SlateURL:    "https://cdn.ystv.co.uk/ystv-holding.mp4",
		Outputs: []Output{
			{
				Type:        "hls",
				Destination: "live-media.ystv.co.uk-test123-manifest.m3u8",
				Renditions: []Rendition{
					{
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
			},
			{
				Passthrough: true,
				Type:        "rtmp",
				Destination: "rtmp://stream.ystv.co.uk/internal/test",
			},
		},
		Archive: true,
		DVR:     true,
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
