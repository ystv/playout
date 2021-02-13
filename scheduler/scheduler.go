package scheduler

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/jmoiron/sqlx"
	"github.com/ystv/playout/channel"
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
	ch   *channel.Channel
	sch  *gocron.Scheduler
	prog *programming.Programmer
	play *vt.Player
	log  *log.Logger
}

type (
	// Schedule handles assigning jobs to the player
	Schedule interface {
		MainLoop(ctx context.Context) error
		NewPlayout(ctx context.Context, b NewPlayout) (int, error)
		Schedule(ctx context.Context, b Playout) error
		Delete(ctx context.Context, playoutID int) error
		// Our gets are always arrays since it isn't channel specific
		GetCurrent(ctx context.Context) ([]Playout, error)
		GetRange(ctx context.Context, start time.Time, end time.Time) ([]Playout, error)
		GetAmount(ctx context.Context, amount int) ([]Playout, error)
		Reload(ctx context.Context) error
	}
	// Health handles ensuring the schedule is in a healthy state such as no gaps in playout
	// and ingest_url's have data when required
	Health interface {
		FindIslands(ctx context.Context, channelID int) ([]Island, error)
	}
	// NewPlayout object required for adding to the schedule
	//
	// Doesn't contain the historic broadcast date
	// in comparision to the main object
	NewPlayout struct {
		ChannelID   int       `db:"channel_id" json:"channelID"`
		ProgrammeID int       `db:"programme_id" json:"programmeID"`
		IngestURL   string    `db:"ingest_url" json:"ingestURL"`
		IngestType  string    `db:"ingest_type" json:"ingestType"`
		Start       time.Time `db:"scheduled_start" json:"start"`
		End         time.Time `db:"scheduled_end" json:"end"`
	}
	// Playout the individual video stream that is played out as part of a channel
	Playout struct {
		PlayoutID   int `db:"playout_id" json:"playoutID"`
		ChannelID   int `db:"channel_id" json:"channelID"`
		ProgrammeID int `db:"programme_id" json:"programmeID"`
		// IngestURL where the player should broadcast to, where it is then picked up
		// by either channel or piper
		IngestURL      string    `db:"ingest_url" json:"ingestURL"`
		IngestType     string    `db:"ingest_type" json:"ingestType"`
		ScheduledStart time.Time `db:"scheduled_start" json:"scheduledStart"`
		BroadcastStart time.Time `db:"broadcast_start" json:"broadcastStart"`
		ScheduledEnd   time.Time `db:"scheduled_end" json:"scheduledEnd"`
		BroadcastEnd   time.Time `db:"broadcast_end" json:"broadcastEnd"`
		VODURL         string    `db:"vod_url" json:"vodURL"`
		DVR            bool      `db:"dvr" json:"dvr"`
		Archive        bool      `db:"archive" json:"archive"`
	}
)

// New creates a new scheduler instance
//
// Scheduler handles assigning jobs to the player
func New(db *sqlx.DB, ch *channel.Channel) (*Scheduler, error) {
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
		ch:   ch,
		sch:  gocron.NewScheduler(time.Local),
		prog: programming.New(db),
		play: p,
	}
	return s, nil
}

// Reload will add a queueSize amount of playouts to the scheduler cache
func (s *Scheduler) Reload(ctx context.Context) error {
	// Get the playouts to be played next
	playouts, err := s.GetAmount(ctx, s.queueSize)
	if err != nil {
		return fmt.Errorf("Reload failed to get playouts: %w", err)
	}
	// Empty scheduler of old playouts
	s.sch.Clear()
	// Add new playouts
	for _, playout := range playouts {
		s.Schedule(ctx, playout)
	}
	return nil
}

// Schedule will add a schedule item to the internal jon scheduler
// to be played out
func (s *Scheduler) Schedule(ctx context.Context, b Playout) error {
	_, err := s.sch.StartAt(b.ScheduledStart).SetTag([]string{fmt.Sprint(b.PlayoutID)}).Do(s.ExecEvent, b)
	if err != nil {
		return fmt.Errorf("failed to schedule event \"%d\": %w", b.PlayoutID, err)
	}
	return nil
}

// ExecEvent trigger a Playout to be played out
func (s *Scheduler) ExecEvent(ctx context.Context, i Playout) error {
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
		return fmt.Errorf("failed to play playout: %w", err)
	}
	return nil
}

// GetRange gets a range of items from a time range
func (s *Scheduler) GetRange(ctx context.Context, start, end time.Time) ([]Playout, error) {
	items := []Playout{}
	err := s.db.SelectContext(ctx, &items, `
	
	SELECT playout_id, channel_id, programme_id, ingest_url,
		scheduled_start, broadcast_start, scheduled_end, broadcast_end
		
	FROM playout.schedule_playouts
	
	WHERE broadcast_start BETWEEN $1 AND $2;`, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to select get range: %w", err)
	}
	return items, nil
}

// GetAmount gets a certain amount of playouts from the current time
func (s *Scheduler) GetAmount(ctx context.Context, amount int) ([]Playout, error) {
	playouts := []Playout{}
	err := s.db.SelectContext(ctx, &playouts, `
	
	SELECT playout_id, channel_id, programme_id, ingest_url,
		scheduled_start, broadcast_start, scheduled_end, broadcast_end
		
	FROM playout.schedule_playouts
	
	WHERE broadcast_start > $1
	LIMIT $2;`, time.Now(), amount)
	if err != nil {
		return nil, fmt.Errorf("failed to select get amount: %w", err)
	}
	return playouts, nil
}

// GetCurrent gets the currently playing playout
func (s *Scheduler) GetCurrent(ctx context.Context) ([]Playout, error) {
	playouts := []Playout{}
	err := s.db.SelectContext(ctx, &playouts, `
	
	SELECT playout_id, channel_id, programme_id, ingest_url,
		scheduled_start, broadcast_start, scheduled_end, broadcast_end
		
	FROM playout.schedule_playouts
	
	WHERE $1 >= broadcast_start
	AND (broadcast_end <= $1 OR broadcast_end IS NULL)
	AND (scheduled_time <= $1) AND scheduled_time >= $1;`, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to select get range: %w", err)
	}
	return playouts, nil
}

// Delete will remove an item from the schedule from the DB and in-memory store
func (s *Scheduler) Delete(ctx context.Context, playoutID int) error {
	err := s.deleteDB(ctx, playoutID)
	err = s.deleteCron(ctx, playoutID)
	if err != nil {
		return fmt.Errorf("failed to delete playout: %w", err)
	}
	return nil
}

func (s *Scheduler) deleteDB(ctx context.Context, playoutID int) error {
	res, err := s.db.ExecContext(ctx, `
	DELETE FROM playout.schedule
	WHERE playout_id = $1`, playoutID)
	if err != nil {
		return fmt.Errorf("failed to delete playout from database: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to calculate rows affected: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("playoutID doesn't exist: %w", err)
	}
	return nil
}

func (s *Scheduler) deleteCron(ctx context.Context, playoutID int) error {
	// TODO: if there are multiple playout instances, we can delete the DB record,
	// but it will still exist in the other instances memory.
	err := s.sch.RemoveJobByTag(fmt.Sprint(playoutID))
	if err != nil {
		return fmt.Errorf("failed to delete playout from memory: %w", err)
	}
	return nil
}
