# Development Guide

This document is the local workflow and CI/release reference for `confluence-cli`.

## Prerequisites

- Go `1.23`
- `make`
- `golangci-lint` for local lint runs
- Optional for releases: `gh` plus access to this repo and `Prisma-Labs-Dev/homebrew-tap`

## Local workflow

```sh
# safely format tracked Go files only
make fmt

# build local binary
make build

# run fixture/unit/integration tests
make test

# run lint + file-size guardrails
make lint

# install the repo build into ~/.local/bin and verify it
make install-local
make verify-local
```

Fast dogfood loop:

```sh
make dev-refresh
```

## Local install and update model

This repository supports two different update stories:

- **end-user update path:** Homebrew, GitHub Releases, or `go install ...@latest`
- **developer local checkout path:** `git pull --ff-only` + `make install-local`

If you are editing the repo itself, treat the checkout flow as a developer workflow, not as the primary end-user install/upgrade contract.

Optional live golden validation against a real workspace:

```sh
make test-live
```

Refresh the redacted live golden snapshot on purpose:

```sh
make test-live-update
```

The live golden is intentionally sanitized before it is written to `testdata/golden/live/contract.json`, so repo snapshots never contain raw workspace content. If you want a more stable target page or query, set `CONFLUENCE_LIVE_SPACE_ID`, `CONFLUENCE_LIVE_SPACE_KEY`, `CONFLUENCE_LIVE_PAGE_ID`, and `CONFLUENCE_LIVE_SEARCH_QUERY`.

## CI

Workflow: `.github/workflows/ci.yml`

CI runs on pushes and pull requests to `main` and executes:

```sh
go build ./...
go test ./... -count=1
go vet ./...
golangci-lint run ./...
./scripts/check-file-length.sh
```

## Linting policy

Linting is enforced in CI and locally.

Current guardrails:
- `gofmt` via `make fmt` and `make fmt-check`
- `govet`
- `staticcheck`
- `gosimple`
- `unused`
- `ineffassign`
- `errcheck`
- `misspell`
- `unconvert`
- Go file size limits via `scripts/check-file-length.sh`

To avoid common shell paper cuts, prefer `make fmt` or `./scripts/fmt-go.sh` over raw `gofmt` when your file list may include Markdown or other non-Go files.

Default file-size limits:
- non-test Go files: `300` lines max
- test Go files: `700` lines max

If a file naturally wants to exceed those limits, split it instead of raising the cap unless there is a strong reason.

## Releases

Workflow: `.github/workflows/release.yml`

Release flow:
1. Checkout full git history and tags.
2. Run build, test, vet, lint, and file-size checks.
3. Compute the next semantic version from tags.
4. Create and push the tag.
5. Run GoReleaser.
6. Update the Homebrew tap.

Automatic bump markers in commit messages:
- `#minor` or `release:minor`
- `#major` or `release:major`

Manual trigger examples:

```sh
gh workflow run release.yml -f bump=patch
gh workflow run release.yml -f bump=minor
gh workflow run release.yml -f bump=major
```

## Homebrew verification

```sh
brew info Prisma-Labs-Dev/tap/confluence-cli
brew update
brew upgrade Prisma-Labs-Dev/tap/confluence-cli
confluence version
```

## End-user upgrade paths

Document and preserve a clean end-user upgrade story:

### Homebrew

```sh
brew update
brew upgrade Prisma-Labs-Dev/tap/confluence-cli
confluence version
```

### GitHub Releases

Replace the installed binary with the latest release asset and verify:

```sh
confluence version
```

### `go install`

```sh
go install github.com/Prisma-Labs-Dev/confluence-cli/cmd/confluence@latest
confluence version
```

### Local developer checkout

```sh
git pull --ff-only
make build
./bin/confluence version
```

## Agent contributor notes

- Prefer the CLI's documented JSON envelopes over raw upstream payload assumptions.
- Keep auth non-interactive.
- Keep help text aligned with the actual live contract.
- Update `AGENTS.md` for shared entrypoint changes and `AGENT_PROMPT.md` for rebuild-brief changes.
