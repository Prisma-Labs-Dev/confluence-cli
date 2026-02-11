package confluence_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
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

	var loginResp struct {
		StoredIn string `json:"storedIn"`
	}
	if err := json.Unmarshal([]byte(stdout), &loginResp); err != nil {
		t.Fatalf("auth login output not valid JSON: %v\nstdout=%s", err, stdout)
	}
	if loginResp.StoredIn != "file" {
		t.Fatalf("storedIn = %q, want %q", loginResp.StoredIn, "file")
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
	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("expected empty stderr, got: %s", stderr)
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

	var loginResp struct {
		StoredIn string `json:"storedIn"`
	}
	if err := json.Unmarshal([]byte(stdout), &loginResp); err != nil {
		t.Fatalf("auth login output not valid JSON: %v\nstdout=%s", err, stdout)
	}
	if loginResp.StoredIn != "file" {
		t.Fatalf("storedIn = %q, want %q", loginResp.StoredIn, "file")
	}

	stdout, stderr, err = runBinary(binPath, []string{"spaces", "list"}, "", envForIntegration(configDir)...)
	if err != nil {
		t.Fatalf("spaces list failed: %v\nstdout=%s\nstderr=%s", err, stdout, stderr)
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
	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("expected empty stderr, got: %s", stderr)
	}
}

func TestAuthLoginNoPrompt_Integration(t *testing.T) {
	tmp := t.TempDir()
	configDir := filepath.Join(tmp, "config")
	binPath := buildBinary(t, tmp)

	stdout, stderr, err := runBinary(binPath, []string{"auth", "login", "--no-prompt"}, "", envForIntegration(configDir)...)
	if err == nil {
		t.Fatalf("expected auth login --no-prompt to fail; stdout=%s stderr=%s", stdout, stderr)
	}

	var errResp struct {
		Error string `json:"error"`
		Code  string `json:"code"`
	}
	if err := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &errResp); err != nil {
		t.Fatalf("stderr is not valid error JSON: %v\nstderr=%s", err, stderr)
	}
	if errResp.Code != "VALIDATION" {
		t.Fatalf("error code = %q, want %q", errResp.Code, "VALIDATION")
	}
	if !strings.Contains(errResp.Error, "missing required fields") {
		t.Fatalf("unexpected error message: %s", errResp.Error)
	}
}

func TestGlobalHelpDoesNotExposeColorFlag_Integration(t *testing.T) {
	tmp := t.TempDir()
	binPath := buildBinary(t, tmp)

	stdout, stderr, err := runBinary(binPath, []string{"--help"}, "", envForIntegration(filepath.Join(tmp, "config"))...)
	if err != nil {
		t.Fatalf("help failed: %v\nstdout=%s\nstderr=%s", err, stdout, stderr)
	}
	if strings.Contains(stdout, "--color") {
		t.Fatalf("help should not expose --color: %s", stdout)
	}
}

func TestAuthLoginHelpIncludesPromptFlag_Integration(t *testing.T) {
	tmp := t.TempDir()
	binPath := buildBinary(t, tmp)

	stdout, stderr, err := runBinary(binPath, []string{"auth", "login", "--help"}, "", envForIntegration(filepath.Join(tmp, "config"))...)
	if err != nil {
		t.Fatalf("auth login help failed: %v\nstdout=%s\nstderr=%s", err, stdout, stderr)
	}
	if !strings.Contains(stdout, "--prompt") {
		t.Fatalf("auth login help should include --prompt: %s", stdout)
	}
}

func buildBinary(t *testing.T, dir string) string {
	t.Helper()
	binPath := filepath.Join(dir, "confluence")
	build := exec.Command("go", "build", "-o", binPath, "./cmd/confluence")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build binary: %v\n%s", err, string(out))
	}
	return binPath
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

func envForIntegration(configDir string) []string {
	return []string{
		"CONFLUENCE_DISABLE_KEYCHAIN=1",
		"CONFLUENCE_CONFIG_DIR=" + configDir,
		"CONFLUENCE_URL=",
		"CONFLUENCE_EMAIL=",
		"CONFLUENCE_API_TOKEN=",
	}
}

func runBinary(bin string, args []string, stdin string, extraEnv ...string) (stdout, stderr string, runErr error) {
	cmd := exec.Command(bin, args...)
	cmd.Env = append(os.Environ(), extraEnv...)
	cmd.Stdin = strings.NewReader(stdin)

	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	runErr = cmd.Run()
	return outBuf.String(), errBuf.String(), runErr
}
