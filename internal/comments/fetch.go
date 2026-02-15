package comments

import (
	"context"
	"sync"
)

// FetchComments fetches all comments for a specific beads issue from the Hypergoat indexer.
// Orchestrates: fetch records (comments + likes in parallel), filter, resolve profiles, assemble, thread, filter by ID.
// Returns threaded comments for the given beadsID, newest-first at root level.
func FetchComments(ctx context.Context, indexerURL, profileAPIURL, beadsID string) ([]BeadsComment, error) {
	// Fetch comment records and like records in parallel
	var commentRecords, likeRecords []IndexerRecord
	var commentErr, likeErr error
	var wg sync.WaitGroup

	wg.Add(2)

	// Fetch comments
	go func() {
		defer wg.Done()
		commentRecords, commentErr = FetchRecordsByCollection(ctx, indexerURL, CommentCollection)
	}()

	// Fetch likes
	go func() {
		defer wg.Done()
		likeRecords, likeErr = FetchRecordsByCollection(ctx, indexerURL, LikeCollection)
	}()

	wg.Wait()

	// If comment fetch fails, return error
	if commentErr != nil {
		return nil, commentErr
	}

	// If like fetch fails, continue with empty likes
	if likeErr != nil {
		likeRecords = []IndexerRecord{}
	}

	// Filter comment records to only beads: URIs
	filteredComments := FilterBeadsComments(commentRecords)

	// Collect unique DIDs from filtered comment records
	didSet := make(map[string]bool)
	for _, record := range filteredComments {
		didSet[record.DID] = true
	}

	dids := make([]string, 0, len(didSet))
	for did := range didSet {
		dids = append(dids, did)
	}

	// Resolve profiles
	profiles := ResolveProfiles(ctx, profileAPIURL, dids)

	// Assemble comments
	assembled := AssembleComments(filteredComments, likeRecords, profiles)

	// Build threads
	threaded := BuildThreads(assembled)

	// Filter by nodeID
	filtered := FilterByNodeID(threaded, beadsID)

	return filtered, nil
}
