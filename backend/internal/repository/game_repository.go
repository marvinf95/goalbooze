package repository

import (
	"database/sql"
	"time"

	"github.com/marvinf95/goalbooze/internal/model"
)

type GameRepository struct {
	db *sql.DB
}

func NewGameRepository(db *sql.DB) *GameRepository {
	return &GameRepository{db: db}
}

func (r *GameRepository) Create(game *model.Game) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	result, err := tx.Exec(`INSERT INTO games (created_at) VALUES (?)`, game.CreatedAt.Format(time.RFC3339))
	if err != nil {
		return err
	}

	gameID, _ := result.LastInsertId()
	game.ID = int(gameID)

	for _, p := range game.Players {
		if _, err := tx.Exec(`INSERT INTO game_players (game_id, player_name) VALUES (?, ?)`, gameID, p.Name); err != nil {
			return err
		}
	}

	for _, e := range game.Events {
		if _, err := tx.Exec(`INSERT INTO game_events (game_id, event_id, league_id, home_team, away_team, date) VALUES (?, ?, ?, ?, ?, ?)`,
			gameID, e.ID, e.LeagueID, e.HomeTeam, e.AwayTeam, e.Date.Format(time.RFC3339)); err != nil {
			return err
		}
	}

	for _, a := range game.Assignments {
		if _, err := tx.Exec(`INSERT INTO game_assignments (game_id, player_name, athlete_name, team_name, event_id, position) VALUES (?, ?, ?, ?, ?, ?)`,
			gameID, a.PlayerName, a.AthleteName, a.TeamName, a.EventID, a.Position); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *GameRepository) GetAll() ([]model.Game, error) {
	rows, err := r.db.Query(`SELECT id, created_at FROM games ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}

	games := []model.Game{} // nicht-nil, damit leere Liste als [] statt null serialisiert wird
	for rows.Next() {
		var g model.Game
		var createdAt string
		if err := rows.Scan(&g.ID, &createdAt); err != nil {
			rows.Close()
			return nil, err
		}
		g.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		games = append(games, g)
	}
	rows.Close()

	for i := range games {
		players, err := r.getPlayers(games[i].ID)
		if err != nil {
			return nil, err
		}
		games[i].Players = players

		events, err := r.getEvents(games[i].ID)
		if err != nil {
			return nil, err
		}
		games[i].Events = events
	}
	return games, nil
}

func (r *GameRepository) GetByID(id int) (*model.Game, error) {
	var g model.Game
	var createdAt string

	err := r.db.QueryRow(`SELECT id, created_at FROM games WHERE id = ?`, id).Scan(&g.ID, &createdAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	g.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)

	players, err := r.getPlayers(g.ID)
	if err != nil {
		return nil, err
	}
	g.Players = players

	events, err := r.getEvents(g.ID)
	if err != nil {
		return nil, err
	}
	g.Events = events

	g.Assignments, err = r.getAssignments(g.ID)
	if err != nil {
		return nil, err
	}

	return &g, nil
}

func (r *GameRepository) Delete(id int) error {
	_, err := r.db.Exec(`DELETE FROM games WHERE id = ?`, id)
	return err
}

func (r *GameRepository) getPlayers(gameID int) ([]model.Player, error) {
	rows, err := r.db.Query(`SELECT player_name FROM game_players WHERE game_id = ?`, gameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var players []model.Player
	for rows.Next() {
		var p model.Player
		if err := rows.Scan(&p.Name); err != nil {
			return nil, err
		}
		players = append(players, p)
	}
	return players, nil
}

func (r *GameRepository) getEvents(gameID int) ([]model.Event, error) {
	rows, err := r.db.Query(`SELECT event_id, league_id, home_team, away_team, date FROM game_events WHERE game_id = ?`, gameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []model.Event
	for rows.Next() {
		var e model.Event
		var date string
		if err := rows.Scan(&e.ID, &e.LeagueID, &e.HomeTeam, &e.AwayTeam, &date); err != nil {
			return nil, err
		}
		e.Date, _ = time.Parse(time.RFC3339, date)
		events = append(events, e)
	}
	return events, nil
}

func (r *GameRepository) getAssignments(gameID int) ([]model.Assignment, error) {
	rows, err := r.db.Query(`SELECT player_name, athlete_name, team_name, event_id, position FROM game_assignments WHERE game_id = ?`, gameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var assignments []model.Assignment
	for rows.Next() {
		var a model.Assignment
		if err := rows.Scan(&a.PlayerName, &a.AthleteName, &a.TeamName, &a.EventID, &a.Position); err != nil {
			return nil, err
		}
		assignments = append(assignments, a)
	}
	return assignments, nil
}
