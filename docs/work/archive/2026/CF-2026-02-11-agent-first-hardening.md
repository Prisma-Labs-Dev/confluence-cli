# Work Packet: CF-2026-02-11-agent-first-hardening

Status: done  
Owner: agent  
Started: 2026-02-11  
Repo SHA: e428c25

## Scope

Close agent-first reliability gaps found in review: non-interactive safety, validation taxonomy, URL normalization, and contract drift.

## Findings

| ID | Severity | Problem | Status | Result | Verification | Link | Notes |
|---|---|---|---|---|---|---|---|
| F-001 | high | `auth login --stdin-json` and `--token-stdin` can block indefinitely in PTY runs with no piped stdin | done | pass | `go test ./...` (includes `TestAuthLoginStdinJSONFailsFastOnTTY`, `TestAuthLoginTokenStdinFailsFastOnTTY`) | beb6c98 | now fails fast with explicit error |
| F-002 | high | auth input validation currently returns generic `ERROR`/exit 1 instead of `VALIDATION`/exit 2 | done | pass | `go test ./...` (includes `TestAuthLoginNoPromptFailsFastWithoutInput` + integration no-prompt test) | af50aa6 | command validation now maps to `VALIDATION`/exit 2 |
| F-003 | medium | base URL normalization can produce `/wiki/wiki/api/v2` when input already contains `/wiki` | done | pass | `go test ./...` (includes `TestClientBaseURLWithWikiPath`, `TestClientBaseURLAlreadyV2Path`, `TestClientBaseURLFromRestAPIPath`) | af50aa6 | base URL normalization added in client constructor |
| F-004 | medium | interactive auth is still default behavior on terminal sessions | done | pass | `go test ./...` (includes `TestAuthLoginDefaultFailsFastWithoutPromptOnTTY`) + manual `go run ./cmd/confluence auth login` in TTY | af50aa6 | default is now non-interactive; `--prompt` required for prompts |
| F-005 | low | `--color` flag is exposed but unused by renderers | done | pass | help output check + README/global flag cleanup | af50aa6 | removed unused flag and dead state |

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
| Validation/auth/runtime exit taxonomy is explicit and tested | pass |
| URL normalization handles `/wiki` and base host forms | pass |
| Flags/docs align with implementation | pass |

## Decisions

1. Track these as a single hardening packet first, then split follow-up packets only if scope expands.

## Handoff

1. Next command: `go test ./...`
2. Next file(s): `internal/cli/auth.go`, `internal/cli/run.go`, `client.go`, `README.md`
3. Remaining risk: hangs and ambiguous error taxonomy can break agent orchestration.
