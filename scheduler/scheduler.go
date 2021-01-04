package scheduler

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/jmoiron/sqlx"
	"github.com/ystv/playout/player"
	"github.com/ystv/playout/player/vt"
	"github.com/ystv/playout/programming"
)

var _ Schedule = &Scheduler{}

// Scheduler wrapper around key dependencies
type Scheduler struct {
	queueSize int
	// dependencies
	db   *sqlx.DB
	sch  *gocron.Scheduler
	prog *programming.Store
	play *vt.Player
	log  *log.Logger
}

type (
	// Schedule handles assigning jobs to the player
	Schedule interface {
		MainLoop(ctx context.Context) error
		Schedule(ctx context.Context, b Block) error
		Delete(ctx context.Context, scheduleItemID int) error
		// Our gets are always arrays since it isn't channel specific
		GetCurrent(ctx context.Context) ([]Block, error)
		GetRange(ctx context.Context, start time.Time, end time.Time) ([]Block, error)
		GetAmount(ctx context.Context, amount int) ([]Block, error)
		Reload(ctx context.Context) error
	}
	// NewScheduleItem object required for adding to the schedule
	NewScheduleItem struct {
		ChannelID   int       `db:"channel_id" json:"channelID"`
		ProgrammeID int       `db:"programme_id" json:"programmeID"`
		IngestURL   string    `db:"ingest_url" json:"ingestURL"`
		Start       time.Time `db:"scheduled_start" json:"start"`
		End         time.Time `db:"scheduled_end" json:"end"`
	}
	// Block the individual block that is played out as part of a channel
	Block struct {
		BlockID     int `db:"block_id" json:"blockID"`
		ChannelID   int `db:"channel_id" json:"channelID"`
		ProgrammeID int `db:"programme_id" json:"programmeID"`
		// IngestURL where the playout should be outputted too
		IngestURL      string    `db:"ingest_url" json:"url"`
		ScheduledStart time.Time `db:"scheduled_start" json:"scheduledStart"`
		BroadcastStart time.Time `db:"broadcast_start" json:"broadcastStart"`
		ScheduledEnd   time.Time `db:"scheduled_end" json:"scheduledEnd"`
		BroadcastEnd   time.Time `db:"broadcast_end" json:"broadcastEnd"`
	}
)

// New creates a new scheduler instance
//
// Scheduler handles assigning jobs to the player
func New(db *sqlx.DB) (*Scheduler, error) {
	err := db.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to ping DB: %w", err)
	}
	p, err := vt.New("localhost:8080")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to vt: %w", err)
	}
	s := &Scheduler{
		db:   db,
		sch:  gocron.NewScheduler(time.Local),
		prog: programming.New(db),
		play: p,
	}
	return s, nil
}

// Reload will add a queueSize amount of jobs to the list
func (s *Scheduler) Reload(ctx context.Context) error {
	return nil
}

// Schedule will add a schedule item to the internal jon scheduler
// to be played out
func (s *Scheduler) Schedule(ctx context.Context, b Block) error {
	_, err := s.sch.StartAt(b.ScheduledStart).SetTag([]string{fmt.Sprint(b.BlockID)}).Do(s.ExecEvent, b)
	if err != nil {
		return fmt.Errorf("failed to schedule event \"%d\": %w", b.BlockID, err)
	}
	return nil
}

// ExecEvent trigger a Block to be played out
func (s *Scheduler) ExecEvent(ctx context.Context, i Block) error {
	i.BroadcastStart = time.Now()
	p, err := s.prog.Get(ctx, i.ProgrammeID)
	if err != nil {
		return fmt.Errorf("failed to get programme: %w", err)
	}
	videos := []string{}
	for _, video := range p.Videos {
		videos = append(videos, video.URL)
	}
	c := player.Config{
		DstURL:    i.IngestURL,
		Width:     1920,
		Height:    1080,
		Bitrate:   8000,
		VideoURLs: videos,
	}
	err = s.play.Play(ctx, c)
	if err != nil {
		return fmt.Errorf("failed to play block: %w", err)
	}
	return nil
}

// GetRange gets a range of items from a time range
func (s *Scheduler) GetRange(ctx context.Context, start, end time.Time) ([]Block, error) {
	items := []Block{}
	err := s.db.SelectContext(ctx, &items, `
	
	SELECT block_id, channel_id, programme_id, ingest_url,
		scheduled_start, broadcast_start, scheduled_end, broadcast_end
		
	FROM playout.schedule
	
	WHERE broadcast_start BETWEEN $1 AND $2;`, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to select get range: %w", err)
	}
	return items, nil
}

// GetAmount gets a certain amount of blocks from the current time
func (s *Scheduler) GetAmount(ctx context.Context, amount int) ([]Block, error) {
	blocks := []Block{}
	err := s.db.SelectContext(ctx, &blocks, `
	
	SELECT block_id, channel_id, programme_id, ingest_url,
		scheduled_start, broadcast_start, scheduled_end, broadcast_end
		
	FROM playout.schedule
	
	WHERE broadcast_start > $1
	LIMIT $2`, time.Now(), amount)
	if err != nil {
		return nil, fmt.Errorf("failed to select get amount: %w", err)
	}
	return blocks, nil
}

// GetCurrent gets the currently playing block
func (s *Scheduler) GetCurrent(ctx context.Context) ([]Block, error) {
	blocks := []Block{}
	err := s.db.SelectContext(ctx, &blocks, `
	
	SELECT block_id, channel_id, programme_id, ingest_url,
		scheduled_start, broadcast_start, scheduled_end, broadcast_end
		
	FROM playout.schedule
	
	WHERE $1 >= broadcast_start AND broadcast_end <= $1;`, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to select get range: %w", err)
	}
	return blocks, nil
}

// Delete will remove an item from the schedule from the DB and in-memory store
func (s *Scheduler) Delete(ctx context.Context, blockID int) error {
	err := s.deleteDB(ctx, blockID)
	err = s.deleteCron(ctx, blockID)
	if err != nil {
		return fmt.Errorf("failed to delete block: %w", err)
	}
	return nil
}

func (s *Scheduler) deleteDB(ctx context.Context, blockID int) error {
	res, err := s.db.ExecContext(ctx, `
	DELETE FROM playout.schedule
	WHERE block_id = $1`, blockID)
	if err != nil {
		return fmt.Errorf("failed to delete block from database: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to calculate rows affected: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("blockID doesn't exist: %w", err)
	}
	return nil
}

func (s *Scheduler) deleteCron(ctx context.Context, blockID int) error {
	// TODO: if there are multiple playout instances, we can delete the DB record,
	// but it will still exist in the other instances memory.
	err := s.sch.RemoveJobByTag(fmt.Sprint(blockID))
	if err != nil {
		return fmt.Errorf("failed to delete block from memory: %w", err)
	}
	return nil
}
