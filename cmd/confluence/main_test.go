package main

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/Prisma-Labs-Dev/confluence-cli/internal/cli"
)

func TestVersionCommand(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{"version"}, &stdout, &stderr, "1.0.0-test")
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d (stderr: %s)", code, stderr.String())
	}

	var result struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse JSON output: %v", err)
	}
	if result.Version != "1.0.0-test" {
		t.Fatalf("expected version 1.0.0-test, got %s", result.Version)
	}
}

func TestVersionPlain(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{"version", "--plain"}, &stdout, &stderr, "1.0.0-test")
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d (stderr: %s)", code, stderr.String())
	}
	if stdout.String() != "confluence 1.0.0-test\n" {
		t.Fatalf("expected 'confluence 1.0.0-test\\n', got %q", stdout.String())
	}
}

func TestMissingAuth(t *testing.T) {
	t.Setenv("CONFLUENCE_DISABLE_KEYCHAIN", "1")
	t.Setenv("CONFLUENCE_CONFIG_DIR", t.TempDir())
	t.Setenv("CONFLUENCE_URL", "")
	t.Setenv("CONFLUENCE_EMAIL", "")
	t.Setenv("CONFLUENCE_API_TOKEN", "")

	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{"spaces", "list"}, &stdout, &stderr, "1.0.0-test")
	if code != cli.ExitValidation {
		t.Fatalf("expected exit code %d, got %d", cli.ExitValidation, code)
	}

	var errResult struct {
		Error string `json:"error"`
		Code  string `json:"code"`
	}
	if err := json.Unmarshal(stderr.Bytes(), &errResult); err != nil {
		t.Fatalf("failed to parse error JSON: %v", err)
	}
	if errResult.Code != "VALIDATION" {
		t.Fatalf("expected code VALIDATION, got %s", errResult.Code)
	}
}

func TestNoCommand(t *testing.T) {
	t.Setenv("CONFLUENCE_DISABLE_KEYCHAIN", "1")
	t.Setenv("CONFLUENCE_CONFIG_DIR", t.TempDir())

	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{}, &stdout, &stderr, "1.0.0-test")
	if code != cli.ExitValidation {
		t.Fatalf("expected exit code %d, got %d", cli.ExitValidation, code)
	}
}
