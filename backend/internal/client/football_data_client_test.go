package client

import (
	"testing"
)

func TestNewFootballDataClient(t *testing.T) {
	c := NewFootballDataClient("test-key-123")
	if c == nil {
		t.Fatal("client should not be nil")
	}
}

func TestFootballDataClient_GetLeagues(t *testing.T) {
	c := NewFootballDataClient("")
	leagues, err := c.GetLeagues()
	if err != nil {
		t.Fatalf("GetLeagues() returned error: %v", err)
	}
	if len(leagues) != 4 {
		t.Fatalf("expected 4 leagues, got %d", len(leagues))
	}
	slugs := make(map[string]bool)
	for _, l := range leagues {
		slugs[l.Slug] = true
	}
	for _, expected := range []string{"BL1", "BL2", "CL", "WC"} {
		if !slugs[expected] {
			t.Errorf("expected league slug '%s' to be present", expected)
		}
	}
}
