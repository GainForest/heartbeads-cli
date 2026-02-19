package comments

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestResolveProfiles(t *testing.T) {
	var requestCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		did := r.URL.Query().Get("actor")

		var profile Profile
		switch did {
		case "did:plc:alice":
			profile = Profile{
				DID:    "did:plc:alice",
				Handle: "alice.bsky.social",
			}
		case "did:plc:bob":
			profile = Profile{
				DID:    "did:plc:bob",
				Handle: "bob.bsky.social",
			}
		default:
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(profile); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}))
	defer server.Close()

	dids := []string{"did:plc:alice", "did:plc:bob"}
	profiles := ResolveProfiles(context.Background(), server.URL, dids)

	if len(profiles) != 2 {
		t.Fatalf("expected 2 profiles, got %d", len(profiles))
	}

	if profiles["did:plc:alice"].Handle != "alice.bsky.social" {
		t.Errorf("expected alice.bsky.social, got %s", profiles["did:plc:alice"].Handle)
	}

	if profiles["did:plc:bob"].Handle != "bob.bsky.social" {
		t.Errorf("expected bob.bsky.social, got %s", profiles["did:plc:bob"].Handle)
	}
}

func TestResolveProfilesFallback(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer server.Close()

	dids := []string{"did:plc:unknown1", "did:plc:unknown2"}
	profiles := ResolveProfiles(context.Background(), server.URL, dids)

	if len(profiles) != 2 {
		t.Fatalf("expected 2 profiles, got %d", len(profiles))
	}

	// Verify fallback profiles (DID as handle)
	if profiles["did:plc:unknown1"].Handle != "did:plc:unknown1" {
		t.Errorf("expected fallback handle did:plc:unknown1, got %s", profiles["did:plc:unknown1"].Handle)
	}

	if profiles["did:plc:unknown1"].DID != "did:plc:unknown1" {
		t.Errorf("expected DID did:plc:unknown1, got %s", profiles["did:plc:unknown1"].DID)
	}

	if profiles["did:plc:unknown2"].Handle != "did:plc:unknown2" {
		t.Errorf("expected fallback handle did:plc:unknown2, got %s", profiles["did:plc:unknown2"].Handle)
	}
}

func TestResolveProfilesDedup(t *testing.T) {
	var requestCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		did := r.URL.Query().Get("actor")

		profile := Profile{
			DID:    did,
			Handle: "alice.bsky.social",
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(profile); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}))
	defer server.Close()

	// Pass same DID twice
	dids := []string{"did:plc:alice", "did:plc:alice"}
	profiles := ResolveProfiles(context.Background(), server.URL, dids)

	// Should only make 1 request
	if atomic.LoadInt32(&requestCount) != 1 {
		t.Errorf("expected 1 request, got %d", requestCount)
	}

	// Should only have 1 entry in map
	if len(profiles) != 1 {
		t.Fatalf("expected 1 profile in map, got %d", len(profiles))
	}

	if profiles["did:plc:alice"].Handle != "alice.bsky.social" {
		t.Errorf("expected alice.bsky.social, got %s", profiles["did:plc:alice"].Handle)
	}
}

func TestResolveProfilesEmpty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("should not make any requests for empty input")
	}))
	defer server.Close()

	profiles := ResolveProfiles(context.Background(), server.URL, []string{})

	if profiles == nil {
		t.Error("expected non-nil map, got nil")
	}

	if len(profiles) != 0 {
		t.Errorf("expected empty map, got %d entries", len(profiles))
	}
}

func TestResolveProfilesDisplayName(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		profile := Profile{
			DID:         "did:plc:alice",
			Handle:      "alice.bsky.social",
			DisplayName: "Alice Wonderland",
			Avatar:      "https://cdn.bsky.app/img/avatar/alice.jpg",
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(profile); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}))
	defer server.Close()

	dids := []string{"did:plc:alice"}
	profiles := ResolveProfiles(context.Background(), server.URL, dids)

	if len(profiles) != 1 {
		t.Fatalf("expected 1 profile, got %d", len(profiles))
	}

	profile := profiles["did:plc:alice"]

	if profile.DisplayName != "Alice Wonderland" {
		t.Errorf("expected displayName 'Alice Wonderland', got '%s'", profile.DisplayName)
	}

	if profile.Avatar != "https://cdn.bsky.app/img/avatar/alice.jpg" {
		t.Errorf("expected avatar URL, got '%s'", profile.Avatar)
	}

	if profile.Handle != "alice.bsky.social" {
		t.Errorf("expected handle 'alice.bsky.social', got '%s'", profile.Handle)
	}

	if profile.DID != "did:plc:alice" {
		t.Errorf("expected DID 'did:plc:alice', got '%s'", profile.DID)
	}
}
