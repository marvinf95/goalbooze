package service

import (
	"crypto/rand"
	"fmt"
	"log"
	"math/big"

	"github.com/marvinf95/goalbooze/internal/model"
	"github.com/marvinf95/goalbooze/internal/repository"
)

// AILineupService resolves a starting lineup for an event by trying, in order:
// the permanent cache, each configured LineupProvider (e.g. Gemini, then Claude),
// and finally a random squad-based fallback.
type AILineupService struct {
	providers   []LineupProvider
	lineupCache *repository.LineupCacheRepository
	squadRepo   *repository.SquadRepository
}

// NewAILineupService wires the service with an ordered provider chain. Providers
// are tried first-to-last; the first one returning a full 11+11 lineup wins.
func NewAILineupService(providers []LineupProvider, lineupCache *repository.LineupCacheRepository, squadRepo *repository.SquadRepository) *AILineupService {
	return &AILineupService{
		providers:   providers,
		lineupCache: lineupCache,
		squadRepo:   squadRepo,
	}
}

// GetLineup fetches the official starting lineup for an event.
// It tries the cache first, then each provider in order, and finally falls back
// to squad-based random selection.
func (s *AILineupService) GetLineup(event model.Event, homeTeamID, awayTeamID int) (home []model.Athlete, away []model.Athlete, isSquadPick bool, err error) {
	// 1. Check lineup cache
	if s.lineupCache != nil {
		cachedHome, cachedAway, found, cacheErr := s.lineupCache.Get(event.ID)
		if cacheErr == nil && found && len(cachedHome) > 0 && len(cachedAway) > 0 {
			log.Printf("lineup cache hit for event %d", event.ID)
			return cachedHome, cachedAway, false, nil
		}
	}

	// 2. Try each lineup provider in order
	for _, p := range s.providers {
		ph, pa, perr := p.FetchLineup(event)
		if perr == nil && len(ph) >= 11 && len(pa) >= 11 {
			log.Printf("lineup from %s for event %d", p.Name(), event.ID)
			if s.lineupCache != nil {
				if err := s.lineupCache.Set(event.ID, ph, pa); err != nil {
					log.Printf("failed to cache lineup for event %d: %v", event.ID, err)
				}
			}
			return ph, pa, false, nil
		}
		log.Printf("lineup provider %s failed for event %d: %v — trying next", p.Name(), event.ID, perr)
	}

	// 3. Squad-based fallback
	home, away, err = s.buildSquadLineup(homeTeamID, awayTeamID, event)
	if err != nil {
		return nil, nil, false, err
	}
	return home, away, true, nil
}

func (s *AILineupService) buildSquadLineup(homeTeamID, awayTeamID int, event model.Event) ([]model.Athlete, []model.Athlete, error) {
	homeSquad, err := s.squadRepo.GetSquad(homeTeamID)
	if err != nil {
		return nil, nil, fmt.Errorf("home squad not available: %w", err)
	}
	awaySquad, err := s.squadRepo.GetSquad(awayTeamID)
	if err != nil {
		return nil, nil, fmt.Errorf("away squad not available: %w", err)
	}

	if len(homeSquad) == 0 || len(awaySquad) == 0 {
		return nil, nil, fmt.Errorf("squad data not cached for teams %d/%d — fetch squads first", homeTeamID, awayTeamID)
	}

	home := pickEleven(homeSquad, event.HomeTeam)
	away := pickEleven(awaySquad, event.AwayTeam)
	return home, away, nil
}

func pickEleven(squad []model.Athlete, teamName string) []model.Athlete {
	cp := make([]model.Athlete, len(squad))
	copy(cp, squad)
	for i := range cp {
		cp[i].Team = teamName
	}
	// Fisher-Yates shuffle using crypto/rand for unbiased randomness
	for i := len(cp) - 1; i > 0; i-- {
		jBig, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			break
		}
		j := int(jBig.Int64())
		cp[i], cp[j] = cp[j], cp[i]
	}
	if len(cp) > 11 {
		cp = cp[:11]
	}
	return cp
}
