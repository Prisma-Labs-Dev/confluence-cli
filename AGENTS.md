# AGENTS.md

Agent entrypoint for work in `confluence-cli`.

## Source of truth

Use `AGENT_PROMPT.md` as the active implementation brief for the rebuild.

## Current direction

This repository is being rebuilt as an agent-first Confluence Cloud CLI.

- optimize for agents, not human-first terminal habits
- prefer explicit non-interactive flows
- make `--help` sufficient for contract discovery
- keep stdout/stderr behavior predictable and machine-friendly
- keep output bounded by default
- strengthen golden and end-to-end contract validation

## Repo-local workflow paper cuts to avoid

Use the repo tooling instead of ad hoc shell commands when possible:

- `make fmt` — safely format tracked Go files only
- `make fmt-check` — fail if tracked Go files are not gofmt-clean
- `make test` — run offline/unit/integration coverage
- `make test-live` — run the live redacted contract golden against a real workspace
- `make test-live-update` — intentionally refresh the live golden
- `make install-local` — replace the active `~/.local/bin/confluence` dogfood binary
- `make verify-local` — verify the installed local binary contract
- `make dev-refresh` — preferred one-command local dogfood loop after code changes

Do not run raw `gofmt` over mixed file lists that may include Markdown or docs. Use `make fmt` or `./scripts/fmt-go.sh` instead.

## Compatibility stance

Historical guidance in older agent docs is obsolete.

If legacy behavior conflicts with the cleaner agent-first design in `AGENT_PROMPT.md`, prefer the new design.
