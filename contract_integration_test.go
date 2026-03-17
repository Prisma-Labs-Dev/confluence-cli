package confluence_test

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
)

func TestHelpGolden_Integration(t *testing.T) {
	tmp := t.TempDir()
	binPath := buildBinary(t, tmp)
	configDir := filepath.Join(tmp, "config")

	tests := []struct {
		name   string
		args   []string
		golden string
	}{
		{name: "root", args: []string{"--help"}, golden: "help/root.txt"},
		{name: "spaces", args: []string{"spaces", "--help"}, golden: "help/spaces.txt"},
		{name: "spaces_list", args: []string{"spaces", "list", "--help"}, golden: "help/spaces_list.txt"},
		{name: "pages", args: []string{"pages", "--help"}, golden: "help/pages.txt"},
		{name: "pages_list", args: []string{"pages", "list", "--help"}, golden: "help/pages_list.txt"},
		{name: "pages_get", args: []string{"pages", "get", "--help"}, golden: "help/pages_get.txt"},
		{name: "pages_tree", args: []string{"pages", "tree", "--help"}, golden: "help/pages_tree.txt"},
		{name: "pages_search", args: []string{"pages", "search", "--help"}, golden: "help/pages_search.txt"},
		{name: "auth", args: []string{"auth", "--help"}, golden: "help/auth.txt"},
		{name: "auth_login", args: []string{"auth", "login", "--help"}, golden: "help/auth_login.txt"},
		{name: "version", args: []string{"version", "--help"}, golden: "help/version.txt"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			stdout, stderr, err := runBinary(binPath, tc.args, "", envForIntegration(configDir)...)
			if err != nil {
				t.Fatalf("help command failed: %v\nstdout=%s\nstderr=%s", err, stdout, stderr)
			}
			if stderr != "" {
				t.Fatalf("expected empty stderr, got %q", stderr)
			}

			want := readGoldenFile(t, tc.golden)
			if stdout != want {
				t.Fatalf("golden mismatch for %s\n--- want ---\n%s--- got ---\n%s", tc.name, want, stdout)
			}
		})
	}
}

func TestVersionContract_Integration(t *testing.T) {
	tmp := t.TempDir()
	binPath := buildBinary(t, tmp)

	stdout, stderr, err := runBinary(binPath, []string{"version"}, "", envForIntegration(filepath.Join(tmp, "config"))...)
	if err != nil {
		t.Fatalf("version failed: %v\nstdout=%s\nstderr=%s", err, stdout, stderr)
	}
	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("expected empty stderr, got %s", stderr)
	}

	var result struct {
		Item struct {
			Version string `json:"version"`
		} `json:"item"`
		Schema struct {
			ItemType string   `json:"itemType"`
			Fields   []string `json:"fields"`
		} `json:"schema"`
	}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("version output invalid JSON: %v\nstdout=%s", err, stdout)
	}
	if result.Item.Version == "" {
		t.Fatalf("expected non-empty version: %s", stdout)
	}
	if result.Schema.ItemType != "version" {
		t.Fatalf("unexpected schema itemType: %s", result.Schema.ItemType)
	}
}

func TestSpacesListJSONContract_Integration(t *testing.T) {
	srv := newBinaryTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/wiki/api/v2/spaces" {
			t.Errorf("path = %q, want %q", r.URL.Path, "/wiki/api/v2/spaces")
			http.Error(w, "bad path", http.StatusBadRequest)
			return
		}
		if got := r.URL.Query().Get("limit"); got != "10" {
			t.Errorf("limit = %q, want %q", got, "10")
			http.Error(w, "bad limit", http.StatusBadRequest)
			return
		}
		if got := r.URL.Query().Get("cursor"); got != "cur-1" {
			t.Errorf("cursor = %q, want %q", got, "cur-1")
			http.Error(w, "bad cursor", http.StatusBadRequest)
			return
		}
		writeFixtureResponse(t, w, "spaces_list.json")
	})
	defer srv.Close()

	tmp := t.TempDir()
	binPath := buildBinary(t, tmp)

	stdout, stderr, err := runBinary(binPath, []string{
		"--url", srv.URL, "--email", "a@b.com", "--token", "tok",
		"spaces", "list", "--cursor", "cur-1",
	}, "", envForIntegration(filepath.Join(tmp, "config"))...)
	if err != nil {
		t.Fatalf("spaces list failed: %v\nstdout=%s\nstderr=%s", err, stdout, stderr)
	}
	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("expected empty stderr, got %s", stderr)
	}

	var result struct {
		Results []struct {
			ID  string `json:"id"`
			Key string `json:"key"`
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
		t.Fatalf("spaces list output not valid JSON: %v\nstdout=%s", err, stdout)
	}
	if len(result.Results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(result.Results))
	}
	if result.Results[0].Key != "CF" {
		t.Fatalf("first key = %q, want %q", result.Results[0].Key, "CF")
	}
	if result.Page.Limit != 10 {
		t.Fatalf("page.limit = %d, want %d", result.Page.Limit, 10)
	}
	if result.Page.NextCursor != "eyJpZCI6MzIwOTc5MzMwNX0=" {
		t.Fatalf("page.nextCursor = %q, want %q", result.Page.NextCursor, "eyJpZCI6MzIwOTc5MzMwNX0=")
	}
	if result.Schema.ItemType != "space-summary" {
		t.Fatalf("schema.itemType = %q, want %q", result.Schema.ItemType, "space-summary")
	}
}

