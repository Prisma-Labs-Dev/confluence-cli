package cli

func rootHelp() string {
	return `Usage: confluence <command> [flags]

Agent-first Confluence Cloud CLI for compact, read-oriented access.

Default behavior:
  - JSON envelopes on stdout
  - structured JSON errors on stderr
  - explicit bounded output defaults
  - credentials from flags/env first, then stored credentials

Global flags:
  -h, --help              Show command help.
      --url=STRING        Confluence base URL ($CONFLUENCE_URL)
      --email=STRING      Atlassian account email ($CONFLUENCE_EMAIL)
      --token=STRING      Atlassian API token ($CONFLUENCE_API_TOKEN)
      --format=json       Output format: json or plain
      --timeout=30s       HTTP timeout

Commands:
  spaces list [flags]
    List spaces with compact summaries and cursor pagination.

  pages list --space-id=STRING [flags]
    List pages in a space with bounded summaries.

  pages get --page-id=STRING [flags]
    Get page metadata by default; request body content explicitly.

  pages tree --page-id=STRING [flags]
    Traverse a bounded page tree with per-level limits.

  pages search --query=STRING [flags]
    Search pages with safe query inputs.

  auth login [flags]
    Store credentials for later non-interactive use.

  version [flags]
    Print the CLI version.

Run "confluence <command> --help" for output shapes, pagination rules, and examples.
`
}
