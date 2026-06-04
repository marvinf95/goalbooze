package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/marvinf95/goalbooze/internal/model"
	"github.com/marvinf95/goalbooze/internal/repository"
	"github.com/marvinf95/goalbooze/internal/service"
)

type GameHandler struct {
	repo     *repository.GameRepository
	svc      *service.AssignmentService
	aiLineup *service.AILineupService
}

func NewGameHandler(repo *repository.GameRepository, svc *service.AssignmentService, aiLineup *service.AILineupService) *GameHandler {
	return &GameHandler{repo: repo, svc: svc, aiLineup: aiLineup}
}

type createEventReq struct {
	ID         int             `json:"id"`
	LeagueID   int             `json:"league_id"`
	HomeTeam   string          `json:"home_team"`
	HomeTeamID int             `json:"home_team_id"`
	AwayTeam   string          `json:"away_team"`
	AwayTeamID int             `json:"away_team_id"`
	Date       string          `json:"date,omitempty"`
	HomeLineup []model.Athlete `json:"home_lineup,omitempty"`
	AwayLineup []model.Athlete `json:"away_lineup,omitempty"`
	// Manual marks a self-created match: the provided lineups are used as-is
	// (>=1 athlete per team), bypassing the AI/squad lookup.
	Manual bool `json:"manual,omitempty"`
}

type createGameRequest struct {
	Players []model.Player   `json:"players"`
	Events  []createEventReq `json:"events"`
}

func (h *GameHandler) CreateGame(w http.ResponseWriter, r *http.Request) {
	var req createGameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.Players) == 0 {
		jsonError(w, "at least one player required", http.StatusBadRequest)
		return
	}
	if len(req.Players) > 10 {
		jsonError(w, "maximum 10 players allowed", http.StatusBadRequest)
		return
	}
	if len(req.Events) == 0 {
		jsonError(w, "at least one event required", http.StatusBadRequest)
		return
	}
	if len(req.Events) > 10 {
		jsonError(w, "maximum 10 events allowed", http.StatusBadRequest)
		return
	}

	seen := make(map[string]bool)
	for _, p := range req.Players {
		if p.Name == "" {
			jsonError(w, "player name must not be empty", http.StatusBadRequest)
			return
		}
		if seen[p.Name] {
			jsonError(w, "duplicate player name", http.StatusBadRequest)
			return
		}
		seen[p.Name] = true
	}

	events := make([]model.Event, 0, len(req.Events))
	lineupMap := make(map[int]service.LineupPair)

	for _, evReq := range req.Events {
		if evReq.HomeTeam == "" || evReq.AwayTeam == "" {
			jsonError(w, "event home_team and away_team are required", http.StatusBadRequest)
			return
		}

		var eventDate time.Time
		if evReq.Date != "" {
			parsed, err := time.Parse(time.RFC3339, evReq.Date)
			if err != nil {
				jsonError(w, fmt.Sprintf("invalid date format for event %d — use RFC3339", evReq.ID), http.StatusBadRequest)
				return
			}
			eventDate = parsed
		}

		event := model.Event{
			ID:       evReq.ID,
			LeagueID: evReq.LeagueID,
			HomeTeam: evReq.HomeTeam,
			AwayTeam: evReq.AwayTeam,
			Date:     eventDate,
		}

		var home, away []model.Athlete

		if evReq.Manual {
			// Self-created match: use the entered athletes directly, no AI/squad fallback.
			if len(evReq.HomeLineup) == 0 || len(evReq.AwayLineup) == 0 {
				jsonError(w, "manual event requires at least one athlete per team", http.StatusBadRequest)
				return
			}
			home = evReq.HomeLineup
			away = evReq.AwayLineup
		} else if len(evReq.HomeLineup) >= 11 && len(evReq.AwayLineup) >= 11 {
			home = evReq.HomeLineup
			away = evReq.AwayLineup
		} else {
			var isSquadPick bool
			var err error
			home, away, isSquadPick, err = h.aiLineup.GetLineup(event, evReq.HomeTeamID, evReq.AwayTeamID)
			if err != nil {
				log.Printf("lineup fetch failed for event %d: %v", evReq.ID, err)
				jsonError(w, "lineup not available — try again or select players manually", http.StatusBadGateway)
				return
			}
			_ = isSquadPick
		}

		for i := range home {
			if home[i].Team == "" {
				home[i].Team = evReq.HomeTeam
			}
		}
		for i := range away {
			if away[i].Team == "" {
				away[i].Team = evReq.AwayTeam
			}
		}

		events = append(events, event)
		lineupMap[evReq.ID] = service.LineupPair{Home: home, Away: away}
	}

	assignments := h.svc.Assign(req.Players, events, lineupMap)

	game := &model.Game{
		CreatedAt:   time.Now(),
		Players:     req.Players,
		Events:      events,
		Assignments: assignments,
	}

	if err := h.repo.Create(game); err != nil {
		log.Printf("failed to save game: %v", err)
		jsonError(w, "failed to save game", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(game)
}

func (h *GameHandler) GetLineupStatus(w http.ResponseWriter, r *http.Request) {
	eventID, err := strconv.Atoi(chi.URLParam(r, "eventID"))
	if err != nil {
		jsonError(w, "invalid event ID", http.StatusBadRequest)
		return
	}

	homeTeamID, _ := strconv.Atoi(r.URL.Query().Get("home_team_id"))
	awayTeamID, _ := strconv.Atoi(r.URL.Query().Get("away_team_id"))
	homeTeam := r.URL.Query().Get("home_team")
	awayTeam := r.URL.Query().Get("away_team")

	var eventDate time.Time
	if dateStr := r.URL.Query().Get("date"); dateStr != "" {
		eventDate, _ = time.Parse(time.RFC3339, dateStr)
	}

	event := model.Event{
		ID:       eventID,
		HomeTeam: homeTeam,
		AwayTeam: awayTeam,
		Date:     eventDate,
	}

	home, away, isSquadPick, err := h.aiLineup.GetLineup(event, homeTeamID, awayTeamID)
	if err != nil {
		log.Printf("lineup status failed for event %d: %v", eventID, err)
		jsonError(w, "lineup not available", http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"home":          home,
		"away":          away,
		"is_squad_pick": isSquadPick,
	})
}

func (h *GameHandler) GetGames(w http.ResponseWriter, r *http.Request) {
	games, err := h.repo.GetAll()
	if err != nil {
		log.Printf("GetGames failed: %v", err)
		jsonError(w, "failed to load games", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(games)
}

func (h *GameHandler) GetGame(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		jsonError(w, "invalid game ID", http.StatusBadRequest)
		return
	}
	game, err := h.repo.GetByID(id)
	if err != nil {
		log.Printf("GetGame(%d) failed: %v", id, err)
		jsonError(w, "failed to load game", http.StatusInternalServerError)
		return
	}
	if game == nil {
		jsonError(w, "game not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(game)
}

func (h *GameHandler) DeleteGame(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		jsonError(w, "invalid game ID", http.StatusBadRequest)
		return
	}
	if err := h.repo.Delete(id); err != nil {
		log.Printf("DeleteGame(%d) failed: %v", id, err)
		jsonError(w, "failed to delete game", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
