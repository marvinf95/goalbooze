package service

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/marvinf95/goalbooze/internal/model"
)

func geminiReply(text string) string {
	resp := geminiResponse{}
	resp.Candidates = []struct {
		Content geminiContent `json:"content"`
	}{{Content: geminiContent{Parts: []geminiPart{{Text: text}}}}}
	b, _ := json.Marshal(resp)
	return string(b)
}

const sampleLineupJSON = `{"home":[{"name":"A","position":"GK","number":1},{"name":"B","position":"RB","number":2},{"name":"C","position":"CB","number":3},{"name":"D","position":"CB","number":4},{"name":"E","position":"LB","number":5},{"name":"F","position":"DM","number":6},{"name":"G","position":"CM","number":7},{"name":"H","position":"AM","number":8},{"name":"I","position":"RW","number":9},{"name":"J","position":"CF","number":10},{"name":"K","position":"LW","number":11}],"away":[{"name":"a","position":"GK","number":1},{"name":"b","position":"RB","number":2},{"name":"c","position":"CB","number":3},{"name":"d","position":"CB","number":4},{"name":"e","position":"LB","number":5},{"name":"f","position":"DM","number":6},{"name":"g","position":"CM","number":7},{"name":"h","position":"AM","number":8},{"name":"i","position":"RW","number":9},{"name":"j","position":"CF","number":10},{"name":"k","position":"LW","number":11}]}`

func TestGeminiProvider_DirectJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(geminiReply("Here you go: " + sampleLineupJSON)))
	}))
	defer srv.Close()

	p := NewGeminiLineupProvider("fake-key", "gemini-test")
	p.baseURL = srv.URL

	home, away, err := p.FetchLineup(model.Event{ID: 1, HomeTeam: "H", AwayTeam: "A", Date: time.Now()})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(home) != 11 || len(away) != 11 {
		t.Fatalf("expected 11+11, got %d+%d", len(home), len(away))
	}
}

func TestGeminiProvider_TwoStepFormatting(t *testing.T) {
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		calls++
		// First call uses the google_search tool and returns prose only;
		// second call (no tool) returns the JSON.
		if strings.Contains(string(body), "google_search") {
			w.Write([]byte(geminiReply("The lineup is A in goal, B at right back, ...")))
			return
		}
		w.Write([]byte(geminiReply(sampleLineupJSON)))
	}))
	defer srv.Close()

	p := NewGeminiLineupProvider("fake-key", "gemini-test")
	p.baseURL = srv.URL

	home, away, err := p.FetchLineup(model.Event{ID: 2, HomeTeam: "H", AwayTeam: "A", Date: time.Now()})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls != 2 {
		t.Errorf("expected 2 calls (search + format), got %d", calls)
	}
	if len(home) != 11 || len(away) != 11 {
		t.Fatalf("expected 11+11, got %d+%d", len(home), len(away))
	}
}

func TestGeminiProvider_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{"error":{"code":429,"message":"rate limited"}}`))
	}))
	defer srv.Close()

	p := NewGeminiLineupProvider("fake-key", "gemini-test")
	p.baseURL = srv.URL

	if _, _, err := p.FetchLineup(model.Event{ID: 3, HomeTeam: "H", AwayTeam: "A", Date: time.Now()}); err == nil {
		t.Fatal("expected an error on HTTP 429")
	}
}
