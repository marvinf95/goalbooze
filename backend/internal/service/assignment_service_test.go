package service

import (
	"testing"

	"github.com/marvinf95/goalbooze/internal/model"
)

func TestAssign_EmptyPlayers(t *testing.T) {
	svc := NewAssignmentService()
	assignments := svc.Assign(nil, nil, nil)

	if len(assignments) != 0 {
		t.Errorf("expected 0 assignments, got %d", len(assignments))
	}
}

func TestAssign_SingleEventTwoPlayers(t *testing.T) {
	svc := NewAssignmentService()

	players := []model.Player{
		{Name: "Alice"},
		{Name: "Bob"},
	}

	events := []model.Event{
		{ID: 1, HomeTeam: "FCB", AwayTeam: "BVB"},
	}

	lineupMap := map[int]LineupPair{
		1: {
			Home: []model.Athlete{
				{ID: 1, Name: "Kane", Team: "FCB", Position: "Forward"},
				{ID: 2, Name: "Müller", Team: "FCB", Position: "Midfielder"},
			},
			Away: []model.Athlete{
				{ID: 3, Name: "Brandt", Team: "BVB", Position: "Midfielder"},
				{ID: 4, Name: "Adeyemi", Team: "BVB", Position: "Forward"},
			},
		},
	}

	assignments := svc.Assign(players, events, lineupMap)

	if len(assignments) != 4 {
		t.Errorf("expected 4 assignments (2 players × 2 teams), got %d", len(assignments))
	}

	for _, a := range assignments {
		if a.EventID != 1 {
			t.Errorf("expected EventID 1, got %d for assignment %s", a.EventID, a.AthleteName)
		}
	}

	seen := make(map[string]bool)
	for _, a := range assignments {
		key := a.AthleteName
		if seen[key] {
			t.Errorf("athlete %s was assigned twice", a.AthleteName)
		}
		seen[key] = true
	}
}

func TestAssign_MultipleEvents(t *testing.T) {
	svc := NewAssignmentService()

	players := []model.Player{
		{Name: "Alice"},
		{Name: "Bob"},
		{Name: "Charlie"},
	}

	events := []model.Event{
		{ID: 1},
		{ID: 2},
		{ID: 3},
	}

	lineupMap := map[int]LineupPair{}
	for _, e := range events {
		homeAthletes := make([]model.Athlete, 11)
		awayAthletes := make([]model.Athlete, 11)
		for i := range 11 {
			homeAthletes[i] = model.Athlete{ID: i + e.ID*100, Name: "H_Athlete", Team: "Home"}
			awayAthletes[i] = model.Athlete{ID: i + e.ID*1000, Name: "A_Athlete", Team: "Away"}
		}
		lineupMap[e.ID] = LineupPair{Home: homeAthletes, Away: awayAthletes}
	}

	assignments := svc.Assign(players, events, lineupMap)

	expectedCount := len(players) * len(events) * 2
	if len(assignments) != expectedCount {
		t.Errorf("expected %d assignments (3 players × 3 events × 2 teams), got %d",
			expectedCount, len(assignments))
	}

	perPlayer := make(map[string]int)
	for _, a := range assignments {
		perPlayer[a.PlayerName]++
	}
	for _, p := range players {
		count := perPlayer[p.Name]
		expected := len(events) * 2
		if count != expected {
			t.Errorf("player %s got %d assignments, expected %d", p.Name, count, expected)
		}
	}
}

func TestAssign_MorePlayersThanAthletes(t *testing.T) {
	svc := NewAssignmentService()

	players := []model.Player{
		{Name: "P1"}, {Name: "P2"}, {Name: "P3"},
		{Name: "P4"}, {Name: "P5"}, {Name: "P6"},
		{Name: "P7"}, {Name: "P8"}, {Name: "P9"},
		{Name: "P10"},
	}

	events := []model.Event{{ID: 1}}

	lineupMap := map[int]LineupPair{
		1: {
			Home: []model.Athlete{
				{ID: 1, Name: "H1", Team: "Home"},
				{ID: 2, Name: "H2", Team: "Home"},
				{ID: 3, Name: "H3", Team: "Home"},
			},
			Away: []model.Athlete{
				{ID: 4, Name: "A1", Team: "Away"},
				{ID: 5, Name: "A2", Team: "Away"},
				{ID: 6, Name: "A3", Team: "Away"},
			},
		},
	}

	assignments := svc.Assign(players, events, lineupMap)

	if len(assignments) != 6 {
		t.Errorf("expected 6 assignments (3 home + 3 away, capped), got %d", len(assignments))
	}
}

func TestShuffleCopy_PreservesLength(t *testing.T) {
	original := []model.Athlete{
		{ID: 1, Name: "A"},
		{ID: 2, Name: "B"},
		{ID: 3, Name: "C"},
	}

	shuffled := shuffleCopy(original, nil)

	if len(shuffled) != len(original) {
		t.Errorf("expected length %d, got %d", len(original), len(shuffled))
	}

	for _, a := range original {
		found := false
		for _, b := range shuffled {
			if a.ID == b.ID {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("athlete %s (ID %d) missing from shuffled copy", a.Name, a.ID)
		}
	}

	if &shuffled == &original {
		t.Error("shuffleCopy should return a different slice")
	}
}
