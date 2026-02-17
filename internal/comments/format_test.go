package comments

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestFormatText(t *testing.T) {
	comments := []BeadsComment{
		{
			Handle:    "alice.bsky.social",
			Text:      "First comment",
			CreatedAt: "2025-01-15T10:00:00Z",
			Likes:     2,
		},
		{
			Handle:    "bob.bsky.social",
			Text:      "Second comment",
			CreatedAt: "2025-01-15T11:00:00Z",
			Likes:     0,
		},
	}

	var buf bytes.Buffer
	FormatText(&buf, comments)
	output := buf.String()

	// Verify contains @handle
	if !strings.Contains(output, "@alice.bsky.social") {
		t.Errorf("output should contain @alice.bsky.social")
	}
	if !strings.Contains(output, "@bob.bsky.social") {
		t.Errorf("output should contain @bob.bsky.social")
	}

	// Verify contains [2 likes]
	if !strings.Contains(output, "[2 likes]") {
		t.Errorf("output should contain [2 likes]")
	}

	// Verify does NOT contain [0 likes]
	if strings.Contains(output, "[0 likes]") {
		t.Errorf("output should not contain [0 likes]")
	}

	// Verify contains text content
	if !strings.Contains(output, "First comment") {
		t.Errorf("output should contain 'First comment'")
	}
	if !strings.Contains(output, "Second comment") {
		t.Errorf("output should contain 'Second comment'")
	}
}

func TestFormatTextThreaded(t *testing.T) {
	comments := []BeadsComment{
		{
			Handle:    "alice.bsky.social",
			Text:      "Root comment",
			CreatedAt: "2025-01-15T10:00:00Z",
			Replies: []BeadsComment{
				{
					Handle:    "bob.bsky.social",
					Text:      "Reply to root",
					CreatedAt: "2025-01-15T11:00:00Z",
				},
			},
		},
	}

	var buf bytes.Buffer
	FormatText(&buf, comments)
	output := buf.String()

	// Verify reply is indented by 2 additional spaces
	// Root should have no indent, reply should have 2 spaces
	lines := strings.Split(output, "\n")

	// Find the reply line (should start with 2 spaces then [nodeID])
	foundReply := false
	for _, line := range lines {
		if strings.Contains(line, "@bob.bsky.social") {
			if strings.HasPrefix(line, "  [") {
				foundReply = true
				break
			}
		}
	}

	if !foundReply {
		t.Errorf("reply should be indented by 2 spaces, got output:\n%s", output)
	}
}

func TestFormatTextEmpty(t *testing.T) {
	var buf bytes.Buffer
	FormatText(&buf, []BeadsComment{})
	output := buf.String()

	if output != "No comments found.\n" {
		t.Errorf("expected 'No comments found.\\n', got %q", output)
	}
}

func TestFormatTextDisplayName(t *testing.T) {
	comments := []BeadsComment{
		{
			Handle:      "alice.bsky.social",
			DisplayName: "Alice Smith",
			Text:        "Comment with display name",
			CreatedAt:   "2025-01-15T10:00:00Z",
		},
	}

	var buf bytes.Buffer
	FormatText(&buf, comments)
	output := buf.String()

	// Verify displayName appears before @handle
	if !strings.Contains(output, "Alice Smith @alice.bsky.social") {
		t.Errorf("output should contain 'Alice Smith @alice.bsky.social', got:\n%s", output)
	}
}

func TestFormatJSON(t *testing.T) {
	comments := []BeadsComment{
		{
			Handle:    "alice.bsky.social",
			Text:      "Test comment",
			CreatedAt: "2025-01-15T10:00:00Z",
			Likes:     1,
		},
	}

	var buf bytes.Buffer
	err := FormatJSON(&buf, comments)
	if err != nil {
		t.Fatalf("FormatJSON failed: %v", err)
	}

	output := buf.String()

	// Verify valid JSON array
	var parsed []BeadsComment
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Errorf("output is not valid JSON: %v", err)
	}

	if len(parsed) != 1 {
		t.Errorf("expected 1 comment, got %d", len(parsed))
	}

	if parsed[0].Handle != "alice.bsky.social" {
		t.Errorf("expected handle 'alice.bsky.social', got %q", parsed[0].Handle)
	}
}

func TestFormatJSONEmpty(t *testing.T) {
	var buf bytes.Buffer
	err := FormatJSON(&buf, []BeadsComment{})
	if err != nil {
		t.Fatalf("FormatJSON failed: %v", err)
	}

	output := buf.String()

	if output != "[]\n" {
		t.Errorf("expected '[]\\n', got %q", output)
	}
}
