package comments

import (
	"context"
	"time"

	"github.com/bluesky-social/indigo/atproto/atclient"
	"github.com/bluesky-social/indigo/atproto/syntax"
)

// CreateCommentInput holds the parameters for creating a comment.
type CreateCommentInput struct {
	BeadsID string // The beads issue ID to comment on
	Text    string // Comment text
	ReplyTo string // Optional AT-URI of parent comment (for replies)
}

// CreateCommentOutput holds the result of creating a comment.
type CreateCommentOutput struct {
	URI string `json:"uri"`
	CID string `json:"cid"`
}

// createRecordRequest is the request body for com.atproto.repo.createRecord
type createRecordRequest struct {
	Repo       string        `json:"repo"`
	Collection string        `json:"collection"`
	Record     commentRecord `json:"record"`
}

// commentRecord is the record structure for a beads comment
type commentRecord struct {
	Type      string         `json:"$type"`
	Subject   commentSubject `json:"subject"`
	Text      string         `json:"text"`
	CreatedAt string         `json:"createdAt"`
	ReplyTo   string         `json:"replyTo,omitempty"`
}

// commentSubject identifies what the comment is about
type commentSubject struct {
	URI  string `json:"uri"`
	Type string `json:"type"`
}

// createRecordResponse is the response from com.atproto.repo.createRecord
type createRecordResponse struct {
	URI string `json:"uri"`
	CID string `json:"cid"`
}

// CreateComment posts a comment to the authenticated user's ATProto repo.
// Uses atclient.APIClient.Post() to call com.atproto.repo.createRecord.
// Requires an authenticated client (from auth.LoadClient).
func CreateComment(ctx context.Context, client *atclient.APIClient, did string, input CreateCommentInput) (*CreateCommentOutput, error) {
	// Build the request body
	reqBody := createRecordRequest{
		Repo:       did,
		Collection: CommentCollection,
		Record: commentRecord{
			Type: CommentCollection,
			Subject: commentSubject{
				URI:  BeadsURIPrefix + input.BeadsID,
				Type: "record",
			},
			Text:      input.Text,
			CreatedAt: time.Now().UTC().Format(time.RFC3339),
			ReplyTo:   input.ReplyTo,
		},
	}

	// Call com.atproto.repo.createRecord
	var output createRecordResponse
	err := client.Post(ctx, syntax.NSID("com.atproto.repo.createRecord"), reqBody, &output)
	if err != nil {
		return nil, err
	}

	return &CreateCommentOutput{URI: output.URI, CID: output.CID}, nil
}
