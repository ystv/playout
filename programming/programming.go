package programming

import "context"

type (
	// ProgrammeStore handles managing programmes
	ProgrammeStore interface {
		New(ctx context.Context, p Programme) error
		Get(ctx context.Context, programmeID int) (*Programme, error)
		Delete(ctx context.Context, programmeID int) error
	}
	// Programme is the stream to be played out
	//
	// Metadata is to be displayed in the schedule.
	//
	// The list of videos is given to the player object which
	// will play them all in sequence giving the illusion of one
	// large video stream.
	//
	// Videos can be made up of any accessible URLs that the
	// player can encode.
	//
	// If the length of Videos is 0. We presume that it is live
	// content, and a player won't be made.
	Programme struct {
		ID          int     `db:"programme_id" json:"id"`
		Title       string  `db:"title"`
		Description string  `db:"description"`
		Thumbnail   string  `db:"thumbnail"`
		Videos      []Video `json:"videos"`
	}
	// Video is the individual video to be played out
	Video struct {
		ID  int    `db:"programme_video_id" json:"id"`
		URL string `db:"url" json:"url"`
	}
)

// Run will execute a program to be played out
func (p *Programme) Run(ctx context.Context) error {

	return nil
}
