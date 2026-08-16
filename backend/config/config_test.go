package config

import (
	"testing"
	"time"
)

func TestDFBPokalLeagueConfigured(t *testing.T) {
	var found *LeagueConfig
	for i := range Leagues {
		if Leagues[i].ID == DFBPokalLeagueID {
			found = &Leagues[i]
			break
		}
	}
	if found == nil {
		t.Fatalf("DFB Pokal (ID %d) not present in Leagues", DFBPokalLeagueID)
	}
	if found.Name != "DFB Pokal" || found.Slug != "DFB" {
		t.Errorf("DFB Pokal config = %+v, want name 'DFB Pokal' / slug 'DFB'", *found)
	}
	// The DFB-Pokal is sourced from OpenLigaDB, not football-data.
	if found.Provider != ProviderOpenLigaDB || found.OpenLigaShortcut != "dfb" {
		t.Errorf("DFB Pokal provider = %q/%q, want openligadb/dfb", found.Provider, found.OpenLigaShortcut)
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
