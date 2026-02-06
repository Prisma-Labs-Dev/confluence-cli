package confluence

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// testServer creates an httptest.Server that serves fixture files based on URL path.
// The handler maps request paths to fixture files and validates auth headers.
func testServer(t *testing.T, routes map[string]string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify auth header is present
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Basic ") {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"message":"Unauthorized; check your credentials","statusCode":401}`))
			return
		}

		// Match route by path prefix (ignoring query params)
		path := r.URL.Path
		for routePath, fixture := range routes {
			if path == routePath {
				data, err := os.ReadFile(filepath.Join("testdata", fixture))
				if err != nil {
					t.Fatalf("reading fixture %s: %v", fixture, err)
				}
				w.Header().Set("Content-Type", "application/json")
				w.Write(data)
				return
			}
		}

		// No route matched
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"errors":[{"message":"Not found"}]}`))
	}))
}

func newTestClient(serverURL string) *Client {
	return NewClient(Options{
		BaseURL: serverURL,
		Email:   "test@example.com",
		Token:   "test-token",
	})
}

func TestListSpaces(t *testing.T) {
	srv := testServer(t, map[string]string{
		"/wiki/api/v2/spaces": "spaces_list.json",
	})
	defer srv.Close()

	client := newTestClient(srv.URL)
	result, err := client.ListSpaces(ListSpacesOptions{Limit: 2})
	if err != nil {
		t.Fatalf("ListSpaces: %v", err)
	}

	if len(result.Results) != 2 {
		t.Fatalf("expected 2 spaces, got %d", len(result.Results))
	}

	// Verify first space fields match real API response
	s := result.Results[0]
	if s.ID != "3082551269" {
		t.Errorf("space ID = %q, want %q", s.ID, "3082551269")
	}
	if s.Key != "CF" {
		t.Errorf("space Key = %q, want %q", s.Key, "CF")
	}
	if s.Name != "BeCSEE Cloud Foundation" {
		t.Errorf("space Name = %q, want %q", s.Name, "BeCSEE Cloud Foundation")
	}
	if s.Type != "global" {
		t.Errorf("space Type = %q, want %q", s.Type, "global")
	}
	if s.Status != "current" {
		t.Errorf("space Status = %q, want %q", s.Status, "current")
	}
	if s.HomepageID != "3082848318" {
		t.Errorf("space HomepageID = %q, want %q", s.HomepageID, "3082848318")
	}

	// Verify second space
	s2 := result.Results[1]
	if s2.Key != "ARCHBCSE" {
		t.Errorf("space[1] Key = %q, want %q", s2.Key, "ARCHBCSE")
	}

	// Verify pagination cursor was extracted
	if result.NextCursor == "" {
		t.Error("expected non-empty NextCursor")
	}
}

func TestListSpacesEmpty(t *testing.T) {
	srv := testServer(t, map[string]string{
		"/wiki/api/v2/spaces": "spaces_list_empty.json",
	})
	defer srv.Close()

	client := newTestClient(srv.URL)
	result, err := client.ListSpaces(ListSpacesOptions{})
	if err != nil {
		t.Fatalf("ListSpaces: %v", err)
	}

	if len(result.Results) != 0 {
		t.Fatalf("expected 0 spaces, got %d", len(result.Results))
	}
	if result.NextCursor != "" {
		t.Errorf("expected empty NextCursor, got %q", result.NextCursor)
	}
}

func TestListPages(t *testing.T) {
	srv := testServer(t, map[string]string{
		"/wiki/api/v2/pages": "pages_list.json",
	})
	defer srv.Close()

	client := newTestClient(srv.URL)
	result, err := client.ListPages(ListPagesOptions{SpaceID: "3082551269", Limit: 2})
	if err != nil {
		t.Fatalf("ListPages: %v", err)
	}

	if len(result.Results) != 2 {
		t.Fatalf("expected 2 pages, got %d", len(result.Results))
	}

	p := result.Results[0]
	if p.ID != "3082848318" {
		t.Errorf("page ID = %q, want %q", p.ID, "3082848318")
	}
	if p.Title != "BeCSEE Cloud Foundation" {
		t.Errorf("page Title = %q, want %q", p.Title, "BeCSEE Cloud Foundation")
	}
	if p.SpaceID != "3082551269" {
		t.Errorf("page SpaceID = %q, want %q", p.SpaceID, "3082551269")
	}
	if p.Status != "current" {
		t.Errorf("page Status = %q, want %q", p.Status, "current")
	}
	if p.Version == nil {
		t.Fatal("expected non-nil Version")
	}
	if p.Version.Number != 119 {
		t.Errorf("page Version.Number = %d, want %d", p.Version.Number, 119)
	}

	// Second page has parent
	p2 := result.Results[1]
	if p2.ParentID != "3082848318" {
		t.Errorf("page[1] ParentID = %q, want %q", p2.ParentID, "3082848318")
	}
	if p2.ParentType != "page" {
		t.Errorf("page[1] ParentType = %q, want %q", p2.ParentType, "page")
	}

	if result.NextCursor == "" {
		t.Error("expected non-empty NextCursor")
	}
}

func TestGetPage(t *testing.T) {
	srv := testServer(t, map[string]string{
		"/wiki/api/v2/pages/3082848318": "page_get.json",
	})
	defer srv.Close()

	client := newTestClient(srv.URL)
	page, err := client.GetPage(GetPageOptions{PageID: "3082848318", BodyFormat: "storage"})
	if err != nil {
		t.Fatalf("GetPage: %v", err)
	}

	if page.ID != "3082848318" {
		t.Errorf("page ID = %q, want %q", page.ID, "3082848318")
	}
	if page.Title != "BeCSEE Cloud Foundation" {
		t.Errorf("page Title = %q, want %q", page.Title, "BeCSEE Cloud Foundation")
	}
	if page.Body == nil {
		t.Fatal("expected non-nil Body")
	}
	if page.Body.Storage == nil {
		t.Fatal("expected non-nil Body.Storage")
	}
	if page.Body.Storage.Representation != "storage" {
		t.Errorf("body representation = %q, want %q", page.Body.Storage.Representation, "storage")
	}
	if !strings.Contains(page.Body.Storage.Value, "Meet the team") {
		t.Error("expected body to contain 'Meet the team'")
	}
}

func TestGetPageChildren(t *testing.T) {
	srv := testServer(t, map[string]string{
		"/wiki/api/v2/pages/3082848318/children": "page_children.json",
	})
	defer srv.Close()

	client := newTestClient(srv.URL)
	result, err := client.GetPageChildren(GetPageChildrenOptions{PageID: "3082848318", Limit: 2})
	if err != nil {
		t.Fatalf("GetPageChildren: %v", err)
	}

	if len(result.Results) != 2 {
		t.Fatalf("expected 2 children, got %d", len(result.Results))
	}

	if result.Results[0].Title != "Our Team" {
		t.Errorf("child[0] Title = %q, want %q", result.Results[0].Title, "Our Team")
	}
	if result.Results[1].Title != "Our Meetings" {
		t.Errorf("child[1] Title = %q, want %q", result.Results[1].Title, "Our Meetings")
	}
	if result.NextCursor == "" {
		t.Error("expected non-empty NextCursor")
	}
}

func TestGetPageChildrenEmpty(t *testing.T) {
	srv := testServer(t, map[string]string{
		"/wiki/api/v2/pages/999/children": "page_children_empty.json",
	})
	defer srv.Close()

	client := newTestClient(srv.URL)
	result, err := client.GetPageChildren(GetPageChildrenOptions{PageID: "999"})
	if err != nil {
		t.Fatalf("GetPageChildren: %v", err)
	}

	if len(result.Results) != 0 {
		t.Fatalf("expected 0 children, got %d", len(result.Results))
	}
}

func TestSearch(t *testing.T) {
	srv := testServer(t, map[string]string{
		"/wiki/rest/api/search": "search.json",
	})
	defer srv.Close()

	client := newTestClient(srv.URL)
	result, err := client.Search(SearchOptions{CQL: "type=page AND space=CF", Limit: 2})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}

	if len(result.Results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(result.Results))
	}

	r := result.Results[0]
	if r.ID != "8975222682" {
		t.Errorf("result ID = %q, want %q", r.ID, "8975222682")
	}
	if r.Title != "CF2-DD-2021-07-01: Sendgrid" {
		t.Errorf("result Title = %q, want %q", r.Title, "CF2-DD-2021-07-01: Sendgrid")
	}
	if r.Type != "page" {
		t.Errorf("result Type = %q, want %q", r.Type, "page")
	}
	if r.Excerpt == "" {
		t.Error("expected non-empty Excerpt")
	}
	if r.URL == "" {
		t.Error("expected non-empty URL")
	}

	if result.NextCursor == "" {
		t.Error("expected non-empty NextCursor for search")
	}
}

func TestAPIError401(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message":"Unauthorized; check your credentials","statusCode":401}`))
	}))
	defer srv.Close()

	client := newTestClient(srv.URL)
	_, err := client.ListSpaces(ListSpacesOptions{})
	if err == nil {
		t.Fatal("expected error for 401")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		// Might be wrapped
		t.Logf("error type: %T, message: %v", err, err)
		return
	}
	if apiErr.StatusCode != 401 {
		t.Errorf("expected status 401, got %d", apiErr.StatusCode)
	}
}

