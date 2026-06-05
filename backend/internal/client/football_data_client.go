package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/marvinf95/goalbooze/config"
	"github.com/marvinf95/goalbooze/internal/model"
)

const footballDataBaseURL = "https://api.football-data.org"

var leagueCodeMap = map[int]string{
	1: "BL1",
	2: "BL2",
	3: "CL",
	4: "WC",
}

type FootballDataClient struct {
	apiKey  string
	baseURL string
	http    *http.Client
}

func NewFootballDataClient(apiKey string) *FootballDataClient {
	return &FootballDataClient{
		apiKey:  apiKey,
		baseURL: footballDataBaseURL,
		http:    &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *FootballDataClient) GetLeagues() ([]model.League, error) {
	clubSeason := config.CurrentSeason()
	var leagues []model.League
	for id, code := range leagueCodeMap {
		leagues = append(leagues, model.League{
			ID:     id,
			Name:   leagueNameForCode(code),
			Slug:   code,
			Season: config.SeasonForLeague(id, clubSeason),
		})
	}
	return leagues, nil
}

func (c *FootballDataClient) GetEvents(leagueID, season int) ([]model.Event, error) {
	code, ok := leagueCodeMap[leagueID]
	if !ok {
		return nil, fmt.Errorf("unknown league ID: %d", leagueID)
	}

	url := fmt.Sprintf("%s/v4/competitions/%s/matches?season=%d", c.baseURL, code, season)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}
	if c.apiKey != "" {
		req.Header.Set("X-Auth-Token", c.apiKey)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("football-data request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("football-data rate limit exceeded, try again later")
	}
	if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusUnauthorized {
		if c.apiKey == "" {
			// No key configured — return dev fixtures so the app is testable without an API key
			return devFixtures(leagueID), nil
		}
		return nil, fmt.Errorf("football-data auth failed (HTTP %d) — check FOOTBALL_DATA_API_KEY", resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("football-data returned status %d", resp.StatusCode)
	}

	var result struct {
		Matches []struct {
			ID       int    `json:"id"`
			UTCDate  string `json:"utcDate"`
			Status   string `json:"status"`
			Matchday int    `json:"matchday"`
			HomeTeam struct {
				ID   int    `json:"id"`
				Name string `json:"name"`
			} `json:"homeTeam"`
			AwayTeam struct {
				ID   int    `json:"id"`
				Name string `json:"name"`
			} `json:"awayTeam"`
		} `json:"matches"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode matches: %w", err)
	}

	events := make([]model.Event, 0, len(result.Matches))
	for _, m := range result.Matches {
		t, _ := time.Parse(time.RFC3339, m.UTCDate)
		events = append(events, model.Event{
			ID:         m.ID,
			LeagueID:   leagueID,
			HomeTeam:   m.HomeTeam.Name,
			HomeTeamID: m.HomeTeam.ID,
			AwayTeam:   m.AwayTeam.Name,
			AwayTeamID: m.AwayTeam.ID,
			Date:       t,
			Status:     normalizeStatus(m.Status),
		})
	}

	return events, nil
}

func (c *FootballDataClient) GetTeams(leagueID, season int) ([]model.Team, error) {
	code, ok := leagueCodeMap[leagueID]
	if !ok {
		return nil, fmt.Errorf("unknown league ID: %d", leagueID)
	}

	url := fmt.Sprintf("%s/v4/competitions/%s/teams?season=%d", c.baseURL, code, season)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}
	if c.apiKey != "" {
		req.Header.Set("X-Auth-Token", c.apiKey)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("football-data teams request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("football-data rate limit exceeded")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("football-data teams returned status %d", resp.StatusCode)
	}

	var result struct {
		Teams []struct {
			ID        int    `json:"id"`
			Name      string `json:"name"`
			ShortName string `json:"shortName"`
			Squad     []struct {
				ID          int    `json:"id"`
				Name        string `json:"name"`
				Position    string `json:"position"`
				ShirtNumber int    `json:"shirtNumber"`
			} `json:"squad"`
		} `json:"teams"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode teams: %w", err)
	}

	teams := make([]model.Team, 0, len(result.Teams))
	for _, t := range result.Teams {
		squad := make([]model.Athlete, 0, len(t.Squad))
		for _, p := range t.Squad {
			squad = append(squad, model.Athlete{
				ID:       p.ID,
				Name:     p.Name,
				Position: normalizePosition(p.Position),
				Number:   p.ShirtNumber,
				Team:     t.Name,
			})
		}
		teams = append(teams, model.Team{
			ID:       t.ID,
			Name:     t.Name,
			LeagueID: leagueID,
			Squad:    squad,
		})
	}

	return teams, nil
}

func normalizeStatus(s string) string {
	switch s {
	case "FINISHED":
		return "finished"
	case "IN_PLAY", "PAUSED", "LIVE":
		return "live"
	case "SCHEDULED", "TIMED":
		return "scheduled"
	case "POSTPONED":
		return "postponed"
	case "CANCELLED":
		return "cancelled"
	default:
		return "scheduled"
	}
}

// devFixtures returns hardcoded test events used when no API key is configured.
func devFixtures(leagueID int) []model.Event {
	base := time.Now().Add(24 * time.Hour).Truncate(time.Hour)
	switch leagueID {
	case 1:
		return []model.Event{
			{ID: 900001, LeagueID: 1, HomeTeam: "FC Bayern München", HomeTeamID: 5, AwayTeam: "Borussia Dortmund", AwayTeamID: 4, Date: base, Status: "scheduled"},
			{ID: 900002, LeagueID: 1, HomeTeam: "Bayer 04 Leverkusen", HomeTeamID: 3, AwayTeam: "RB Leipzig", AwayTeamID: 721, Date: base.Add(3 * time.Hour), Status: "scheduled"},
			{ID: 900003, LeagueID: 1, HomeTeam: "VfB Stuttgart", HomeTeamID: 10, AwayTeam: "Eintracht Frankfurt", AwayTeamID: 19, Date: base.Add(6 * time.Hour), Status: "scheduled"},
		}
	case 2:
		return []model.Event{
			{ID: 900011, LeagueID: 2, HomeTeam: "Hamburger SV", HomeTeamID: 1, AwayTeam: "1. FC Köln", AwayTeamID: 1, Date: base, Status: "scheduled"},
		}
	case 3:
		return []model.Event{
			{ID: 900021, LeagueID: 3, HomeTeam: "FC Bayern München", HomeTeamID: 5, AwayTeam: "Real Madrid CF", AwayTeamID: 86, Date: base, Status: "scheduled"},
		}
	case 4:
		// World Cup 2026 fixtures. Team IDs match wc2026_squads.json so the
		// squad-based fallback resolves correctly in mock/dev mode.
		return []model.Event{
			{ID: 900401, LeagueID: 4, HomeTeam: "Germany", HomeTeamID: 759, AwayTeam: "Brazil", AwayTeamID: 764, Date: base, Status: "scheduled"},
			{ID: 900402, LeagueID: 4, HomeTeam: "France", HomeTeamID: 773, AwayTeam: "Argentina", AwayTeamID: 762, Date: base.Add(3 * time.Hour), Status: "scheduled"},
			{ID: 900403, LeagueID: 4, HomeTeam: "Spain", HomeTeamID: 760, AwayTeam: "England", AwayTeamID: 770, Date: base.Add(6 * time.Hour), Status: "scheduled"},
		}
	}
	return nil
}

func normalizePosition(p string) string {
	switch p {
	case "Goalkeeper":
		return "GK"
	case "Defence":
		return "DEF"
	case "Midfield":
		return "MF"
	case "Offence":
		return "FW"
	default:
		return p
	}
}

func leagueNameForCode(code string) string {
	switch code {
	case "BL1":
		return "1. Bundesliga"
	case "BL2":
		return "2. Bundesliga"
	case "CL":
		return "Champions League"
	case "WC":
		return "WM 2026"
	default:
		return code
	}
}