func TestSpacesListPlainContract_Integration(t *testing.T) {
	srv := newBinaryTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		writeFixtureResponse(t, w, "spaces_list.json")
	})
	defer srv.Close()

	tmp := t.TempDir()
	binPath := buildBinary(t, tmp)

	stdout, stderr, err := runBinary(binPath, []string{
		"--url", srv.URL, "--email", "a@b.com", "--token", "tok",
		"--format", "plain", "spaces", "list",
	}, "", envForIntegration(filepath.Join(tmp, "config"))...)
	if err != nil {
		t.Fatalf("spaces list --format plain failed: %v\nstdout=%s\nstderr=%s", err, stdout, stderr)
	}
	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("expected empty stderr, got %s", stderr)
	}
	if !strings.Contains(stdout, "ID") || !strings.Contains(stdout, "KEY") {
		t.Fatalf("expected tabular header, got:\n%s", stdout)
	}
	if !strings.Contains(stdout, "CF") || !strings.Contains(stdout, "ARCHBCSE") {
		t.Fatalf("expected space keys in output, got:\n%s", stdout)
	}
	if !strings.Contains(stdout, "Next cursor: eyJpZCI6MzIwOTc5MzMwNX0=") {
		t.Fatalf("expected next cursor line, got:\n%s", stdout)
	}
}

func TestPagesListJSONContract_Integration(t *testing.T) {
	srv := newBinaryTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/wiki/api/v2/pages" {
			t.Errorf("path = %q, want %q", r.URL.Path, "/wiki/api/v2/pages")
			http.Error(w, "bad path", http.StatusBadRequest)
			return
		}
		q := r.URL.Query()
		if got := q.Get("space-id"); got != "3082551269" {
			t.Errorf("space-id = %q, want %q", got, "3082551269")
			http.Error(w, "bad space-id", http.StatusBadRequest)
			return
		}
		if got := q.Get("limit"); got != "10" {
			t.Errorf("limit = %q, want %q", got, "10")
			http.Error(w, "bad limit", http.StatusBadRequest)
			return
		}
		if got := q.Get("cursor"); got != "cur-2" {
			t.Errorf("cursor = %q, want %q", got, "cur-2")
			http.Error(w, "bad cursor", http.StatusBadRequest)
			return
		}
		if got := q.Get("sort"); got != "title" {
			t.Errorf("sort = %q, want %q", got, "title")
			http.Error(w, "bad sort", http.StatusBadRequest)
			return
		}
		writeFixtureResponse(t, w, "pages_list.json")
	})
	defer srv.Close()

	tmp := t.TempDir()
	binPath := buildBinary(t, tmp)

	stdout, stderr, err := runBinary(binPath, []string{
		"--url", srv.URL, "--email", "a@b.com", "--token", "tok",
		"pages", "list", "--space-id", "3082551269", "--cursor", "cur-2", "--sort", "title",
	}, "", envForIntegration(filepath.Join(tmp, "config"))...)
	if err != nil {
		t.Fatalf("pages list failed: %v\nstdout=%s\nstderr=%s", err, stdout, stderr)
	}
	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("expected empty stderr, got %s", stderr)
	}

	var result struct {
		Results []struct {
			ID            string `json:"id"`
			Title         string `json:"title"`
			SpaceID       string `json:"spaceId"`
			VersionNumber int    `json:"versionNumber"`
		} `json:"results"`
		Page struct {
			Limit      int    `json:"limit"`
			NextCursor string `json:"nextCursor"`
		} `json:"page"`
	}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("pages list output not valid JSON: %v\nstdout=%s", err, stdout)
	}
	if len(result.Results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(result.Results))
	}
	if result.Results[0].Title != "BeCSEE Cloud Foundation" {
		t.Fatalf("first title = %q, want %q", result.Results[0].Title, "BeCSEE Cloud Foundation")
	}
	if result.Results[0].VersionNumber != 119 {
		t.Fatalf("versionNumber = %d, want %d", result.Results[0].VersionNumber, 119)
	}
	if result.Page.NextCursor != "eyJpZCI6IjMwOTAzMzQzOTQifQ==" {
		t.Fatalf("page.nextCursor = %q, want %q", result.Page.NextCursor, "eyJpZCI6IjMwOTAzMzQzOTQifQ==")
	}
}

