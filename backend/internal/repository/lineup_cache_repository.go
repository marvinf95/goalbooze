package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/marvinf95/goalbooze/internal/model"
)

type LineupCacheRepository struct {
	db *sql.DB
}

func NewLineupCacheRepository(db *sql.DB) *LineupCacheRepository {
	return &LineupCacheRepository{db: db}
}

func (r *LineupCacheRepository) Get(eventID int) (home []model.Athlete, away []model.Athlete, found bool, err error) {
	var homeJSON, awayJSON string
	err = r.db.QueryRow(
		`SELECT home_json, away_json FROM lineup_cache WHERE event_id = ?`,
		eventID,
	).Scan(&homeJSON, &awayJSON)
	if err == sql.ErrNoRows {
		return nil, nil, false, nil
	}
	if err != nil {
		return nil, nil, false, err
	}
	if err := json.Unmarshal([]byte(homeJSON), &home); err != nil {
		return nil, nil, false, fmt.Errorf("failed to parse home lineup: %w", err)
	}
	if err := json.Unmarshal([]byte(awayJSON), &away); err != nil {
		return nil, nil, false, fmt.Errorf("failed to parse away lineup: %w", err)
	}
	return home, away, true, nil
}

func (r *LineupCacheRepository) Set(eventID int, home, away []model.Athlete) error {
	homeJSON, err := json.Marshal(home)
	if err != nil {
		return err
	}
	awayJSON, err := json.Marshal(away)
	if err != nil {
		return err
	}
	now := time.Now().UTC().Format(time.RFC3339)
	_, err = r.db.Exec(
		`INSERT OR REPLACE INTO lineup_cache (event_id, home_json, away_json, fetched_at)
		 VALUES (?, ?, ?, ?)`,
		eventID, string(homeJSON), string(awayJSON), now,
	)
	return err
}
