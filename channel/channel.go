// Package channel deals with the technical aspect of the channel only. For the information (schedule)
// part of the channel check the scheduler package
package channel

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/ystv/playout/piper"
	"github.com/ystv/playout/scheduler"
)

type (
	// Channel represents a video feed
	Channel struct {
		// Core
		ID          int      `db:"channel_id"`
		ShortName   string   `db:"short_name"` // URL name
		ChannelType string   `db:"type"`       // event / linear
		IngestURL   string   `db:"ingest_url"`
		IngestType  string   `db:"ingest_type"` // RTP / RTMP / HLS
		SlateURL    string   `db:"slate_url"`   // Fallback video
		Outputs     []Output // Configured outputs
		Status      string   // The state of channel ready / running / starting / stopping / pending

		// Options
		Visibilty string `db:"visibility"`
		Archive   bool   `db:"archive"` // Add to VOD after
		DVR       bool   `db:"dvr"`     // Allow rewind on outputs

		// Frontend
		Name        string    `db:"name"` // Display name
		Description string    `db:"description"`
		Thumbnail   string    `db:"thumbnail"`
		CreatedAt   time.Time `db:"created_at"`

		// Modules
		hasScheduler bool `db:"has_scheduler"`
		sch          *scheduler.Scheduler
		hasPiper     bool `db:"has_piper"`
		piper        *piper.Piper

		// Dependencies
		conf *Config
	}

	// NewChannelStruct represnets the required channel config
	NewChannelStruct struct {
		ShortName    string
		Name         string
		Description  string
		ChannelType  string // event / linear
		IngestType   string // RTSP / RTMP / HLS
		SlateURL     string // fallback video
		Visible      string // public / internal / private. TOOD: Will it stay?
		Archive      bool   // Add to VOD after
		DVR          bool   // Allow rewind on outputs, is a default value
		HasScheduler bool
		HasPiper     bool
		Outputs      []Output
	}

	// Outputs

	// Output is an video output
	Output struct {
		Name        string
		Type        string // RTP / RTMP / HLS / DASH / CMAF
		Passthrough bool
		DVR         bool   // can rewind
		Destination string // URL endpoint
		Renditions  []Rendition
	}
	// Output types

	// Renditions

	// Rendition represents a generic rendition of the source video
	Rendition struct {
		Width   int
		Height  int
		Bitrate int // Kb/s
		FPS     int
		Codec   string // h264 / h265
	}

	// We might have some custom renditions for the fancier outputs
)

// Start the channel
//
// Will create new tasks for VT, starting a new endpoint on the id/main.m3u8
func (ch *Channel) Start() error {
	for _, output := range ch.Outputs {
		if output.Passthrough {
			// just copy

			t := Task{
				SrcURL:  ch.IngestURL,
				DstURL:  output.Destination,
				DstArgs: "",
			}

			switch output.Type {
			case "rtmp":
				t.DstArgs = `-c copy -f flv`
			case "hls":
				isDVR := ""
				if output.DVR {
					isDVR = "-hls_playlist_type event"
				}
				t.DstArgs = fmt.Sprintf(`-c copy -f hls -hls_time %d %s -hls_segment_type fmp4 -method PUT`, 4, isDVR)
			default:
				return errors.New("unknown output type")
			}

			log.Printf("%+v", t)
			// err := ch.NewStream(context.Background(), t)
			// if err != nil {
			// 	return err
			// }
			return nil
		}
		inputArgs := ""
		switch ch.IngestType {
		case "rtmp":
			inputArgs = fmt.Sprintf(`-re -stream_loop 1 -f flv -i "%s" `, ch.IngestURL)

		default:
			inputArgs = fmt.Sprintf(`-re -stream_loop 1 -f %s -i "%s" `, ch.IngestType, ch.IngestURL)
		}
		mapString := ""
		if len(output.Renditions) == 0 {
			mapString = "-map 0"
		} else {
			mapString = strings.Repeat("-map 0 ", len(output.Renditions)-1)
		}
		inputArgs = fmt.Sprintf("%s %s", inputArgs, mapString)

		// Audio mux
		audioString := "-c:a aac -ar 48000"

		// Video mux
		encode := ""
		for idx, rendition := range output.Renditions {
			videoCodec := ""
			switch rendition.Codec {
			case "h264":
				videoCodec = "libx264"
			case "h265":
				videoCodec = "libx265"
				// TODO: nvenc codec's
			default:
				return errors.New("unknown codec")
			}
			encode = fmt.Sprintf(`
					-vf "scale=w=%d:h=%d:force_original_aspect_ratio=decrease"
					-b:v:%d %dk
					-c:v:%d %s
					-s:v:%d %dx%d 
					-profile:v:%d main
					-pix_fmt yuv420p`,
				rendition.Width, rendition.Height, // scale resolution
				idx, rendition.Bitrate, // bitrate
				idx, videoCodec, // codec
				idx, rendition.Width, rendition.Height, // resolution
				idx, // profile
			)
		}
		outputString := ""
		switch output.Type {
		case "rtp":
			log.Println("rtp out")
			outputString = fmt.Sprintf(`-f rtp "%s"`, output.Destination)

		case "rtmp":
			log.Println("rtmp out")
			outputString = fmt.Sprintf(`-f flv "%s"`, output.Destination)
		case "hls":
			log.Println("hls out")
			// Output
			outputString = fmt.Sprintf(`
				-keyint_min 120 -g 120 -sc_threshold 0 -use_timeline 1
				-use_template 1 -window_size 5
				-adaptation_sets "id=0,streams=v id=1,streams=a"
				-hls_playlist 1 -seg_duration 4 -streaming 1
				-remove_at_exit 1 -method PUT -f hls "%s"`, output.Destination)

		case "dash":
			return errors.New("dash not implemented")

		case "cmaf":
			return errors.New("cmaf not implemented")

		default:
			return errors.New("unknown output type")
		}
		// Combine to an executable string
		removeInsideWhitespace := regexp.MustCompile(`[\s\p{Zs}]{2,}`)
		cmd := strings.ReplaceAll(fmt.Sprintf("%s %s %s %s", inputArgs, audioString, encode, outputString), "\n", "")
		cmd = removeInsideWhitespace.ReplaceAllString(cmd, " ")
		log.Println(cmd)
	}
	ch.Status = "playing"
	return nil
}

// Stop the channel
//
// Will cancel VT jobs, triggering archiving if enabled
func (ch *Channel) Stop() error {
	ch.Status = "stopping"
	return nil
}

// Stat returns the current status of the channel
//
// Used by http api to allow VT to check if the stream still needs to be up
func (ch *Channel) Stat() (string, error) {
	return ch.Status, nil
}
