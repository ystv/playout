// Package channel deals with the technical aspect of the channel only. For the information (schedule)
// part of the channel check the scheduler package
package channel

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
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
		OriginURL   string
		OriginType  string // RTSP / RTMP / HLS
		SlateURL    string // fallback video
		PlaybackURL string
		Renditions  []Rendition
		Outputs     []string // Configured outputs
		Archive     bool     // Add to VOD after
		DVR         bool     // can rewind
		OriginOnly  bool     // Encoding needed
		CreatedAt   time.Time
		Status      string // The state of channel ready / running / starting / stopping / pending

		conf *Config
	}
	// Rendition represents a rendition of the source video
	Rendition struct {
		Width   int
		Height  int
		Bitrate int // Kb/s
		FPS     int
		Codec   string // h264 / h265
	}
	// NewChannelStruct represnets the required channel config
	NewChannelStruct struct {
		ID          string
		Name        string
		Description string
		ChannelType string // event / linear
		OriginURL   string
		OriginType  string // RTSP / RTMP / HLS
		SlateURL    string // fallback video
		Renditions  []Rendition
		Archive     bool // Add to VOD after
		DVR         bool // can rewind
		OriginOnly  bool // Encoding needed
	}
)

// Start the channel
//
// Will create new tasks for VT, starting a new endpoint on the id/main.m3u8
func (ch *Channel) Start() error {

	ch.PlaybackURL = ch.conf.OutputEndpoint + "/" + ch.ID + "/main.m3u8"

	if ch.OriginOnly {
		// just copy

		isDVR := ""
		if ch.DVR {
			isDVR = "-hls_playlist_type event"
		}

		t := Task{
			SrcURL:  ch.OriginURL,
			DstURL:  ch.PlaybackURL,
			DstArgs: fmt.Sprintf("-c:v h264 -f hls -hls_time %d %s -hls_segment_type fmp4 -method PUT", 4, isDVR),
		}
		err := ch.NewStream(context.Background(), t)
		if err != nil {
			return err
		}
		ch.Status = "playing"
		return nil
	}
	// there is renditions to be made
	inputArgs := fmt.Sprintf(`
		-re -stream_loop 1 -i %s `, ch.OriginURL)
	mapString := strings.Repeat("-map 0 ", len(ch.Renditions))
	audioString := "-c:a aac -ar 48000"

	encode := ""
	for idx, rendition := range ch.Renditions {
		videoCodec := ""
		switch rendition.Codec {
		case "h264":
			videoCodec = "libx264"
		case "h265":
			videoCodec = "libx265"
			// TODO: nvenc codec's
		}
		encode += fmt.Sprintf(`
			-vf scale=w=%d:h=%d:force_original_aspect_ratio=decrease
			-b:v:%d %dk
			-c:v:%d %s
			-s:v:%d %dx%d 
			-profile:v:%d main
			-pix_fmt yuv420p
			-b:v %dk`, rendition.Width, rendition.Height,
			idx, rendition.Bitrate,
			idx, videoCodec,
			idx, rendition.Width, rendition.Height,
		)

	}

	outputString := ""
	switch ch.OriginType {
	case "hls":
		outputString = fmt.Sprintf(`
		-keyint_min 120 -g 120 -sc_threshold 0 -use_timeline 1
		-use_template 1 -window_size 5
		-adaptation_sets "id=0,streams=v id=1,streams=a
		-hls_playlist 1 -seg_duration 4 -streaming 1
		-remove_at_exit 1 -method PUT -f hls %s`, ch.conf.OutputEndpoint)
	case "cmaf":
		log.Println("lol")
		panic("get out of here lol")
	}

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
