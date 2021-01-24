package public

import "context"

type Channel struct {
	Name         string
	Description  string
	ThumbnailURL string
	OutputURL    string
}

func (s *Store) GetAll(ctx context.Context) ([]Channel, error) {
	return nil, nil
}
