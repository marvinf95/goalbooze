package client

import (
	"testing"

	"github.com/marvinf95/goalbooze/internal/model"
)

// tagClient is a fake SportsAPIClient that stamps its tag onto returned data so
// tests can assert which upstream a request was routed to.
type tagClient struct{ tag string }

func (f tagClient) GetLeagues() ([]model.League, error) { return nil, nil }
func (f tagClient) GetEvents(leagueID, season int) ([]model.Event, error) {
	return []model.Event{{ID: leagueID, HomeTeam: f.tag}}, nil
}
func (f tagClient) GetTeams(leagueID, season int) ([]model.Team, error) {
	return []model.Team{{ID: leagueID, Name: f.tag}}, nil
}

func TestRoutingClient_RoutesByProvider(t *testing.T) {
	r := NewRoutingClient(tagClient{tag: "footballdata"}, tagClient{tag: "openligadb"})

	cases := map[int]string{
		1: "footballdata", // 1. Bundesliga
		3: "footballdata", // Champions League
		2: "openligadb",   // 2. Bundesliga
		5: "openligadb",   // DFB-Pokal
	}
	for leagueID, want := range cases {
		events, err := r.GetEvents(leagueID, 2026)
		if err != nil || len(events) != 1 {
			t.Fatalf("league %d: unexpected result %v / err %v", leagueID, events, err)
		}
		if events[0].HomeTeam != want {
			t.Errorf("league %d routed to %q, want %q", leagueID, events[0].HomeTeam, want)
		}
	}
}

func TestRoutingClient_GetLeaguesListsAll(t *testing.T) {
	r := NewRoutingClient(tagClient{}, tagClient{})
	leagues, err := r.GetLeagues()
	if err != nil {
		t.Fatalf("GetLeagues error: %v", err)
	}
	slugs := make(map[string]bool)
	for _, l := range leagues {
		slugs[l.Slug] = true
	}
	for _, expected := range []string{"BL1", "BL2", "CL", "DFB"} {
		if !slugs[expected] {
			t.Errorf("expected slug %q in league list", expected)
		}
	}
}
