// Package playout provides an interace for managing and creating
// playouts. A playout is a video stream of content to be played on
// a channel.
package playout

import "context"

type (
	Playouter interface {
		Get(ctx context.Context)
	}
)
