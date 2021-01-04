package programming

import (
	"context"
	"errors"
	"fmt"

	"github.com/ystv/playout/utils"

	"github.com/jmoiron/sqlx"
)

// Store encapsulates our dependency
type Store struct {
	db *sqlx.DB
}

var _ ProgrammeStore = &Store{}

// New will create a new programme
//
// This requires at least one video in the Videos slice
func (r *Store) New(ctx context.Context, p Programme) error {
	err := utils.Transact(r.db, func(tx *sqlx.Tx) error {
		programmeID := 0
		err := tx.QueryRowContext(ctx, `
		INSERT INTO playout.programmes(title, description, thumbnail)
		VALUES ($1, $2, $3)
		RETURNING programme_id;`,
			p.Title, p.Description, p.Thumbnail).Scan(&programmeID)
		if err != nil {
			return fmt.Errorf("failed to insert meta: %w", err)
		}
		if len(p.Videos) == 0 {
			return errors.New("no videos in playlist")
		}
		stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO playout.programme_videos(programme_id, url)
		VALUES ($1, $2);`)
		for _, video := range p.Videos {
			_, err = stmt.ExecContext(ctx, programmeID, video.URL)
			if err != nil {
				return fmt.Errorf("failed to link videos to programme: %w", err)
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to insert programme: %w", err)
	}
	return nil
}

// Get retrives a programme by it's programmeID
func (r *Store) Get(ctx context.Context, programmeID int) (*Programme, error) {
	p := Programme{}
	err := utils.Transact(r.db, func(tx *sqlx.Tx) error {
		err := tx.SelectContext(ctx, &p, `
		SELECT title, description, thumbnail
		FROM playout.programmes
		WHERE programme_id = $1;`, programmeID)
		if err != nil {
			return fmt.Errorf("failed to select meta: %w", err)
		}
		err = tx.SelectContext(ctx, &p.Videos, `
		SELECT programme_video_id, url
		FROM playout.programme_videos
		WHERE programme_id = $1
		;`, programmeID)
		if err != nil {
			return fmt.Errorf("failed to select videos: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve programme: %w", err)
	}
	return &p, nil
}

// Delete removes a programme by programmeID
func (r *Store) Delete(ctx context.Context, programmeID int) error {
	res, err := r.db.ExecContext(ctx, `
	DELETE FROM playout.programmes
	WHERE programme_id = $1`, programmeID)
	if err != nil {
		return fmt.Errorf("failed to delete programme: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to find rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("no programme with that ID")
	}
	return nil
}
