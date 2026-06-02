package service

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"github.com/marvinf95/goalbooze/internal/model"
	"github.com/marvinf95/goalbooze/internal/repository"
)

func newTestRepos(t *testing.T) (*repository.LineupCacheRepository, *repository.SquadRepository) {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	db.SetMaxOpenConns(1)
	if err := repository.RunMigrations(db); err != nil {
		t.Fatalf("migrations: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return repository.NewLineupCacheRepository(db), repository.NewSquadRepository(db)
}

func eleven(team string) []model.Athlete {
	out := make([]model.Athlete, 11)
	for i := range out {
		out[i] = model.Athlete{ID: i, Name: team + " P", Position: "CM", Number: i + 1, Team: team}
	}
	return out
}

// fakeProvider returns a fixed result or an error, and records whether it ran.
type fakeProvider struct {
	name   string
	home   []model.Athlete
	away   []model.Athlete
	err    error
	called *bool
}

func (f fakeProvider) Name() string { return f.name }
func (f fakeProvider) FetchLineup(model.Event) ([]model.Athlete, []model.Athlete, error) {
	if f.called != nil {
		*f.called = true
	}
	return f.home, f.away, f.err
}

func testEvent() model.Event {
	return model.Event{ID: 5001, HomeTeam: "Germany", AwayTeam: "Brazil", Date: time.Now()}
}

func TestGetLineup_FirstProviderWins(t *testing.T) {
	cache, squads := newTestRepos(t)
	secondCalled := false
	svc := NewAILineupService([]LineupProvider{
		fakeProvider{name: "first", home: eleven("Germany"), away: eleven("Brazil")},
		fakeProvider{name: "second", home: eleven("X"), away: eleven("Y"), called: &secondCalled},
	}, cache, squads)

	home, away, squadPick, err := svc.GetLineup(testEvent(), 1, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if squadPick {
		t.Error("expected is_squad_pick=false for a provider hit")
	}
	if len(home) != 11 || len(away) != 11 {
		t.Fatalf("expected 11+11, got %d+%d", len(home), len(away))
	}
	if secondCalled {
		t.Error("second provider should not be called after first succeeds")
	}
}

func TestGetLineup_FallsThroughToNextProvider(t *testing.T) {
	cache, squads := newTestRepos(t)
	svc := NewAILineupService([]LineupProvider{
		fakeProvider{name: "broken", err: errors.New("boom")},
		fakeProvider{name: "good", home: eleven("Germany"), away: eleven("Brazil")},
	}, cache, squads)

	_, _, squadPick, err := svc.GetLineup(testEvent(), 1, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if squadPick {
		t.Error("expected is_squad_pick=false from the second provider")
	}
}

func TestGetLineup_SquadFallback(t *testing.T) {
	cache, squads := newTestRepos(t)
	// Seed two squads so the fallback can build lineups.
	if err := squads.SaveTeams([]model.Team{
		{ID: 759, Name: "Germany", LeagueID: 4, Squad: eleven("Germany")},
		{ID: 764, Name: "Brazil", LeagueID: 4, Squad: eleven("Brazil")},
	}, 4, 2026); err != nil {
		t.Fatalf("seed squads: %v", err)
	}

	svc := NewAILineupService([]LineupProvider{
		fakeProvider{name: "broken", err: errors.New("boom")},
	}, cache, squads)

	home, away, squadPick, err := svc.GetLineup(testEvent(), 759, 764)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !squadPick {
		t.Error("expected is_squad_pick=true for squad fallback")
	}
	if len(home) != 11 || len(away) != 11 {
		t.Fatalf("expected 11+11 from squad, got %d+%d", len(home), len(away))
	}
}

func TestGetLineup_CacheHitSkipsProviders(t *testing.T) {
	cache, squads := newTestRepos(t)
	ev := testEvent()
	if err := cache.Set(ev.ID, eleven("Germany"), eleven("Brazil")); err != nil {
		t.Fatalf("seed cache: %v", err)
	}
	called := false
	svc := NewAILineupService([]LineupProvider{
		fakeProvider{name: "should-not-run", home: eleven("X"), away: eleven("Y"), called: &called},
	}, cache, squads)

	_, _, squadPick, err := svc.GetLineup(ev, 1, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if squadPick {
		t.Error("cache hit should report is_squad_pick=false")
	}
	if called {
		t.Error("providers must not run on a cache hit")
	}
}

func TestMockProvider_GeneratesPlaceholdersWithoutFixture(t *testing.T) {
	m := NewMockLineupProvider("") // no file → generated placeholders
	home, away, err := m.FetchLineup(testEvent())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(home) != 11 || len(away) != 11 {
		t.Fatalf("expected 11+11 placeholders, got %d+%d", len(home), len(away))
	}
	if home[0].Team != "Germany" || away[0].Team != "Brazil" {
		t.Errorf("placeholder teams not set: %q / %q", home[0].Team, away[0].Team)
	}
}

func TestMockProvider_LoadsFixtureFile(t *testing.T) {
	m := NewMockLineupProvider("../../testdata/mock_lineups.json")
	ev := model.Event{ID: 900401, HomeTeam: "Germany", AwayTeam: "Brazil", Date: time.Now()}
	home, away, err := m.FetchLineup(ev)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(home) != 11 || len(away) != 11 {
		t.Fatalf("expected 11+11 from fixture, got %d+%d", len(home), len(away))
	}
	if home[0].Name != "Marc-André ter Stegen" {
		t.Errorf("unexpected first home player: %q", home[0].Name)
	}
}
