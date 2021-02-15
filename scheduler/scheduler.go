// Package scheduler handles scheduling playouts
// to the internal schedule, executing them to
// a player.
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
	"github.com/ystv/playout/playout"
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
	po   *playout.Playouter
	prog *programming.Programmer
	play *vt.Player
	log  *log.Logger
}

type (
	// Schedule handles assigning jobs to the player
	Schedule interface {
		MainLoop(ctx context.Context) error
		Schedule(ctx context.Context, b playout.Playout) error
		Delete(ctx context.Context, playoutID int) error
		Reload(ctx context.Context) error
	}
	// Health handles ensuring the schedule is in a healthy state such as no gaps in playout
	// and ingest_url's have data when required
	Health interface {
		FindIslands(ctx context.Context, channelID int) ([]Island, error)
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
	playouts, err := s.po.GetAmount(ctx, s.queueSize)
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
func (s *Scheduler) Schedule(ctx context.Context, b playout.Playout) error {
	_, err := s.sch.StartAt(b.ScheduledStart).SetTag([]string{fmt.Sprint(b.PlayoutID)}).Do(s.ExecEvent, b)
	if err != nil {
		return fmt.Errorf("failed to schedule event \"%d\": %w", b.PlayoutID, err)
	}
	return nil
}

// ExecEvent trigger a Playout to be played out
func (s *Scheduler) ExecEvent(ctx context.Context, po playout.Playout) error {
	po.BroadcastStart = time.Now()
	prog, err := s.prog.Get(ctx, po.ProgrammeID)
	if err != nil {
		return fmt.Errorf("failed to get programme: %w", err)
	}
	videos := []string{}
	for _, video := range prog.Videos {
		videos = append(videos, video.URL)
	}
	c := player.Config{
		DstURL:    po.IngestURL,
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

// Delete will remove an item from the schedule from the DB and in-memory store
func (s *Scheduler) Delete(ctx context.Context, playoutID int) error {
	err := s.po.Delete(ctx, playoutID)
	err = s.deleteCron(ctx, playoutID)
	if err != nil {
		return fmt.Errorf("failed to delete playout: %w", err)
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
