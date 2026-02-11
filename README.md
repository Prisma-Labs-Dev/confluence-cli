# confluence-cli

Read-only CLI for [Confluence Cloud API](https://developer.atlassian.com/cloud/confluence/rest/v2/intro/), designed for AI agent consumption.

- JSON output by default, parseable by any tool or agent
- `--plain` flag for human-readable tables
- `pages get --body-format view --plain` renders cleaned Markdown (not raw HTML) to reduce context size
- Structured error messages to stderr with exit codes
- Single binary with optional local credential fallback file

## Installation

### Homebrew (macOS / Linux)

```sh
brew install Prisma-Labs-Dev/tap/confluence-cli
```

### Binary download

Download from [Releases](https://github.com/Prisma-Labs-Dev/confluence-cli/releases) and place in your `PATH`.

### From source

```sh
go install github.com/Prisma-Labs-Dev/confluence-cli/cmd/confluence@latest
```

## Setup

Run `auth login` once to store credentials using an explicit non-interactive path:

```sh
printf '{"url":"https://your-domain.atlassian.net","email":"you@example.com","token":"your-token"}' | confluence auth login --stdin-json
```

Credentials are stored in macOS Keychain. If Keychain is unavailable, the CLI falls back to a local file at:

```text
~/Library/Application Support/confluence-cli/credentials.json
```

You can still pass credentials as flags or environment variables when needed.
See `confluence auth login` examples below for non-interactive agent flows.

**Get your API token:** https://id.atlassian.com/manage-profile/security/api-tokens

## Commands

### `confluence spaces list`

List all Confluence spaces.

```sh
# JSON (default)
confluence spaces list

# Limit results
confluence spaces list --limit 10

# Paginate with cursor from previous response
confluence spaces list --cursor "eyJpZCI6..."

# Human-readable table
confluence spaces list --plain
```

### `confluence pages list`

List pages in a space.

```sh
# All pages in a space
confluence pages list --space-id 12345

# Sort by recently modified (descending)
confluence pages list --space-id 12345 --sort -modified-date

# Paginate
confluence pages list --space-id 12345 --limit 20 --cursor "eyJpZCI6..."
```

### `confluence pages get`

Get a single page by ID, optionally with body content.

```sh
# Metadata only
confluence pages get --page-id 67890

# With rendered HTML body
confluence pages get --page-id 67890 --body-format view

# In --plain mode, the view body is converted to Markdown for readability/context efficiency
confluence --plain pages get --page-id 67890 --body-format view

# With raw storage format (Confluence XML)
confluence pages get --page-id 67890 --body-format storage

# With Atlas Doc Format (structured JSON)
confluence pages get --page-id 67890 --body-format atlas_doc_format
```

### `confluence pages tree`

Show child pages as a tree.

```sh
# Direct children only (depth 1)
confluence pages tree --page-id 67890

# Recurse 3 levels deep
confluence pages tree --page-id 67890 --depth 3

# Plain text tree view
confluence pages tree --page-id 67890 --depth 2 --plain
```

Plain output renders a tree:

```
Page 67890
├── Getting Started (id:11111)
│   ├── Installation (id:22222)
│   └── Configuration (id:33333)
└── API Reference (id:44444)
```

### `confluence pages search`

Search pages using [CQL (Confluence Query Language)](https://developer.atlassian.com/cloud/confluence/advanced-searching-using-cql/).

```sh
# Search by space
confluence pages search --cql "type=page AND space=MYSPACE"

# Search by title
confluence pages search --cql "title ~ 'meeting notes'"

# Full-text search
confluence pages search --cql "text ~ 'deployment process'"

# Combine with space filter flag
confluence pages search --cql "type=page" --space-id 12345

# Limit results
confluence pages search --cql "type=page AND space=DEV" --limit 5
```

### `confluence version`

```sh
confluence version
# {"version":"1.0.0"}

confluence version --plain
# confluence 1.0.0
```

### `confluence auth login`

Store credentials for future commands:

```sh
# non-interactive
confluence --url https://your-domain.atlassian.net --email you@example.com --token your-token auth login

# non-interactive with JSON credentials from stdin (agent-friendly)
printf '{"url":"https://your-domain.atlassian.net","email":"you@example.com","token":"your-token"}' | confluence auth login --stdin-json

# non-interactive with token from stdin
printf '%s' "$CONFLUENCE_API_TOKEN" | confluence --url https://your-domain.atlassian.net --email you@example.com auth login --token-stdin

# optional human-only prompt mode
confluence auth login --prompt

# disable prompts entirely (fail fast if anything is missing)
confluence auth login --no-prompt
```

## Global Flags

| Flag | Env var | Description |
|------|---------|-------------|
| `--url` | `CONFLUENCE_URL` | Confluence base URL (host, `/wiki`, or `/wiki/api/v2`) |
| `--email` | `CONFLUENCE_EMAIL` | Atlassian account email |
| `--token` | `CONFLUENCE_API_TOKEN` | Atlassian API token |
| `--plain` | | Human-readable table output |
| `--timeout` | | HTTP timeout (default: 30s) |

## Output

### JSON (default)

All commands output JSON to stdout. Paginated responses include a `nextCursor` field:

```json
{
  "results": [
    {"id": "123", "key": "DEV", "name": "Development", "type": "global", "status": "current"}
  ],
  "nextCursor": "eyJpZCI6..."
}
```

### Errors

Errors are JSON on stderr with a non-zero exit code:

```json
{"error": "missing required flags: --email (or CONFLUENCE_EMAIL)", "code": "VALIDATION"}
```

## Exit Codes

| Code | Meaning | Example |
|------|---------|---------|
| 0 | Success | Command completed |
| 1 | Runtime error | API returned 404/500, network failure |
| 2 | Validation error | Missing required flag, unknown command |
| 3 | Auth error | Invalid credentials (401/403) |

## Releasing

Every push to `main` creates a release. By default this is a patch bump.

To bump minor/major directly on push (without an intermediate patch), include one of these markers in the commit message:

- `#minor` or `release:minor`
- `#major` or `release:major`

You can also trigger releases manually:

```sh
gh workflow run release.yml -f bump=patch
gh workflow run release.yml -f bump=minor
gh workflow run release.yml -f bump=major
```

The workflow auto-tags from the highest existing semantic version, builds binaries for macOS and Linux (amd64/arm64), creates a GitHub release, and updates the Homebrew tap.

**First-time setup:** Add a `HOMEBREW_TAP_GITHUB_TOKEN` secret to the repo with a GitHub PAT that has push access to `Prisma-Labs-Dev/homebrew-tap`.

## CI

CI runs on push and pull requests to `main` and executes:

```sh
go build ./...
go test ./... -count=1
go vet ./...
```

## Development

See [`DEVELOPMENT.md`](DEVELOPMENT.md) for local workflow, CI, release details, and Homebrew verification steps.

## Work Tracking

Current agent work packets and history live under:

- `docs/work/active/`
- `docs/work/archive/`
- `docs/work/INDEX.md`

## License

MIT
