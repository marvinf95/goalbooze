package service

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/marvinf95/goalbooze/internal/model"
)

const (
	// PlayersPerTeam is a full starting eleven; a provider lineup must have at
	// least this many players per side to be accepted.
	PlayersPerTeam = 11

	// Synthetic athlete IDs are derived from the event ID so they stay unique
	// across events. athleteIDStride spaces each event's IDs; awayIDOffset keeps
	// the away team in a separate range from the home team.
	athleteIDStride = 1000
	awayIDOffset    = 100
)

// athleteID builds a stable synthetic ID for the index-th athlete of an event.
func athleteID(eventID, index int, away bool) int {
	id := eventID*athleteIDStride + index
	if away {
		id += awayIDOffset
	}
	return id
}

// LineupProvider fetches the official starting lineup for an event from some
// external source (an AI with web search, a mock, etc.). Providers are tried in
// order by AILineupService; the first one returning a full 11+11 lineup wins.
type LineupProvider interface {
	// Name identifies the provider in logs (e.g. "gemini", "claude", "mock").
	Name() string
	// FetchLineup returns the home and away starting elevens for the event.
	FetchLineup(event model.Event) (home, away []model.Athlete, err error)
}

// lineupSearchPrompt is the shared natural-language prompt used to ask a
// web-search-capable model for a match lineup. Kept identical across providers
// so results are comparable.
func lineupSearchPrompt(event model.Event) string {
	dateStr := event.Date.Format("02.01.2006")
	return fmt.Sprintf(
		"Search for the official starting lineup (Startaufstellung) for the football match: "+
			"%s vs %s on %s. "+
			"List all 11 starting players for each team with their jersey number and position.",
		event.HomeTeam, event.AwayTeam, dateStr,
	)
}

// lineupFormatPrompt asks a model (without web search) to convert a prose lineup
// description into the strict JSON shape parseLineupJSON expects.
func lineupFormatPrompt(lineupText, homeTeam, awayTeam string) string {
	return fmt.Sprintf(
		"Convert the following lineup information to JSON. "+
			"Return ONLY the JSON object, no explanation, no markdown:\n\n"+
			"%s\n\n"+
			`{"home":[{"name":"Name","position":"GK","number":1}],"away":[{"name":"Name","position":"GK","number":1}]}`+
			"\nHome team: %s  Away team: %s  "+
			"Use positions: GK, CB, RB, LB, CM, DM, RM, LM, AM, RW, LW, CF. "+
			"Exactly 11 players per team.",
		lineupText, homeTeam, awayTeam,
	)
}

// parseLineupJSON turns the model JSON ({"home":[...],"away":[...]}) into
// athlete slices, assigning stable IDs and team names from the event. Requires
// at least 11 players per side.
func parseLineupJSON(jsonStr string, event model.Event) ([]model.Athlete, []model.Athlete, error) {
	var data struct {
		Home []struct {
			Name     string `json:"name"`
			Position string `json:"position"`
			Number   int    `json:"number"`
		} `json:"home"`
		Away []struct {
			Name     string `json:"name"`
			Position string `json:"position"`
			Number   int    `json:"number"`
		} `json:"away"`
	}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return nil, nil, fmt.Errorf("JSON parse failed: %w", err)
	}

	home := make([]model.Athlete, 0, len(data.Home))
	for i, p := range data.Home {
		home = append(home, model.Athlete{
			ID: athleteID(event.ID, i, false), Name: p.Name,
			Position: p.Position, Number: p.Number, Team: event.HomeTeam,
		})
	}
	away := make([]model.Athlete, 0, len(data.Away))
	for i, p := range data.Away {
		away = append(away, model.Athlete{
			ID: athleteID(event.ID, i, true), Name: p.Name,
			Position: p.Position, Number: p.Number, Team: event.AwayTeam,
		})
	}

	if len(home) < PlayersPerTeam || len(away) < PlayersPerTeam {
		return nil, nil, fmt.Errorf("incomplete lineup: home=%d away=%d", len(home), len(away))
	}
	return home, away, nil
}

// extractJSON returns the first balanced {...} object found in text, or "".
func extractJSON(text string) string {
	start := strings.Index(text, "{")
	if start < 0 {
		return ""
	}
	depth := 0
	for i := start; i < len(text); i++ {
		switch text[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return text[start : i+1]
			}
		}
	}
	return ""
}
