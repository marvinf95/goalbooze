package service

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/marvinf95/goalbooze/internal/model"
	"github.com/marvinf95/goalbooze/internal/repository"
)

const anthropicAPIURL = "https://api.anthropic.com/v1/messages"

type AILineupService struct {
	anthropicKey   string
	anthropicModel string
	httpClient     *http.Client
	lineupCache    *repository.LineupCacheRepository
	squadRepo      *repository.SquadRepository
}

func NewAILineupService(anthropicKey, anthropicModel string, lineupCache *repository.LineupCacheRepository, squadRepo *repository.SquadRepository) *AILineupService {
	if anthropicModel == "" {
		anthropicModel = "claude-haiku-4-5-20251001"
	}
	return &AILineupService{
		anthropicKey:   anthropicKey,
		anthropicModel: anthropicModel,
		httpClient:     &http.Client{Timeout: 60 * time.Second},
		lineupCache:    lineupCache,
		squadRepo:      squadRepo,
	}
}

type claudeContentBlock struct {
	Type      string          `json:"type"`
	Text      string          `json:"text,omitempty"`
	ID        string          `json:"id,omitempty"`
	Name      string          `json:"name,omitempty"`
	Input     json.RawMessage `json:"input,omitempty"`
	ToolUseID string          `json:"tool_use_id,omitempty"`
	Content   json.RawMessage `json:"content,omitempty"`
}

type claudeMsg struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"`
}

type claudeRequest struct {
	Model     string      `json:"model"`
	MaxTokens int         `json:"max_tokens"`
	Tools     []claudeTool `json:"tools,omitempty"`
	Messages  []claudeMsg `json:"messages"`
}

type claudeTool struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

type claudeResponse struct {
	Content    []claudeContentBlock `json:"content"`
	StopReason string               `json:"stop_reason"`
	Error      *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// GetLineup fetches the official starting lineup for an event.
// It tries the cache first, then calls Claude with web search.
// Falls back to squad-based random selection if Claude is not configured.
func (s *AILineupService) GetLineup(event model.Event, homeTeamID, awayTeamID int) (home []model.Athlete, away []model.Athlete, isSquadPick bool, err error) {
	// 1. Check lineup cache
	if s.lineupCache != nil {
		cachedHome, cachedAway, found, cacheErr := s.lineupCache.Get(event.ID)
		if cacheErr == nil && found && len(cachedHome) > 0 && len(cachedAway) > 0 {
			log.Printf("lineup cache hit for event %d", event.ID)
			return cachedHome, cachedAway, false, nil
		}
	}

	// 2. Try Claude AI lineup if configured
	if s.anthropicKey != "" {
		home, away, aiErr := s.fetchLineupFromClaude(event)
		if aiErr == nil && len(home) >= 11 && len(away) >= 11 {
			// Cache permanently
			if s.lineupCache != nil {
				if err := s.lineupCache.Set(event.ID, home, away); err != nil {
					log.Printf("failed to cache lineup for event %d: %v", event.ID, err)
				}
			}
			return home, away, false, nil
		}
		log.Printf("AI lineup failed for event %d: %v — using squad fallback", event.ID, aiErr)
	}

	// 3. Squad-based fallback
	home, away, err = s.buildSquadLineup(homeTeamID, awayTeamID, event)
	if err != nil {
		return nil, nil, false, err
	}
	return home, away, true, nil
}

func (s *AILineupService) fetchLineupFromClaude(event model.Event) ([]model.Athlete, []model.Athlete, error) {
	dateStr := event.Date.Format("02.01.2006")
	prompt := fmt.Sprintf(
		"Search for the official starting lineup (Startaufstellung) for the football match: "+
			"%s vs %s on %s. "+
			"Return ONLY valid JSON with exactly 11 players per team and NO other text:\n"+
			`{"home":[{"name":"Player Name","position":"GK","number":1}],"away":[{"name":"Player Name","position":"GK","number":1}]}`+
			"\nUse position codes: GK, CB, RB, LB, CM, DM, RM, LM, AM, RW, LW, CF.",
		event.HomeTeam, event.AwayTeam, dateStr,
	)

	text, err := s.callClaude(prompt)
	if err != nil {
		return nil, nil, err
	}

	jsonStr := extractJSON(text)
	if jsonStr == "" {
		return nil, nil, fmt.Errorf("no JSON in Claude response")
	}

	var data struct {
		Home []struct {
			Name     string `json:"name"`
			Position string `json:"position"`
			Number   int    `json:"number"`
		} `json:"home"`
		Away []struct {
			Name     string `json:"name"`
			Position string `json:"position"`
			Number   int    `json:"number"`
		} `json:"away"`
	}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return nil, nil, fmt.Errorf("JSON parse failed: %w", err)
	}

	home := make([]model.Athlete, 0, len(data.Home))
	for i, p := range data.Home {
		home = append(home, model.Athlete{
			ID: event.ID*1000 + i, Name: p.Name,
			Position: p.Position, Number: p.Number, Team: event.HomeTeam,
		})
	}
	away := make([]model.Athlete, 0, len(data.Away))
	for i, p := range data.Away {
		away = append(away, model.Athlete{
			ID: event.ID*1000 + 100 + i, Name: p.Name,
			Position: p.Position, Number: p.Number, Team: event.AwayTeam,
		})
	}

	if len(home) < 11 || len(away) < 11 {
		return nil, nil, fmt.Errorf("incomplete lineup: home=%d away=%d", len(home), len(away))
	}
	return home, away, nil
}

