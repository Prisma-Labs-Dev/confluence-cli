# Work Packet: CF-2026-02-11-agent-first-hardening

Status: in_progress  
Owner: agent  
Started: 2026-02-11  
Repo SHA: e428c25

## Scope

Close agent-first reliability gaps found in review: non-interactive safety, validation taxonomy, URL normalization, and contract drift.

## Findings

| ID | Severity | Problem | Status | Result | Verification | Link | Notes |
|---|---|---|---|---|---|---|---|
| F-001 | high | `auth login --stdin-json` and `--token-stdin` can block indefinitely in PTY runs with no piped stdin | done | pass | `go test ./...` (includes `TestAuthLoginStdinJSONFailsFastOnTTY`, `TestAuthLoginTokenStdinFailsFastOnTTY`) | pending | now fails fast with explicit error |
| F-002 | high | auth input validation currently returns generic `ERROR`/exit 1 instead of `VALIDATION`/exit 2 | todo | fail | CLI test for missing fields on `auth login --no-prompt` | - | currently coded as generic error |
| F-003 | medium | base URL normalization can produce `/wiki/wiki/api/v2` when input already contains `/wiki` | todo | fail | unit test for base URL variants | - | normalize URL path before API suffix |
| F-004 | medium | interactive auth is still default behavior on terminal sessions | todo | fail | command behavior test on TTY defaults | - | should default fail-fast for agent mode |
| F-005 | low | `--color` flag is exposed but unused by renderers | todo | fail | grep + output test | - | docs/flag behavior drift |

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
| Non-interactive auth flows cannot hang | pass |
| Validation/auth/runtime exit taxonomy is explicit and tested | fail |
| URL normalization handles `/wiki` and base host forms | fail |
| Flags/docs align with implementation | fail |

## Decisions

1. Track these as a single hardening packet first, then split follow-up packets only if scope expands.

## Handoff

1. Next command: `go test ./...`
2. Next file(s): `internal/cli/auth.go`, `internal/cli/run.go`, `client.go`, `README.md`
3. Remaining risk: hangs and ambiguous error taxonomy can break agent orchestration.
