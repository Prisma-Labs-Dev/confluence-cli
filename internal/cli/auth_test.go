package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"
)

func runCLIForTest(t *testing.T, args []string, stdinIsTTY bool) (stdout, stderr string, exitCode int) {
	t.Helper()
	oldTerminal := isTerminal
	isTerminal = func(int) bool { return stdinIsTTY }
	defer func() { isTerminal = oldTerminal }()

	envKeys := []string{"CONFLUENCE_URL", "CONFLUENCE_EMAIL", "CONFLUENCE_API_TOKEN"}
	oldEnv := map[string]string{}
	for _, key := range envKeys {
		oldEnv[key] = os.Getenv(key)
		if err := os.Unsetenv(key); err != nil {
			t.Fatalf("unset %s: %v", key, err)
		}
	}
	defer func() {
		for _, key := range envKeys {
			if oldEnv[key] == "" {
				_ = os.Unsetenv(key)
				continue
			}
			_ = os.Setenv(key, oldEnv[key])
		}
	}()

	var outBuf, errBuf bytes.Buffer
	code := Run(args, &outBuf, &errBuf, "test-version")
	return outBuf.String(), errBuf.String(), code
}

func TestAuthLoginStdinJSONRejectsTTY(t *testing.T) {
	_, stderr, exitCode := runCLIForTest(t, []string{"auth", "login", "--stdin-json"}, true)
	if exitCode != ExitValidation {
		t.Fatalf("exit code = %d, want %d", exitCode, ExitValidation)
	}

	var payload struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal([]byte(stderr), &payload); err != nil {
		t.Fatalf("parse error JSON: %v\nstderr=%s", err, stderr)
	}
	if payload.Error.Code != "VALIDATION" {
		t.Fatalf("code = %q, want %q", payload.Error.Code, "VALIDATION")
	}
	if payload.Error.Message != "--stdin-json requires piped stdin" {
		t.Fatalf("unexpected message: %q", payload.Error.Message)
	}
}

func TestAuthLoginTokenStdinRejectsTTY(t *testing.T) {
	_, stderr, exitCode := runCLIForTest(t, []string{"--url", "https://example.atlassian.net", "--email", "a@b.com", "auth", "login", "--token-stdin"}, true)
	if exitCode != ExitValidation {
		t.Fatalf("exit code = %d, want %d", exitCode, ExitValidation)
	}

	var payload struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal([]byte(stderr), &payload); err != nil {
		t.Fatalf("parse error JSON: %v\nstderr=%s", err, stderr)
	}
	if payload.Error.Code != "VALIDATION" {
		t.Fatalf("code = %q, want %q", payload.Error.Code, "VALIDATION")
	}
	if payload.Error.Message != "--token-stdin requires piped stdin" {
		t.Fatalf("unexpected message: %q", payload.Error.Message)
	}
}
