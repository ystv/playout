package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jmoiron/sqlx"

	"github.com/ystv/playout/channel"
	"github.com/ystv/playout/scheduler"
)

func main() {
	mcr := channel.NewMCR()
	ch, err := mcr.NewChannel(context.Background(), channel.NewChannelStruct{
		Name:        "Cooking time",
		Description: "Very cool cooking show",
		ChannelType: "linear",
		IngestURL:   "rtmp://stream.ystv.co.uk/internal/rhys",
		IngestType:  "rtmp",
		SlateURL:    "https://cdn.ystv.co.uk/ystv-holding.mp4",
		Outputs: []channel.Output{
			{
				Name:        "website stream",
				Type:        "hls",
				Passthrough: false,
				DVR:         true,
				Destination: "https://live-media.ystv.co.uk/test123-manifest.m3u8",
				Renditions: []channel.Rendition{
					{
						Width:   1920,
						Height:  1080,
						Bitrate: 8000,
						FPS:     25,
						Codec:   "h264",
					}, {
						Width:   1280,
						Height:  720,
						Bitrate: 4000,
						FPS:     25,
						Codec:   "h264",
					},
				},
			},
			{
				Name:        "signage stream",
				Type:        "rtmp",
				Passthrough: true,
				DVR:         false,
				Destination: "rtmp://stream.ystv.co.uk/internal/test",
				// RTMP type doesn't have renditions
			},
		},
		Archive: true,
	})
	if err != nil {
		log.Fatalf("failed to create new channel: %+v", err)
	}

	db, err := newDatabase()
	if err != nil {
		log.Fatalf("failed to start database: %+v", err)
	}

	sch, err := scheduler.New(db, ch)

	err = sch.MainLoop(context.Background())
	if err != nil {
		log.Fatalf("scheduling failed: %+v", err)
	}

	err = ch.Start()
	if err != nil {
		log.Fatalf("%+v", err)
	}
}

// newDatabase creates a new database connection
func newDatabase() (*sqlx.DB, error) {
	host := os.Getenv("PLAYOUT_DB_HOST")
	port := os.Getenv("PLAYOUT_DB_PORT")
	sslMode := os.Getenv("PLAYOUT_DB_SSLMODE")
	name := os.Getenv("PLAYOUT_DB_NAME")
	username := os.Getenv("PLAYOUT_DB_USER")
	password := os.Getenv("PLAYOUT_DB_PASS")

	dbURI := fmt.Sprintf("dbname=%s host=%s user=%s password=%s port=%s sslmode=%s", name, host, username, password, port, sslMode) // Build connection string

	db, err := sqlx.Open("postgres", dbURI)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to DB: %w", err)
	}
	return db, nil
}
