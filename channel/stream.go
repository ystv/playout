package channel

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

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
