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

// App carries runtime state used by command implementations.
type App struct {
	Client  *confluence.Client
	Stdout  io.Writer
	Stderr  io.Writer
	Format  string
	Version string
	URL     string
	Email   string
	Token   string
}

func (app *App) IsPlain() bool {
	return app.Format == "plain"
}

type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Hint    string `json:"hint,omitempty"`
}

type ErrorEnvelope struct {
	Error ErrorDetail `json:"error"`
}

func writeError(w io.Writer, code, message, hint string) {
	payload := ErrorEnvelope{
		Error: ErrorDetail{
			Code:    code,
			Message: message,
			Hint:    hint,
		},
	}
	b, _ := json.Marshal(payload)
	discardWrite(fmt.Fprintln(w, string(b)))
}

func Run(args []string, stdout, stderr io.Writer, version string) int {
	if maybeWriteHelp(args, stdout) {
		return ExitOK
	}
	args = stripHelpFlags(args)

	var cli CLI
	parser, err := kong.New(&cli,
		kong.Name("confluence"),
		kong.Description("Agent-first Confluence Cloud CLI"),
		kong.Writers(stdout, stderr),
	)
	if err != nil {
		writeError(stderr, "INTERNAL", err.Error(), "Rebuild the binary or inspect the CLI wiring.")
		return ExitError
	}

	ctx, err := parser.Parse(args)
	if err != nil {
		writeError(stderr, "VALIDATION", err.Error(), helpHint(commandHint(args)))
		return ExitValidation
	}

	app := &App{
		Stdout:  stdout,
		Stderr:  stderr,
		Format:  cli.Format,
		Version: version,
		URL:     strings.TrimSpace(cli.URL),
		Email:   strings.TrimSpace(cli.Email),
		Token:   strings.TrimSpace(cli.Token),
	}

	if commandNeedsClient(ctx.Command()) {
		creds, err := resolveCredentials(Credentials{URL: app.URL, Email: app.Email, Token: app.Token})
		if err != nil {
			writeError(stderr, "AUTH_STORE", err.Error(), "Use explicit flags/env vars or rerun `confluence auth login`.")
			return ExitError
		}
		if creds.URL == "" || creds.Email == "" || creds.Token == "" {
			writeError(stderr, "VALIDATION", "missing credentials: provide --url, --email, and --token, or store them with `confluence auth login`", helpHint("auth login"))
			return ExitValidation
		}
		app.URL = creds.URL
		app.Email = creds.Email
		app.Token = creds.Token
		app.Client = confluence.NewClient(confluence.Options{
			BaseURL: creds.URL,
			Email:   creds.Email,
			Token:   creds.Token,
			Timeout: cli.Timeout,
		})
	}

	if err := ctx.Run(app); err != nil {
		var validationErr *ValidationError
		var apiErr *confluence.APIError
		switch {
		case errors.As(err, &validationErr):
			writeError(stderr, "VALIDATION", validationErr.Message, validationErr.Hint)
			return ExitValidation
		case errors.As(err, &apiErr):
			if apiErr.StatusCode == 401 || apiErr.StatusCode == 403 {
				writeError(stderr, "AUTH_FAILED", apiErr.Error(), "Verify your Confluence credentials or rerun `confluence auth login`.")
				return ExitAuth
			}
			writeError(stderr, "API_ERROR", apiErr.Error(), "Retry the request or inspect the upstream Confluence response.")
			return ExitError
		default:
			writeError(stderr, "ERROR", err.Error(), "Inspect the command inputs or retry with a smaller request.")
			return ExitError
		}
	}

	return ExitOK
}

func commandNeedsClient(command string) bool {
	return command != "version" && command != "auth login"
}

func commandHint(args []string) string {
	parts := make([]string, 0, 2)
	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
			continue
		}
		parts = append(parts, arg)
		if len(parts) == 2 {
			break
		}
	}
	return strings.Join(parts, " ")
}