func TestPagesGetDefaultMetadataOnly_Integration(t *testing.T) {
	srv := newBinaryTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("body-format"); got != "" {
			t.Errorf("body-format = %q, want empty", got)
			http.Error(w, "unexpected body-format", http.StatusBadRequest)
			return
		}
		writeJSONResponse(w, []byte(`{"id":"123","title":"Overview","spaceId":"S1","status":"current","version":{"number":7}}`))
	})
	defer srv.Close()

	tmp := t.TempDir()
	binPath := buildBinary(t, tmp)

	stdout, stderr, err := runBinary(binPath, []string{
		"--url", srv.URL, "--email", "a@b.com", "--token", "tok",
		"pages", "get", "--page-id", "123",
	}, "", envForIntegration(filepath.Join(tmp, "config"))...)
	if err != nil {
		t.Fatalf("pages get failed: %v\nstdout=%s\nstderr=%s", err, stdout, stderr)
	}
	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("expected empty stderr, got %s", stderr)
	}

	var result struct {
		Item map[string]any `json:"item"`
	}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("pages get output invalid JSON: %v\nstdout=%s", err, stdout)
	}
	if _, ok := result.Item["body"]; ok {
		t.Fatalf("expected metadata-only page output, got %s", stdout)
	}
}

func TestPagesGetBodyFormatView_Integration(t *testing.T) {
	pageJSON := `{"id":"123","title":"Overview","spaceId":"S1","status":"current","version":{"number":7},"body":{"view":{"representation":"view","value":"<h2>Overview</h2><p>Status: <strong>LIVE</strong></p>"}}}`

	srv := newBinaryTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("body-format"); got != "view" {
			t.Errorf("body-format = %q, want %q", got, "view")
			http.Error(w, "bad body-format", http.StatusBadRequest)
			return
		}
		writeJSONResponse(w, []byte(pageJSON))
	})
	defer srv.Close()

	tmp := t.TempDir()
	binPath := buildBinary(t, tmp)
	env := envForIntegration(filepath.Join(tmp, "config"))

	t.Run("json", func(t *testing.T) {
		stdout, stderr, err := runBinary(binPath, []string{
			"--url", srv.URL, "--email", "a@b.com", "--token", "tok",
			"pages", "get", "--page-id", "123", "--body-format", "view",
		}, "", env...)
		if err != nil {
			t.Fatalf("pages get JSON failed: %v\nstdout=%s\nstderr=%s", err, stdout, stderr)
		}
		if strings.TrimSpace(stderr) != "" {
			t.Fatalf("expected empty stderr, got %s", stderr)
		}

		var result struct {
			Item struct {
				Body struct {
					Format string `json:"format"`
					Value  string `json:"value"`
				} `json:"body"`
			} `json:"item"`
		}
		if err := json.Unmarshal([]byte(stdout), &result); err != nil {
			t.Fatalf("pages get JSON output invalid: %v\nstdout=%s", err, stdout)
		}
		if result.Item.Body.Format != "view" {
			t.Fatalf("body.format = %q, want %q", result.Item.Body.Format, "view")
		}
		if result.Item.Body.Value != "<h2>Overview</h2><p>Status: <strong>LIVE</strong></p>" {
			t.Fatalf("raw view body mismatch: %q", result.Item.Body.Value)
		}
	})

	t.Run("plain", func(t *testing.T) {
		stdout, stderr, err := runBinary(binPath, []string{
			"--url", srv.URL, "--email", "a@b.com", "--token", "tok",
			"--format", "plain", "pages", "get", "--page-id", "123", "--body-format", "view",
		}, "", env...)
		if err != nil {
			t.Fatalf("pages get plain failed: %v\nstdout=%s\nstderr=%s", err, stdout, stderr)
		}
		if strings.TrimSpace(stderr) != "" {
			t.Fatalf("expected empty stderr, got %s", stderr)
		}
		if !strings.Contains(stdout, "Body (view):") {
			t.Fatalf("expected body header, got:\n%s", stdout)
		}
		if !strings.Contains(stdout, "## Overview") {
			t.Fatalf("expected converted markdown heading, got:\n%s", stdout)
		}
		if !strings.Contains(stdout, "**LIVE**") {
			t.Fatalf("expected converted markdown emphasis, got:\n%s", stdout)
		}
		if strings.Contains(stdout, "<h2>") {
			t.Fatalf("expected HTML to be converted away, got:\n%s", stdout)
		}
	})
}

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

