package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/marvinf95/goalbooze/internal/model"
)

const defaultAnthropicAPIURL = "https://api.anthropic.com/v1/messages"

// ClaudeLineupProvider fetches lineups via the Anthropic Messages API using the
// server-side web_search tool, then a second (non-search) call to coerce the
// prose result into JSON when needed.
type ClaudeLineupProvider struct {
	apiKey     string
	model      string
	apiURL     string
	httpClient *http.Client
	mu         sync.Mutex // serialize calls so concurrent fetches don't hit rate limits
}

func NewClaudeLineupProvider(apiKey, model string) *ClaudeLineupProvider {
	if model == "" {
		model = "claude-haiku-4-5-20251001"
	}
	return &ClaudeLineupProvider{
		apiKey:     apiKey,
		model:      model,
		apiURL:     defaultAnthropicAPIURL,
		httpClient: &http.Client{Timeout: 60 * time.Second},
	}
}

func (p *ClaudeLineupProvider) Name() string { return "claude" }

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
	Model     string       `json:"model"`
	MaxTokens int          `json:"max_tokens"`
	Tools     []claudeTool `json:"tools,omitempty"`
	Messages  []claudeMsg  `json:"messages"`
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

func (p *ClaudeLineupProvider) FetchLineup(event model.Event) ([]model.Athlete, []model.Athlete, error) {
	// Serialize Claude calls so concurrent lineup requests don't hit rate limits
	p.mu.Lock()
	defer p.mu.Unlock()

	// Step 1: web search — let Claude answer naturally, no forced JSON format
	lineupText, err := p.callClaude(lineupSearchPrompt(event))
	if err != nil {
		return nil, nil, err
	}
	if lineupText == "" {
		return nil, nil, fmt.Errorf("empty response from web search")
	}

	// Step 2: try to extract JSON directly; if Claude already returned it, we're done
	jsonStr := extractJSON(lineupText)
	if jsonStr == "" {
		// Claude returned prose — ask a second time to format as JSON (no web search)
		jsonStr, err = p.formatLineupAsJSON(lineupText, event.HomeTeam, event.AwayTeam)
		if err != nil {
			return nil, nil, fmt.Errorf("JSON formatting failed: %w", err)
		}
	}

	return parseLineupJSON(jsonStr, event)
}

// formatLineupAsJSON converts a prose lineup description to structured JSON via a
// simple (non-search) Claude call. Reliable because it's just reformatting known data.
func (p *ClaudeLineupProvider) formatLineupAsJSON(lineupText, homeTeam, awayTeam string) (string, error) {
	reqData := claudeRequest{
		Model:     p.model,
		MaxTokens: 1024,
		Messages:  []claudeMsg{{Role: "user", Content: lineupFormatPrompt(lineupText, homeTeam, awayTeam)}},
	}
	body, err := json.Marshal(reqData)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", p.apiURL, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("x-api-key", p.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("content-type", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("http request failed: %w", err)
	}
	respBody, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return "", err
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

	var text string
	for _, block := range claudeResp.Content {
		if block.Type == "text" {
			text += block.Text
		}
	}

	jsonStr := extractJSON(text)
	if jsonStr == "" {
		return "", fmt.Errorf("no JSON in formatting response")
	}
	return jsonStr, nil
}

func (p *ClaudeLineupProvider) callClaude(prompt string) (string, error) {
	messages := []claudeMsg{{Role: "user", Content: prompt}}

	for attempt := 0; attempt < 5; attempt++ {
		reqData := claudeRequest{
			Model:     p.model,
			MaxTokens: 1024,
			Tools:     []claudeTool{{Type: "web_search_20250305", Name: "web_search"}},
			Messages:  messages,
		}

		body, err := json.Marshal(reqData)
		if err != nil {
			return "", err
		}

		req, err := http.NewRequest("POST", p.apiURL, bytes.NewReader(body))
		if err != nil {
			return "", err
		}
		req.Header.Set("x-api-key", p.apiKey)
		req.Header.Set("anthropic-version", "2023-06-01")
		req.Header.Set("anthropic-beta", "web-search-2025-03-05")
		req.Header.Set("content-type", "application/json")

		resp, err := p.httpClient.Do(req)
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
