package service

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/marvinf95/goalbooze/internal/model"
)

// mockPositions is a plausible 11-player formation used to generate placeholder
// lineups when no fixture is provided for an event.
var mockPositions = []string{"GK", "RB", "CB", "CB", "LB", "DM", "CM", "AM", "RW", "CF", "LW"}

// MockLineupProvider returns deterministic lineups without any network access.
// It looks up fixtures by event ID in a JSON file (event ID -> {home,away});
// when an event is missing it synthesizes 11+11 placeholder players from the
// team names. Enable it via LINEUP_MOCK=true to test the full flow token-free.
type MockLineupProvider struct {
	fixtures map[string]json.RawMessage
}

// NewMockLineupProvider loads optional fixtures from path. A missing file is fine
// — the provider then synthesizes placeholder lineups for every event.
func NewMockLineupProvider(path string) *MockLineupProvider {
	m := &MockLineupProvider{fixtures: map[string]json.RawMessage{}}
	if path == "" {
		return m
	}
	data, err := os.ReadFile(path)
	if err != nil {
		log.Printf("mock lineup fixtures not loaded (%s): %v — using generated placeholders", path, err)
		return m
	}
	if err := json.Unmarshal(data, &m.fixtures); err != nil {
		log.Printf("mock lineup fixtures invalid (%s): %v — using generated placeholders", path, err)
		return m
	}
	log.Printf("mock lineup fixtures loaded: %d events from %s", len(m.fixtures), path)
	return m
}

func (m *MockLineupProvider) Name() string { return "mock" }

func (m *MockLineupProvider) FetchLineup(event model.Event) ([]model.Athlete, []model.Athlete, error) {
	if raw, ok := m.fixtures[strconv.Itoa(event.ID)]; ok {
		return parseLineupJSON(string(raw), event)
	}
	return generatePlaceholderLineup(event)
}

// generatePlaceholderLineup builds a deterministic 11+11 lineup from team names.
func generatePlaceholderLineup(event model.Event) ([]model.Athlete, []model.Athlete, error) {
	home := make([]model.Athlete, 0, len(mockPositions))
	away := make([]model.Athlete, 0, len(mockPositions))
	for i, pos := range mockPositions {
		home = append(home, model.Athlete{
			ID:       event.ID*1000 + i,
			Name:     fmt.Sprintf("%s Spieler %d", event.HomeTeam, i+1),
			Position: pos, Number: i + 1, Team: event.HomeTeam,
		})
		away = append(away, model.Athlete{
			ID:       event.ID*1000 + 100 + i,
			Name:     fmt.Sprintf("%s Spieler %d", event.AwayTeam, i+1),
			Position: pos, Number: i + 1, Team: event.AwayTeam,
		})
	}
	return home, away, nil
}
