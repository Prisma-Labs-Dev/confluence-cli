package cli

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"
)

type AuthCmd struct {
	Login AuthLoginCmd `cmd:"" help:"Store credentials securely"`
}

type AuthLoginCmd struct {
	StdinJSON  bool `name:"stdin-json" help:"Read credentials JSON from stdin (url,email,token)"`
	TokenStdin bool `name:"token-stdin" help:"Read token from stdin"`
	NoPrompt   bool `name:"no-prompt" help:"Do not prompt for missing fields; fail instead" default:"false"`
}

func (cmd *AuthLoginCmd) Run(app *App) error {
	creds := Credentials{
		URL:   strings.TrimSpace(app.URL),
		Email: strings.TrimSpace(app.Email),
		Token: strings.TrimSpace(app.Token),
	}

	if cmd.StdinJSON {
		stdinCreds, err := readCredentialsJSON(os.Stdin)
		if err != nil {
			return fmt.Errorf("read stdin credentials: %w", err)
		}
		if creds.URL == "" {
			creds.URL = stdinCreds.URL
		}
		if creds.Email == "" {
			creds.Email = stdinCreds.Email
		}
		if creds.Token == "" {
			creds.Token = stdinCreds.Token
		}
	}
	if cmd.TokenStdin && creds.Token == "" {
		token, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("read token from stdin: %w", err)
		}
		creds.Token = strings.TrimSpace(string(token))
	}

	interactive := term.IsTerminal(int(os.Stdin.Fd())) && !cmd.NoPrompt

	var err error
	if creds.URL == "" && interactive {
		creds.URL, err = promptLine("Confluence URL: ")
		if err != nil {
			return fmt.Errorf("read url: %w", err)
		}
	}
	if creds.Email == "" && interactive {
		creds.Email, err = promptLine("Atlassian email: ")
		if err != nil {
			return fmt.Errorf("read email: %w", err)
		}
	}
	if creds.Token == "" && interactive {
		creds.Token, err = promptSecret("Atlassian API token: ")
		if err != nil {
			return fmt.Errorf("read token: %w", err)
		}
	}

	creds.URL = strings.TrimSpace(creds.URL)
	creds.Email = strings.TrimSpace(creds.Email)
	creds.Token = strings.TrimSpace(creds.Token)
	if creds.URL == "" || creds.Email == "" || creds.Token == "" {
		return fmt.Errorf("missing required fields: --url, --email, --token (or use --stdin-json / --token-stdin)")
	}

	location, err := saveStoredCredentials(creds)
	if err != nil {
		return err
	}

	if app.Plain {
		fmt.Fprintf(app.Stdout, "Stored credentials in %s\n", location)
		return nil
	}
	return renderJSON(app.Stdout, struct {
		StoredIn string `json:"storedIn"`
	}{
		StoredIn: location,
	})
}

func readCredentialsJSON(r io.Reader) (Credentials, error) {
	var c Credentials
	b, err := io.ReadAll(r)
	if err != nil {
		return Credentials{}, err
	}
	if err := json.Unmarshal(b, &c); err != nil {
		return Credentials{}, err
	}
	return c, nil
}

func promptLine(label string) (string, error) {
	fmt.Fprint(os.Stderr, label)
	r := bufio.NewReader(os.Stdin)
	v, err := r.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}
	return strings.TrimSpace(v), nil
}

func promptSecret(label string) (string, error) {
	fmt.Fprint(os.Stderr, label)
	b, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stderr)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