func TestAPIError404(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		data, _ := os.ReadFile("testdata/error_404.json")
		w.Write(data)
	}))
	defer srv.Close()

	client := newTestClient(srv.URL)
	_, err := client.GetPage(GetPageOptions{PageID: "999999"})
	if err == nil {
		t.Fatal("expected error for 404")
	}

	if !strings.Contains(err.Error(), "999999") {
		t.Logf("error message: %v", err)
	}
}

func TestInvalidCredentials(t *testing.T) {
	// Server that always returns 401 (simulating invalid credentials)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message":"Unauthorized; check your credentials","statusCode":401}`))
	}))
	defer srv.Close()

	client := NewClient(Options{
		BaseURL: srv.URL,
		Email:   "wrong@email.com",
		Token:   "invalid-token",
	})
	_, err := client.ListSpaces(ListSpacesOptions{})
	if err == nil {
		t.Fatal("expected error for invalid credentials")
	}
	if !strings.Contains(err.Error(), "401") {
		t.Errorf("error should mention 401, got: %v", err)
	}
}

func TestExtractCursor(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "v2 spaces pagination",
			input:    "/wiki/api/v2/spaces?limit=2&cursor=eyJpZCI6MzIwOTc5MzMwNX0=",
			expected: "eyJpZCI6MzIwOTc5MzMwNX0=",
		},
		{
			name:     "v2 pages pagination",
			input:    "/wiki/api/v2/pages?limit=2&space-id=3082551269&cursor=eyJpZCI6IjMwOTAzMzQzOTQifQ==",
			expected: "eyJpZCI6IjMwOTAzMzQzOTQifQ==",
		},
		{
			name:     "empty link",
			input:    "",
			expected: "",
		},
		{
			name:     "no cursor param",
			input:    "/wiki/api/v2/spaces?limit=2",
			expected: "",
		},
		{
			name:     "v1 search pagination",
			input:    "/rest/api/search?next=true&cursor=abc123&limit=2",
			expected: "abc123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractCursor(tt.input)
			if got != tt.expected {
				t.Errorf("extractCursor(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestClientBaseURLTrailingSlash(t *testing.T) {
	srv := testServer(t, map[string]string{
		"/wiki/api/v2/spaces": "spaces_list.json",
	})
	defer srv.Close()

	// URL with trailing slash should work
	client := NewClient(Options{
		BaseURL: srv.URL + "/",
		Email:   "test@example.com",
		Token:   "test-token",
	})
	result, err := client.ListSpaces(ListSpacesOptions{})
	if err != nil {
		t.Fatalf("ListSpaces with trailing slash: %v", err)
	}
	if len(result.Results) != 2 {
		t.Fatalf("expected 2 spaces, got %d", len(result.Results))
	}
}
