package confluence_test

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
)

func TestPagesSearchJSONContract_Integration(t *testing.T) {
	srv := newBinaryTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/wiki/rest/api/search" {
			t.Errorf("path = %q, want %q", r.URL.Path, "/wiki/rest/api/search")
			http.Error(w, "bad path", http.StatusBadRequest)
			return
		}
		q := r.URL.Query()
		if got := q.Get("cql"); got != `type=page AND title ~ "runbook" AND space.id=3082551269` {
			t.Errorf("cql = %q", got)
			http.Error(w, "bad cql", http.StatusBadRequest)
			return
		}
		if got := q.Get("limit"); got != "10" {
			t.Errorf("limit = %q, want %q", got, "10")
			http.Error(w, "bad limit", http.StatusBadRequest)
			return
		}
		if got := q.Get("cursor"); got != "cur-3" {
			t.Errorf("cursor = %q, want %q", got, "cur-3")
			http.Error(w, "bad cursor", http.StatusBadRequest)
			return
		}
		writeFixtureResponse(t, w, "search.json")
	})
	defer srv.Close()

	tmp := t.TempDir()
	binPath := buildBinary(t, tmp)

	stdout, stderr, err := runBinary(binPath, []string{
		"--url", srv.URL, "--email", "a@b.com", "--token", "tok",
		"pages", "search", "--query", "runbook", "--title-only", "--space-id", "3082551269", "--cursor", "cur-3",
	}, "", envForIntegration(filepath.Join(tmp, "config"))...)
	if err != nil {
		t.Fatalf("pages search failed: %v\nstdout=%s\nstderr=%s", err, stdout, stderr)
	}
	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("expected empty stderr, got %s", stderr)
	}

	var result struct {
		Results []struct {
			ID      string `json:"id"`
			Title   string `json:"title"`
			SpaceID string `json:"spaceId"`
			Type    string `json:"type"`
		} `json:"results"`
		Page struct {
			Limit      int    `json:"limit"`
			NextCursor string `json:"nextCursor"`
		} `json:"page"`
		Schema struct {
			ItemType string `json:"itemType"`
		} `json:"schema"`
	}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("pages search output not valid JSON: %v\nstdout=%s", err, stdout)
	}
	if len(result.Results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(result.Results))
	}
	if result.Results[0].ID != "8975222682" {
		t.Fatalf("first id = %q, want %q", result.Results[0].ID, "8975222682")
	}
	if result.Page.NextCursor != "abc123" {
		t.Fatalf("page.nextCursor = %q, want %q", result.Page.NextCursor, "abc123")
	}
	if result.Schema.ItemType != "page-search-result" {
		t.Fatalf("schema.itemType = %q, want %q", result.Schema.ItemType, "page-search-result")
	}
}

func TestPagesSearchRawCQLContract_Integration(t *testing.T) {
	srv := newBinaryTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/wiki/rest/api/search" {
			t.Errorf("path = %q, want %q", r.URL.Path, "/wiki/rest/api/search")
			http.Error(w, "bad path", http.StatusBadRequest)
			return
		}
		q := r.URL.Query()
		if got := q.Get("cql"); got != `space = "SC" AND title ~ "slotting"` {
			t.Errorf("cql = %q", got)
			http.Error(w, "bad cql", http.StatusBadRequest)
			return
		}
		writeFixtureResponse(t, w, "search.json")
	})
	defer srv.Close()

	tmp := t.TempDir()
	binPath := buildBinary(t, tmp)

	stdout, stderr, err := runBinary(binPath, []string{
		"--url", srv.URL, "--email", "a@b.com", "--token", "tok",
		"pages", "search", "--cql", `space = "SC" AND title ~ "slotting"`, "--limit", "1",
	}, "", envForIntegration(filepath.Join(tmp, "config"))...)
	if err != nil {
		t.Fatalf("pages search raw cql failed: %v\nstdout=%s\nstderr=%s", err, stdout, stderr)
	}
	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("expected empty stderr, got %s", stderr)
	}

	var result struct {
		Results []struct {
			ID string `json:"id"`
		} `json:"results"`
		Page struct {
			Limit int `json:"limit"`
		} `json:"page"`
	}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("pages search raw cql output not valid JSON: %v\nstdout=%s", err, stdout)
	}
	if len(result.Results) == 0 {
		t.Fatal("expected at least one search result")
	}
	if result.Page.Limit != 1 {
		t.Fatalf("page.limit = %d, want %d", result.Page.Limit, 1)
	}
}

