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

const defaultGeminiBaseURL = "https://generativelanguage.googleapis.com/v1beta"

// GeminiLineupProvider fetches lineups via the Google Gemini API using the
// built-in google_search grounding tool, then a second (non-search) call to
// coerce the prose result into JSON when needed. Gemini's free tier makes this a
// cheap first stop before falling back to Claude.
type GeminiLineupProvider struct {
	apiKey     string
	model      string
	baseURL    string
	httpClient *http.Client
	mu         sync.Mutex // serialize calls so concurrent fetches don't hit rate limits
}

func NewGeminiLineupProvider(apiKey, model string) *GeminiLineupProvider {
	if model == "" {
		model = "gemini-2.5-flash"
	}
	return &GeminiLineupProvider{
		apiKey:     apiKey,
		model:      model,
		baseURL:    defaultGeminiBaseURL,
		httpClient: &http.Client{Timeout: 60 * time.Second},
	}
}

func (p *GeminiLineupProvider) Name() string { return "gemini" }

type geminiPart struct {
	Text string `json:"text"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
	Role  string       `json:"role,omitempty"`
}

type geminiTool struct {
	GoogleSearch *struct{} `json:"google_search,omitempty"`
}

type geminiRequest struct {
	Contents []geminiContent `json:"contents"`
	Tools    []geminiTool    `json:"tools,omitempty"`
}

type geminiResponse struct {
	Candidates []struct {
		Content geminiContent `json:"content"`
	} `json:"candidates"`
	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (p *GeminiLineupProvider) FetchLineup(event model.Event) ([]model.Athlete, []model.Athlete, error) {
	// Serialize calls so concurrent lineup requests don't hit rate limits
	p.mu.Lock()
	defer p.mu.Unlock()

	// Step 1: web search via google_search grounding — answer naturally
	lineupText, err := p.generate(lineupSearchPrompt(event), true)
	if err != nil {
		return nil, nil, err
	}
	if lineupText == "" {
		return nil, nil, fmt.Errorf("empty response from gemini search")
	}

	// Step 2: try to extract JSON directly; otherwise reformat (no search tool)
	jsonStr := extractJSON(lineupText)
	if jsonStr == "" {
		formatted, ferr := p.generate(lineupFormatPrompt(lineupText, event.HomeTeam, event.AwayTeam), false)
		if ferr != nil {
			return nil, nil, fmt.Errorf("JSON formatting failed: %w", ferr)
		}
		jsonStr = extractJSON(formatted)
		if jsonStr == "" {
			return nil, nil, fmt.Errorf("no JSON in gemini formatting response")
		}
	}

	return parseLineupJSON(jsonStr, event)
}

// generate calls the Gemini generateContent endpoint. When useSearch is true the
// google_search grounding tool is enabled.
func (p *GeminiLineupProvider) generate(prompt string, useSearch bool) (string, error) {
	reqData := geminiRequest{
		Contents: []geminiContent{{
			Role:  "user",
			Parts: []geminiPart{{Text: prompt}},
		}},
	}
	if useSearch {
		reqData.Tools = []geminiTool{{GoogleSearch: &struct{}{}}}
	}

	body, err := json.Marshal(reqData)
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("%s/models/%s:generateContent?key=%s", p.baseURL, p.model, p.apiKey)
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
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
		return "", fmt.Errorf("gemini API error %d: %s", resp.StatusCode, string(respBody))
	}

	var gResp geminiResponse
	if err := json.Unmarshal(respBody, &gResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}
	if gResp.Error != nil {
		return "", fmt.Errorf("gemini error: %s", gResp.Error.Message)
	}

	var text string
	for _, cand := range gResp.Candidates {
		for _, part := range cand.Content.Parts {
			text += part.Text
		}
	}
	return text, nil
}
