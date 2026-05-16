package repository

import (
	"database/sql"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"github.com/marvinf95/goalbooze/internal/model"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory db: %v", err)
	}
	db.SetMaxOpenConns(1)

	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		t.Fatalf("failed to enable foreign keys: %v", err)
	}

	if err := RunMigrations(db); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	return db
}

func TestRunMigrations_ShouldCreateTables(t *testing.T) {
	db := setupTestDB(t)

	tables := []string{"leagues", "games", "game_players", "game_events", "game_assignments"}
	for _, table := range tables {
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&count)
		if err != nil {
			t.Fatalf("failed to check table %s: %v", table, err)
		}
		if count != 1 {
			t.Errorf("table %s should exist", table)
		}
	}
}

func TestRunMigrations_ShouldSeedLeagues(t *testing.T) {
	db := setupTestDB(t)

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM leagues").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count < 2 {
		t.Errorf("expected at least 2 leagues seeded, got %d", count)
	}

	var bl1Name string
	if err := db.QueryRow("SELECT name FROM leagues WHERE slug='bl1'").Scan(&bl1Name); err != nil {
		t.Fatal(err)
	}
	if bl1Name != "1. Bundesliga" {
		t.Errorf("expected '1. Bundesliga', got '%s'", bl1Name)
	}
}

func TestCreateGame(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGameRepository(db)

	game := &model.Game{
		CreatedAt: time.Now(),
		Players: []model.Player{
			{Name: "Alice"},
			{Name: "Bob"},
		},
		Events: []model.Event{
			{ID: 101, LeagueID: 78, HomeTeam: "FCB", AwayTeam: "BVB"},
		},
		Assignments: []model.Assignment{
			{PlayerName: "Alice", AthleteName: "Kane", TeamName: "FCB", EventID: 101, Position: "Forward"},
			{PlayerName: "Alice", AthleteName: "Brandt", TeamName: "BVB", EventID: 101, Position: "Midfielder"},
			{PlayerName: "Bob", AthleteName: "Müller", TeamName: "FCB", EventID: 101, Position: "Midfielder"},
			{PlayerName: "Bob", AthleteName: "Adeyemi", TeamName: "BVB", EventID: 101, Position: "Forward"},
		},
	}

	if err := repo.Create(game); err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	if game.ID == 0 {
		t.Error("expected game ID to be set after creation")
	}
}

func TestGetAllGames(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGameRepository(db)

	game := &model.Game{
		Players: []model.Player{
			{Name: "Alice"},
		},
		Events: []model.Event{
			{ID: 101, LeagueID: 78, HomeTeam: "FCB", AwayTeam: "BVB"},
		},
		Assignments: []model.Assignment{
			{PlayerName: "Alice", AthleteName: "Kane", TeamName: "FCB", EventID: 101, Position: "Forward"},
		},
	}
	_ = repo.Create(game)

	games, err := repo.GetAll()
	if err != nil {
		t.Fatalf("GetAll() failed: %v", err)
	}

	if len(games) < 1 {
		t.Error("expected at least 1 game")
	}

	found := games[0]
	if len(found.Players) != 1 || found.Players[0].Name != "Alice" {
		t.Errorf("expected player Alice, got %v", found.Players)
	}
	if len(found.Events) != 1 {
		t.Errorf("expected 1 event, got %d", len(found.Events))
	}
}

func TestGetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGameRepository(db)

	game := &model.Game{
		Players: []model.Player{
			{Name: "Charlie"},
		},
		Events: []model.Event{
			{ID: 202, LeagueID: 79, HomeTeam: "S04", AwayTeam: "HSV"},
		},
		Assignments: []model.Assignment{
			{PlayerName: "Charlie", AthleteName: "Terodde", TeamName: "S04", EventID: 202, Position: "Forward"},
		},
	}
	_ = repo.Create(game)

	retrieved, err := repo.GetByID(game.ID)
	if err != nil {
		t.Fatalf("GetByID() failed: %v", err)
	}
	if retrieved == nil {
		t.Fatal("game should exist")
	}

	if len(retrieved.Assignments) != 1 {
		t.Errorf("expected 1 assignment, got %d", len(retrieved.Assignments))
	}
	if retrieved.Assignments[0].PlayerName != "Charlie" {
		t.Errorf("expected player Charlie, got %s", retrieved.Assignments[0].PlayerName)
	}
}

func TestGetByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGameRepository(db)

	game, err := repo.GetByID(99999)
	if err != nil {
		t.Fatalf("GetByID() should not error for not found: %v", err)
	}
	if game != nil {
		t.Error("expected nil for non-existent game")
	}
}

func TestDeleteGame(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGameRepository(db)

	game := &model.Game{
		Players:   []model.Player{{Name: "DeleteMe"}},
		Events:    []model.Event{{ID: 303, LeagueID: 78, HomeTeam: "FCB", AwayTeam: "BVB"}},
		Assignments: []model.Assignment{
			{PlayerName: "DeleteMe", AthleteName: "Kane", TeamName: "FCB", EventID: 303, Position: "Forward"},
		},
	}
	_ = repo.Create(game)

	if err := repo.Delete(game.ID); err != nil {
		t.Fatalf("Delete() failed: %v", err)
	}

	retrieved, _ := repo.GetByID(game.ID)
	if retrieved != nil {
		t.Error("game should be deleted")
	}

	var count int
	db.QueryRow("SELECT COUNT(*) FROM game_assignments WHERE game_id = ?", game.ID).Scan(&count)
	if count != 0 {
		t.Error("assignments should be cascade-deleted")
	}

	db.QueryRow("SELECT COUNT(*) FROM game_players WHERE game_id = ?", game.ID).Scan(&count)
	if count != 0 {
		t.Error("players should be cascade-deleted")
	}
}
