// Package channel deals with the technical aspect of the channel only. For the information (schedule)
// part of the channel check the scheduler package
package channel

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type (
	// Channel represents a video feed
	Channel struct {
		ID          string
		Name        string
		Description string
		ChannelType string // event / linear
		IngestURL   string
		IngestType  string   // RTP / RTMP / HLS
		SlateURL    string   // Fallback video
		Outputs     []Output // Configured outputs
		Archive     bool     // Add to VOD after
		DVR         bool     // Can rewind
		Passthrough bool     // Encoding needed
		CreatedAt   time.Time
		Status      string // The state of channel ready / running / starting / stopping / pending

		conf *Config
	}

	// NewChannelStruct represnets the required channel config
	NewChannelStruct struct {
		ID          string
		Name        string
		Description string
		ChannelType string // event / linear
		IngestURL   string
		IngestType  string // RTSP / RTMP / HLS
		SlateURL    string // fallback video
		Outputs     []Output
		Archive     bool // Add to VOD after
		DVR         bool // can rewind
	}
)

// Outputs
type (
	// Output is an video output
	Output struct {
		Name        string
		Type        string // RTP / RTMP / HLS / DASH / CMAF
		Destination string // URL endpoint
		Renditions  []Rendition
		Passthrough bool
	}
	// Output types
)

// Renditions
type (
	// Rendition represents a rendition of the source video
	Rendition struct {
		Width   int
		Height  int
		Bitrate int // Kb/s
		FPS     int
		Codec   string // h264 / h265
	}
)

// Start the channel
//
// Will create new tasks for VT, starting a new endpoint on the id/main.m3u8
func (ch *Channel) Start() error {
	for _, output := range ch.Outputs {
		if ch.Passthrough {
			// just copy

			isDVR := ""
			if ch.DVR {
				isDVR = "-hls_playlist_type event"
			}

			t := Task{
				SrcURL:  ch.IngestURL,
				DstURL:  output.Destination,
				DstArgs: fmt.Sprintf("-c:v h264 -f hls -hls_time %d %s -hls_segment_type fmp4 -method PUT", 4, isDVR),
			}
			err := ch.NewStream(context.Background(), t)
			if err != nil {
				return err
			}
			ch.Status = "playing"
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
	}

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

// Task represents a task in VT
type Task struct {
	ID      string `json:"id"`      // Task UUID
	Args    string `json:"args"`    // Global arguments
	SrcArgs string `json:"srcArgs"` // Input file options
	SrcURL  string `json:"srcURL"`  // Location of source file on CDN
	DstArgs string `json:"dstArgs"` // Output file options
	DstURL  string `json:"dstURL"`  // Destination of finished encode on CDN
}

// NewStream will create a new stream
func (ch *Channel) NewStream(ctx context.Context, t Task) error {
	postBody, err := json.Marshal(t)
	if err != nil {
		return err
	}
	reqBody := bytes.NewBuffer(postBody)
	res, err := http.Post(ch.conf.VTEndpoint+"/new_live", "application/json", reqBody)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	log.Printf("%s", body)
	return nil
}
