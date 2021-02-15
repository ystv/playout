package scheduler

import (
	"context"
	"errors"
	"fmt"
)

// NewPlayout adds a playout to the schedule
func (s *Scheduler) NewPlayout(ctx context.Context, b NewPlayout) (int, error) {
	/*
		First check the validity of the playout,
		* Channel exists
		* Programme exists
		* Time isn't overlapping existing schedule
	*/
	playoutID := 0
	_, err := s.prog.Get(ctx, b.ProgrammeID)
	if err != nil {
		return playoutID, fmt.Errorf("failed to get programme: %w", err)
	}
	playouts, err := s.GetRange(ctx, b.Start, b.End)
	if err != nil {
		return playoutID, fmt.Errorf("failed to get range: %w", err)
	}
	if len(playouts) != 0 {
		return playoutID, errors.New("time already scheduled: %w")
	}
	err = s.db.GetContext(ctx, &playoutID, `
		INSERT INTO schedule_playouts
		(channel_id, programme_id, ingest_url, ingest_type, scheduled_start,
		scheduled_end)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING playout_id;`, b.ChannelID, b.ProgrammeID, b.IngestURL, b.IngestType, b.Start, b.End)
	if err != nil {
		return playoutID, fmt.Errorf("failed to insert new playout")
	}
	return playoutID, nil
}

// UpdatePlayout changes a playout to the updated parameters
func (s *Scheduler) UpdatePlayout(ctx context.Context, b Playout) error {
	// TOOD: Validate this query. Are we allowing overlaps? FindIslands supports overlapping
	// Ideally we need to validate each field
	res, err := s.db.ExecContext(ctx, `
		UPDATE playout.schedule_playouts SET
			channel_id = $1,
			programme_id = $2,
			ingest_url = $3,
			ingest_type = $4,
			scheduled_start = $5,
			broadcast_start = $6,
			scheduled_end = $7,
			broadcast_end = $8
			vod_url = $9,
			dvr = $10,
			archive = $11
		WHERE playout_id = $12;`, b.ChannelID, b.ProgrammeID, b.IngestURL, b.IngestType,
		b.ScheduledStart, b.BroadcastStart, b.ScheduledEnd, b.BroadcastEnd,
		b.VODURL, b.DVR, b.Archive, b.PlayoutID)
	if err != nil {
		return fmt.Errorf("failed to update playout: %w", err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to determine rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return errors.New("playout doesn't exist")
	}
	return nil
}
