package comments

import (
	"sort"
	"strings"
)

// FilterBeadsComments returns only records whose value.subject.uri starts with "beads:".
func FilterBeadsComments(records []IndexerRecord) []IndexerRecord {
	filtered := make([]IndexerRecord, 0)
	for _, record := range records {
		// Check if value["subject"] exists and is a map
		subject, ok := record.Value["subject"].(map[string]interface{})
		if !ok {
			continue
		}

		// Check if subject["uri"] exists and is a string
		uri, ok := subject["uri"].(string)
		if !ok {
			continue
		}

		// Check if uri starts with "beads:"
		if strings.HasPrefix(uri, BeadsURIPrefix) {
			filtered = append(filtered, record)
		}
	}
	return filtered
}

// ExtractNodeID extracts the beads issue ID from a record's value.subject.uri field.
// For subject.uri="beads:my-issue-123", returns "my-issue-123".
// Returns empty string if the record has no valid beads subject URI.
func ExtractNodeID(record IndexerRecord) string {
	// Try to access value["subject"]
	subject, ok := record.Value["subject"].(map[string]interface{})
	if !ok {
		return ""
	}

	// Try to access subject["uri"]
	uri, ok := subject["uri"].(string)
	if !ok {
		return ""
	}

	// Strip "beads:" prefix
	return strings.TrimPrefix(uri, BeadsURIPrefix)
}

// AssembleComments converts raw IndexerRecords into BeadsComment structs.
// Attaches like counts and resolved profile info.
// Parameters:
//   - commentRecords: filtered comment records (already filtered to beads: URIs)
//   - likeRecords: all like records (will be matched by comment URI)
//   - profiles: resolved profiles (DID -> Profile map)
//
// Returns a flat list of BeadsComment (not yet threaded).
func AssembleComments(commentRecords, likeRecords []IndexerRecord, profiles map[string]Profile) []BeadsComment {
	// Build a map of comment URI -> count of likes
	likeCounts := make(map[string]int)
	for _, likeRecord := range likeRecords {
		subject, ok := likeRecord.Value["subject"].(map[string]interface{})
		if !ok {
			continue
		}
		uri, ok := subject["uri"].(string)
		if !ok {
			continue
		}
		likeCounts[uri]++
	}

	// Build BeadsComment list
	comments := make([]BeadsComment, 0, len(commentRecords))
	for _, record := range commentRecords {
		nodeID := ExtractNodeID(record)

		// Extract text
		text, _ := record.Value["text"].(string)

		// Extract createdAt
		createdAt, _ := record.Value["createdAt"].(string)

		// Extract replyTo (optional)
		replyTo, _ := record.Value["replyTo"].(string)

		// Look up profile
		profile, ok := profiles[record.DID]
		if !ok {
			// Fallback: use DID as handle
			profile = Profile{
				DID:    record.DID,
				Handle: record.DID,
			}
		}

		// Get like count for this comment's URI
		likes := likeCounts[record.URI]

		comment := BeadsComment{
			DID:         record.DID,
			Handle:      profile.Handle,
			DisplayName: profile.DisplayName,
			Text:        text,
			CreatedAt:   createdAt,
			URI:         record.URI,
			RKey:        record.RKey,
			NodeID:      nodeID,
			ReplyTo:     replyTo,
			Likes:       likes,
			Replies:     make([]BeadsComment, 0),
		}
		comments = append(comments, comment)
	}

	return comments
}

// BuildThreads takes a flat list of BeadsComment and builds threaded trees.
// Comments with replyTo pointing to another comment's URI become children (nested in Replies).
// Root comments (no replyTo, or orphaned replyTo) are sorted newest-first by CreatedAt.
// Replies within each thread are sorted oldest-first (chronological).
// Returns only root-level comments (replies are nested inside).
func BuildThreads(comments []BeadsComment) []BeadsComment {
	// Index all comments by URI
	commentsByURI := make(map[string]*BeadsComment)
	for i := range comments {
		commentsByURI[comments[i].URI] = &comments[i]
	}

	// Separate root comments and replies
	roots := make([]*BeadsComment, 0)
	for i := range comments {
		comment := &comments[i]
		if comment.ReplyTo == "" {
			// No replyTo -> root comment
			roots = append(roots, comment)
		} else {
			// Has replyTo -> try to attach to parent
			parent, exists := commentsByURI[comment.ReplyTo]
			if exists {
				parent.Replies = append(parent.Replies, *comment)
				// Update the map to point to the reply in the parent's Replies slice
				// so that nested replies can find their parent
				commentsByURI[comment.URI] = &parent.Replies[len(parent.Replies)-1]
			} else {
				// Orphaned reply -> treat as root
				roots = append(roots, comment)
			}
		}
	}

	// Sort root comments: newest first (descending by CreatedAt)
	sort.Slice(roots, func(i, j int) bool {
		return roots[i].CreatedAt > roots[j].CreatedAt
	})

	// Sort replies recursively: oldest first (ascending by CreatedAt)
	for _, root := range roots {
		sortReplies(root)
	}

	// Convert pointers back to values
	result := make([]BeadsComment, len(roots))
	for i, root := range roots {
		result[i] = *root
	}

	return result
}

// sortReplies recursively sorts replies oldest-first (ascending by CreatedAt).
func sortReplies(comment *BeadsComment) {
	if len(comment.Replies) == 0 {
		return
	}

	// Sort replies oldest-first
	sort.Slice(comment.Replies, func(i, j int) bool {
		return comment.Replies[i].CreatedAt < comment.Replies[j].CreatedAt
	})

	// Recursively sort nested replies
	for i := range comment.Replies {
		sortReplies(&comment.Replies[i])
	}
}

// FilterByNodeID returns only comments (root-level) whose NodeID matches the given ID.
func FilterByNodeID(comments []BeadsComment, nodeID string) []BeadsComment {
	filtered := make([]BeadsComment, 0)
	for _, comment := range comments {
		if comment.NodeID == nodeID {
			filtered = append(filtered, comment)
		}
	}
	return filtered
}
