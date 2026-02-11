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

var isTerminal = term.IsTerminal

type AuthCmd struct {
	Login AuthLoginCmd `cmd:"" help:"Store credentials securely"`
}

type AuthLoginCmd struct {
	StdinJSON  bool `name:"stdin-json" help:"Read credentials JSON from stdin (url,email,token)"`
	TokenStdin bool `name:"token-stdin" help:"Read token from stdin"`
	Prompt     bool `name:"prompt" help:"Allow interactive prompts for missing fields (human use)" default:"false"`
	NoPrompt   bool `name:"no-prompt" help:"Do not prompt for missing fields; fail instead" default:"false"`
}

func (cmd *AuthLoginCmd) Run(app *App) error {
	creds := Credentials{
		URL:   strings.TrimSpace(app.URL),
		Email: strings.TrimSpace(app.Email),
		Token: strings.TrimSpace(app.Token),
	}
	stdinIsTerminal := isTerminal(int(os.Stdin.Fd()))

	if cmd.StdinJSON {
		if stdinIsTerminal {
			return validationErrorf("--stdin-json requires piped stdin; refusing terminal input")
		}
		stdinCreds, err := readCredentialsJSON(os.Stdin)
		if err != nil {
			return validationErrorf("read stdin credentials: %v", err)
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
		if stdinIsTerminal {
			return validationErrorf("--token-stdin requires piped stdin; refusing terminal input")
		}
		token, err := io.ReadAll(os.Stdin)
		if err != nil {
			return validationErrorf("read token from stdin: %v", err)
		}
		creds.Token = strings.TrimSpace(string(token))
	}

	interactive := stdinIsTerminal && cmd.Prompt && !cmd.NoPrompt

	var err error
	if creds.URL == "" && interactive {
		creds.URL, err = promptLine("Confluence URL: ")
		if err != nil {
			return validationErrorf("read url: %v", err)
		}
	}
	if creds.Email == "" && interactive {
		creds.Email, err = promptLine("Atlassian email: ")
		if err != nil {
			return validationErrorf("read email: %v", err)
		}
	}
	if creds.Token == "" && interactive {
		creds.Token, err = promptSecret("Atlassian API token: ")
		if err != nil {
			return validationErrorf("read token: %v", err)
		}
	}

	creds.URL = strings.TrimSpace(creds.URL)
	creds.Email = strings.TrimSpace(creds.Email)
	creds.Token = strings.TrimSpace(creds.Token)
	if creds.URL == "" || creds.Email == "" || creds.Token == "" {
		return validationErrorf("missing required fields: --url, --email, --token (or use --stdin-json / --token-stdin)")
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
