package comments

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// FormatText writes human-readable threaded comments to w.
// Format per comment (indented by depth*2 spaces):
//
//	@handle (2025-01-15T10:00:00Z) [N likes]
//	<2+depth*2 spaces>text content here
//
// Replies are indented by additional 2 spaces per depth level.
// Empty comments list: write "No comments found.\n"
func FormatText(w io.Writer, comments []BeadsComment) {
	if len(comments) == 0 {
		fmt.Fprint(w, "No comments found.\n")
		return
	}

	for i, comment := range comments {
		formatComment(w, comment, 0)
		// Blank line between root comments (not after the last one)
		if i < len(comments)-1 {
			fmt.Fprintln(w)
		}
	}
}

// formatComment recursively formats a comment and its replies
func formatComment(w io.Writer, comment BeadsComment, depth int) {
	indent := strings.Repeat(" ", depth*2)

	// Header line: [nodeID] (↩ reply · ) DisplayName @handle (createdAt)
	replyIndicator := ""
	if comment.ReplyTo != "" {
		replyIndicator = "↩ reply · "
	}

	var header string
	if comment.DisplayName != "" {
		header = fmt.Sprintf("%s[%s] %s%s @%s (%s)", indent, comment.NodeID, replyIndicator, comment.DisplayName, comment.Handle, comment.CreatedAt)
	} else {
		header = fmt.Sprintf("%s[%s] %s@%s (%s)", indent, comment.NodeID, replyIndicator, comment.Handle, comment.CreatedAt)
	}

	// Add likes if > 0
	if comment.Likes > 0 {
		if comment.Likes == 1 {
			header += " [1 like]"
		} else {
			header += fmt.Sprintf(" [%d likes]", comment.Likes)
		}
	}

	fmt.Fprintln(w, header)

	// Text line (indented by 2 more spaces)
	textIndent := strings.Repeat(" ", depth*2+2)
	fmt.Fprintf(w, "%s%s\n", textIndent, comment.Text)

	// Recursively format replies
	for _, reply := range comment.Replies {
		formatComment(w, reply, depth+1)
	}
}

// FormatJSON writes comments as a JSON array to w.
// Uses json.MarshalIndent with 2-space indentation.
// Empty comments: write "[]\n"
func FormatJSON(w io.Writer, comments []BeadsComment) error {
	if len(comments) == 0 {
		_, err := fmt.Fprint(w, "[]\n")
		return err
	}

	data, err := json.MarshalIndent(comments, "", "  ")
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(w, "%s\n", data)
	return err
}
