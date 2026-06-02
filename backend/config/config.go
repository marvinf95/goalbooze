package config

import (
	"os"
	"strings"
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

type LeagueConfig struct {
	ID               int
	Name             string
	FootballDataCode string
	Sport            string
}

var Leagues = []LeagueConfig{
	{ID: 1, Name: "1. Bundesliga", FootballDataCode: "BL1", Sport: "football"},
	{ID: 2, Name: "2. Bundesliga", FootballDataCode: "BL2", Sport: "football"},
	{ID: 3, Name: "Champions League", FootballDataCode: "CL", Sport: "football"},
	{ID: 4, Name: "WM 2026", FootballDataCode: "WC", Sport: "football"},
}

// WorldCupLeagueID is the internal league ID for the FIFA World Cup. The cup runs
// in summer, so its season is the tournament year rather than the club-season
// calculation used for the leagues above.
const WorldCupLeagueID = 4

// WorldCupSeason is the football-data season identifier for the World Cup.
const WorldCupSeason = 2026

// SeasonForLeague returns the season to query for a league. Summer tournaments
// (World Cup) use a fixed tournament year; club leagues use the supplied
// club-season fallback (typically "current season" derived from the month).
func SeasonForLeague(leagueID, fallback int) int {
	if leagueID == WorldCupLeagueID {
		return WorldCupSeason
	}
	return fallback
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
