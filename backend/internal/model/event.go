package model

import "time"

type Event struct {
	ID         int       `json:"id"`
	LeagueID   int       `json:"league_id"`
	HomeTeam   string    `json:"home_team"`
	HomeTeamID int       `json:"home_team_id,omitempty"`
	AwayTeam   string    `json:"away_team"`
	AwayTeamID int       `json:"away_team_id,omitempty"`
	Date       time.Time `json:"date"`
	Status     string    `json:"status"`
}

type Athlete struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Number   int    `json:"number"`
	Position string `json:"position"`
	Team     string `json:"team"`
}
