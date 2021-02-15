// Package public offers an API of the channels and their content schedule
package public

import (
	"github.com/ystv/playout/channel"
	"github.com/ystv/playout/playout"
	"github.com/ystv/playout/programming"
)

// Publicer publicises the playout system
type Publicer struct {
	mcr  channel.MCR
	prog programming.ProgrammeStore
	po   playout.Playout
}
