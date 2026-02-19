package comments

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// graphQLRequest represents a GraphQL request payload.
type graphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}

// graphQLResponse represents a GraphQL response.
type graphQLResponse struct {
	Data   *graphQLData   `json:"data"`
	Errors []graphQLError `json:"errors"`
}

// graphQLData represents the data field in a GraphQL response.
type graphQLData struct {
	Records *recordsPage `json:"records"`
}

// recordsPage represents a paginated list of records.
type recordsPage struct {
	Edges    []recordEdge `json:"edges"`
	PageInfo pageInfo     `json:"pageInfo"`
}

// recordEdge represents a single record edge in the GraphQL response.
type recordEdge struct {
	Node IndexerRecord `json:"node"`
}

// pageInfo represents pagination information.
type pageInfo struct {
	HasNextPage bool    `json:"hasNextPage"`
	EndCursor   *string `json:"endCursor"`
}

// graphQLError represents a GraphQL error.
type graphQLError struct {
	Message string `json:"message"`
}

const graphQLQuery = `query FetchRecords($collection: String!, $first: Int, $after: String) {
  records(collection: $collection, first: $first, after: $after) {
    edges {
      node {
        cid
        collection
        did
        rkey
        uri
        value
      }
    }
    pageInfo {
      hasNextPage
      endCursor
    }
  }
}`

// fetchPage fetches a single page of records from the GraphQL indexer.
func fetchPage(ctx context.Context, indexerURL string, variables map[string]interface{}) (*graphQLResponse, error) {
	// Create GraphQL request
	reqBody := graphQLRequest{
		Query:     graphQLQuery,
		Variables: variables,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", indexerURL, bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Check HTTP status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Parse response
	var gqlResp graphQLResponse
	if err := json.NewDecoder(resp.Body).Decode(&gqlResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Check for GraphQL errors
	if len(gqlResp.Errors) > 0 {
		return nil, fmt.Errorf("graphql error: %s", gqlResp.Errors[0].Message)
	}

	return &gqlResp, nil
}

// FetchRecordsByCollection queries the Hypergoat GraphQL indexer for all records
// in the given collection. Paginates automatically (100 per page, max 5 pages).
func FetchRecordsByCollection(ctx context.Context, indexerURL, collection string) ([]IndexerRecord, error) {
	allRecords := make([]IndexerRecord, 0)
	var cursor *string
	const maxPages = 5
	const pageSize = 100

	for page := 0; page < maxPages; page++ {
		// Build request variables
		variables := map[string]interface{}{
			"collection": collection,
			"first":      pageSize,
		}
		if cursor != nil {
			variables["after"] = *cursor
		}

		// Fetch page
		gqlResp, err := fetchPage(ctx, indexerURL, variables)
		if err != nil {
			return nil, err
		}

		// Extract records
		if gqlResp.Data == nil || gqlResp.Data.Records == nil {
			break
		}

		for _, edge := range gqlResp.Data.Records.Edges {
			allRecords = append(allRecords, edge.Node)
		}

		// Check if there are more pages
		if !gqlResp.Data.Records.PageInfo.HasNextPage {
			break
		}

		cursor = gqlResp.Data.Records.PageInfo.EndCursor
		if cursor == nil {
			break
		}
	}

	return allRecords, nil
}
