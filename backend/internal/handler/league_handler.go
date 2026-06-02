package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/marvinf95/goalbooze/config"
	"github.com/marvinf95/goalbooze/internal/client"
)

type LeagueHandler struct {
	apiClient client.SportsAPIClient
}

func NewLeagueHandler(apiClient client.SportsAPIClient) *LeagueHandler {
	return &LeagueHandler{apiClient: apiClient}
}

func (h *LeagueHandler) GetLeagues(w http.ResponseWriter, r *http.Request) {
	clubSeason := currentSeason()
	leagues := make([]map[string]interface{}, 0, len(config.Leagues))
	for _, lc := range config.Leagues {
		leagues = append(leagues, map[string]interface{}{
			"id":     lc.ID,
			"name":   lc.Name,
			"slug":   lc.FootballDataCode,
			"season": config.SeasonForLeague(lc.ID, clubSeason),
			"sport":  lc.Sport,
		})
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(leagues)
}

func (h *LeagueHandler) GetEvents(w http.ResponseWriter, r *http.Request) {
	leagueID, err := strconv.Atoi(chi.URLParam(r, "leagueID"))
	if err != nil {
		jsonError(w, "invalid league ID", http.StatusBadRequest)
		return
	}

	season := config.SeasonForLeague(leagueID, currentSeason())
	if s := r.URL.Query().Get("season"); s != "" {
		parsed, err := strconv.Atoi(s)
		if err != nil || parsed < 2000 || parsed > 2100 {
			jsonError(w, "invalid season — use a year between 2000 and 2100", http.StatusBadRequest)
			return
		}
		season = parsed
	}

	events, err := h.apiClient.GetEvents(leagueID, season)
	if err != nil {
		log.Printf("GetEvents league=%d season=%d: %v", leagueID, season, err)
		jsonError(w, "events not available", http.StatusBadGateway)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}
