package comments

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/urfave/cli/v3"
)

// TestFallbackNoArgs verifies that running "hb comment" with no args shows help
// and does NOT contain "bd" branding
func TestFallbackNoArgs(t *testing.T) {
	var buf bytes.Buffer
	app := &cli.Command{
		Name:   "hb",
		Writer: &buf,
		Commands: []*cli.Command{
			CmdComment,
		},
	}

	err := app.Run(context.Background(), []string{"hb", "comment"})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	output := buf.String()

	// Verify output contains "get" (the subcommand)
	if !strings.Contains(output, "get") {
		t.Errorf("expected output to contain 'get' subcommand, got:\n%s", output)
	}

	// Verify output does NOT contain "bd"
	if strings.Contains(output, "bd") {
		t.Errorf("expected output to NOT contain 'bd', got:\n%s", output)
	}
}

// TestFallbackHelp verifies that running "hb comment --help" shows help
// and contains the "get" subcommand
func TestFallbackHelp(t *testing.T) {
	var buf bytes.Buffer
	app := &cli.Command{
		Name:   "hb",
		Writer: &buf,
		Commands: []*cli.Command{
			CmdComment,
		},
	}

	err := app.Run(context.Background(), []string{"hb", "comment", "--help"})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	output := buf.String()

	// Verify output contains "get"
	if !strings.Contains(output, "get") {
		t.Errorf("expected output to contain 'get' subcommand, got:\n%s", output)
	}
}

// mockIndexerWithComments creates a mock indexer server that returns n comments
// with sequential nodeIDs and timestamps
func mockIndexerWithComments(t *testing.T, n int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req graphQLRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		collection := req.Variables["collection"].(string)

		if collection == CommentCollection {
			edges := make([]recordEdge, n)
			for i := 0; i < n; i++ {
				nodeID := fmt.Sprintf("node-%d", i+1)
				timestamp := fmt.Sprintf("2025-01-15T10:%02d:00Z", i)
				edges[i] = recordEdge{
					Node: IndexerRecord{
						DID:  "did:plc:testuser",
						URI:  fmt.Sprintf("at://did:plc:testuser/org.impactindexer.review.comment/%d", i+1),
						RKey: fmt.Sprintf("%d", i+1),
						Value: map[string]interface{}{
							"subject": map[string]interface{}{
								"uri": fmt.Sprintf("beads:%s", nodeID),
							},
							"text":      fmt.Sprintf("Comment %d", i+1),
							"createdAt": timestamp,
						},
					},
				}
			}
			resp := graphQLResponse{
				Data: &graphQLData{
					Records: &recordsPage{
						Edges:    edges,
						PageInfo: pageInfo{HasNextPage: false},
					},
				},
			}
			json.NewEncoder(w).Encode(resp)
		} else if collection == LikeCollection {
			// Return empty likes
			resp := graphQLResponse{
				Data: &graphQLData{
					Records: &recordsPage{
						Edges:    []recordEdge{},
						PageInfo: pageInfo{HasNextPage: false},
					},
				},
			}
			json.NewEncoder(w).Encode(resp)
		}
	}))
}

// mockProfileServer creates a mock profile server that returns a test profile
func mockProfileServer(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actor := r.URL.Query().Get("actor")
		profile := Profile{DID: actor, Handle: "testuser.bsky.social"}
		json.NewEncoder(w).Encode(profile)
	}))
}

