// Package playout provides an interace for managing and creating
// playouts. A playout is a video stream of content to be played on
// a channel.
package playout

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/ystv/playout/programming"
)

type (
	// Repo handles managing the video streams
	Repo interface {
		New(ctx context.Context, po NewPlayout) (int, error)
		Update(ctx context.Context, b Playout) error
		// Our gets are always arrays since it isn't channel specific
		GetCurrent(ctx context.Context) ([]Playout, error)
		GetRange(ctx context.Context, start time.Time, end time.Time) ([]Playout, error)
		GetAmount(ctx context.Context, amount int) ([]Playout, error)
	}
	// Playouter handles the videostreams
	Playouter struct {
		// prog programming.ProgrammeStore
		prog *programming.Programmer
		db   *sqlx.DB
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

var _ Repo = &Playouter{}

// New adds a playout to the schedule
func (p *Playouter) New(ctx context.Context, po NewPlayout) (int, error) {
	/*
		First check the validity of the playout,
		* Channel exists
		* Programme exists
		* Time isn't overlapping existing schedule
	*/
	playoutID := 0
	_, err := p.prog.Get(ctx, po.ProgrammeID)
	if err != nil {
		return playoutID, fmt.Errorf("failed to get programme: %w", err)
	}
	playouts, err := p.GetRange(ctx, po.Start, po.End)
	if err != nil {
		return playoutID, fmt.Errorf("failed to get range: %w", err)
	}
	if len(playouts) != 0 {
		return playoutID, errors.New("time already scheduled: %w")
	}
	err = p.db.GetContext(ctx, &playoutID, `
		INSERT INTO schedule_playouts
		(channel_id, programme_id, ingest_url, ingest_type, scheduled_start,
		scheduled_end)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING playout_id;`, po.ChannelID, po.ProgrammeID, po.IngestURL, po.IngestType, po.Start, po.End)
	if err != nil {
		return playoutID, fmt.Errorf("failed to insert new playout")
	}
	return playoutID, nil
}

// Update changes a playout to the updated parameters
func (p *Playouter) Update(ctx context.Context, b Playout) error {
	// TOOD: Validate this query. Are we allowing overlaps? FindIslands supports overlapping
	// Ideally we need to validate each field
	res, err := p.db.ExecContext(ctx, `
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

// GetRange gets a range of items from a time range
func (p *Playouter) GetRange(ctx context.Context, start, end time.Time) ([]Playout, error) {
	items := []Playout{}
	err := p.db.SelectContext(ctx, &items, `
	
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
func (p *Playouter) GetAmount(ctx context.Context, amount int) ([]Playout, error) {
	playouts := []Playout{}
	err := p.db.SelectContext(ctx, &playouts, `
	
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
func (p *Playouter) GetCurrent(ctx context.Context) ([]Playout, error) {
	playouts := []Playout{}
	err := p.db.SelectContext(ctx, &playouts, `
	
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

// Delete will remove a playout
func (p *Playouter) Delete(ctx context.Context, playoutID int) error {
	res, err := p.db.ExecContext(ctx, `
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
