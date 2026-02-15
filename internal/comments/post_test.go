package comments

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bluesky-social/indigo/atproto/atclient"
)

// TestCreateComment verifies that CreateComment sends the correct request body
func TestCreateComment(t *testing.T) {
	var receivedReq createRecordRequest

	// Mock PDS server
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/xrpc/com.atproto.repo.createRecord" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		// Decode request body
		if err := json.NewDecoder(r.Body).Decode(&receivedReq); err != nil {
			t.Errorf("failed to decode request: %v", err)
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		// Return success response
		resp := createRecordResponse{
			URI: "at://did:plc:test123/org.impactindexer.review.comment/abc123",
			CID: "bafytest123",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	// Create test client
	client := atclient.NewAPIClient(srv.URL)

	// Call CreateComment
	input := CreateCommentInput{
		BeadsID: "test-issue-1",
		Text:    "This is a test comment",
		ReplyTo: "",
	}
	output, err := CreateComment(context.Background(), client, "did:plc:test123", input)
	if err != nil {
		t.Fatalf("CreateComment failed: %v", err)
	}

	// Verify output
	if output.URI != "at://did:plc:test123/org.impactindexer.review.comment/abc123" {
		t.Errorf("unexpected URI: %s", output.URI)
	}
	if output.CID != "bafytest123" {
		t.Errorf("unexpected CID: %s", output.CID)
	}

	// Verify request body
	if receivedReq.Repo != "did:plc:test123" {
		t.Errorf("unexpected repo: %s", receivedReq.Repo)
	}
	if receivedReq.Collection != CommentCollection {
		t.Errorf("unexpected collection: %s", receivedReq.Collection)
	}
	if receivedReq.Record.Type != CommentCollection {
		t.Errorf("unexpected record type: %s", receivedReq.Record.Type)
	}
	if receivedReq.Record.Subject.URI != "beads:test-issue-1" {
		t.Errorf("unexpected subject URI: %s", receivedReq.Record.Subject.URI)
	}
	if receivedReq.Record.Subject.Type != "record" {
		t.Errorf("unexpected subject type: %s", receivedReq.Record.Subject.Type)
	}
	if receivedReq.Record.Text != "This is a test comment" {
		t.Errorf("unexpected text: %s", receivedReq.Record.Text)
	}
	if receivedReq.Record.CreatedAt == "" {
		t.Error("createdAt should not be empty")
	}
}

// TestCreateCommentReply verifies that replyTo field is included when set
func TestCreateCommentReply(t *testing.T) {
	var receivedReq createRecordRequest

	// Mock PDS server
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&receivedReq); err != nil {
			t.Errorf("failed to decode request: %v", err)
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		resp := createRecordResponse{
			URI: "at://did:plc:test123/org.impactindexer.review.comment/reply456",
			CID: "bafyreply456",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	client := atclient.NewAPIClient(srv.URL)

	// Call CreateComment with ReplyTo set
	input := CreateCommentInput{
		BeadsID: "test-issue-1",
		Text:    "This is a reply",
		ReplyTo: "at://did:plc:other/org.impactindexer.review.comment/parent123",
	}
	output, err := CreateComment(context.Background(), client, "did:plc:test123", input)
	if err != nil {
		t.Fatalf("CreateComment failed: %v", err)
	}

	// Verify replyTo is present
	if receivedReq.Record.ReplyTo != "at://did:plc:other/org.impactindexer.review.comment/parent123" {
		t.Errorf("unexpected replyTo: %s", receivedReq.Record.ReplyTo)
	}

	// Verify output
	if output.URI != "at://did:plc:test123/org.impactindexer.review.comment/reply456" {
		t.Errorf("unexpected URI: %s", output.URI)
	}
}

// TestCreateCommentNoReplyTo verifies that replyTo field is absent when not set
func TestCreateCommentNoReplyTo(t *testing.T) {
	var receivedReq map[string]interface{}

	// Mock PDS server
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&receivedReq); err != nil {
			t.Errorf("failed to decode request: %v", err)
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		resp := createRecordResponse{
			URI: "at://did:plc:test123/org.impactindexer.review.comment/abc123",
			CID: "bafytest123",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	client := atclient.NewAPIClient(srv.URL)

	// Call CreateComment without ReplyTo
	input := CreateCommentInput{
		BeadsID: "test-issue-1",
		Text:    "This is a test comment",
		ReplyTo: "",
	}
	_, err := CreateComment(context.Background(), client, "did:plc:test123", input)
	if err != nil {
		t.Fatalf("CreateComment failed: %v", err)
	}

	// Verify replyTo field is absent (not just empty string)
	record, ok := receivedReq["record"].(map[string]interface{})
	if !ok {
		t.Fatal("record field not found or not a map")
	}
	if _, exists := record["replyTo"]; exists {
		t.Error("replyTo field should be absent when not set, but it exists")
	}
}

// TestCreateCommentError verifies that errors are propagated
func TestCreateCommentError(t *testing.T) {
	// Mock PDS server that returns an error
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}))
	defer srv.Close()

	client := atclient.NewAPIClient(srv.URL)

	// Call CreateComment
	input := CreateCommentInput{
		BeadsID: "test-issue-1",
		Text:    "This should fail",
		ReplyTo: "",
	}
	_, err := CreateComment(context.Background(), client, "did:plc:test123", input)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
