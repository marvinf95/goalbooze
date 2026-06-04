package service

import (
	"math/rand"
	"time"

	"github.com/marvinf95/goalbooze/internal/model"
)

type AssignmentService struct{}

func NewAssignmentService() *AssignmentService {
	return &AssignmentService{}
}

func (s *AssignmentService) Assign(players []model.Player, events []model.Event, lineupMap map[int]LineupPair) []model.Assignment {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	var assignments []model.Assignment

	for _, event := range events {
		pair, ok := lineupMap[event.ID]
		if !ok {
			continue
		}

		shuffledHome := shuffleCopy(pair.Home, rng)
		shuffledAway := shuffleCopy(pair.Away, rng)

		// Assign cyclically so every player gets one home and one away athlete,
		// even when a team has fewer athletes than players (manual games may have
		// as few as 1 per team). When a team has >= players (e.g. the 11-strong
		// football-data lineups), i%len == i, so behaviour is unchanged there.
		for i, player := range players {
			if len(shuffledHome) > 0 {
				a := shuffledHome[i%len(shuffledHome)]
				assignments = append(assignments, model.Assignment{
					PlayerName:  player.Name,
					AthleteName: a.Name,
					TeamName:    a.Team,
					EventID:     event.ID,
					Position:    a.Position,
				})
			}
			if len(shuffledAway) > 0 {
				a := shuffledAway[i%len(shuffledAway)]
				assignments = append(assignments, model.Assignment{
					PlayerName:  player.Name,
					AthleteName: a.Name,
					TeamName:    a.Team,
					EventID:     event.ID,
					Position:    a.Position,
				})
			}
		}
	}

	return assignments
}

type LineupPair struct {
	Home []model.Athlete
	Away []model.Athlete
}

func shuffleCopy(src []model.Athlete, rng *rand.Rand) []model.Athlete {
	dest := make([]model.Athlete, len(src))
	copy(dest, src)
	if rng == nil {
		rng = rand.New(rand.NewSource(time.Now().UnixNano()))
	}
	rng.Shuffle(len(dest), func(i, j int) {
		dest[i], dest[j] = dest[j], dest[i]
	})
	return dest
}