func TestPagesTreeContracts_Integration(t *testing.T) {
	srv := newBinaryTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/wiki/api/v2/pages/root/children" {
			t.Errorf("path = %q, want %q", r.URL.Path, "/wiki/api/v2/pages/root/children")
			http.Error(w, "bad path", http.StatusBadRequest)
			return
		}
		switch r.URL.Query().Get("cursor") {
		case "":
			if got := r.URL.Query().Get("limit"); got != "10" {
				t.Errorf("limit = %q, want %q", got, "10")
				http.Error(w, "bad limit", http.StatusBadRequest)
				return
			}
			writeJSONResponse(w, []byte(`{"results":[{"id":"child-1","title":"One","spaceId":"S1","status":"current"}],"_links":{"next":"/wiki/api/v2/pages/root/children?limit=10&cursor=page-2"}}`))
		case "page-2":
			// not expected in the bounded tree implementation; if it happens the CLI is no longer bounded
			t.Errorf("unexpected root pagination request for cursor=%q", r.URL.Query().Get("cursor"))
			http.Error(w, "unexpected pagination", http.StatusBadRequest)
		default:
			writeJSONResponse(w, []byte(`{"results":[],"_links":{}}`))
		}
	})
	defer srv.Close()

	tmp := t.TempDir()
	binPath := buildBinary(t, tmp)
	env := envForIntegration(filepath.Join(tmp, "config"))

	t.Run("json", func(t *testing.T) {
		stdout, stderr, err := runBinary(binPath, []string{
			"--url", srv.URL, "--email", "a@b.com", "--token", "tok",
			"pages", "tree", "--page-id", "root",
		}, "", env...)
		if err != nil {
			t.Fatalf("pages tree failed: %v\nstdout=%s\nstderr=%s", err, stdout, stderr)
		}
		if strings.TrimSpace(stderr) != "" {
			t.Fatalf("expected empty stderr, got %s", stderr)
		}

		var result struct {
			Item struct {
				RootPageID      string `json:"rootPageId"`
				LimitPerLevel   int    `json:"limitPerLevel"`
				HasMoreChildren bool   `json:"hasMoreChildren"`
				Children        []struct {
					ID string `json:"id"`
				} `json:"children"`
			} `json:"item"`
		}
		if err := json.Unmarshal([]byte(stdout), &result); err != nil {
			t.Fatalf("pages tree output invalid JSON: %v\nstdout=%s", err, stdout)
		}
		if result.Item.RootPageID != "root" {
			t.Fatalf("rootPageId = %q, want %q", result.Item.RootPageID, "root")
		}
		if result.Item.LimitPerLevel != 10 {
			t.Fatalf("limitPerLevel = %d, want %d", result.Item.LimitPerLevel, 10)
		}
		if !result.Item.HasMoreChildren {
			t.Fatalf("expected hasMoreChildren=true, got %s", stdout)
		}
		if len(result.Item.Children) != 1 || result.Item.Children[0].ID != "child-1" {
			t.Fatalf("unexpected children: %s", stdout)
		}
	})

	t.Run("plain", func(t *testing.T) {
		stdout, stderr, err := runBinary(binPath, []string{
			"--url", srv.URL, "--email", "a@b.com", "--token", "tok",
			"--format", "plain", "pages", "tree", "--page-id", "root",
		}, "", env...)
		if err != nil {
			t.Fatalf("pages tree plain failed: %v\nstdout=%s\nstderr=%s", err, stdout, stderr)
		}
		if strings.TrimSpace(stderr) != "" {
			t.Fatalf("expected empty stderr, got %s", stderr)
		}
		if !strings.Contains(stdout, "Root page: root") {
			t.Fatalf("expected root header, got:\n%s", stdout)
		}
		if !strings.Contains(stdout, "More children available: yes") {
			t.Fatalf("expected truncation hint, got:\n%s", stdout)
		}
	})
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
