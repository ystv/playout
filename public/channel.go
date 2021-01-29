package public

import "context"

type Channel struct {
	Name         string
	Description  string
	ThumbnailURL string
	OutputURL    string
}

// GetAll retrieves all channels
func (p *Publicer) GetAll(ctx context.Context) ([]Channel, error) {
	return nil, nil
}
