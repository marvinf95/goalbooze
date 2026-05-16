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

		for i, player := range players {
			if i < len(shuffledHome) {
				assignments = append(assignments, model.Assignment{
					PlayerName:  player.Name,
					AthleteName: shuffledHome[i].Name,
					TeamName:    shuffledHome[i].Team,
					EventID:     event.ID,
					Position:    shuffledHome[i].Position,
				})
			}
			if i < len(shuffledAway) {
				assignments = append(assignments, model.Assignment{
					PlayerName:  player.Name,
					AthleteName: shuffledAway[i].Name,
					TeamName:    shuffledAway[i].Team,
					EventID:     event.ID,
					Position:    shuffledAway[i].Position,
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
