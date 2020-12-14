package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/jmoiron/sqlx"
)

func main() {

}

type Scheduler struct {
	db   sqlx.DB
	sch  *gocron.Scheduler
	prog *ProgrammeRepo
}

type (
	Schedule interface {
		New(ctx context.Context, s NewScheduleItem) error
		Delete(ctx context.Context, scheduleItemID int) error
		GetCurrent(ctx context.Context) (ScheduleItem, error)
		GetRange(ctx context.Context, start time.Time, end time.Time) ([]ScheduleItem, error)
		Recalibrate(ctx context.Context) error
	}
	NewScheduleItem struct {
		ChannelID   int       `db:"channel_id" json:"id"`
		ProgrammeID int       `db:"programme_id" json:"programmeID"`
		IngestURL   string    `db:"ingest_url" json:"ingestURL"`
		Start       time.Time `db:"scheduled_start" json:"start"`
		End         time.Time `db:"scheduled_end" json:"end"`
	}
	ScheduleItem struct {
		ID             int       `db:"schedule_id" json:"id"`
		ChannelID      int       `db:"channel_id" json:"channelID"`
		ProgrammeID    int       `db:"programme" json:"programme"`
		IngestURL      string    `db:"ingest_url" json:"url"`
		ScheduledStart time.Time `db:"scheduled_start" json:"scheduledStart"`
		BroadcastStart time.Time `db:"broadcast_start" json:"broadcastStart"`
		ScheduledEnd   time.Time `db:"scheduled_end" json:"scheduledEnd"`
		BroadcastEnd   time.Time `db:"broadcast_end" json:"broadcastEnd"`
	}
)

// Schedule will add a schedule item to the internal jon scheduler
// to be played out
func (s *Scheduler) Schedule(i ScheduleItem) error {
	s.sch = gocron.NewScheduler(time.UTC)
	_, err := s.sch.StartAt(i.ScheduledStart).Do(s.ExecEvent, i)
	if err != nil {
		return fmt.Errorf("failed to schedule event \"%s\": %w", i.ID, err)
	}
	return nil
}

func (s *Scheduler) ExecEvent(ctx context.Context, i ScheduleItem) error {
	i.BroadcastStart = time.Now()
	p, err := s.prog.Get(ctx, i.ProgrammeID)
	if err != nil {
		return fmt.Errorf("failed to get programme: %w", err)
	}
	return nil
}

func (s *Scheduler) GetRange(ctx context.Context, start, end time.Time) ([]ScheduleItem, error) {
	items := []ScheduleItem{}
	err := s.db.SelectContext(ctx, &items, `
	
	SELECT schedule_id, channel_id, programme_id, ingest_url,
		scheduled_start, broadcast_start, scheduled_end, broadcast_end
		
	FROM playout.schedule
	
	WHERE broadcast_time BETWEEN $1 AND $2;`, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to select get rane: %w", err)
	}
	return items, nil
}
