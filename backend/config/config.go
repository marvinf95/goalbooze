package config

import (
	"os"
	"strings"
)

type Config struct {
	AnthropicAPIKey    string
	FootballDataAPIKey string
	Port               string
	CORSOrigins        []string
	AnthropicModel     string
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
		Port:               getEnv("PORT", "8080"),
		CORSOrigins:        corsOrigins,
		AnthropicModel:     model,
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
