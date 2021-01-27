package scheduler

import (
	"context"
	"time"
)

// MainLoop is the subroutine to manage the schedule
func (s *Scheduler) MainLoop(ctx context.Context) error {
	timer := time.Tick(5 * time.Second)
	hasChanged := false
	for time := range timer {

		if hasChanged {
			s.log.Printf("%s - something happened", time.String())
		}
	}
	return nil
}