// TestGetNoArgsDefaultLimit verifies that "hb comment get" with no args returns only 10 comments
func TestGetNoArgsDefaultLimit(t *testing.T) {
	indexerServer := mockIndexerWithComments(t, 15)
	defer indexerServer.Close()

	profileServer := mockProfileServer(t)
	defer profileServer.Close()

	var buf bytes.Buffer
	app := &cli.Command{
		Name:   "hb",
		Writer: &buf,
		Commands: []*cli.Command{
			CmdComment,
		},
	}

	err := app.Run(context.Background(), []string{
		"hb", "comment", "get",
		"--indexer-url", indexerServer.URL,
		"--profile-api-url", profileServer.URL,
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	output := buf.String()

	// Count occurrences of "[" at start of lines (each comment starts with [nodeID])
	lines := strings.Split(output, "\n")
	commentCount := 0
	for _, line := range lines {
		if strings.HasPrefix(line, "[") {
			commentCount++
		}
	}

	if commentCount != 10 {
		t.Errorf("expected 10 comments in output, got %d", commentCount)
	}
}

// TestGetWithBeadsIDNoLimit verifies that "hb comment get test-id" returns all matching comments
func TestGetWithBeadsIDNoLimit(t *testing.T) {
	// Mock indexer server that returns 5 comments for "test-id"
	indexerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req graphQLRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		collection := req.Variables["collection"].(string)

		if collection == CommentCollection {
			edges := make([]recordEdge, 5)
			for i := 0; i < 5; i++ {
				timestamp := fmt.Sprintf("2025-01-15T10:%02d:00Z", i)
				edges[i] = recordEdge{
					Node: IndexerRecord{
						DID:  "did:plc:testuser",
						URI:  fmt.Sprintf("at://did:plc:testuser/org.impactindexer.review.comment/%d", i+1),
						RKey: fmt.Sprintf("%d", i+1),
						Value: map[string]interface{}{
							"subject": map[string]interface{}{
								"uri": "beads:test-id",
							},
							"text":      fmt.Sprintf("Comment %d", i+1),
							"createdAt": timestamp,
						},
					},
				}
			}
			resp := graphQLResponse{
				Data: &graphQLData{
					Records: &recordsPage{
						Edges:    edges,
						PageInfo: pageInfo{HasNextPage: false},
					},
				},
			}
			json.NewEncoder(w).Encode(resp)
		} else if collection == LikeCollection {
			// Return empty likes
			resp := graphQLResponse{
				Data: &graphQLData{
					Records: &recordsPage{
						Edges:    []recordEdge{},
						PageInfo: pageInfo{HasNextPage: false},
					},
				},
			}
			json.NewEncoder(w).Encode(resp)
		}
	}))
	defer indexerServer.Close()

	profileServer := mockProfileServer(t)
	defer profileServer.Close()

	var buf bytes.Buffer
	app := &cli.Command{
		Name:   "hb",
		Writer: &buf,
		Commands: []*cli.Command{
			CmdComment,
		},
	}

	err := app.Run(context.Background(), []string{
		"hb", "comment", "get", "test-id",
		"--indexer-url", indexerServer.URL,
		"--profile-api-url", profileServer.URL,
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	output := buf.String()

	// Count occurrences of "[" at start of lines (each comment starts with [nodeID])
	lines := strings.Split(output, "\n")
	commentCount := 0
	for _, line := range lines {
		if strings.HasPrefix(line, "[") {
			commentCount++
		}
	}

	// Should return all 5 comments (no limit when beads-id is specified)
	if commentCount != 5 {
		t.Errorf("expected 5 comments in output, got %d", commentCount)
	}
}

// TestGetWithNFlag verifies that "hb comment get -n 3" returns only 3 comments
func TestGetWithNFlag(t *testing.T) {
	indexerServer := mockIndexerWithComments(t, 10)
	defer indexerServer.Close()

	profileServer := mockProfileServer(t)
	defer profileServer.Close()

	var buf bytes.Buffer
	app := &cli.Command{
		Name:   "hb",
		Writer: &buf,
		Commands: []*cli.Command{
			CmdComment,
		},
	}

	err := app.Run(context.Background(), []string{
		"hb", "comment", "get", "-n", "3",
		"--indexer-url", indexerServer.URL,
		"--profile-api-url", profileServer.URL,
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	output := buf.String()

	// Count occurrences of "[" at start of lines (each comment starts with [nodeID])
	lines := strings.Split(output, "\n")
	commentCount := 0
	for _, line := range lines {
		if strings.HasPrefix(line, "[") {
			commentCount++
		}
	}

	if commentCount != 3 {
		t.Errorf("expected 3 comments in output, got %d", commentCount)
	}
}
