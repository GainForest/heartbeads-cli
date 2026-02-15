package comments

import (
	"testing"
)

func TestFilterBeadsComments(t *testing.T) {
	records := []IndexerRecord{
		{
			URI: "at://did:plc:abc/org.impactindexer.review.comment/1",
			Value: map[string]interface{}{
				"subject": map[string]interface{}{
					"uri": "beads:test-id-1",
				},
			},
		},
		{
			URI: "at://did:plc:abc/org.impactindexer.review.comment/2",
			Value: map[string]interface{}{
				"subject": map[string]interface{}{
					"uri": "beads:test-id-2",
				},
			},
		},
		{
			URI: "at://did:plc:abc/org.impactindexer.review.comment/3",
			Value: map[string]interface{}{
				"subject": map[string]interface{}{
					"uri": "at://some-other-uri",
				},
			},
		},
	}

	filtered := FilterBeadsComments(records)

	if len(filtered) != 2 {
		t.Errorf("expected 2 filtered records, got %d", len(filtered))
	}

	for _, record := range filtered {
		subject := record.Value["subject"].(map[string]interface{})
		uri := subject["uri"].(string)
		if uri != "beads:test-id-1" && uri != "beads:test-id-2" {
			t.Errorf("unexpected URI in filtered results: %s", uri)
		}
	}
}

func TestFilterBeadsCommentsEmpty(t *testing.T) {
	records := []IndexerRecord{}
	filtered := FilterBeadsComments(records)

	if filtered == nil {
		t.Error("expected non-nil slice, got nil")
	}

	if len(filtered) != 0 {
		t.Errorf("expected empty slice, got %d records", len(filtered))
	}
}

func TestExtractNodeID(t *testing.T) {
	record := IndexerRecord{
		Value: map[string]interface{}{
			"subject": map[string]interface{}{
				"uri": "beads:test-id",
			},
		},
	}

	nodeID := ExtractNodeID(record)

	if nodeID != "test-id" {
		t.Errorf("expected 'test-id', got '%s'", nodeID)
	}
}

func TestExtractNodeIDInvalid(t *testing.T) {
	// Test with no subject
	record1 := IndexerRecord{
		Value: map[string]interface{}{},
	}
	if nodeID := ExtractNodeID(record1); nodeID != "" {
		t.Errorf("expected empty string for no subject, got '%s'", nodeID)
	}

	// Test with non-beads URI
	record2 := IndexerRecord{
		Value: map[string]interface{}{
			"subject": map[string]interface{}{
				"uri": "at://some-other-uri",
			},
		},
	}
	if nodeID := ExtractNodeID(record2); nodeID != "at://some-other-uri" {
		t.Errorf("expected 'at://some-other-uri', got '%s'", nodeID)
	}
}

