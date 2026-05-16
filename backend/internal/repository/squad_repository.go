package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/marvinf95/goalbooze/internal/model"
)

type SquadRepository struct {
	db *sql.DB
}

func NewSquadRepository(db *sql.DB) *SquadRepository {
	return &SquadRepository{db: db}
}

func (r *SquadRepository) GetTeams(leagueID, season int) ([]model.Team, error) {
	rows, err := r.db.Query(
		`SELECT team_id, team_name, players_json FROM squad_cache
		 WHERE league_id = ? AND season = ?`,
		leagueID, season,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teams []model.Team
	for rows.Next() {
		var teamID int
		var teamName, playersJSON string
		if err := rows.Scan(&teamID, &teamName, &playersJSON); err != nil {
			return nil, err
		}
		var squad []model.Athlete
		if err := json.Unmarshal([]byte(playersJSON), &squad); err != nil {
			continue
		}
		teams = append(teams, model.Team{
			ID:       teamID,
			Name:     teamName,
			LeagueID: leagueID,
			Squad:    squad,
		})
	}
	return teams, rows.Err()
}

func (r *SquadRepository) GetSquad(teamID int) ([]model.Athlete, error) {
	season := currentCacheSeason()
	var playersJSON string
	err := r.db.QueryRow(
		`SELECT players_json FROM squad_cache WHERE team_id = ? AND season = ?`,
		teamID, season,
	).Scan(&playersJSON)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var squad []model.Athlete
	if err := json.Unmarshal([]byte(playersJSON), &squad); err != nil {
		return nil, fmt.Errorf("failed to parse squad JSON: %w", err)
	}
	return squad, nil
}

func (r *SquadRepository) SaveTeams(teams []model.Team, leagueID, season int) error {
	now := time.Now().UTC().Format(time.RFC3339)
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, t := range teams {
		playersJSON, err := json.Marshal(t.Squad)
		if err != nil {
			continue
		}
		_, err = tx.Exec(
			`INSERT OR REPLACE INTO squad_cache
			 (team_id, league_id, season, team_name, players_json, cached_at)
			 VALUES (?, ?, ?, ?, ?, ?)`,
			t.ID, leagueID, season, t.Name, string(playersJSON), now,
		)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *SquadRepository) IsCacheStale(leagueID, season int) (bool, error) {
	var cachedAt string
	err := r.db.QueryRow(
		`SELECT cached_at FROM squad_cache WHERE league_id = ? AND season = ? LIMIT 1`,
		leagueID, season,
	).Scan(&cachedAt)
	if err == sql.ErrNoRows {
		return true, nil
	}
	if err != nil {
		return true, err
	}
	t, err := time.Parse(time.RFC3339, cachedAt)
	if err != nil {
		return true, nil
	}
	return time.Since(t) > 24*time.Hour, nil
}

func currentCacheSeason() int {
	now := time.Now()
	if now.Month() < 7 {
		return now.Year() - 1
	}
	return now.Year()
}
