package client

import "github.com/marvinf95/goalbooze/internal/model"

type SportsAPIClient interface {
	GetLeagues() ([]model.League, error)
	GetEvents(leagueID, season int) ([]model.Event, error)
	GetTeams(leagueID, season int) ([]model.Team, error)
}
