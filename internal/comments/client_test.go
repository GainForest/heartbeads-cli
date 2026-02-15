package comments

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchRecordsByCollection(t *testing.T) {
	// Mock server that returns 2 records
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type: application/json, got %s", r.Header.Get("Content-Type"))
		}

		// Parse request
		var req graphQLRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		// Verify collection variable
		if req.Variables["collection"] != "org.impactindexer.review.comment" {
			t.Errorf("expected collection org.impactindexer.review.comment, got %v", req.Variables["collection"])
		}

		// Return 2 records
		resp := graphQLResponse{
			Data: &graphQLData{
				Records: &recordsPage{
					Edges: []recordEdge{
						{
							Node: IndexerRecord{
								CID:        "cid1",
								Collection: "org.impactindexer.review.comment",
								DID:        "did:plc:user1",
								RKey:       "rkey1",
								URI:        "at://did:plc:user1/org.impactindexer.review.comment/rkey1",
								Value:      map[string]interface{}{"text": "comment 1"},
							},
						},
						{
							Node: IndexerRecord{
								CID:        "cid2",
								Collection: "org.impactindexer.review.comment",
								DID:        "did:plc:user2",
								RKey:       "rkey2",
								URI:        "at://did:plc:user2/org.impactindexer.review.comment/rkey2",
								Value:      map[string]interface{}{"text": "comment 2"},
							},
						},
					},
					PageInfo: pageInfo{
						HasNextPage: false,
						EndCursor:   nil,
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Test
	records, err := FetchRecordsByCollection(context.Background(), server.URL, "org.impactindexer.review.comment")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(records) != 2 {
		t.Errorf("expected 2 records, got %d", len(records))
	}

	if records[0].DID != "did:plc:user1" {
		t.Errorf("expected first DID to be did:plc:user1, got %s", records[0].DID)
	}

	if records[1].DID != "did:plc:user2" {
		t.Errorf("expected second DID to be did:plc:user2, got %s", records[1].DID)
	}
}

func TestFetchRecordsPagination(t *testing.T) {
	pageCount := 0

	// Mock server that returns 2 pages
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req graphQLRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		var resp graphQLResponse

		// First page
		if req.Variables["after"] == nil {
			pageCount++
			cursor := "cursor1"
			resp = graphQLResponse{
				Data: &graphQLData{
					Records: &recordsPage{
						Edges: []recordEdge{
							{
								Node: IndexerRecord{
									CID:        "cid1",
									Collection: "org.impactindexer.review.comment",
									DID:        "did:plc:page1",
									RKey:       "rkey1",
									URI:        "at://did:plc:page1/org.impactindexer.review.comment/rkey1",
									Value:      map[string]interface{}{"text": "page 1"},
								},
							},
						},
						PageInfo: pageInfo{
							HasNextPage: true,
							EndCursor:   &cursor,
						},
					},
				},
			}
		} else {
			// Second page
			pageCount++
			resp = graphQLResponse{
				Data: &graphQLData{
					Records: &recordsPage{
						Edges: []recordEdge{
							{
								Node: IndexerRecord{
									CID:        "cid2",
									Collection: "org.impactindexer.review.comment",
									DID:        "did:plc:page2",
									RKey:       "rkey2",
									URI:        "at://did:plc:page2/org.impactindexer.review.comment/rkey2",
									Value:      map[string]interface{}{"text": "page 2"},
								},
							},
						},
						PageInfo: pageInfo{
							HasNextPage: false,
							EndCursor:   nil,
						},
					},
				},
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Test
	records, err := FetchRecordsByCollection(context.Background(), server.URL, "org.impactindexer.review.comment")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if pageCount != 2 {
		t.Errorf("expected 2 pages to be fetched, got %d", pageCount)
	}

	if len(records) != 2 {
		t.Errorf("expected 2 total records, got %d", len(records))
	}

	if records[0].DID != "did:plc:page1" {
		t.Errorf("expected first record from page 1, got DID %s", records[0].DID)
	}

	if records[1].DID != "did:plc:page2" {
		t.Errorf("expected second record from page 2, got DID %s", records[1].DID)
	}
}

func TestFetchRecordsError(t *testing.T) {
	// Mock server that returns HTTP 500
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	// Test
	_, err := FetchRecordsByCollection(context.Background(), server.URL, "org.impactindexer.review.comment")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestFetchRecordsGraphQLError(t *testing.T) {
	// Mock server that returns GraphQL error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := graphQLResponse{
			Errors: []graphQLError{
				{Message: "collection not found"},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Test
	_, err := FetchRecordsByCollection(context.Background(), server.URL, "org.impactindexer.review.comment")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestFetchRecordsEmpty(t *testing.T) {
	// Mock server that returns empty edges
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := graphQLResponse{
			Data: &graphQLData{
				Records: &recordsPage{
					Edges: []recordEdge{},
					PageInfo: pageInfo{
						HasNextPage: false,
						EndCursor:   nil,
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Test
	records, err := FetchRecordsByCollection(context.Background(), server.URL, "org.impactindexer.review.comment")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if records == nil {
		t.Fatal("expected non-nil slice, got nil")
	}

	if len(records) != 0 {
		t.Errorf("expected empty slice, got %d records", len(records))
	}
}
