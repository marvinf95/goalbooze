package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
	_ "modernc.org/sqlite"

	"github.com/marvinf95/goalbooze/config"
	"github.com/marvinf95/goalbooze/internal/client"
	"github.com/marvinf95/goalbooze/internal/handler"
	"github.com/marvinf95/goalbooze/internal/repository"
	"github.com/marvinf95/goalbooze/internal/service"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found — using environment variables")
	}

	cfg := config.Load()

	db, err := sql.Open("sqlite", "./data/games.db")
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	db.SetMaxOpenConns(1)

	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		log.Fatalf("failed to enable foreign keys: %v", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}
	if err := repository.RunMigrations(db); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	apiClient := client.NewFootballDataClient(cfg.FootballDataAPIKey)
	log.Println("Using football-data.org client")

	gameRepo := repository.NewGameRepository(db)
	squadRepo := repository.NewSquadRepository(db)
	lineupCache := repository.NewLineupCacheRepository(db)

	// Seed static squads for competitions without reliable free roster data
	// (World Cup national teams). Idempotent.
	if n, err := squadRepo.SeedSquadsFromFile("data/wc2026_squads.json", config.WorldCupLeagueID, config.WorldCupSeason); err != nil {
		log.Printf("WM squad seed skipped: %v", err)
	} else {
		log.Printf("WM squad seed: %d national teams cached", n)
	}

	// Build the lineup provider chain. Mock mode replaces all real providers so
	// the full flow is testable without any API tokens.
	var providers []service.LineupProvider
	if cfg.LineupMock {
		providers = append(providers, service.NewMockLineupProvider("testdata/mock_lineups.json"))
		log.Println("lineup mock mode active — no external AI calls")
	} else {
		if cfg.GeminiAPIKey != "" {
			providers = append(providers, service.NewGeminiLineupProvider(cfg.GeminiAPIKey, cfg.GeminiModel))
			log.Println("lineup provider enabled: gemini (free tier, tried first)")
		}
		if cfg.AnthropicAPIKey != "" {
			providers = append(providers, service.NewClaudeLineupProvider(cfg.AnthropicAPIKey, cfg.AnthropicModel))
			log.Println("lineup provider enabled: claude (fallback)")
		}
		if len(providers) == 0 {
			log.Println("no AI lineup providers configured — using squad-based selection")
		}
	}

	assignmentSvc := service.NewAssignmentService()
	aiLineupSvc := service.NewAILineupService(providers, lineupCache, squadRepo)

	leagueHandler := handler.NewLeagueHandler(apiClient)
	squadHandler := handler.NewSquadHandler(squadRepo, apiClient)
	gameHandler := handler.NewGameHandler(gameRepo, assignmentSvc, aiLineupSvc)

	r := chi.NewRouter()
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.RequestID)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.CORSOrigins,
		AllowedMethods:   []string{"GET", "POST", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Content-Type"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	r.Get("/api/v1/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/leagues", leagueHandler.GetLeagues)
		r.Get("/leagues/{leagueID}/events", leagueHandler.GetEvents)
		r.Get("/leagues/{leagueID}/teams", squadHandler.GetTeams)

		r.Get("/teams/{teamID}/squad", squadHandler.GetSquad)

		r.Get("/events/{eventID}/lineup", gameHandler.GetLineupStatus)

		r.Post("/games", gameHandler.CreateGame)
		r.Get("/games", gameHandler.GetGames)
		r.Get("/games/{id}", gameHandler.GetGame)
		r.Delete("/games/{id}", gameHandler.DeleteGame)
	})

	port := cfg.Port
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}

	addr := fmt.Sprintf(":%s", port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 90 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	log.Printf("GoalBooze backend starting on %s", addr)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
