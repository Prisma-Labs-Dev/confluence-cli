package confluence_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAuthLoginStdinJSON_Integration(t *testing.T) {
	srv := integrationServer()
	defer srv.Close()

	tmp := t.TempDir()
	configDir := filepath.Join(tmp, "config")
	binPath := buildBinary(t, tmp)

	stdinJSON := fmt.Sprintf(`{"url":%q,"email":"a@b.com","token":"tok"}`, srv.URL)
	stdout, stderr, err := runBinary(binPath, []string{"auth", "login", "--stdin-json"}, stdinJSON, envForIntegration(configDir)...)
	if err != nil {
		t.Fatalf("auth login failed: %v\nstdout=%s\nstderr=%s", err, stdout, stderr)
	}
	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("expected empty stderr, got: %s", stderr)
	}

	var loginResp struct {
		Item struct {
			StoredIn string `json:"storedIn"`
		} `json:"item"`
	}
	if err := json.Unmarshal([]byte(stdout), &loginResp); err != nil {
		t.Fatalf("auth login output not valid JSON: %v\nstdout=%s", err, stdout)
	}
	if loginResp.Item.StoredIn != "file" {
		t.Fatalf("storedIn = %q, want %q", loginResp.Item.StoredIn, "file")
	}

	credFile := filepath.Join(configDir, "credentials.json")
	info, err := os.Stat(credFile)
	if err != nil {
		t.Fatalf("expected credentials file %s: %v", credFile, err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("credentials file perms = %o, want 600", info.Mode().Perm())
	}

	stdout, stderr, err = runBinary(binPath, []string{"spaces", "list"}, "", envForIntegration(configDir)...)
	if err != nil {
		t.Fatalf("spaces list failed: %v\nstdout=%s\nstderr=%s", err, stdout, stderr)
	}
	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("expected empty stderr, got: %s", stderr)
	}

	var spacesResp struct {
		Results []struct {
			Key string `json:"key"`
		} `json:"results"`
	}
	if err := json.Unmarshal([]byte(stdout), &spacesResp); err != nil {
		t.Fatalf("spaces list output not valid JSON: %v\nstdout=%s", err, stdout)
	}
	if len(spacesResp.Results) != 1 || spacesResp.Results[0].Key != "DEV" {
		t.Fatalf("unexpected spaces list output: %s", stdout)
	}
}

func TestAuthLoginTokenStdin_Integration(t *testing.T) {
	srv := integrationServer()
	defer srv.Close()

	tmp := t.TempDir()
	configDir := filepath.Join(tmp, "config")
	binPath := buildBinary(t, tmp)

	stdout, stderr, err := runBinary(binPath, []string{
		"--url", srv.URL, "--email", "a@b.com", "auth", "login", "--token-stdin",
	}, "tok", envForIntegration(configDir)...)
	if err != nil {
		t.Fatalf("auth login --token-stdin failed: %v\nstdout=%s\nstderr=%s", err, stdout, stderr)
	}
	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("expected empty stderr, got: %s", stderr)
	}

	var loginResp struct {
		Item struct {
			StoredIn string `json:"storedIn"`
		} `json:"item"`
	}
	if err := json.Unmarshal([]byte(stdout), &loginResp); err != nil {
		t.Fatalf("auth login output not valid JSON: %v\nstdout=%s", err, stdout)
	}
	if loginResp.Item.StoredIn != "file" {
		t.Fatalf("storedIn = %q, want %q", loginResp.Item.StoredIn, "file")
	}
}

func TestAuthLoginRejectsMixedInputModes_Integration(t *testing.T) {
	tmp := t.TempDir()
	configDir := filepath.Join(tmp, "config")
	binPath := buildBinary(t, tmp)

	stdout, stderr, exitCode, err := runBinaryWithExitCode(binPath, []string{
		"--url", "https://example.atlassian.net", "auth", "login", "--stdin-json",
	}, `{"url":"https://example.atlassian.net","email":"a@b.com","token":"tok"}`, envForIntegration(configDir)...)
	if err == nil {
		t.Fatalf("expected mixed input mode failure\nstdout=%s\nstderr=%s", stdout, stderr)
	}
	if exitCode != 2 {
		t.Fatalf("exit code = %d, want %d", exitCode, 2)
	}

	var errResp struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &errResp); err != nil {
		t.Fatalf("stderr is not valid error JSON: %v\nstderr=%s", err, stderr)
	}
	if errResp.Error.Code != "VALIDATION" {
		t.Fatalf("error code = %q, want %q", errResp.Error.Code, "VALIDATION")
	}
	if !strings.Contains(errResp.Error.Message, "cannot be combined") {
		t.Fatalf("unexpected error message: %s", errResp.Error.Message)
	}
}

func integrationServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/wiki/api/v2/spaces" {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"errors":[{"message":"Not found"}]}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"results":[{"id":"1","key":"DEV","name":"Development","type":"global","status":"current"}]}`))
	}))
}
