package comments

// DefaultIndexerURL is the default Hypergoat GraphQL indexer endpoint.
const DefaultIndexerURL = "https://hypergoat-app-production.up.railway.app/graphql"

// DefaultProfileAPIURL is the default Bluesky public API endpoint for profile resolution.
const DefaultProfileAPIURL = "https://public.api.bsky.app"

// CommentCollection is the ATProto lexicon for beads review comments.
const CommentCollection = "org.impactindexer.review.comment"

// LikeCollection is the ATProto lexicon for beads review likes.
const LikeCollection = "org.impactindexer.review.like"

// BeadsURIPrefix is the prefix used in comment subject URIs to target beads issues.
const BeadsURIPrefix = "beads:"

// IndexerRecord represents a record returned by the Hypergoat GraphQL indexer.
type IndexerRecord struct {
	CID        string                 `json:"cid"`
	Collection string                 `json:"collection"`
	DID        string                 `json:"did"`
	RKey       string                 `json:"rkey"`
	URI        string                 `json:"uri"`
	Value      map[string]interface{} `json:"value"`
}

// Profile represents a resolved Bluesky profile.
type Profile struct {
	DID         string `json:"did"`
	Handle      string `json:"handle"`
	DisplayName string `json:"displayName,omitempty"`
	Avatar      string `json:"avatar,omitempty"`
}

// BeadsComment represents a processed, threaded comment on a beads issue.
type BeadsComment struct {
	DID         string         `json:"did"`
	Handle      string         `json:"handle"`
	DisplayName string         `json:"displayName,omitempty"`
	Text        string         `json:"text"`
	CreatedAt   string         `json:"createdAt"`
	URI         string         `json:"uri"`
	RKey        string         `json:"rkey"`
	NodeID      string         `json:"nodeId"`
	ReplyTo     string         `json:"replyTo,omitempty"`
	Likes       int            `json:"likes"`
	Replies     []BeadsComment `json:"replies,omitempty"`
}

// FetchOptions controls filtering and limiting of fetched comments.
type FetchOptions struct {
	BeadsID string // exact nodeID match (empty = no exact filter)
	Pattern string // glob pattern match (empty = no pattern filter)
	Limit   int    // max root comments to return (0 = unlimited)
}
