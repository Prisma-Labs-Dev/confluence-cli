package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	confluence "github.com/Prisma-Labs-Dev/confluence-cli"
	"github.com/alecthomas/kong"
)

const (
	ExitOK         = 0
	ExitError      = 1
	ExitValidation = 2
	ExitAuth       = 3
)

type App struct {
	Client  *confluence.Client
	Stdout  io.Writer
	Stderr  io.Writer
	Plain   bool
	Color   bool
	Version string
	URL     string
	Email   string
	Token   string
}

type CLIError struct {
	Message string `json:"error"`
	Code    string `json:"code"`
}

func writeError(w io.Writer, msg, code string) {
	e := CLIError{Message: msg, Code: code}
	b, _ := json.Marshal(e)
	fmt.Fprintln(w, string(b))
}

func Run(args []string, stdout, stderr io.Writer, version string) int {
	var cli CLI
	parser, err := kong.New(&cli,
		kong.Name("confluence"),
		kong.Description("CLI for Confluence Cloud API"),
		kong.UsageOnError(),
		kong.Writers(stdout, stderr),
	)
	if err != nil {
		writeError(stderr, err.Error(), "INTERNAL")
		return ExitError
	}

	ctx, err := parser.Parse(args)
	if err != nil {
		writeError(stderr, err.Error(), "VALIDATION")
		return ExitValidation
	}

	app := &App{
		Stdout:  stdout,
		Stderr:  stderr,
		Plain:   cli.Plain,
		Color:   cli.Color,
		Version: version,
		URL:     cli.URL,
		Email:   cli.Email,
		Token:   cli.Token,
	}

	// Only create client for commands that need it.
	if ctx.Command() != "version" && ctx.Command() != "auth login" {
		creds, err := resolveCredentials(Credentials{
			URL:   cli.URL,
			Email: cli.Email,
			Token: cli.Token,
		})
		if err != nil {
			writeError(stderr, err.Error(), "AUTH_STORE")
			return ExitError
		}
		app.URL = creds.URL
		app.Email = creds.Email
		app.Token = creds.Token

		if creds.URL == "" || creds.Email == "" || creds.Token == "" {
			missing := []string{}
			if creds.URL == "" {
				missing = append(missing, "--url (or CONFLUENCE_URL)")
			}
			if creds.Email == "" {
				missing = append(missing, "--email (or CONFLUENCE_EMAIL)")
			}
			if creds.Token == "" {
				missing = append(missing, "--token (or CONFLUENCE_API_TOKEN)")
			}
			writeError(stderr, fmt.Sprintf("missing required flags: %s", strings.Join(missing, ", ")), "VALIDATION")
			return ExitValidation
		}
		app.Client = confluence.NewClient(confluence.Options{
			BaseURL: creds.URL,
			Email:   creds.Email,
			Token:   creds.Token,
			Timeout: cli.Timeout,
		})
	}

	if err := ctx.Run(app); err != nil {
		var apiErr *confluence.APIError
		if errors.As(err, &apiErr) {
			code := "API_ERROR"
			exitCode := ExitError
			if apiErr.StatusCode == 401 || apiErr.StatusCode == 403 {
				code = "AUTH_FAILED"
				exitCode = ExitAuth
			}
			writeError(stderr, apiErr.Error(), code)
			return exitCode
		}
		writeError(stderr, err.Error(), "ERROR")
		return ExitError
	}

	return ExitOK
}
