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
	CQL       string
	TitleOnly bool
	SpaceID   string
	SpaceKey  string
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
	rawCQL := strings.TrimSpace(opts.CQL)
	query := strings.TrimSpace(opts.Query)
	spaceID := strings.TrimSpace(opts.SpaceID)
	spaceKey := strings.TrimSpace(opts.SpaceKey)

	if rawCQL != "" {
		if query != "" {
			return "", fmt.Errorf("provide exactly one of search query or cql")
		}
		if opts.TitleOnly {
			return "", fmt.Errorf("title-only cannot be combined with raw cql")
		}
		if spaceID != "" || spaceKey != "" {
			return "", fmt.Errorf("space-id and space-key cannot be combined with raw cql")
		}
		return rawCQL, nil
	}

	if query == "" {
		return "", fmt.Errorf("search query or cql is required")
	}
	if spaceID != "" && spaceKey != "" {
		return "", fmt.Errorf("space-id and space-key are mutually exclusive")
	}

	field := "text"
	if opts.TitleOnly {
		field = "title"
	}

	cql := fmt.Sprintf("type=page AND %s ~ \"%s\"", field, escapeCQL(query))
	if spaceID != "" {
		cql += fmt.Sprintf(" AND space.id=%s", spaceID)
	}
	if spaceKey != "" {
		cql += fmt.Sprintf(" AND space=\"%s\"", escapeCQL(spaceKey))
	}
	return cql, nil
}

func escapeCQL(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, `"`, `\\"`)
	return value
}
