// Package public offers an API of the channels and their content schedule
package public

import "github.com/ystv/playout/channel"

// Publicer publicises the playout system
type Publicer struct {
	mcr channel.MCR
}
