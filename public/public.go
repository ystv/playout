// Package public offers an API of the channels and their content schedule
package public

import (
	"github.com/ystv/playout/channel"
	"github.com/ystv/playout/playout"
	"github.com/ystv/playout/programming"
)

// Publicer publicises the playout system
type Publicer struct {
	mcr  *channel.MCR
	prog *programming.Programmer
	po   *playout.Playouter
}

// New creates a new instance of a publicer
//
// Publicer provides an API that is suitable for end user clients
func New(mcr *channel.MCR, prog *programming.Programmer, po *playout.Playouter) *Publicer {
	return &Publicer{
		mcr:  mcr,
		prog: prog,
		po:   po,
	}
}
