package model

type Team struct {
	ID       int       `json:"id"`
	Name     string    `json:"name"`
	LeagueID int       `json:"league_id"`
	Squad    []Athlete `json:"squad,omitempty"`
}
