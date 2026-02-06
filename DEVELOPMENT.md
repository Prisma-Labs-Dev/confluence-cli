# Development Guide

This document is the single source of truth for local development, CI behavior, and release flow.

## Prerequisites

- Go `1.23`
- `make`
- Optional for releases: `gh` (GitHub CLI), access to this repo and `Prisma-Labs-Dev/homebrew-tap`

## Local Workflow

```sh
# build local binary
make build

# run test suite
make test

# optional lint
make lint
```

Equivalent CI checks:

```sh
go build ./...
go test ./... -count=1
go vet ./...
```

## CI

Workflow: `.github/workflows/ci.yml`

- Triggers on:
  - `push` to `main`
  - `pull_request` targeting `main`
- Runs:
  - `go build ./...`
  - `go test ./... -count=1`
  - `go vet ./...`

## Releases

Workflow: `.github/workflows/release.yml`

- Auto patch release:
  - Triggered on every `push` to `main`
  - Uses `patch` bump by default
- Manual minor/major:
  - Trigger via:
    - `gh workflow run release.yml -f bump=minor`
    - `gh workflow run release.yml -f bump=major`

Release workflow order:
1. Checkout full git history/tags.
2. Run build/test/vet gates.
3. Compute next version from the highest semantic version tag (not nearest git tag).
4. Create and push git tag.
5. Run GoReleaser.
6. Update Homebrew tap formula (`Prisma-Labs-Dev/homebrew-tap`).

## Homebrew Verification

```sh
# inspect published formula version
brew info Prisma-Labs-Dev/tap/confluence-cli

# update metadata and install latest release
brew update
brew upgrade Prisma-Labs-Dev/tap/confluence-cli

# validate installed version
confluence version --plain
```

## Notes For Agent Contributors

- Keep CLI behavior stable:
  - JSON output must remain machine-parseable.
  - Human-readable output belongs behind `--plain`.
- If changing plain rendering, add tests under `internal/cli` and unit tests for helper packages.
- Do not edit `CLAUDE.md`; update `agents.mt` for project instructions.
