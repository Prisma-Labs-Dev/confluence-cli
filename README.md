# confluence-cli

CLI for Confluence Cloud API, optimized for AI agent consumption.

## Installation

```sh
brew install Prisma-Labs-Dev/tap/confluence-cli
```

Or download binaries from [Releases](https://github.com/Prisma-Labs-Dev/confluence-cli/releases).

## Configuration

Set environment variables:

```sh
export CONFLUENCE_URL=https://your-domain.atlassian.net
export CONFLUENCE_EMAIL=you@example.com
export CONFLUENCE_API_TOKEN=your-api-token
```

Generate an API token at https://id.atlassian.com/manage-profile/security/api-tokens

## Usage

Output is JSON by default. Use `--plain` for human-readable tables.

### List spaces

```sh
confluence spaces list
confluence spaces list --limit 10
```

### List pages in a space

```sh
confluence pages list --space-id 12345
confluence pages list --space-id 12345 --sort -modified-date --limit 20
```

### Get a page

```sh
confluence pages get --page-id 67890
confluence pages get --page-id 67890 --body-format storage
```

### Page tree (children)

```sh
confluence pages tree --page-id 67890
confluence pages tree --page-id 67890 --depth 3
```

### Search

```sh
confluence pages search --cql "type=page AND space=MYSPACE"
confluence pages search --cql "title ~ 'meeting notes'" --limit 5
```

### Version

```sh
confluence version
```

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Runtime error |
| 2 | Validation/usage error |
| 3 | Authentication error |

## License

MIT
