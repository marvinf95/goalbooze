package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/marvinf95/goalbooze/config"
	"github.com/marvinf95/goalbooze/internal/client"
	"github.com/marvinf95/goalbooze/internal/repository"
)

type SquadHandler struct {
	squadRepo *repository.SquadRepository
	apiClient client.SportsAPIClient
}

func NewSquadHandler(squadRepo *repository.SquadRepository, apiClient client.SportsAPIClient) *SquadHandler {
	return &SquadHandler{squadRepo: squadRepo, apiClient: apiClient}
}

func (h *SquadHandler) GetTeams(w http.ResponseWriter, r *http.Request) {
	leagueID, err := strconv.Atoi(chi.URLParam(r, "leagueID"))
	if err != nil {
		jsonError(w, "invalid league ID", http.StatusBadRequest)
		return
	}

	season := config.SeasonForLeague(leagueID, config.CurrentSeason())
	if s := r.URL.Query().Get("season"); s != "" {
		parsed, err := strconv.Atoi(s)
		if err != nil || parsed < 2000 || parsed > 2100 {
			jsonError(w, "invalid season", http.StatusBadRequest)
			return
		}
		season = parsed
	}

	stale, err := h.squadRepo.IsCacheStale(leagueID, season)
	if err != nil || stale {
		teams, fetchErr := h.apiClient.GetTeams(leagueID, season)
		if fetchErr != nil {
			log.Printf("failed to fetch teams for league %d: %v", leagueID, fetchErr)
		} else {
			if saveErr := h.squadRepo.SaveTeams(teams, leagueID, season); saveErr != nil {
				log.Printf("failed to save teams to cache: %v", saveErr)
			}
		}
	}

	teams, err := h.squadRepo.GetTeams(leagueID, season)
	if err != nil {
		log.Printf("GetTeams league=%d: %v", leagueID, err)
		jsonError(w, "failed to load teams", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(teams)
}

func (h *SquadHandler) GetSquad(w http.ResponseWriter, r *http.Request) {
	teamID, err := strconv.Atoi(chi.URLParam(r, "teamID"))
	if err != nil {
		jsonError(w, "invalid team ID", http.StatusBadRequest)
		return
	}

	squad, err := h.squadRepo.GetSquad(teamID)
	if err != nil {
		log.Printf("GetSquad team=%d: %v", teamID, err)
		jsonError(w, "failed to load squad", http.StatusInternalServerError)
		return
	}
	if squad == nil {
		jsonError(w, "squad not found — fetch teams first", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(squad)
}
