package comments

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchComments(t *testing.T) {
	// Mock indexer server
	indexerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req graphQLRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		collection := req.Variables["collection"].(string)

		if collection == CommentCollection {
			// Return 2 comment records for beads:test-id
			resp := graphQLResponse{
				Data: &graphQLData{
					Records: &recordsPage{
						Edges: []recordEdge{
							{
								Node: IndexerRecord{
									DID:  "did:plc:alice",
									URI:  "at://did:plc:alice/org.impactindexer.review.comment/1",
									RKey: "1",
									Value: map[string]interface{}{
										"subject": map[string]interface{}{
											"uri": "beads:test-id",
										},
										"text":      "First comment",
										"createdAt": "2025-01-15T10:00:00Z",
									},
								},
							},
							{
								Node: IndexerRecord{
									DID:  "did:plc:bob",
									URI:  "at://did:plc:bob/org.impactindexer.review.comment/2",
									RKey: "2",
									Value: map[string]interface{}{
										"subject": map[string]interface{}{
											"uri": "beads:test-id",
										},
										"text":      "Second comment",
										"createdAt": "2025-01-15T11:00:00Z",
									},
								},
							},
						},
						PageInfo: pageInfo{HasNextPage: false},
					},
				},
			}
			json.NewEncoder(w).Encode(resp)
		} else if collection == LikeCollection {
			// Return 1 like for first comment
			resp := graphQLResponse{
				Data: &graphQLData{
					Records: &recordsPage{
						Edges: []recordEdge{
							{
								Node: IndexerRecord{
									DID: "did:plc:charlie",
									URI: "at://did:plc:charlie/org.impactindexer.review.like/1",
									Value: map[string]interface{}{
										"subject": map[string]interface{}{
											"uri": "at://did:plc:alice/org.impactindexer.review.comment/1",
										},
									},
								},
							},
						},
						PageInfo: pageInfo{HasNextPage: false},
					},
				},
			}
			json.NewEncoder(w).Encode(resp)
		}
	}))
	defer indexerServer.Close()

	// Mock profile server
	profileServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actor := r.URL.Query().Get("actor")
		var profile Profile
		if actor == "did:plc:alice" {
			profile = Profile{DID: "did:plc:alice", Handle: "alice.bsky.social"}
		} else if actor == "did:plc:bob" {
			profile = Profile{DID: "did:plc:bob", Handle: "bob.bsky.social"}
		}
		json.NewEncoder(w).Encode(profile)
	}))
	defer profileServer.Close()

	// Fetch comments
	comments, err := FetchComments(context.Background(), indexerServer.URL, profileServer.URL, FetchOptions{BeadsID: "test-id"})
	if err != nil {
		t.Fatalf("FetchComments failed: %v", err)
	}

	// Verify 2 comments returned
	if len(comments) != 2 {
		t.Fatalf("expected 2 comments, got %d", len(comments))
	}

	// Verify nodeID
	for _, comment := range comments {
		if comment.NodeID != "test-id" {
			t.Errorf("expected nodeID 'test-id', got '%s'", comment.NodeID)
		}
	}

	// Verify like count on first comment (newest first, so second in original order)
	// Root comments are sorted newest-first, so bob's comment (11:00) comes before alice's (10:00)
	if comments[0].Handle != "bob.bsky.social" {
		t.Errorf("expected first comment from bob, got %s", comments[0].Handle)
	}
	if comments[1].Handle != "alice.bsky.social" {
		t.Errorf("expected second comment from alice, got %s", comments[1].Handle)
	}
	if comments[1].Likes != 1 {
		t.Errorf("expected alice's comment to have 1 like, got %d", comments[1].Likes)
	}
}

