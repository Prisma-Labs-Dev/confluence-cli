package confluence

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

type SearchOptions struct {
	CQL    string
	Limit  int
	Cursor string
}

func (c *Client) Search(opts SearchOptions) (*ListResult[SearchResult], error) {
	query := url.Values{}
	query.Set("cql", opts.CQL)
	if opts.Limit > 0 {
		query.Set("limit", strconv.Itoa(opts.Limit))
	}
	if opts.Cursor != "" {
		query.Set("cursor", opts.Cursor)
	}

	// Note: Search uses the v1 API endpoint as v2 doesn't have CQL search
	body, err := c.doV1("GET", "/search", query)
	if err != nil {
		return nil, fmt.Errorf("searching: %w", err)
	}

	var raw struct {
		Results []struct {
			Content struct {
				ID    string `json:"id"`
				Title string `json:"title"`
				Type  string `json:"type"`
				Space *struct {
					ID string `json:"id"`
				} `json:"space,omitempty"`
			} `json:"content"`
			Excerpt string `json:"excerpt"`
			URL     string `json:"url"`
		} `json:"results"`
		Links struct {
			Next string `json:"next"`
		} `json:"_links"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("parsing search response: %w", err)
	}

	results := make([]SearchResult, len(raw.Results))
	for i, r := range raw.Results {
		results[i] = SearchResult{
			ID:      r.Content.ID,
			Title:   r.Content.Title,
			Type:    r.Content.Type,
			Excerpt: r.Excerpt,
			URL:     r.URL,
		}
		if r.Content.Space != nil {
			results[i].SpaceID = r.Content.Space.ID
		}
	}

	return &ListResult[SearchResult]{
		Results:    results,
		NextCursor: extractCursor(raw.Links.Next),
	}, nil
}
