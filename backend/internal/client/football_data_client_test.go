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
	// The football-data client only lists the leagues it actually serves
	// (BL1, CL); 2. Bundesliga and DFB-Pokal come from OpenLigaDB.
	c := NewFootballDataClient("")
	leagues, err := c.GetLeagues()
	if err != nil {
		t.Fatalf("GetLeagues() returned error: %v", err)
	}
	slugs := make(map[string]bool)
	for _, l := range leagues {
		slugs[l.Slug] = true
	}
	for _, expected := range []string{"BL1", "CL"} {
		if !slugs[expected] {
			t.Errorf("expected league slug '%s' to be present", expected)
		}
	}
	for _, unexpected := range []string{"BL2", "DFB"} {
		if slugs[unexpected] {
			t.Errorf("slug '%s' should not be served by football-data", unexpected)
		}
	}
}
