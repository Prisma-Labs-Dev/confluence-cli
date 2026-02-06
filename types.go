package confluence

import "time"

type Space struct {
	ID          string    `json:"id"`
	Key         string    `json:"key"`
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	Status      string    `json:"status"`
	HomepageID  string    `json:"homepageId,omitempty"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"createdAt,omitempty"`
}

type Page struct {
	ID         string    `json:"id"`
	Title      string    `json:"title"`
	SpaceID    string    `json:"spaceId"`
	Status     string    `json:"status"`
	ParentID   string    `json:"parentId,omitempty"`
	ParentType string    `json:"parentType,omitempty"`
	AuthorID   string    `json:"authorId,omitempty"`
	CreatedAt  time.Time `json:"createdAt,omitempty"`
	Version    *Version  `json:"version,omitempty"`
	Body       *Body     `json:"body,omitempty"`
}

type Version struct {
	Number    int       `json:"number"`
	Message   string    `json:"message,omitempty"`
	CreatedAt time.Time `json:"createdAt,omitempty"`
	AuthorID  string    `json:"authorId,omitempty"`
}

type Body struct {
	View           *BodyRepresentation `json:"view,omitempty"`
	Storage        *BodyRepresentation `json:"storage,omitempty"`
	AtlasDocFormat *BodyRepresentation `json:"atlas_doc_format,omitempty"`
}

type BodyRepresentation struct {
	Value          string `json:"value"`
	Representation string `json:"representation"`
}

type SearchResult struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Type    string `json:"type"`
	SpaceID string `json:"spaceId,omitempty"`
	Excerpt string `json:"excerpt,omitempty"`
	URL     string `json:"url,omitempty"`
}

// ListResult is a generic paginated result
type ListResult[T any] struct {
	Results    []T    `json:"results"`
	NextCursor string `json:"nextCursor,omitempty"`
}
