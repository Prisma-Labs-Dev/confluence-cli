# confluence-cli

Agent-first CLI for compact, read-oriented access to Confluence Cloud.

`confluence` is designed for automation first:
- JSON envelopes on stdout by default
- structured JSON errors on stderr
- bounded output by default
- non-interactive auth only
- help text that explains the live contract

## Install

### Homebrew

```sh
brew install Prisma-Labs-Dev/tap/confluence-cli
```

### Release binary

Download the latest release asset from GitHub Releases and place `confluence` on your `PATH`.

### From source

```sh
go install github.com/Prisma-Labs-Dev/confluence-cli/cmd/confluence@latest
```

### Local repository checkout

This is the developer path, not the recommended end-user install path:

```sh
git clone https://github.com/Prisma-Labs-Dev/confluence-cli.git
cd confluence-cli
make build
./bin/confluence version
```

## Upgrade

Use the same channel you installed from.

### Homebrew

```sh
brew update
brew upgrade Prisma-Labs-Dev/tap/confluence-cli
confluence version
```

### Release binary

Download the latest release asset and replace the existing `confluence` binary on your `PATH`, then verify:

```sh
confluence version
```

### Go install

```sh
go install github.com/Prisma-Labs-Dev/confluence-cli/cmd/confluence@latest
confluence version
```

### Local repository checkout

This is a developer workflow rather than an end-user upgrade path:

```sh
git pull --ff-only
make build
./bin/confluence version
```

If you want end users to receive updates cleanly, prefer Homebrew or GitHub Releases over local source checkouts.

## Authentication

The CLI supports non-interactive setup only.

### 1. Flags or environment variables

```sh
confluence \
  --url https://your-domain.atlassian.net \
  --email you@example.com \
  --token "$CONFLUENCE_API_TOKEN" \
  auth login
```

### 2. Full credential JSON from stdin

```sh
printf '{"url":"https://your-domain.atlassian.net","email":"you@example.com","token":"TOKEN"}' \
  | confluence auth login --stdin-json
```

### 3. Token from stdin

```sh
printf '%s' "$CONFLUENCE_API_TOKEN" \
  | confluence --url https://your-domain.atlassian.net --email you@example.com auth login --token-stdin
```

Credential resolution for read commands is:
1. explicit flags / environment variables
2. stored credentials

Stored credentials go to macOS Keychain first, then to a local config file fallback if Keychain is unavailable.

## Command surface

Primary commands:
- `confluence spaces list`
- `confluence pages list`
- `confluence pages get`
- `confluence pages tree`
- `confluence pages search`
- `confluence auth login`
- `confluence version`

Run `confluence <command> --help` for the authoritative contract, including output shape, pagination behavior, defaults, and examples.

## Output contract

### List commands

List commands return a CLI-owned envelope:

```json
{
  "results": [
    {"id": "123", "key": "DEV", "name": "Development", "type": "global", "status": "current"}
  ],
  "page": {
    "limit": 10,
    "nextCursor": "eyJpZCI6..."
  },
  "schema": {
    "itemType": "space-summary",
    "fields": ["id", "key", "name", "type", "status"]
  }
}
```

### Single-object commands

Single-object commands return a wrapped item:

```json
{
  "item": {
    "id": "67890",
    "title": "Runbook",
    "spaceId": "12345",
    "status": "current"
  },
  "schema": {
    "itemType": "page-detail",
    "fields": ["id", "title", "spaceId", "status", "version"]
  }
}
```

### Errors

Errors are always JSON on stderr:

```json
{
  "error": {
    "code": "VALIDATION",
    "message": "missing credentials: provide --url, --email, and --token, or store them with `confluence auth login`",
    "hint": "Run `confluence auth login --help` for usage."
  }
}
```

### Exit codes

| Code | Meaning |
|---|---|
| `0` | success |
| `1` | runtime / API failure |
| `2` | validation failure |
| `3` | authentication / authorization failure |

## Examples

### List spaces

```sh
confluence spaces list
confluence spaces list --limit 25
confluence --format plain spaces list
```

### List pages in a space

```sh
confluence pages list --space-id 12345
confluence pages list --space-id 12345 --sort -modified-date
```

### Get page metadata or body

```sh
# metadata only
confluence pages get --page-id 67890

# explicit body request
confluence pages get --page-id 67890 --body-format view

# plain output converts view HTML to Markdown
confluence --format plain pages get --page-id 67890 --body-format view
```

### Traverse a bounded tree

```sh
confluence pages tree --page-id 67890
confluence pages tree --page-id 67890 --depth 2 --limit-per-level 5
confluence --format plain pages tree --page-id 67890
```

`pages tree` is intentionally bounded. It fetches only the first page of children per node and marks `hasMoreChildren` when more data exists.

### Search pages

```sh
confluence pages search --query "deployment"
confluence pages search --query "meeting notes" --title-only
confluence pages search --query "runbook" --space-id 12345
```

Search uses the current supported Confluence Cloud REST v1 search endpoint internally because REST v2 does not yet provide equivalent CQL search.

## Development and validation

```sh
make build
make test
make lint
```

Optional live golden validation against a real Confluence workspace:

```sh
zsh -lc 'CONFLUENCE_LIVE_E2E=1 go test -run LiveAPI ./...'
```

Refresh the redacted live golden snapshot intentionally:

```sh
zsh -lc 'CONFLUENCE_LIVE_E2E=1 CONFLUENCE_LIVE_E2E_UPDATE=1 go test -run LiveAPI ./...'
```

The checked-in live golden stores only redacted contract summaries, not raw page content or titles. For more stable selection you can optionally set `CONFLUENCE_LIVE_SPACE_ID`, `CONFLUENCE_LIVE_PAGE_ID`, and `CONFLUENCE_LIVE_SEARCH_QUERY`.

See `DEVELOPMENT.md` for local workflow, CI, release flow, and linting details.
