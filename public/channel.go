package public

type Channel struct {
	Name         string
	Description  string
	ThumbnailURL string
	OutputURL    string
	Schedule     Schedule
}

func (s *Store)GetAll(ctx context.Context) ([]Channel, error) {
	s.chs[]
	return nil, nil
}
