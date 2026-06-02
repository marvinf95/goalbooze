package repository

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/marvinf95/goalbooze/internal/model"
)

// seedTeam mirrors one entry of the bundled static squad JSON files.
type seedTeam struct {
	TeamID int    `json:"team_id"`
	Name   string `json:"name"`
	Squad  []struct {
		Name     string `json:"name"`
		Position string `json:"position"`
		Number   int    `json:"number"`
	} `json:"squad"`
}

// SeedSquadsFromFile loads a static squad JSON file and writes it into the squad
// cache under the given league/season. Used for competitions whose rosters are
// not reliably available from the free football-data tier (e.g. the World Cup).
// Idempotent thanks to INSERT OR REPLACE in SaveTeams.
func (r *SquadRepository) SeedSquadsFromFile(path string, leagueID, season int) (int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, fmt.Errorf("read squad seed %s: %w", path, err)
	}

	var seeds []seedTeam
	if err := json.Unmarshal(data, &seeds); err != nil {
		return 0, fmt.Errorf("parse squad seed %s: %w", path, err)
	}

	teams := make([]model.Team, 0, len(seeds))
	for _, s := range seeds {
		squad := make([]model.Athlete, 0, len(s.Squad))
		for i, p := range s.Squad {
			squad = append(squad, model.Athlete{
				ID:       s.TeamID*100 + i,
				Name:     p.Name,
				Position: p.Position,
				Number:   p.Number,
				Team:     s.Name,
			})
		}
		teams = append(teams, model.Team{
			ID:       s.TeamID,
			Name:     s.Name,
			LeagueID: leagueID,
			Squad:    squad,
		})
	}

	if err := r.SaveTeams(teams, leagueID, season); err != nil {
		return 0, fmt.Errorf("save seeded squads: %w", err)
	}
	return len(teams), nil
}
