package model

import "time"

type Game struct {
	ID          int          `json:"id"`
	CreatedAt   time.Time    `json:"created_at"`
	Players     []Player     `json:"players"`
	Events      []Event      `json:"events"`
	Assignments []Assignment `json:"assignments"`
}

type Assignment struct {
	PlayerName  string `json:"player_name"`
	AthleteName string `json:"athlete_name"`
	TeamName    string `json:"team_name"`
	EventID     int    `json:"event_id"`
	Position    string `json:"position"`
}
