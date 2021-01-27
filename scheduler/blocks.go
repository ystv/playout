package scheduler

import (
	"context"
	"errors"
	"fmt"
)

// NewBlock adds a block to the schedule
func (s *Scheduler) NewBlock(ctx context.Context, b NewBlock) (int, error) {
	/*
		First check the validity of the block,
		* Channel exists
		* Programme exists
		* Time isn't overlapping existing schedule
	*/
	blockID := 0
	_, err := s.prog.Get(ctx, b.ProgrammeID)
	if err != nil {
		return blockID, fmt.Errorf("failed to get programme: %w", err)
	}
	blocks, err := s.GetRange(ctx, b.Start, b.End)
	if err != nil {
		return blockID, fmt.Errorf("failed to get range: %w", err)
	}
	if len(blocks) != 0 {
		return blockID, errors.New("time already scheduled: %w")
	}
	err = s.db.SelectContext(ctx, &blockID, `
		INSERT INTO schedule_blocks
		(channel_id, programme_id, ingest_url, ingest_type, scheduled_start,
		scheduled_end)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING block_id;`, b.ChannelID, b.ProgrammeID, b.IngestURL, b.IngestType, b.Start, b.End)
	if err != nil {
		return blockID, fmt.Errorf("failed to insert new block")
	}
	return blockID, nil
}
