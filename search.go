package confluence

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

type PageSearchOptions struct {
	Query     string
	TitleOnly bool
	SpaceID   string
	Limit     int
	Cursor    string
}

func (c *Client) SearchPages(opts PageSearchOptions) (*ListResult[SearchResult], error) {
	cql, err := buildPageSearchCQL(opts)
	if err != nil {
		return nil, err
	}

	query := url.Values{}
	query.Set("cql", cql)
	if opts.Limit > 0 {
		query.Set("limit", strconv.Itoa(opts.Limit))
	}
	if opts.Cursor != "" {
		query.Set("cursor", opts.Cursor)
	}

	body, err := c.doV1("GET", "/search", query)
	if err != nil {
		return nil, fmt.Errorf("searching pages: %w", err)
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
	for i, result := range raw.Results {
		results[i] = SearchResult{
			ID:      result.Content.ID,
			Title:   result.Content.Title,
			Type:    result.Content.Type,
			Excerpt: result.Excerpt,
			URL:     result.URL,
		}
		if result.Content.Space != nil {
			results[i].SpaceID = result.Content.Space.ID
		}
	}

	return &ListResult[SearchResult]{
		Results:    results,
		NextCursor: extractCursor(raw.Links.Next),
	}, nil
}

func buildPageSearchCQL(opts PageSearchOptions) (string, error) {
	query := strings.TrimSpace(opts.Query)
	if query == "" {
		return "", fmt.Errorf("search query is required")
	}

	field := "text"
	if opts.TitleOnly {
		field = "title"
	}

	cql := fmt.Sprintf("type=page AND %s ~ \"%s\"", field, escapeCQL(query))
	if strings.TrimSpace(opts.SpaceID) != "" {
		cql += fmt.Sprintf(" AND space.id=%s", strings.TrimSpace(opts.SpaceID))
	}
	return cql, nil
}

func escapeCQL(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, `"`, `\\"`)
	return value
}
