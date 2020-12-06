package main

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
)

func main() {

}

type Scheduler struct {
	db sqlx.DB
}

func (s *Scheduler) GetRange(ctx context.Context, start, end time.Time) ([]ScheduleItem, error) {
	items := []ScheduleItem{}
	s.db.SelectContext(ctx, &items, `
	
	SELECT schedule_id, channel_id, programme_id, ingest_url,
		scheduled_start, broadcast_start, scheduled_end, broadcast_end
		
	FROM playout.schedule
	
	WHERE broadcast_time;`)
	return nil, nil
}

type (
	Schedule interface {
		New(ctx context.Context, s NewScheduleItem) error
		Delete(ctx context.Context, scheduleItemID int) error
		GetCurrent(ctx context.Context) (ScheduleItem, error)
		GetRange(ctx context.Context, start time.Time, end time.Time) ([]ScheduleItem, error)
		Recalibrate(ctx context.Context) error
	}

	IProgramme interface {
		New(ctx context.Context, p Programme) error
		Delete(ctx context.Context, programmeID int) error
		ExtendRuntime(ctx context.Context, programmeID int, duration int) error
	}
)

type (
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
		ProgrammeID    int       `db:"programme_id" json:"programmeID"`
		IngestURL      string    `db:"ingest_url" json:"url"`
		ScheduledStart time.Time `db:"scheduled_start" json:"scheduledStart"`
		BroadcastStart time.Time `db:"broadcast_start" json:"broadcastStart"`
		ScheduledEnd   time.Time `db:"scheduled_end" json:"scheduledEnd"`
		BroadcastEnd   time.Time `db:"broadcast_end" json:"broadcastEnd"`
	}
	Programme struct {
		ID          int    `db:"programme_id" json:"id"`
		Title       string `db:"title"`
		Description string `db:"description"`
		Thumbnail   string `db:"thumbnail"`
		Items       []Item `json:"items"`
	}
	Item struct {
		ID  int    `db:"programme_item_id" json:"id"`
		URL string `db:"url" json:"url"`
	}
)
