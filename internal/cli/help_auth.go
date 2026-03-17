package cli

func authHelp() string {
	return `Usage: confluence auth <command>

Credential management commands.

Commands:
  login [flags]
    Store credentials for later non-interactive use.

Run "confluence auth login --help" for setup flows.
`
}

func authLoginHelp() string {
	return `Usage: confluence auth login [flags]

Store credentials for later non-interactive use.

Supported input modes:
  1. Flags or environment variables:
       confluence --url https://example.atlassian.net --email you@example.com --token TOKEN auth login

  2. JSON credentials from stdin:
       printf '{"url":"https://example.atlassian.net","email":"you@example.com","token":"TOKEN"}' | confluence auth login --stdin-json

  3. Token from stdin with URL/email from flags or env:
       printf '%s' "$CONFLUENCE_API_TOKEN" | confluence --url https://example.atlassian.net --email you@example.com auth login --token-stdin

Notes:
  - No interactive prompts are supported.
  - Stored credentials are used only when running read commands, not while logging in.

Output (json):
  {
    "item": {"storedIn":"keychain"},
    "schema": {"itemType":"auth-login","fields":["storedIn"]}
  }

Flags:
  -h, --help            Show command help.
      --url=STRING      Confluence base URL ($CONFLUENCE_URL)
      --email=STRING    Atlassian account email ($CONFLUENCE_EMAIL)
      --token=STRING    Atlassian API token ($CONFLUENCE_API_TOKEN)
      --format=json     Output format: json or plain
      --timeout=30s     HTTP timeout
      --stdin-json      Read {url,email,token} JSON from piped stdin
      --token-stdin     Read token from piped stdin; requires --url and --email
`
}
