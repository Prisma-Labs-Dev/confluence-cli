package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"
)

var isTerminal = term.IsTerminal

type AuthLoginCmd struct {
	StdinJSON  bool `name:"stdin-json" help:"Read {url,email,token} JSON from piped stdin"`
	TokenStdin bool `name:"token-stdin" help:"Read token from piped stdin"`
}

func (cmd *AuthLoginCmd) Run(app *App) error {
	if cmd.StdinJSON && cmd.TokenStdin {
		return validationError("--stdin-json and --token-stdin are mutually exclusive", helpHint("auth login"))
	}

	stdinIsTerminal := isTerminal(int(os.Stdin.Fd()))
	creds := Credentials{
		URL:   strings.TrimSpace(app.URL),
		Email: strings.TrimSpace(app.Email),
		Token: strings.TrimSpace(app.Token),
	}

	switch {
	case cmd.StdinJSON:
		if creds.URL != "" || creds.Email != "" || creds.Token != "" {
			return validationError("--stdin-json cannot be combined with --url, --email, or --token", helpHint("auth login"))
		}
		if stdinIsTerminal {
			return validationError("--stdin-json requires piped stdin", helpHint("auth login"))
		}
		stdinCreds, err := readCredentialsJSON(os.Stdin)
		if err != nil {
			return validationErrorf(helpHint("auth login"), "read stdin credentials: %v", err)
		}
		creds = stdinCreds
	case cmd.TokenStdin:
		if creds.Token != "" {
			return validationError("--token-stdin cannot be combined with --token", helpHint("auth login"))
		}
		if creds.URL == "" || creds.Email == "" {
			return validationError("--token-stdin requires --url and --email (or matching env vars)", helpHint("auth login"))
		}
		if stdinIsTerminal {
			return validationError("--token-stdin requires piped stdin", helpHint("auth login"))
		}
		token, err := io.ReadAll(os.Stdin)
		if err != nil {
			return validationErrorf(helpHint("auth login"), "read token from stdin: %v", err)
		}
		creds.Token = strings.TrimSpace(string(token))
	default:
		// flags/env-only mode
	}

	creds.URL = strings.TrimSpace(creds.URL)
	creds.Email = strings.TrimSpace(creds.Email)
	creds.Token = strings.TrimSpace(creds.Token)
	if creds.URL == "" || creds.Email == "" || creds.Token == "" {
		return validationError("auth login requires url, email, and token; use flags/env vars, --stdin-json, or --token-stdin", helpHint("auth login"))
	}

	storedIn, err := saveStoredCredentials(creds)
	if err != nil {
		return fmt.Errorf("store credentials: %w", err)
	}

	response := itemEnvelope(AuthLoginInfo{StoredIn: storedIn}, "auth-login", []string{"storedIn"})
	if app.IsPlain() {
		discardWrite(fmt.Fprintf(app.Stdout, "Stored credentials in %s\n", storedIn))
		return nil
	}
	return renderJSON(app.Stdout, response)
}

func readCredentialsJSON(r io.Reader) (Credentials, error) {
	var creds Credentials
	data, err := io.ReadAll(r)
	if err != nil {
		return Credentials{}, err
	}
	if err := json.Unmarshal(data, &creds); err != nil {
		return Credentials{}, err
	}
	return creds, nil
}
