package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	defaultKeychainService = "com.prismalabs.confluence-cli"
	keychainAccount        = "default"
)

var (
	errNotFound = errors.New("not found")
)

type Credentials struct {
	URL   string `json:"url"`
	Email string `json:"email"`
	Token string `json:"token"`
}

func resolveCredentials(in Credentials) (Credentials, error) {
	if in.URL != "" && in.Email != "" && in.Token != "" {
		return in, nil
	}

	stored, err := loadStoredCredentials()
	if err != nil {
		return Credentials{}, err
	}
	if in.URL == "" {
		in.URL = stored.URL
	}
	if in.Email == "" {
		in.Email = stored.Email
	}
	if in.Token == "" {
		in.Token = stored.Token
	}
	return in, nil
}

func loadStoredCredentials() (Credentials, error) {
	creds, err := loadFromKeychain()
	if err == nil {
		return creds, nil
	}
	if !errors.Is(err, errNotFound) {
		// If keychain exists but failed unexpectedly, still allow file fallback.
	}

	creds, err = loadFromFile()
	if err == nil || errors.Is(err, errNotFound) {
		return creds, nil
	}
	return Credentials{}, err
}

func saveStoredCredentials(creds Credentials) (string, error) {
	if err := saveToKeychain(creds); err == nil {
		return "keychain", nil
	}

	if err := saveToFile(creds); err == nil {
		return "file", nil
	}
	return "", fmt.Errorf("unable to store credentials in keychain or file")
}

func keychainService() string {
	if v := strings.TrimSpace(os.Getenv("CONFLUENCE_KEYCHAIN_SERVICE")); v != "" {
		return v
	}
	return defaultKeychainService
}

func keychainDisabled() bool {
	return os.Getenv("CONFLUENCE_DISABLE_KEYCHAIN") == "1"
}

func loadFromKeychain() (Credentials, error) {
	if keychainDisabled() {
		return Credentials{}, errNotFound
	}

	cmd := exec.Command("security", "find-generic-password", "-s", keychainService(), "-a", keychainAccount, "-w")
	out, err := cmd.CombinedOutput()
	if err != nil {
		msg := string(out)
		if strings.Contains(msg, "could not be found") {
			return Credentials{}, errNotFound
		}
		return Credentials{}, errNotFound
	}

	var creds Credentials
	if err := json.Unmarshal([]byte(strings.TrimSpace(string(out))), &creds); err != nil {
		return Credentials{}, fmt.Errorf("parse keychain credentials: %w", err)
	}
	return creds, nil
}

func saveToKeychain(creds Credentials) error {
	if keychainDisabled() {
		return errNotFound
	}

	b, err := json.Marshal(creds)
	if err != nil {
		return err
	}
	cmd := exec.Command("security", "add-generic-password", "-U", "-s", keychainService(), "-a", keychainAccount, "-w", string(b))
	if out, err := cmd.CombinedOutput(); err != nil {
		_ = out
		return err
	}
	return nil
}

func configDir() (string, error) {
	if v := strings.TrimSpace(os.Getenv("CONFLUENCE_CONFIG_DIR")); v != "" {
		return v, nil
	}
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "confluence-cli"), nil
}

func credentialsPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "credentials.json"), nil
}

func loadFromFile() (Credentials, error) {
	p, err := credentialsPath()
	if err != nil {
		return Credentials{}, err
	}
	b, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return Credentials{}, errNotFound
		}
		return Credentials{}, err
	}

	var creds Credentials
	if err := json.Unmarshal(b, &creds); err != nil {
		return Credentials{}, fmt.Errorf("parse %s: %w", p, err)
	}
	return creds, nil
}

func saveToFile(creds Credentials) error {
	p, err := credentialsPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
		return err
	}

	b, err := json.MarshalIndent(creds, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	return os.WriteFile(p, b, 0o600)
}