func TestFetchCommentsNotFound(t *testing.T) {
	// Mock indexer server that returns comments for different beads IDs
	indexerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req graphQLRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		collection := req.Variables["collection"].(string)

		if collection == CommentCollection {
			// Return comments for different beads ID
			resp := graphQLResponse{
				Data: &graphQLData{
					Records: &recordsPage{
						Edges: []recordEdge{
							{
								Node: IndexerRecord{
									DID:  "did:plc:alice",
									URI:  "at://did:plc:alice/org.impactindexer.review.comment/1",
									RKey: "1",
									Value: map[string]interface{}{
										"subject": map[string]interface{}{
											"uri": "beads:other-id",
										},
										"text":      "Comment for other issue",
										"createdAt": "2025-01-15T10:00:00Z",
									},
								},
							},
						},
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

	// Mock profile server
	profileServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		profile := Profile{DID: "did:plc:alice", Handle: "alice.bsky.social"}
		json.NewEncoder(w).Encode(profile)
	}))
	defer profileServer.Close()

	// Fetch comments for test-id (should be empty)
	comments, err := FetchComments(context.Background(), indexerServer.URL, profileServer.URL, FetchOptions{BeadsID: "test-id"})
	if err != nil {
		t.Fatalf("FetchComments failed: %v", err)
	}

	// Verify empty result
	if len(comments) != 0 {
		t.Errorf("expected 0 comments, got %d", len(comments))
	}
}

func TestFetchCommentsLikesFail(t *testing.T) {
	// Mock indexer server
	indexerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req graphQLRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		collection := req.Variables["collection"].(string)

		if collection == CommentCollection {
			// Return 1 comment
			resp := graphQLResponse{
				Data: &graphQLData{
					Records: &recordsPage{
						Edges: []recordEdge{
							{
								Node: IndexerRecord{
									DID:  "did:plc:alice",
									URI:  "at://did:plc:alice/org.impactindexer.review.comment/1",
									RKey: "1",
									Value: map[string]interface{}{
										"subject": map[string]interface{}{
											"uri": "beads:test-id",
										},
										"text":      "Comment",
										"createdAt": "2025-01-15T10:00:00Z",
									},
								},
							},
						},
						PageInfo: pageInfo{HasNextPage: false},
					},
				},
			}
			json.NewEncoder(w).Encode(resp)
		} else if collection == LikeCollection {
			// Return 500 error for likes
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer indexerServer.Close()

	// Mock profile server
	profileServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		profile := Profile{DID: "did:plc:alice", Handle: "alice.bsky.social"}
		json.NewEncoder(w).Encode(profile)
	}))
	defer profileServer.Close()

	// Fetch comments (should succeed despite likes failure)
	comments, err := FetchComments(context.Background(), indexerServer.URL, profileServer.URL, FetchOptions{BeadsID: "test-id"})
	if err != nil {
		t.Fatalf("FetchComments failed: %v", err)
	}

	// Verify 1 comment returned with 0 likes
	if len(comments) != 1 {
		t.Fatalf("expected 1 comment, got %d", len(comments))
	}

	if comments[0].Likes != 0 {
		t.Errorf("expected 0 likes (likes fetch failed), got %d", comments[0].Likes)
	}
}

func TestFetchCommentsNoFilter(t *testing.T) {
	// Mock indexer server
	indexerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req graphQLRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		collection := req.Variables["collection"].(string)

		if collection == CommentCollection {
			// Return 2 comment records for beads:test-id
			resp := graphQLResponse{
				Data: &graphQLData{
					Records: &recordsPage{
						Edges: []recordEdge{
							{
								Node: IndexerRecord{
									DID:  "did:plc:alice",
									URI:  "at://did:plc:alice/org.impactindexer.review.comment/1",
									RKey: "1",
									Value: map[string]interface{}{
										"subject": map[string]interface{}{
											"uri": "beads:test-id",
										},
										"text":      "First comment",
										"createdAt": "2025-01-15T10:00:00Z",
									},
								},
							},
							{
								Node: IndexerRecord{
									DID:  "did:plc:bob",
									URI:  "at://did:plc:bob/org.impactindexer.review.comment/2",
									RKey: "2",
									Value: map[string]interface{}{
										"subject": map[string]interface{}{
											"uri": "beads:test-id",
										},
										"text":      "Second comment",
										"createdAt": "2025-01-15T11:00:00Z",
									},
								},
							},
						},
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

	// Mock profile server
	profileServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actor := r.URL.Query().Get("actor")
		var profile Profile
		if actor == "did:plc:alice" {
			profile = Profile{DID: "did:plc:alice", Handle: "alice.bsky.social"}
		} else if actor == "did:plc:bob" {
			profile = Profile{DID: "did:plc:bob", Handle: "bob.bsky.social"}
		}
		json.NewEncoder(w).Encode(profile)
	}))
	defer profileServer.Close()

	// Fetch comments with no filter (empty FetchOptions)
	comments, err := FetchComments(context.Background(), indexerServer.URL, profileServer.URL, FetchOptions{})
	if err != nil {
		t.Fatalf("FetchComments failed: %v", err)
	}

	// Verify 2 comments returned (both test-id ones, not filtered)
	if len(comments) != 2 {
		t.Fatalf("expected 2 comments, got %d", len(comments))
	}

	// Verify both have nodeID "test-id"
	for _, comment := range comments {
		if comment.NodeID != "test-id" {
			t.Errorf("expected nodeID 'test-id', got '%s'", comment.NodeID)
		}
	}
}
