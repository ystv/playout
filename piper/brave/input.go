package brave

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type (
	// NewInput information required to create a new Brave input
	NewInput struct {
		URI      string
		Type     string
		HasAudio bool
		HasVideo bool
		Volume   float64
		Position int // defaut is auto. -1 for live
		Width    int
		Height   int
	}
	// NewInputResponse response from Brave
	NewInputResponse struct {
		InputID     int    `json:"id"`
		UniversalID string `json:"uid"`
	}
)

// New creates a new input
func (b *Brave) New(ctx context.Context, i NewInput) (NewInputResponse, error) {
	reqBody, err := json.Marshal(i)
	if err != nil {
		return NewInputResponse{}, fmt.Errorf("failed to marshal new pipe json: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, b.endpoint+"/api/inputs",
		bytes.NewReader(reqBody))
	if err != nil {
		return NewInputResponse{}, fmt.Errorf("failed to make request: %w", err)
	}
	httpRes, err := b.c.Do(req)
	if err != nil {
		return NewInputResponse{}, fmt.Errorf("failed to do request: %w", err)
	}
	defer httpRes.Body.Close()

	body, err := io.ReadAll(httpRes.Body)
	if err != nil {
		return NewInputResponse{}, fmt.Errorf("failed to read response body: %w", err)
	}
	res := NewInputResponse{}
	err = json.Unmarshal(body, &res)
	if err != nil {
		return NewInputResponse{}, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	return res, nil
}

// Delete removes the input
func (b *Brave) Delete(ctx context.Context, inputID int) error {
	url := fmt.Sprintf("%s/api/inputs/%d", b.endpoint, inputID)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("failed to make delete input request: %w", err)
	}
	_, err = b.c.Do(req)
	if err != nil {
		return fmt.Errorf("failed to do delete input request: %w", err)
	}
	return nil
}
