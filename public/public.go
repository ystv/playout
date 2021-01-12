package public

import "github.com/ystv/playout/channel"

// This offers an API of the channels and their content schedule

type Store struct {
	chs channel.Channels
}
