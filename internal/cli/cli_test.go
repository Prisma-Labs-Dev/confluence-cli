package cli

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	confluence "github.com/Prisma-Labs-Dev/confluence-cli"
)

// fixtureServer creates a test server serving fixture files from testdata.
// Routes map URL paths to fixture filenames. If a route value starts with "!",
// it's treated as inline JSON content instead of a filename.
func fixtureServer(t *testing.T, routes map[string]string) *httptest.Server {
	t.Helper()
	fixtureDir := filepath.Join("..", "..", "testdata")
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		for routePath, fixture := range routes {
			if path == routePath {
				var data []byte
				var err error
				if strings.HasPrefix(fixture, "!") {
					data = []byte(fixture[1:])
				} else {
					data, err = os.ReadFile(filepath.Join(fixtureDir, fixture))
					if err != nil {
						t.Fatalf("reading fixture %s: %v", fixture, err)
					}
				}
				w.Header().Set("Content-Type", "application/json")
				w.Write(data)
				return
			}
		}
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"errors":[{"message":"Not found"}]}`))
	}))
}

func runCLI(args []string, version string) (stdout, stderr string, exitCode int) {
	var outBuf, errBuf bytes.Buffer
	code := Run(args, &outBuf, &errBuf, version)
	return outBuf.String(), errBuf.String(), code
}

// --- Version command tests ---

func TestVersionJSON(t *testing.T) {
	stdout, _, code := runCLI([]string{"version"}, "1.2.3")
	if code != ExitOK {
		t.Fatalf("exit code = %d, want %d", code, ExitOK)
	}

	var result struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("parse JSON: %v\noutput: %s", err, stdout)
	}
	if result.Version != "1.2.3" {
		t.Errorf("version = %q, want %q", result.Version, "1.2.3")
	}
}

func TestVersionPlain(t *testing.T) {
	stdout, _, code := runCLI([]string{"version", "--plain"}, "2.0.0")
	if code != ExitOK {
		t.Fatalf("exit code = %d, want %d", code, ExitOK)
	}
	if stdout != "confluence 2.0.0\n" {
		t.Errorf("output = %q, want %q", stdout, "confluence 2.0.0\n")
	}
}

func TestVersionNoAuth(t *testing.T) {
	// version should work without auth credentials
	_, _, code := runCLI([]string{"version"}, "1.0.0")
	if code != ExitOK {
		t.Fatalf("version should work without auth, exit code = %d", code)
	}
}

// --- Validation tests ---

func TestNoCommand(t *testing.T) {
	_, stderr, code := runCLI([]string{}, "1.0.0")
	if code != ExitValidation {
		t.Fatalf("exit code = %d, want %d", code, ExitValidation)
	}

	var errResult CLIError
	if err := json.Unmarshal([]byte(stderr), &errResult); err != nil {
		t.Fatalf("parse error JSON: %v\nstderr: %s", err, stderr)
	}
	if errResult.Code != "VALIDATION" {
		t.Errorf("error code = %q, want %q", errResult.Code, "VALIDATION")
	}
}

func TestMissingAuth(t *testing.T) {
	// Clear env vars so they don't leak into the test
	t.Setenv("CONFLUENCE_URL", "")
	t.Setenv("CONFLUENCE_EMAIL", "")
	t.Setenv("CONFLUENCE_API_TOKEN", "")

	_, stderr, code := runCLI([]string{"spaces", "list"}, "1.0.0")
	if code != ExitValidation {
		t.Fatalf("exit code = %d, want %d", code, ExitValidation)
	}

	var errResult CLIError
	if err := json.Unmarshal([]byte(stderr), &errResult); err != nil {
		t.Fatalf("parse error JSON: %v\nstderr: %s", err, stderr)
	}
	if errResult.Code != "VALIDATION" {
		t.Errorf("error code = %q, want %q", errResult.Code, "VALIDATION")
	}
	if !strings.Contains(errResult.Message, "CONFLUENCE_URL") {
		t.Errorf("error should mention CONFLUENCE_URL, got: %s", errResult.Message)
	}
	if !strings.Contains(errResult.Message, "CONFLUENCE_EMAIL") {
		t.Errorf("error should mention CONFLUENCE_EMAIL, got: %s", errResult.Message)
	}
	if !strings.Contains(errResult.Message, "CONFLUENCE_API_TOKEN") {
		t.Errorf("error should mention CONFLUENCE_API_TOKEN, got: %s", errResult.Message)
	}
}

func TestMissingPartialAuth(t *testing.T) {
	t.Setenv("CONFLUENCE_URL", "")
	t.Setenv("CONFLUENCE_EMAIL", "")
	t.Setenv("CONFLUENCE_API_TOKEN", "")

	// Only URL provided, missing email and token
	_, stderr, code := runCLI([]string{"--url", "https://test.atlassian.net", "spaces", "list"}, "1.0.0")
	if code != ExitValidation {
		t.Fatalf("exit code = %d, want %d", code, ExitValidation)
	}

	var errResult CLIError
	json.Unmarshal([]byte(stderr), &errResult)
	if !strings.Contains(errResult.Message, "CONFLUENCE_EMAIL") {
		t.Errorf("should mention missing email, got: %s", errResult.Message)
	}
	if strings.Contains(errResult.Message, "CONFLUENCE_URL") {
		t.Errorf("should NOT mention URL (it was provided), got: %s", errResult.Message)
	}
}

func TestMissingRequiredFlags(t *testing.T) {
	srv := fixtureServer(t, map[string]string{})
	defer srv.Close()

	// pages list without --space-id
	_, _, code := runCLI([]string{
		"--url", srv.URL, "--email", "a@b.com", "--token", "tok",
		"pages", "list",
	}, "1.0.0")
	if code != ExitValidation {
		t.Fatalf("exit code = %d, want %d for missing --space-id", code, ExitValidation)
	}

	// pages get without --page-id
	_, _, code = runCLI([]string{
		"--url", srv.URL, "--email", "a@b.com", "--token", "tok",
		"pages", "get",
	}, "1.0.0")
	if code != ExitValidation {
		t.Fatalf("exit code = %d, want %d for missing --page-id", code, ExitValidation)
	}

	// pages search without --cql
	_, _, code = runCLI([]string{
		"--url", srv.URL, "--email", "a@b.com", "--token", "tok",
		"pages", "search",
	}, "1.0.0")
	if code != ExitValidation {
		t.Fatalf("exit code = %d, want %d for missing --cql", code, ExitValidation)
	}
}

// --- Spaces command tests ---

func TestSpacesListJSON(t *testing.T) {
	srv := fixtureServer(t, map[string]string{
		"/wiki/api/v2/spaces": "spaces_list.json",
	})
	defer srv.Close()

	stdout, _, code := runCLI([]string{
		"--url", srv.URL, "--email", "a@b.com", "--token", "tok",
		"spaces", "list",
	}, "1.0.0")
	if code != ExitOK {
		t.Fatalf("exit code = %d, want %d", code, ExitOK)
	}

	var result confluence.ListResult[confluence.Space]
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("parse JSON: %v\noutput: %s", err, stdout)
	}
	if len(result.Results) != 2 {
		t.Fatalf("expected 2 spaces, got %d", len(result.Results))
	}
	if result.Results[0].Key != "CF" {
		t.Errorf("first space key = %q, want CF", result.Results[0].Key)
	}
	if result.NextCursor == "" {
		t.Error("expected non-empty NextCursor")
	}
}

func TestSpacesListPlain(t *testing.T) {
	srv := fixtureServer(t, map[string]string{
		"/wiki/api/v2/spaces": "spaces_list.json",
	})
	defer srv.Close()

	stdout, _, code := runCLI([]string{
		"--url", srv.URL, "--email", "a@b.com", "--token", "tok",
		"--plain", "spaces", "list",
	}, "1.0.0")
	if code != ExitOK {
		t.Fatalf("exit code = %d, want %d", code, ExitOK)
	}

	// Should contain table headers
	if !strings.Contains(stdout, "ID") || !strings.Contains(stdout, "KEY") || !strings.Contains(stdout, "NAME") {
		t.Errorf("plain output missing table headers: %s", stdout)
	}
	// Should contain space data
	if !strings.Contains(stdout, "CF") {
		t.Errorf("plain output missing space key CF: %s", stdout)
	}
	if !strings.Contains(stdout, "BeCSEE Cloud Foundation") {
		t.Errorf("plain output missing space name: %s", stdout)
	}
	// Should contain cursor info
	if !strings.Contains(stdout, "Next cursor:") {
		t.Errorf("plain output missing cursor info: %s", stdout)
	}
}

// --- Pages command tests ---

func TestPagesListJSON(t *testing.T) {
	srv := fixtureServer(t, map[string]string{
		"/wiki/api/v2/pages": "pages_list.json",
	})
	defer srv.Close()

	stdout, _, code := runCLI([]string{
		"--url", srv.URL, "--email", "a@b.com", "--token", "tok",
		"pages", "list", "--space-id", "3082551269",
	}, "1.0.0")
	if code != ExitOK {
		t.Fatalf("exit code = %d, want %d", code, ExitOK)
	}

	var result confluence.ListResult[confluence.Page]
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("parse JSON: %v\noutput: %s", err, stdout)
	}
	if len(result.Results) != 2 {
		t.Fatalf("expected 2 pages, got %d", len(result.Results))
	}
	if result.Results[0].Title != "BeCSEE Cloud Foundation" {
		t.Errorf("first page title = %q", result.Results[0].Title)
	}
}

func TestPagesListPlain(t *testing.T) {
	srv := fixtureServer(t, map[string]string{
		"/wiki/api/v2/pages": "pages_list.json",
	})
	defer srv.Close()

	stdout, _, code := runCLI([]string{
		"--url", srv.URL, "--email", "a@b.com", "--token", "tok",
		"--plain", "pages", "list", "--space-id", "3082551269",
	}, "1.0.0")
	if code != ExitOK {
		t.Fatalf("exit code = %d, want %d", code, ExitOK)
	}

	if !strings.Contains(stdout, "TITLE") {
		t.Errorf("missing table header TITLE: %s", stdout)
	}
	if !strings.Contains(stdout, "BeCSEE Cloud Foundation") {
		t.Errorf("missing page title: %s", stdout)
	}
}

func TestPagesGetJSON(t *testing.T) {
	srv := fixtureServer(t, map[string]string{
		"/wiki/api/v2/pages/3082848318": "page_get.json",
	})
	defer srv.Close()

	stdout, _, code := runCLI([]string{
		"--url", srv.URL, "--email", "a@b.com", "--token", "tok",
		"pages", "get", "--page-id", "3082848318", "--body-format", "storage",
	}, "1.0.0")
	if code != ExitOK {
		t.Fatalf("exit code = %d, want %d", code, ExitOK)
	}

	var page confluence.Page
	if err := json.Unmarshal([]byte(stdout), &page); err != nil {
		t.Fatalf("parse JSON: %v\noutput: %s", err, stdout)
	}
	if page.ID != "3082848318" {
		t.Errorf("page ID = %q, want %q", page.ID, "3082848318")
	}
	if page.Body == nil || page.Body.Storage == nil {
		t.Fatal("expected body.storage to be present")
	}
	if !strings.Contains(page.Body.Storage.Value, "Meet the team") {
		t.Error("body should contain 'Meet the team'")
	}
}

func TestPagesGetPlain(t *testing.T) {
	srv := fixtureServer(t, map[string]string{
		"/wiki/api/v2/pages/3082848318": "page_get.json",
	})
	defer srv.Close()

	stdout, _, code := runCLI([]string{
		"--url", srv.URL, "--email", "a@b.com", "--token", "tok",
		"--plain", "pages", "get", "--page-id", "3082848318",
	}, "1.0.0")
	if code != ExitOK {
		t.Fatalf("exit code = %d, want %d", code, ExitOK)
	}

	if !strings.Contains(stdout, "ID:") {
		t.Errorf("missing ID label: %s", stdout)
	}
	if !strings.Contains(stdout, "3082848318") {
		t.Errorf("missing page ID value: %s", stdout)
	}
	if !strings.Contains(stdout, "BeCSEE Cloud Foundation") {
		t.Errorf("missing page title: %s", stdout)
	}
}

func TestPagesTreeJSON(t *testing.T) {
	// Use inline JSON without _links.next to avoid pagination loops
	childrenJSON := `{"results":[{"id":"148954617101","status":"current","title":"Our Team","spaceId":"3082551269"},{"id":"148948554949","status":"current","title":"Our Meetings","spaceId":"3082551269"}],"_links":{}}`
	emptyChildren := `{"results":[],"_links":{}}`

	srv := fixtureServer(t, map[string]string{
		"/wiki/api/v2/pages/3082848318/children":   "!" + childrenJSON,
		"/wiki/api/v2/pages/148954617101/children":  "!" + emptyChildren,
		"/wiki/api/v2/pages/148948554949/children":  "!" + emptyChildren,
	})
	defer srv.Close()

	stdout, _, code := runCLI([]string{
		"--url", srv.URL, "--email", "a@b.com", "--token", "tok",
		"pages", "tree", "--page-id", "3082848318", "--depth", "2",
	}, "1.0.0")
	if code != ExitOK {
		t.Fatalf("exit code = %d, want %d", code, ExitOK)
	}

	var result struct {
		PageID   string `json:"pageId"`
		Children []struct {
			ID       string `json:"id"`
			Title    string `json:"title"`
			Children []struct {
				ID    string `json:"id"`
				Title string `json:"title"`
			} `json:"children"`
		} `json:"children"`
	}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("parse JSON: %v\noutput: %s", err, stdout)
	}

	if result.PageID != "3082848318" {
		t.Errorf("pageId = %q, want %q", result.PageID, "3082848318")
	}
	if len(result.Children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(result.Children))
	}
	if result.Children[0].Title != "Our Team" {
		t.Errorf("child[0] title = %q, want %q", result.Children[0].Title, "Our Team")
	}
}

func TestPagesTreePlain(t *testing.T) {
	childrenJSON := `{"results":[{"id":"148954617101","status":"current","title":"Our Team","spaceId":"3082551269"},{"id":"148948554949","status":"current","title":"Our Meetings","spaceId":"3082551269"}],"_links":{}}`
	emptyChildren := `{"results":[],"_links":{}}`

	srv := fixtureServer(t, map[string]string{
		"/wiki/api/v2/pages/3082848318/children":   "!" + childrenJSON,
		"/wiki/api/v2/pages/148954617101/children":  "!" + emptyChildren,
		"/wiki/api/v2/pages/148948554949/children":  "!" + emptyChildren,
	})
	defer srv.Close()

	stdout, _, code := runCLI([]string{
		"--url", srv.URL, "--email", "a@b.com", "--token", "tok",
		"--plain", "pages", "tree", "--page-id", "3082848318",
	}, "1.0.0")
	if code != ExitOK {
		t.Fatalf("exit code = %d, want %d", code, ExitOK)
	}

	if !strings.Contains(stdout, "Our Team") {
		t.Errorf("missing child 'Our Team': %s", stdout)
	}
	if !strings.Contains(stdout, "Our Meetings") {
		t.Errorf("missing child 'Our Meetings': %s", stdout)
	}
	// Should have tree connectors
	if !strings.Contains(stdout, "├──") && !strings.Contains(stdout, "└──") {
		t.Errorf("missing tree connectors: %s", stdout)
	}
}

func TestPagesSearchJSON(t *testing.T) {
	srv := fixtureServer(t, map[string]string{
		"/wiki/rest/api/search": "search.json",
	})
	defer srv.Close()

	stdout, _, code := runCLI([]string{
		"--url", srv.URL, "--email", "a@b.com", "--token", "tok",
		"pages", "search", "--cql", "type=page AND space=CF",
	}, "1.0.0")
	if code != ExitOK {
		t.Fatalf("exit code = %d, want %d", code, ExitOK)
	}

	var result confluence.ListResult[confluence.SearchResult]
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("parse JSON: %v\noutput: %s", err, stdout)
	}
	if len(result.Results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(result.Results))
	}
	if result.Results[0].Title != "CF2-DD-2021-07-01: Sendgrid" {
		t.Errorf("result[0] title = %q", result.Results[0].Title)
	}
}

func TestPagesSearchPlain(t *testing.T) {
	srv := fixtureServer(t, map[string]string{
		"/wiki/rest/api/search": "search.json",
	})
	defer srv.Close()

	stdout, _, code := runCLI([]string{
		"--url", srv.URL, "--email", "a@b.com", "--token", "tok",
		"--plain", "pages", "search", "--cql", "type=page",
	}, "1.0.0")
	if code != ExitOK {
		t.Fatalf("exit code = %d, want %d", code, ExitOK)
	}

	if !strings.Contains(stdout, "Sendgrid") {
		t.Errorf("missing search result: %s", stdout)
	}
}

// --- Error handling tests ---

func TestAuthError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message":"Unauthorized; check your credentials","statusCode":401}`))
	}))
	defer srv.Close()

	_, stderr, code := runCLI([]string{
		"--url", srv.URL, "--email", "bad@email.com", "--token", "bad-token",
		"spaces", "list",
	}, "1.0.0")
	if code != ExitAuth {
		t.Fatalf("exit code = %d, want %d", code, ExitAuth)
	}

	var errResult CLIError
	if err := json.Unmarshal([]byte(stderr), &errResult); err != nil {
		t.Fatalf("parse error JSON: %v\nstderr: %s", err, stderr)
	}
	if errResult.Code != "AUTH_FAILED" {
		t.Errorf("error code = %q, want %q", errResult.Code, "AUTH_FAILED")
	}
}

