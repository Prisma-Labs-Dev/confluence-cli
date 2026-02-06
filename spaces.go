package confluence

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

type ListSpacesOptions struct {
	Limit  int
	Cursor string
}

func (c *Client) ListSpaces(opts ListSpacesOptions) (*ListResult[Space], error) {
	query := url.Values{}
	if opts.Limit > 0 {
		query.Set("limit", strconv.Itoa(opts.Limit))
	}
	if opts.Cursor != "" {
		query.Set("cursor", opts.Cursor)
	}

	body, err := c.do("GET", "/spaces", query)
	if err != nil {
		return nil, fmt.Errorf("listing spaces: %w", err)
	}

	var raw paginatedResponse
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("parsing spaces response: %w", err)
	}

	var spaces []Space
	if err := json.Unmarshal(raw.Results, &spaces); err != nil {
		return nil, fmt.Errorf("parsing spaces: %w", err)
	}

	return &ListResult[Space]{
		Results:    spaces,
		NextCursor: extractCursor(raw.Links.Next),
	}, nil
}
