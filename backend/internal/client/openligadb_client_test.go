package client

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// dfbMatchesJSON mirrors the shape OpenLigaDB returns for getmatchdata, trimmed
// to the fields the client consumes. Includes one finished and one undrawn
// (team ID 0) fixture to exercise status mapping and the placeholder skip.
const dfbMatchesJSON = `[
  {
    "matchID": 81832,
    "matchDateTimeUTC": "2026-08-21T16:00:00Z",
    "matchIsFinished": false,
    "team1": {"teamId": 5712, "teamName": "SC St. Tönis"},
    "team2": {"teamId": 91, "teamName": "Eintracht Frankfurt"}
  },
  {
    "matchID": 81833,
    "matchDateTimeUTC": "2026-08-21T18:45:00Z",
    "matchIsFinished": true,
    "team1": {"teamId": 100, "teamName": "Hansa Rostock"},
    "team2": {"teamId": 10, "teamName": "VfB Stuttgart"}
  },
  {
    "matchID": 81834,
    "matchDateTimeUTC": "2026-09-01T18:00:00Z",
    "matchIsFinished": false,
    "team1": {"teamId": 0, "teamName": "Sieger Spiel A"},
    "team2": {"teamId": 0, "teamName": "Sieger Spiel B"}
  }
]`

const dfbTeamsJSON = `[
  {"teamId": 91, "teamName": "Eintracht Frankfurt"},
  {"teamId": 5712, "teamName": "SC St. Tönis"}
]`

func newTestOpenLigaClient(handler http.Handler) (*OpenLigaDBClient, *httptest.Server) {
	srv := httptest.NewServer(handler)
	return &OpenLigaDBClient{baseURL: srv.URL, http: srv.Client()}, srv
}

func TestOpenLigaDBClient_GetEventsMapsFixtures(t *testing.T) {
	c, srv := newTestOpenLigaClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/getmatchdata/dfb/2026" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Write([]byte(dfbMatchesJSON))
	}))
	defer srv.Close()

	events, err := c.GetEvents(5, 2026)
	if err != nil {
		t.Fatalf("GetEvents error: %v", err)
	}
	// The undrawn placeholder fixture (team IDs 0) must be skipped.
	if len(events) != 2 {
		t.Fatalf("expected 2 mapped events, got %d", len(events))
	}
	e := events[0]
	if e.ID != 81832 || e.HomeTeam != "SC St. Tönis" || e.AwayTeamID != 91 {
		t.Errorf("unexpected mapping: %+v", e)
	}
	if e.LeagueID != 5 || e.Status != "scheduled" {
		t.Errorf("expected league 5 / scheduled, got league %d / %s", e.LeagueID, e.Status)
	}
	if e.Date.IsZero() {
		t.Error("expected parsed match date")
	}
	if events[1].Status != "finished" {
		t.Errorf("expected finished status, got %s", events[1].Status)
	}
}

func TestOpenLigaDBClient_GetTeamsEmptySquads(t *testing.T) {
	c, srv := newTestOpenLigaClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/getavailableteams/dfb/2026" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Write([]byte(dfbTeamsJSON))
	}))
	defer srv.Close()

	teams, err := c.GetTeams(5, 2026)
	if err != nil {
		t.Fatalf("GetTeams error: %v", err)
	}
	if len(teams) != 2 {
		t.Fatalf("expected 2 teams, got %d", len(teams))
	}
	for _, tm := range teams {
		if tm.LeagueID != 5 {
			t.Errorf("team league = %d, want 5", tm.LeagueID)
		}
		if len(tm.Squad) != 0 {
			t.Errorf("OpenLigaDB teams carry no squad, got %d players", len(tm.Squad))
		}
	}
}

func TestOpenLigaDBClient_UnknownLeague(t *testing.T) {
	c := NewOpenLigaDBClient()
	if _, err := c.GetEvents(1, 2026); err == nil {
		t.Error("expected error for a non-OpenLigaDB league (football-data league 1)")
	}
}
