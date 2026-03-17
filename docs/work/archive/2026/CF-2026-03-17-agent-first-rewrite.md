# Work Packet: CF-2026-03-17-agent-first-rewrite

Status: done  
Owner: agent  
Started: 2026-03-17  
Repo SHA: cdd1975

## Scope

Rewrite `confluence` into an explicitly agent-first CLI with stronger machine contracts, richer help, bounded defaults, non-interactive auth only, current documentation, and enforced linting/size guardrails.

## Findings

| ID | Severity | Problem | Status | Result | Verification | Link | Notes |
|---|---|---|---|---|---|---|---|
| F-001 | high | Auth still includes interactive prompt-oriented paths that do not fit the clean-slate agent-first contract | done | pass | `go test ./...` + live auth-backed CLI runs | working tree | removed prompt behavior and kept only explicit non-interactive flows |
| F-002 | high | Root and subcommand help do not explain output shapes, pagination, or examples well enough for agents to rely on `--help` alone | done | pass | `go test ./...` (help goldens) | working tree | custom help now documents output shapes, pagination, defaults, and examples |
| F-003 | high | JSON success envelopes are too implicit and page tree traversal is not bounded enough for agent context windows | done | pass | `go test ./...` + live API smoke test | working tree | list/item envelopes now include schema metadata and tree traversal is bounded per node |
| F-004 | medium | Repo docs and contract docs describe the old behavior and are no longer suitable as a handoff surface for other agents | done | pass | README + contract docs rewrite review | working tree | docs now match `--format`, nested errors, bounded defaults, and live validation workflow |
| F-005 | medium | Linting and size constraints are not enforced strongly enough to keep the codebase maintainable for future agent edits | done | pass | `golangci-lint run ./...` + `./scripts/check-file-length.sh` | working tree | added lint config, CI enforcement, release enforcement, and Go file-size checks |

Status values:

- `todo`
- `in_progress`
- `blocked`
- `done`
- `wontfix`

Result values:

- `pass`
- `fail`
- `n/a`

## Exit Criteria

| Check | Status |
|---|---|
| Non-interactive agent execution only for auth and command flows | pass |
| Help text explains output shape, pagination, defaults, and examples | pass |
| JSON/stdout/stderr/exit-code contracts are rewritten and validated end to end | pass |
| Documentation matches the rewritten CLI exactly | pass |
| Linting and file-size checks run in CI and locally | pass |

## Decisions

1. Compatibility-breaking changes are allowed; optimize for the cleaner agent-first contract instead of legacy behavior.
2. Search support stays, but the CLI surface should be safer than exposing raw CQL directly.
3. The rewritten CLI should prefer bounded defaults even if callers can still opt into larger payloads explicitly.
4. `--format json|plain` is clearer for agents than a boolean `--plain` switch and became the single output-mode flag.
5. Live API validation is optional in the test suite and is enabled with `CONFLUENCE_LIVE_E2E=1`.

## Handoff

1. Next command: `go test ./...`
2. Next file(s): `README.md`, `contract/README.md`, `internal/cli/help_*.go`, `.golangci.yml`
3. Remaining risk: live API behavior can still change if Atlassian introduces a v2 search endpoint; `pages search` should migrate when that happens.
