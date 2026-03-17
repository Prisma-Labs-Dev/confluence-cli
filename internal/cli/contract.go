package cli

import (
	"time"

	confluence "github.com/Prisma-Labs-Dev/confluence-cli"
)

const (
	defaultListLimit         = 10
	maxListLimit             = 100
	defaultTreeDepth         = 1
	maxTreeDepth             = 5
	defaultTreeLimitPerLevel = 10
	maxTreeLimitPerLevel     = 25
)

// Schema describes the stable CLI-owned shape of successful JSON output.
type Schema struct {
	ItemType string   `json:"itemType"`
	Fields   []string `json:"fields"`
}

// PageWindow describes list pagination state owned by the CLI contract.
type PageWindow struct {
	Limit      int    `json:"limit"`
	NextCursor string `json:"nextCursor,omitempty"`
}

// ListEnvelope is the default success contract for paginated commands.
type ListEnvelope[T any] struct {
	Results []T        `json:"results"`
	Page    PageWindow `json:"page"`
	Schema  Schema     `json:"schema"`
}

// ItemEnvelope is the default success contract for single-object commands.
type ItemEnvelope[T any] struct {
	Item   T      `json:"item"`
	Schema Schema `json:"schema"`
}

// VersionInfo is the CLI-owned version payload.
type VersionInfo struct {
	Version string `json:"version"`
}

// AuthLoginInfo is the CLI-owned auth login payload.
type AuthLoginInfo struct {
	StoredIn string `json:"storedIn"`
}

// SpaceSummary is the CLI-owned list shape for spaces.
type SpaceSummary struct {
	ID     string `json:"id"`
	Key    string `json:"key"`
	Name   string `json:"name"`
	Type   string `json:"type"`
	Status string `json:"status"`
}

// PageSummary is the CLI-owned list shape for pages.
type PageSummary struct {
	ID            string `json:"id"`
	Title         string `json:"title"`
	SpaceID       string `json:"spaceId"`
	Status        string `json:"status"`
	ParentID      string `json:"parentId,omitempty"`
	VersionNumber int    `json:"versionNumber,omitempty"`
}

// PageVersionInfo is the CLI-owned version metadata for page detail.
type PageVersionInfo struct {
	Number    int       `json:"number"`
	Message   string    `json:"message,omitempty"`
	CreatedAt time.Time `json:"createdAt,omitempty"`
	AuthorID  string    `json:"authorId,omitempty"`
}

// PageBody is the CLI-owned body contract for page detail.
type PageBody struct {
	Format string `json:"format"`
	Value  string `json:"value"`
}

// PageDetail is the CLI-owned detail shape for pages.
type PageDetail struct {
	ID         string           `json:"id"`
	Title      string           `json:"title"`
	SpaceID    string           `json:"spaceId"`
	Status     string           `json:"status"`
	ParentID   string           `json:"parentId,omitempty"`
	ParentType string           `json:"parentType,omitempty"`
	AuthorID   string           `json:"authorId,omitempty"`
	CreatedAt  time.Time        `json:"createdAt,omitempty"`
	Version    *PageVersionInfo `json:"version,omitempty"`
	Body       *PageBody        `json:"body,omitempty"`
}

// SearchSummary is the CLI-owned page search shape.
type SearchSummary struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	SpaceID string `json:"spaceId,omitempty"`
	Excerpt string `json:"excerpt,omitempty"`
	URL     string `json:"url,omitempty"`
}

// PageTreeNode is the CLI-owned bounded tree shape.
type PageTreeNode struct {
	ID              string         `json:"id"`
	Title           string         `json:"title"`
	SpaceID         string         `json:"spaceId"`
	Status          string         `json:"status"`
	HasMoreChildren bool           `json:"hasMoreChildren,omitempty"`
	Children        []PageTreeNode `json:"children,omitempty"`
}

// PageTree is the CLI-owned root payload for pages tree.
type PageTree struct {
	RootPageID      string         `json:"rootPageId"`
	Depth           int            `json:"depth"`
	LimitPerLevel   int            `json:"limitPerLevel"`
	HasMoreChildren bool           `json:"hasMoreChildren,omitempty"`
	Children        []PageTreeNode `json:"children"`
}

func listEnvelope[T any](results []T, limit int, nextCursor string, itemType string, fields []string) ListEnvelope[T] {
	return ListEnvelope[T]{
		Results: results,
		Page: PageWindow{
			Limit:      limit,
			NextCursor: nextCursor,
		},
		Schema: Schema{
			ItemType: itemType,
			Fields:   fields,
		},
	}
}

func itemEnvelope[T any](item T, itemType string, fields []string) ItemEnvelope[T] {
	return ItemEnvelope[T]{
		Item: item,
		Schema: Schema{
			ItemType: itemType,
			Fields:   fields,
		},
	}
}

func newSpaceSummary(space confluence.Space) SpaceSummary {
	return SpaceSummary{
		ID:     space.ID,
		Key:    space.Key,
		Name:   space.Name,
		Type:   space.Type,
		Status: space.Status,
	}
}

func newPageSummary(page confluence.Page) PageSummary {
	summary := PageSummary{
		ID:       page.ID,
		Title:    page.Title,
		SpaceID:  page.SpaceID,
		Status:   page.Status,
		ParentID: page.ParentID,
	}
	if page.Version != nil {
		summary.VersionNumber = page.Version.Number
	}
	return summary
}

func newPageDetail(page *confluence.Page, bodyFormat string) PageDetail {
	detail := PageDetail{
		ID:         page.ID,
		Title:      page.Title,
		SpaceID:    page.SpaceID,
		Status:     page.Status,
		ParentID:   page.ParentID,
		ParentType: page.ParentType,
		AuthorID:   page.AuthorID,
		CreatedAt:  page.CreatedAt,
	}
	if page.Version != nil {
		detail.Version = &PageVersionInfo{
			Number:    page.Version.Number,
			Message:   page.Version.Message,
			CreatedAt: page.Version.CreatedAt,
			AuthorID:  page.Version.AuthorID,
		}
	}
	if bodyFormat != "" {
		detail.Body = bodyFromPage(page, bodyFormat)
	}
	return detail
}

func bodyFromPage(page *confluence.Page, bodyFormat string) *PageBody {
	if page.Body == nil {
		return nil
	}

	switch bodyFormat {
	case "view":
		if page.Body.View == nil {
			return nil
		}
		return &PageBody{Format: "view", Value: page.Body.View.Value}
	case "storage":
		if page.Body.Storage == nil {
			return nil
		}
		return &PageBody{Format: "storage", Value: page.Body.Storage.Value}
	case "atlas_doc_format":
		if page.Body.AtlasDocFormat == nil {
			return nil
		}
		return &PageBody{Format: "atlas_doc_format", Value: page.Body.AtlasDocFormat.Value}
	default:
		return nil
	}
}

func newSearchSummary(result confluence.SearchResult) SearchSummary {
	return SearchSummary{
		ID:      result.ID,
		Title:   result.Title,
		SpaceID: result.SpaceID,
		Excerpt: result.Excerpt,
		URL:     result.URL,
	}
}
