package comments

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"sync"
)

// ResolveProfiles resolves Bluesky profiles for a list of DIDs.
// Makes HTTP requests to the Bluesky public API. Falls back to a stub profile
// (DID as handle) on any error. Never returns an error — all failures are graceful.
// The returned map has one entry per unique input DID.
func ResolveProfiles(ctx context.Context, apiURL string, dids []string) map[string]Profile {
	if len(dids) == 0 {
		return make(map[string]Profile)
	}

	// Deduplicate DIDs
	uniqueDIDs := make(map[string]bool)
	for _, did := range dids {
		uniqueDIDs[did] = true
	}

	// Result map
	profiles := make(map[string]Profile)
	var mu sync.Mutex

	// Semaphore to limit concurrent requests to 5
	sem := make(chan struct{}, 5)

	var wg sync.WaitGroup
	for did := range uniqueDIDs {
		wg.Add(1)
		go func(d string) {
			defer wg.Done()

			// Acquire semaphore
			sem <- struct{}{}
			defer func() { <-sem }()

			profile := fetchProfile(ctx, apiURL, d)
			mu.Lock()
			profiles[d] = profile
			mu.Unlock()
		}(did)
	}

	wg.Wait()
	return profiles
}

// fetchProfile fetches a single profile from the Bluesky API.
// Returns a fallback profile (DID as handle) on any error.
func fetchProfile(ctx context.Context, apiURL, did string) Profile {
	u, _ := url.Parse(apiURL + "/xrpc/app.bsky.actor.getProfile")
	q := u.Query()
	q.Set("actor", did)
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return Profile{DID: did, Handle: did}
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return Profile{DID: did, Handle: did}
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return Profile{DID: did, Handle: did}
	}

	var profile Profile
	if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		return Profile{DID: did, Handle: did}
	}

	return profile
}