func (s *AILineupService) callClaude(prompt string) (string, error) {
	messages := []claudeMsg{{Role: "user", Content: prompt}}

	for attempt := 0; attempt < 5; attempt++ {
		reqData := claudeRequest{
			Model:     s.anthropicModel,
			MaxTokens: 1024,
			Tools:     []claudeTool{{Type: "web_search_20250305", Name: "web_search"}},
			Messages:  messages,
		}

		body, err := json.Marshal(reqData)
		if err != nil {
			return "", err
		}

		req, err := http.NewRequest("POST", anthropicAPIURL, bytes.NewReader(body))
		if err != nil {
			return "", err
		}
		req.Header.Set("x-api-key", s.anthropicKey)
		req.Header.Set("anthropic-version", "2023-06-01")
		req.Header.Set("anthropic-beta", "web-search-2025-03-05")
		req.Header.Set("content-type", "application/json")

		resp, err := s.httpClient.Do(req)
		if err != nil {
			return "", fmt.Errorf("http request failed: %w", err)
		}
		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return "", fmt.Errorf("failed to read response: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("claude API error %d: %s", resp.StatusCode, string(respBody))
		}

		var claudeResp claudeResponse
		if err := json.Unmarshal(respBody, &claudeResp); err != nil {
			return "", fmt.Errorf("failed to decode response: %w", err)
		}
		if claudeResp.Error != nil {
			return "", fmt.Errorf("claude error: %s", claudeResp.Error.Message)
		}

		var textContent string
		var toolUseBlocks []claudeContentBlock
		for _, block := range claudeResp.Content {
			switch block.Type {
			case "text":
				textContent += block.Text
			case "tool_use":
				toolUseBlocks = append(toolUseBlocks, block)
			}
		}

		if claudeResp.StopReason == "end_turn" || len(toolUseBlocks) == 0 {
			return textContent, nil
		}

		// Handle tool_use: add assistant message and send tool results
		messages = append(messages, claudeMsg{Role: "assistant", Content: claudeResp.Content})

		toolResults := make([]claudeContentBlock, 0, len(toolUseBlocks))
		for _, tu := range toolUseBlocks {
			// Find matching web_search_tool_result in response (Anthropic server-side)
			var resultContent json.RawMessage
			for _, block := range claudeResp.Content {
				if block.Type == "web_search_tool_result" {
					resultContent = block.Content
					break
				}
			}
			if resultContent == nil {
				resultContent = json.RawMessage(`"Search completed"`)
			}
			toolResults = append(toolResults, claudeContentBlock{
				Type:      "tool_result",
				ToolUseID: tu.ID,
				Content:   resultContent,
			})
		}
		messages = append(messages, claudeMsg{Role: "user", Content: toolResults})
	}

	return "", fmt.Errorf("max iterations reached without response")
}

func (s *AILineupService) buildSquadLineup(homeTeamID, awayTeamID int, event model.Event) ([]model.Athlete, []model.Athlete, error) {
	homeSquad, err := s.squadRepo.GetSquad(homeTeamID)
	if err != nil {
		return nil, nil, fmt.Errorf("home squad not available: %w", err)
	}
	awaySquad, err := s.squadRepo.GetSquad(awayTeamID)
	if err != nil {
		return nil, nil, fmt.Errorf("away squad not available: %w", err)
	}

	if len(homeSquad) == 0 || len(awaySquad) == 0 {
		return nil, nil, fmt.Errorf("squad data not cached for teams %d/%d — fetch squads first", homeTeamID, awayTeamID)
	}

	home := pickEleven(homeSquad, event.HomeTeam)
	away := pickEleven(awaySquad, event.AwayTeam)
	return home, away, nil
}

func pickEleven(squad []model.Athlete, teamName string) []model.Athlete {
	cp := make([]model.Athlete, len(squad))
	copy(cp, squad)
	for i := range cp {
		cp[i].Team = teamName
	}
	// Fisher-Yates shuffle using crypto/rand for unbiased randomness
	for i := len(cp) - 1; i > 0; i-- {
		jBig, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			break
		}
		j := int(jBig.Int64())
		cp[i], cp[j] = cp[j], cp[i]
	}
	if len(cp) > 11 {
		cp = cp[:11]
	}
	return cp
}

func extractJSON(text string) string {
	start := strings.Index(text, "{")
	if start < 0 {
		return ""
	}
	depth := 0
	for i := start; i < len(text); i++ {
		switch text[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return text[start : i+1]
			}
		}
	}
	return ""
}
