package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/marvinf95/goalbooze/config"
	"github.com/marvinf95/goalbooze/internal/model"
)

const openLigaDBBaseURL = "https://api.openligadb.de"

// OpenLigaDBClient sources fixtures and teams from OpenLigaDB (api.openligadb.de),
// a free, key-less German football API. It covers competitions the free
// football-data tier lacks — the DFB-Pokal (incl. the fully-drawn cup rounds)
// and the 2. Bundesliga. OpenLigaDB exposes no squad data, so GetTeams returns
// teams with empty rosters; starting elevens come from the AI lineup providers.
type OpenLigaDBClient struct {
	baseURL string
	http    *http.Client
}

func NewOpenLigaDBClient() *OpenLigaDBClient {
	return &OpenLigaDBClient{
		baseURL: openLigaDBBaseURL,
		http:    &http.Client{Timeout: 30 * time.Second},
	}
}

// shortcutFor resolves the OpenLigaDB league shortcut configured for an internal
// league ID.
func shortcutFor(leagueID int) (string, error) {
	lc, ok := config.LeagueByID(leagueID)
	if !ok || lc.OpenLigaShortcut == "" {
		return "", fmt.Errorf("no OpenLigaDB shortcut for league ID: %d", leagueID)
	}
	return lc.OpenLigaShortcut, nil
}

func (c *OpenLigaDBClient) GetLeagues() ([]model.League, error) {
	season := config.CurrentSeason()
	var leagues []model.League
	for _, lc := range config.Leagues {
		if lc.Provider != config.ProviderOpenLigaDB {
			continue
		}
		leagues = append(leagues, model.League{
			ID:     lc.ID,
			Name:   lc.Name,
			Slug:   lc.Slug,
			Season: season,
		})
	}
	return leagues, nil
}

type openLigaTeam struct {
	TeamID   int    `json:"teamId"`
	TeamName string `json:"teamName"`
}

type openLigaMatch struct {
	MatchID          int          `json:"matchID"`
	MatchDateTimeUTC string       `json:"matchDateTimeUTC"`
	MatchDateTime    string       `json:"matchDateTime"`
	MatchIsFinished  bool         `json:"matchIsFinished"`
	Team1            openLigaTeam `json:"team1"`
	Team2            openLigaTeam `json:"team2"`
}

func (c *OpenLigaDBClient) GetEvents(leagueID, season int) ([]model.Event, error) {
	shortcut, err := shortcutFor(leagueID)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/getmatchdata/%s/%d", c.baseURL, shortcut, season)
	resp, err := c.http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("openligadb request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("openligadb returned status %d for %s", resp.StatusCode, shortcut)
	}

	var matches []openLigaMatch
	if err := json.NewDecoder(resp.Body).Decode(&matches); err != nil {
		return nil, fmt.Errorf("failed to decode openligadb matches: %w", err)
	}

	events := make([]model.Event, 0, len(matches))
	for _, m := range matches {
		// Skip placeholder fixtures where a pairing is not yet drawn.
		if m.Team1.TeamID == 0 || m.Team2.TeamID == 0 {
			continue
		}
		t, _ := time.Parse(time.RFC3339, m.MatchDateTimeUTC)
		status := "scheduled"
		if m.MatchIsFinished {
			status = "finished"
		}
		events = append(events, model.Event{
			ID:         m.MatchID,
			LeagueID:   leagueID,
			HomeTeam:   m.Team1.TeamName,
			HomeTeamID: m.Team1.TeamID,
			AwayTeam:   m.Team2.TeamName,
			AwayTeamID: m.Team2.TeamID,
			Date:       t,
			Status:     status,
		})
	}
	return events, nil
}

func (c *OpenLigaDBClient) GetTeams(leagueID, season int) ([]model.Team, error) {
	shortcut, err := shortcutFor(leagueID)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/getavailableteams/%s/%d", c.baseURL, shortcut, season)
	resp, err := c.http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("openligadb teams request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("openligadb teams returned status %d for %s", resp.StatusCode, shortcut)
	}

	var teams []openLigaTeam
	if err := json.NewDecoder(resp.Body).Decode(&teams); err != nil {
		return nil, fmt.Errorf("failed to decode openligadb teams: %w", err)
	}

	// OpenLigaDB does not expose squads, so rosters are left empty; the AI lineup
	// providers supply starting elevens by team name at match time.
	result := make([]model.Team, 0, len(teams))
	for _, t := range teams {
		if t.TeamID == 0 {
			continue
		}
		result = append(result, model.Team{
			ID:       t.TeamID,
			Name:     t.TeamName,
			LeagueID: leagueID,
			Squad:    []model.Athlete{},
		})
	}
	return result, nil
}
