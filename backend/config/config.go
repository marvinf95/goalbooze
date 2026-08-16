package config

import (
	"os"
	"strings"
	"time"
)

type Config struct {
	AnthropicAPIKey    string
	FootballDataAPIKey string
	GeminiAPIKey       string
	GeminiModel        string
	Port               string
	CORSOrigins        []string
	AnthropicModel     string
	LineupMock         bool
}

// Data providers a league's fixtures and teams can be sourced from.
const (
	ProviderFootballData = "footballdata"
	ProviderOpenLigaDB   = "openligadb"
)

type LeagueConfig struct {
	ID   int
	Name string
	// Slug is the public code returned to the frontend (e.g. "BL1", "DFB").
	Slug string
	// Provider selects the upstream data source (see the Provider* constants).
	Provider string
	// FootballDataCode is the competition code for football-data.org
	// (Provider == ProviderFootballData).
	FootballDataCode string
	// OpenLigaShortcut is the league shortcut for OpenLigaDB
	// (Provider == ProviderOpenLigaDB), e.g. "bl2", "dfb".
	OpenLigaShortcut string
	Sport            string
}

// Leagues is the canonical competition list. The DFB-Pokal and 2. Bundesliga are
// not on the free football-data tier, so they are sourced from OpenLigaDB (free,
// no key), which also covers the fully-drawn cup rounds.
var Leagues = []LeagueConfig{
	{ID: 1, Name: "1. Bundesliga", Slug: "BL1", Provider: ProviderFootballData, FootballDataCode: "BL1", Sport: "football"},
	{ID: 2, Name: "2. Bundesliga", Slug: "BL2", Provider: ProviderOpenLigaDB, OpenLigaShortcut: "bl2", Sport: "football"},
	{ID: 3, Name: "Champions League", Slug: "CL", Provider: ProviderFootballData, FootballDataCode: "CL", Sport: "football"},
	{ID: 5, Name: "DFB Pokal", Slug: "DFB", Provider: ProviderOpenLigaDB, OpenLigaShortcut: "dfb", Sport: "football"},
}

// DFBPokalLeagueID is the internal league ID for the DFB-Pokal (sourced from
// OpenLigaDB). It runs across the normal club season, so no special season
// handling is required.
const DFBPokalLeagueID = 5

// LeagueByID returns the configured league with the given ID, or false.
func LeagueByID(id int) (LeagueConfig, bool) {
	for _, lc := range Leagues {
		if lc.ID == id {
			return lc, true
		}
	}
	return LeagueConfig{}, false
}

// CurrentSeason returns the current club season as football-data identifies it:
// a season started in autumn is labelled by its starting year, so before July
// the season is the previous calendar year.
func CurrentSeason() int {
	now := time.Now()
	if now.Month() < 7 {
		return now.Year() - 1
	}
	return now.Year()
}

func Load() *Config {
	corsOrigins := []string{"*"}
	if v := os.Getenv("CORS_ALLOWED_ORIGINS"); v != "" {
		corsOrigins = strings.Split(v, ",")
	}
	model := getEnv("ANTHROPIC_MODEL", "claude-haiku-4-5-20251001")
	return &Config{
		AnthropicAPIKey:    getEnv("ANTHROPIC_API_KEY", ""),
		FootballDataAPIKey: getEnv("FOOTBALL_DATA_API_KEY", ""),
		GeminiAPIKey:       getEnv("GEMINI_API_KEY", ""),
		GeminiModel:        getEnv("GEMINI_MODEL", "gemini-2.5-flash"),
		Port:               getEnv("PORT", "8080"),
		CORSOrigins:        corsOrigins,
		AnthropicModel:     model,
		LineupMock:         getEnv("LINEUP_MOCK", "") == "true",
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
