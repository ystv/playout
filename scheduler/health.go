package scheduler

import (
	"context"
	"fmt"
	"time"
)

// Island a group of continious videos between two times
type Island struct {
	PlayoutIDs  []int     `db:"playout_ids" json:"playoutIDs"`
	IslandStart time.Time `db:"island_start" json:"islandStart"`
	IslandEnd   time.Time `db:"island_end" json:"islandEnd"`
}

// FindIslands Validates DB schedules to see if islands have formed
//
// There should only be one island, more than
// one are caused by gaps in the schedule
func (s *Scheduler) FindIslands(ctx context.Context) ([]Island, error) {
	// We want to ensure that there will always be
	// something playing, so we will check that there
	// are playouts present.

	// * Check DB are there empty spaces, if so indicate where and duration
	// * Provide warnings for blank spaces but for spaces located within <24hr of playout, add content

	i := []Island{}
	err := s.db.SelectContext(ctx, &i, `
		SELECT
		array_agg(playout_id) AS playout_ids,
		MIN(scheduled_start) AS island_start,
		MAX(scheduled_end) AS island_end
		FROM
		(
			SELECT
				*,
				CASE WHEN groups.prev_item_sched_end >= scheduled_start THEN false ELSE true END AS island_start_indicator,
				SUM(CASE WHEN groups.prev_item_sched_end >= scheduled_start THEN 0 ELSE 1 END) OVER (ORDER BY groups.RN) AS island_id
			FROM
			(
				SELECT
					ROW_NUMBER() OVER(ORDER BY scheduled_start, scheduled_end) AS RN,
					playout_id,
					scheduled_start,
					scheduled_end,
					LAG(scheduled_end, 1) OVER (ORDER BY scheduled_start, scheduled_end) AS prev_item_sched_end,
					LAG(playout_id, 1) OVER (ORDER BY scheduled_start, scheduled_end) AS prev_playout_id
				FROM
					playout.schedule_playouts
				WHERE channel_id = $1
			) groups
		) islands
		GROUP BY
			island_id
		ORDER BY
			island_start;`, s.channel)
	if err != nil {
		return nil, fmt.Errorf("failed to find islands: %w", err)
	}
	return i, nil
}

// CheckSource will validate an IngestURL to see if
// there is content available.
func CheckSource() error {
	// This will probably be something VT dependent where it uses ffprobe?
	return nil
}
