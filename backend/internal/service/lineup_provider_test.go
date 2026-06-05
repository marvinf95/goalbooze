package service

import (
	"testing"
	"time"

	"github.com/marvinf95/goalbooze/internal/model"
)

func TestExtractJSON(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"plain", `{"a":1}`, `{"a":1}`},
		{"with prefix/suffix", "here you go: {\"a\":1} thanks", `{"a":1}`},
		{"nested", `text {"a":{"b":2},"c":3} end`, `{"a":{"b":2},"c":3}`},
		{"no brace", "no json here", ""},
		{"empty", "", ""},
		{"unbalanced", `{"a":1`, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := extractJSON(c.in); got != c.want {
				t.Errorf("extractJSON(%q) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}

func lineupJSON(homeN, awayN int) string {
	build := func(team string, n int) string {
		s := ""
		for i := 0; i < n; i++ {
			if i > 0 {
				s += ","
			}
			s += `{"name":"` + team + `","position":"CM","number":1}`
		}
		return s
	}
	return `{"home":[` + build("H", homeN) + `],"away":[` + build("A", awayN) + `]}`
}

func TestParseLineupJSON_Valid(t *testing.T) {
	event := model.Event{ID: 42, HomeTeam: "Rot", AwayTeam: "Blau", Date: time.Now()}
	home, away, err := parseLineupJSON(lineupJSON(11, 11), event)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(home) != 11 || len(away) != 11 {
		t.Fatalf("expected 11+11, got %d+%d", len(home), len(away))
	}
	if home[0].Team != "Rot" || away[0].Team != "Blau" {
		t.Errorf("team names not assigned: %q / %q", home[0].Team, away[0].Team)
	}
	// IDs derived from event id, away offset keeps the sides apart.
	if home[0].ID != athleteID(42, 0, false) || away[0].ID != athleteID(42, 0, true) {
		t.Errorf("unexpected athlete IDs: home=%d away=%d", home[0].ID, away[0].ID)
	}
	if away[0].ID == home[0].ID {
		t.Error("home and away athlete IDs must differ")
	}
}

func TestParseLineupJSON_Incomplete(t *testing.T) {
	event := model.Event{ID: 1, HomeTeam: "H", AwayTeam: "A", Date: time.Now()}
	if _, _, err := parseLineupJSON(lineupJSON(10, 11), event); err == nil {
		t.Error("expected error for fewer than 11 home players")
	}
	if _, _, err := parseLineupJSON(lineupJSON(11, 9), event); err == nil {
		t.Error("expected error for fewer than 11 away players")
	}
}

func TestParseLineupJSON_BadJSON(t *testing.T) {
	event := model.Event{ID: 1, HomeTeam: "H", AwayTeam: "A", Date: time.Now()}
	if _, _, err := parseLineupJSON(`not json`, event); err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestAthleteID(t *testing.T) {
	if got := athleteID(5, 3, false); got != 5*athleteIDStride+3 {
		t.Errorf("home id = %d", got)
	}
	if got := athleteID(5, 3, true); got != 5*athleteIDStride+awayIDOffset+3 {
		t.Errorf("away id = %d", got)
	}
}
