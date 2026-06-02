package repository

import "database/sql"

func RunMigrations(db *sql.DB) error {
	schemas := []string{
		`CREATE TABLE IF NOT EXISTS leagues (
			id     INTEGER PRIMARY KEY,
			name   TEXT NOT NULL,
			slug   TEXT NOT NULL UNIQUE,
			season INTEGER NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS games (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			created_at TEXT NOT NULL DEFAULT (datetime('now'))
		)`,
		`CREATE TABLE IF NOT EXISTS game_players (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			game_id     INTEGER NOT NULL REFERENCES games(id) ON DELETE CASCADE,
			player_name TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS game_events (
			id        INTEGER PRIMARY KEY AUTOINCREMENT,
			game_id   INTEGER NOT NULL REFERENCES games(id) ON DELETE CASCADE,
			event_id  INTEGER NOT NULL,
			league_id INTEGER NOT NULL,
			home_team TEXT NOT NULL,
			away_team TEXT NOT NULL,
			date      TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS game_assignments (
			id           INTEGER PRIMARY KEY AUTOINCREMENT,
			game_id      INTEGER NOT NULL REFERENCES games(id) ON DELETE CASCADE,
			player_name  TEXT NOT NULL,
			athlete_name TEXT NOT NULL,
			team_name    TEXT NOT NULL,
			event_id     INTEGER NOT NULL,
			position     TEXT,
			is_squad_pick INTEGER NOT NULL DEFAULT 0
		)`,
		`CREATE TABLE IF NOT EXISTS squad_cache (
			team_id     INTEGER NOT NULL,
			league_id   INTEGER NOT NULL,
			season      INTEGER NOT NULL,
			team_name   TEXT NOT NULL,
			players_json TEXT NOT NULL,
			cached_at   TEXT NOT NULL,
			PRIMARY KEY (team_id, season)
		)`,
		`CREATE TABLE IF NOT EXISTS lineup_cache (
			event_id   INTEGER PRIMARY KEY,
			home_json  TEXT NOT NULL,
			away_json  TEXT NOT NULL,
			fetched_at TEXT NOT NULL
		)`,
		`INSERT OR IGNORE INTO leagues (id, name, slug, season) VALUES
			(1, '1. Bundesliga',    'bl1', 2025),
			(2, '2. Bundesliga',    'bl2', 2025),
			(3, 'Champions League', 'cl',  2025),
			(4, 'WM 2026',          'wc',  2026)`,
	}

	for _, s := range schemas {
		if _, err := db.Exec(s); err != nil {
			return err
		}
	}
	return nil
}
