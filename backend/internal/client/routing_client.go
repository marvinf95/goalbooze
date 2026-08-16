package client

import (
	"github.com/marvinf95/goalbooze/config"
	"github.com/marvinf95/goalbooze/internal/model"
)

// RoutingClient dispatches each request to the upstream configured for the
// league (config.Provider): football-data.org for the leagues on its free tier,
// OpenLigaDB for the DFB-Pokal and 2. Bundesliga.
type RoutingClient struct {
	footballData SportsAPIClient
	openLiga     SportsAPIClient
}

func NewRoutingClient(footballData, openLiga SportsAPIClient) *RoutingClient {
	return &RoutingClient{footballData: footballData, openLiga: openLiga}
}

func (r *RoutingClient) clientFor(leagueID int) SportsAPIClient {
	if lc, ok := config.LeagueByID(leagueID); ok && lc.Provider == config.ProviderOpenLigaDB {
		return r.openLiga
	}
	return r.footballData
}

func (r *RoutingClient) GetLeagues() ([]model.League, error) {
	season := config.CurrentSeason()
	leagues := make([]model.League, 0, len(config.Leagues))
	for _, lc := range config.Leagues {
		leagues = append(leagues, model.League{
			ID:     lc.ID,
			Name:   lc.Name,
			Slug:   lc.Slug,
			Season: season,
		})
	}
	return leagues, nil
}

func (r *RoutingClient) GetEvents(leagueID, season int) ([]model.Event, error) {
	return r.clientFor(leagueID).GetEvents(leagueID, season)
}

func (r *RoutingClient) GetTeams(leagueID, season int) ([]model.Team, error) {
	return r.clientFor(leagueID).GetTeams(leagueID, season)
}
