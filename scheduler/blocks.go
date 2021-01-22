package scheduler

import (
	"context"
	"fmt"
)

// NewBlock adds a block to the schedule
func (s *Scheduler) NewBlock(ctx context.Context, b NewBlock) error {
	/*
		First check the validity of the block,
		* Channel exists
		* Programme exists
		* Time isn't overlapping existing schedule
	*/
	_, err := s.prog.Get(ctx, b.ProgrammeID)
	if err != nil {
		return fmt.Errorf("failed to get programme: %w", err)
	}
	blocks, err := s.GetRange(ctx, b.Start, b.End)
	if err != nil {
		return fmt.Errorf("failed to get range: %w", err)
	}
	return nil
}