func TestPagesSearchValidationError_Integration(t *testing.T) {
	tmp := t.TempDir()
	binPath := buildBinary(t, tmp)

	stdout, stderr, exitCode, err := runBinaryWithExitCode(binPath, []string{
		"--url", "https://example.atlassian.net", "--email", "a@b.com", "--token", "tok",
		"pages", "search", "--query", "slotting", "--cql", `space = "SC"`,
	}, "", envForIntegration(filepath.Join(tmp, "config"))...)
	if err == nil {
		t.Fatalf("expected pages search validation failure\nstdout=%s\nstderr=%s", stdout, stderr)
	}
	if exitCode != 2 {
		t.Fatalf("exit code = %d, want %d", exitCode, 2)
	}
	if stdout != "" {
		t.Fatalf("expected empty stdout on error, got %q", stdout)
	}

	var result struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &result); err != nil {
		t.Fatalf("stderr is not valid JSON: %v\nstderr=%s", err, stderr)
	}
	if result.Error.Code != "VALIDATION" {
		t.Fatalf("code = %q, want %q", result.Error.Code, "VALIDATION")
	}
	if !strings.Contains(result.Error.Message, "exactly one of --query or --cql") {
		t.Fatalf("unexpected error message: %q", result.Error.Message)
	}
}

func TestMissingAuthErrorContract_Integration(t *testing.T) {
	tmp := t.TempDir()
	binPath := buildBinary(t, tmp)

	stdout, stderr, exitCode, err := runBinaryWithExitCode(binPath, []string{"spaces", "list"}, "", envForIntegration(filepath.Join(tmp, "config"))...)
	if err == nil {
		t.Fatalf("expected spaces list without auth to fail\nstdout=%s\nstderr=%s", stdout, stderr)
	}
	if exitCode != 2 {
		t.Fatalf("exit code = %d, want %d", exitCode, 2)
	}
	if stdout != "" {
		t.Fatalf("expected empty stdout on error, got %q", stdout)
	}

	var result struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
			Hint    string `json:"hint"`
		} `json:"error"`
	}
	if err := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &result); err != nil {
		t.Fatalf("stderr is not valid JSON: %v\nstderr=%s", err, stderr)
	}
	if result.Error.Code != "VALIDATION" {
		t.Fatalf("code = %q, want %q", result.Error.Code, "VALIDATION")
	}
	if !strings.Contains(result.Error.Message, "missing credentials") {
		t.Fatalf("unexpected error message: %q", result.Error.Message)
	}
	if !strings.Contains(result.Error.Hint, "auth login") {
		t.Fatalf("unexpected error hint: %q", result.Error.Hint)
	}
}

func TestAuthFailureExitCode_Integration(t *testing.T) {
	srv := newBinaryTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		writeJSONResponse(w, []byte(`{"message":"Unauthorized; check your credentials","statusCode":401}`))
	})
	defer srv.Close()

	tmp := t.TempDir()
	binPath := buildBinary(t, tmp)

	stdout, stderr, exitCode, err := runBinaryWithExitCode(binPath, []string{
		"--url", srv.URL, "--email", "a@b.com", "--token", "tok",
		"spaces", "list",
	}, "", envForIntegration(filepath.Join(tmp, "config"))...)
	if err == nil {
		t.Fatalf("expected auth failure\nstdout=%s\nstderr=%s", stdout, stderr)
	}
	if exitCode != 3 {
		t.Fatalf("exit code = %d, want %d", exitCode, 3)
	}
	if stdout != "" {
		t.Fatalf("expected empty stdout on error, got %q", stdout)
	}

	var result struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &result); err != nil {
		t.Fatalf("stderr is not valid JSON: %v\nstderr=%s", err, stderr)
	}
	if result.Error.Code != "AUTH_FAILED" {
		t.Fatalf("code = %q, want %q", result.Error.Code, "AUTH_FAILED")
	}
}

func TestAPIErrorExitCode_Integration(t *testing.T) {
	srv := newBinaryTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		writeJSONResponse(w, []byte(`{"message":"server exploded","statusCode":500}`))
	})
	defer srv.Close()

	tmp := t.TempDir()
	binPath := buildBinary(t, tmp)

	stdout, stderr, exitCode, err := runBinaryWithExitCode(binPath, []string{
		"--url", srv.URL, "--email", "a@b.com", "--token", "tok",
		"spaces", "list",
	}, "", envForIntegration(filepath.Join(tmp, "config"))...)
	if err == nil {
		t.Fatalf("expected API failure\nstdout=%s\nstderr=%s", stdout, stderr)
	}
	if exitCode != 1 {
		t.Fatalf("exit code = %d, want %d", exitCode, 1)
	}
	if stdout != "" {
		t.Fatalf("expected empty stdout on error, got %q", stdout)
	}

	var result struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &result); err != nil {
		t.Fatalf("stderr is not valid JSON: %v\nstderr=%s", err, stderr)
	}
	if result.Error.Code != "API_ERROR" {
		t.Fatalf("code = %q, want %q", result.Error.Code, "API_ERROR")
	}
	if !strings.Contains(result.Error.Message, "server exploded") {
		t.Fatalf("unexpected error message: %q", result.Error.Message)
	}
}
