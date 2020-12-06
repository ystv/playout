package brave

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ystv/playout/piper"
)

var _ piper.InputStore = &Brave{}

// New creates a new input
func (b *Brave) New(ctx context.Context, i piper.NewInput) error {

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
