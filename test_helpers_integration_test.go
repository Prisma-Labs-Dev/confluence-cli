package confluence_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func buildBinary(t *testing.T, dir string) string {
	t.Helper()
	binPath := filepath.Join(dir, "confluence")
	build := exec.Command("go", "build", "-o", binPath, "./cmd/confluence")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build binary: %v\n%s", err, string(out))
	}
	return binPath
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

func runBinaryWithExitCode(bin string, args []string, stdin string, extraEnv ...string) (stdout, stderr string, exitCode int, runErr error) {
	stdout, stderr, runErr = runBinary(bin, args, stdin, extraEnv...)
	if runErr == nil {
		return stdout, stderr, 0, nil
	}

	var exitErr *exec.ExitError
	if errors.As(runErr, &exitErr) {
		return stdout, stderr, exitErr.ExitCode(), runErr
	}
	return stdout, stderr, -1, runErr
}

func newBinaryTestServer(t *testing.T, handler func(http.ResponseWriter, *http.Request)) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.Header.Get("Authorization"), "Basic ") {
			t.Errorf("missing basic auth header")
			http.Error(w, "missing auth", http.StatusUnauthorized)
			return
		}
		handler(w, r)
	}))
}

func writeFixtureResponse(t *testing.T, w http.ResponseWriter, name string) {
	t.Helper()
	body, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	writeJSONResponse(w, body)
}

func writeJSONResponse(w http.ResponseWriter, body []byte) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(body)
}

func readGoldenFile(t *testing.T, name string) string {
	t.Helper()
	body, err := os.ReadFile(filepath.Join("testdata", "golden", name))
	if err != nil {
		t.Fatalf("read golden %s: %v", name, err)
	}
	return string(body)
}

func writePrettyJSONGolden(t *testing.T, path string, value any) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir golden dir: %v", err)
	}
	body, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		t.Fatalf("marshal golden JSON: %v", err)
	}
	body = append(body, '\n')
	if err := os.WriteFile(path, body, 0o644); err != nil {
		t.Fatalf("write golden %s: %v", path, err)
	}
}

func readJSONGolden(t *testing.T, path string, out any) {
	t.Helper()
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden %s: %v", path, err)
	}
	if err := json.Unmarshal(body, out); err != nil {
		t.Fatalf("parse golden %s: %v", path, err)
	}
}
