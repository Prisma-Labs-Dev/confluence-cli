package confluence

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

type ListPagesOptions struct {
	SpaceID string
	Limit   int
	Cursor  string
	Sort    string
}

func (c *Client) ListPages(opts ListPagesOptions) (*ListResult[Page], error) {
	query := url.Values{}
	if opts.SpaceID != "" {
		query.Set("space-id", opts.SpaceID)
	}
	if opts.Limit > 0 {
		query.Set("limit", strconv.Itoa(opts.Limit))
	}
	if opts.Cursor != "" {
		query.Set("cursor", opts.Cursor)
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	body, err := c.do("GET", "/pages", query)
	if err != nil {
		return nil, fmt.Errorf("listing pages: %w", err)
	}

	var raw paginatedResponse
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("parsing pages response: %w", err)
	}

	var pages []Page
	if err := json.Unmarshal(raw.Results, &pages); err != nil {
		return nil, fmt.Errorf("parsing pages: %w", err)
	}

	return &ListResult[Page]{
		Results:    pages,
		NextCursor: extractCursor(raw.Links.Next),
	}, nil
}

type GetPageOptions struct {
	PageID     string
	BodyFormat string // "view", "storage", "atlas_doc_format"
}

func (c *Client) GetPage(opts GetPageOptions) (*Page, error) {
	query := url.Values{}
	if opts.BodyFormat != "" {
		query.Set("body-format", opts.BodyFormat)
	}

	body, err := c.do("GET", "/pages/"+opts.PageID, query)
	if err != nil {
		return nil, fmt.Errorf("getting page: %w", err)
	}

	var page Page
	if err := json.Unmarshal(body, &page); err != nil {
		return nil, fmt.Errorf("parsing page: %w", err)
	}

	return &page, nil
}

type GetPageChildrenOptions struct {
	PageID string
	Limit  int
	Cursor string
}

func (c *Client) GetPageChildren(opts GetPageChildrenOptions) (*ListResult[Page], error) {
	query := url.Values{}
	if opts.Limit > 0 {
		query.Set("limit", strconv.Itoa(opts.Limit))
	}
	if opts.Cursor != "" {
		query.Set("cursor", opts.Cursor)
	}

	body, err := c.do("GET", "/pages/"+opts.PageID+"/children", query)
	if err != nil {
		return nil, fmt.Errorf("getting page children: %w", err)
	}

	var raw paginatedResponse
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("parsing children response: %w", err)
	}

	var pages []Page
	if err := json.Unmarshal(raw.Results, &pages); err != nil {
		return nil, fmt.Errorf("parsing children: %w", err)
	}

	return &ListResult[Page]{
		Results:    pages,
		NextCursor: extractCursor(raw.Links.Next),
	}, nil
}