func TestAssembleComments(t *testing.T) {
	commentRecords := []IndexerRecord{
		{
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
		{
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
	}

	likeRecords := []IndexerRecord{
		{
			Value: map[string]interface{}{
				"subject": map[string]interface{}{
					"uri": "at://did:plc:alice/org.impactindexer.review.comment/1",
				},
			},
		},
	}

	profiles := map[string]Profile{
		"did:plc:alice": {DID: "did:plc:alice", Handle: "alice.bsky.social"},
		"did:plc:bob":   {DID: "did:plc:bob", Handle: "bob.bsky.social"},
	}

	comments := AssembleComments(commentRecords, likeRecords, profiles)

	if len(comments) != 2 {
		t.Fatalf("expected 2 comments, got %d", len(comments))
	}

	// Check first comment has 1 like
	if comments[0].Likes != 1 {
		t.Errorf("expected first comment to have 1 like, got %d", comments[0].Likes)
	}

	// Check second comment has 0 likes
	if comments[1].Likes != 0 {
		t.Errorf("expected second comment to have 0 likes, got %d", comments[1].Likes)
	}

	// Check handles
	if comments[0].Handle != "alice.bsky.social" {
		t.Errorf("expected handle 'alice.bsky.social', got '%s'", comments[0].Handle)
	}
	if comments[1].Handle != "bob.bsky.social" {
		t.Errorf("expected handle 'bob.bsky.social', got '%s'", comments[1].Handle)
	}
}

func TestBuildThreads(t *testing.T) {
	comments := []BeadsComment{
		{
			URI:       "at://root",
			Text:      "Root comment",
			CreatedAt: "2025-01-15T10:00:00Z",
			Replies:   []BeadsComment{},
		},
		{
			URI:       "at://reply1",
			Text:      "Reply to root",
			CreatedAt: "2025-01-15T11:00:00Z",
			ReplyTo:   "at://root",
			Replies:   []BeadsComment{},
		},
		{
			URI:       "at://reply2",
			Text:      "Reply to reply",
			CreatedAt: "2025-01-15T12:00:00Z",
			ReplyTo:   "at://reply1",
			Replies:   []BeadsComment{},
		},
	}

	threaded := BuildThreads(comments)

	// Should have 1 root comment
	if len(threaded) != 1 {
		t.Fatalf("expected 1 root comment, got %d", len(threaded))
	}

	root := threaded[0]
	if root.URI != "at://root" {
		t.Errorf("expected root URI 'at://root', got '%s'", root.URI)
	}

	// Root should have 1 reply
	if len(root.Replies) != 1 {
		t.Fatalf("expected root to have 1 reply, got %d", len(root.Replies))
	}

	reply1 := root.Replies[0]
	if reply1.URI != "at://reply1" {
		t.Errorf("expected reply URI 'at://reply1', got '%s'", reply1.URI)
	}

	// Reply1 should have 1 reply
	if len(reply1.Replies) != 1 {
		t.Fatalf("expected reply1 to have 1 reply, got %d", len(reply1.Replies))
	}

	reply2 := reply1.Replies[0]
	if reply2.URI != "at://reply2" {
		t.Errorf("expected reply2 URI 'at://reply2', got '%s'", reply2.URI)
	}
}

func TestBuildThreadsSortOrder(t *testing.T) {
	comments := []BeadsComment{
		{
			URI:       "at://root1",
			Text:      "Older root",
			CreatedAt: "2025-01-15T10:00:00Z",
			Replies:   []BeadsComment{},
		},
		{
			URI:       "at://root2",
			Text:      "Newer root",
			CreatedAt: "2025-01-15T12:00:00Z",
			Replies:   []BeadsComment{},
		},
		{
			URI:       "at://reply1",
			Text:      "Newer reply",
			CreatedAt: "2025-01-15T11:30:00Z",
			ReplyTo:   "at://root2",
			Replies:   []BeadsComment{},
		},
		{
			URI:       "at://reply2",
			Text:      "Older reply",
			CreatedAt: "2025-01-15T11:00:00Z",
			ReplyTo:   "at://root2",
			Replies:   []BeadsComment{},
		},
	}

	threaded := BuildThreads(comments)

	// Should have 2 root comments
	if len(threaded) != 2 {
		t.Fatalf("expected 2 root comments, got %d", len(threaded))
	}

	// Root comments should be sorted newest-first
	if threaded[0].URI != "at://root2" {
		t.Errorf("expected first root to be 'at://root2', got '%s'", threaded[0].URI)
	}
	if threaded[1].URI != "at://root1" {
		t.Errorf("expected second root to be 'at://root1', got '%s'", threaded[1].URI)
	}

	// Replies should be sorted oldest-first
	root2 := threaded[0]
	if len(root2.Replies) != 2 {
		t.Fatalf("expected root2 to have 2 replies, got %d", len(root2.Replies))
	}

	if root2.Replies[0].URI != "at://reply2" {
		t.Errorf("expected first reply to be 'at://reply2', got '%s'", root2.Replies[0].URI)
	}
	if root2.Replies[1].URI != "at://reply1" {
		t.Errorf("expected second reply to be 'at://reply1', got '%s'", root2.Replies[1].URI)
	}
}

func TestBuildThreadsOrphanReply(t *testing.T) {
	comments := []BeadsComment{
		{
			URI:       "at://orphan",
			Text:      "Orphan reply",
			CreatedAt: "2025-01-15T10:00:00Z",
			ReplyTo:   "at://nonexistent",
			Replies:   []BeadsComment{},
		},
	}

	threaded := BuildThreads(comments)

	// Orphan should be treated as root
	if len(threaded) != 1 {
		t.Fatalf("expected 1 root comment (orphan), got %d", len(threaded))
	}

	if threaded[0].URI != "at://orphan" {
		t.Errorf("expected orphan URI 'at://orphan', got '%s'", threaded[0].URI)
	}
}

func TestFilterByNodeID(t *testing.T) {
	comments := []BeadsComment{
		{NodeID: "test-id-1", Text: "Comment 1"},
		{NodeID: "test-id-2", Text: "Comment 2"},
		{NodeID: "test-id-1", Text: "Comment 3"},
	}

	filtered := FilterByNodeID(comments, "test-id-1")

	if len(filtered) != 2 {
		t.Fatalf("expected 2 filtered comments, got %d", len(filtered))
	}

	for _, comment := range filtered {
		if comment.NodeID != "test-id-1" {
			t.Errorf("expected nodeID 'test-id-1', got '%s'", comment.NodeID)
		}
	}
}