func TestForbiddenError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"message":"Forbidden"}`))
	}))
	defer srv.Close()

	_, _, code := runCLI([]string{
		"--url", srv.URL, "--email", "a@b.com", "--token", "tok",
		"spaces", "list",
	}, "1.0.0")
	if code != ExitAuth {
		t.Fatalf("exit code = %d, want %d for 403", code, ExitAuth)
	}
}

func TestNotFoundError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"errors":[{"message":"No page found with id: 999999"}]}`))
	}))
	defer srv.Close()

	_, stderr, code := runCLI([]string{
		"--url", srv.URL, "--email", "a@b.com", "--token", "tok",
		"pages", "get", "--page-id", "999999",
	}, "1.0.0")
	if code != ExitError {
		t.Fatalf("exit code = %d, want %d", code, ExitError)
	}

	var errResult CLIError
	if err := json.Unmarshal([]byte(stderr), &errResult); err != nil {
		t.Fatalf("parse error JSON: %v\nstderr: %s", err, stderr)
	}
	if errResult.Code != "API_ERROR" {
		t.Errorf("error code = %q, want %q", errResult.Code, "API_ERROR")
	}
}

func TestServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"Internal server error"}`))
	}))
	defer srv.Close()

	_, _, code := runCLI([]string{
		"--url", srv.URL, "--email", "a@b.com", "--token", "tok",
		"spaces", "list",
	}, "1.0.0")
	if code != ExitError {
		t.Fatalf("exit code = %d, want %d", code, ExitError)
	}
}

// --- Output format tests ---

func TestJSONOutputIsValidJSON(t *testing.T) {
	srv := fixtureServer(t, map[string]string{
		"/wiki/api/v2/spaces": "spaces_list.json",
	})
	defer srv.Close()

	stdout, _, _ := runCLI([]string{
		"--url", srv.URL, "--email", "a@b.com", "--token", "tok",
		"spaces", "list",
	}, "1.0.0")

	if !json.Valid([]byte(stdout)) {
		t.Errorf("output is not valid JSON: %s", stdout)
	}
}

func TestErrorOutputIsValidJSON(t *testing.T) {
	t.Setenv("CONFLUENCE_URL", "")
	t.Setenv("CONFLUENCE_EMAIL", "")
	t.Setenv("CONFLUENCE_API_TOKEN", "")

	_, stderr, _ := runCLI([]string{"spaces", "list"}, "1.0.0")

	// Trim newline
	stderr = strings.TrimSpace(stderr)
	if !json.Valid([]byte(stderr)) {
		t.Errorf("error output is not valid JSON: %s", stderr)
	}
}

func TestErrorsGoToStderr(t *testing.T) {
	t.Setenv("CONFLUENCE_URL", "")
	t.Setenv("CONFLUENCE_EMAIL", "")
	t.Setenv("CONFLUENCE_API_TOKEN", "")

	stdout, stderr, code := runCLI([]string{"spaces", "list"}, "1.0.0")
	if code == ExitOK {
		t.Fatal("expected non-zero exit code")
	}

	// stdout should be empty on error
	if stdout != "" {
		t.Errorf("stdout should be empty on error, got: %s", stdout)
	}
	// stderr should have the error
	if stderr == "" {
		t.Error("stderr should contain error message")
	}
}
