package config

import (
	"testing"
	"time"
)

func TestSeasonForLeague(t *testing.T) {
	if got := SeasonForLeague(WorldCupLeagueID, 2025); got != WorldCupSeason {
		t.Errorf("World Cup season = %d, want %d", got, WorldCupSeason)
	}
	// Club leagues fall back to the supplied club season.
	if got := SeasonForLeague(1, 2025); got != 2025 {
		t.Errorf("club league season = %d, want 2025", got)
	}
	if got := SeasonForLeague(3, 2024); got != 2024 {
		t.Errorf("club league season = %d, want 2024", got)
	}
}

func TestCurrentSeason(t *testing.T) {
	now := time.Now()
	want := now.Year()
	if now.Month() < 7 {
		want = now.Year() - 1
	}
	if got := CurrentSeason(); got != want {
		t.Errorf("CurrentSeason() = %d, want %d", got, want)
	}
}

func TestLoad_DefaultsAndCORS(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGINS", "")
	cfg := Load()
	if len(cfg.CORSOrigins) != 1 || cfg.CORSOrigins[0] != "*" {
		t.Errorf("default CORS = %v, want [*]", cfg.CORSOrigins)
	}
	if cfg.Port != "8080" {
		t.Errorf("default port = %q, want 8080", cfg.Port)
	}
	if cfg.LineupMock {
		t.Error("LineupMock should default to false")
	}

	t.Setenv("CORS_ALLOWED_ORIGINS", "https://a.com,https://b.com")
	t.Setenv("LINEUP_MOCK", "true")
	cfg = Load()
	if len(cfg.CORSOrigins) != 2 {
		t.Errorf("CORS split = %v, want 2 entries", cfg.CORSOrigins)
	}
	if !cfg.LineupMock {
		t.Error("LineupMock should be true when LINEUP_MOCK=true")
	}
}
