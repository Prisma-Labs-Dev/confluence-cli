You are working in `confluence-cli`, a Go CLI for Confluence Cloud v2 read APIs.

Primary user is other agents, not humans.

Operating rules:
- Prefer deterministic, non-interactive behavior.
- Never require prompts when a stdin/flag/env path can be used.
- Keep JSON output and structured stderr error contracts stable.
- Avoid adding broad auth layers; keep auth simple and explicit.
- Treat macOS as the only target platform unless asked otherwise.

Auth model to preserve:
- `confluence auth login` stores `url/email/token`.
- Storage priority: macOS Keychain first, local file fallback second.
- Non-interactive setup paths must remain first-class:
  - `--stdin-json` for `{url,email,token}`
  - `--token-stdin` for token-only stdin
  - `--no-prompt` to fail fast
- Runtime credential resolution order:
  1. explicit flags/env
  2. stored credentials

Engineering standards:
- Make minimal, composable changes.
- Add or update tests with each behavior change.
- If a feature is removed, remove or rewrite only the tests that specifically covered that removed behavior.
- Never delete tests just to make the test suite pass.
- Run `gofmt -w` and `go test ./...` before finishing.
- Do not weaken existing exit code behavior without explicit approval.

Mandatory cleanup before finishing:
- Remove duplicated code paths, stale docs, dead comments, and abandoned TODOs created during the task.
- Ensure one clear source of truth per behavior; avoid documenting the same flow in multiple conflicting places.
- Verify changed docs match implemented behavior and flag names exactly.
- Leave the workspace in a coherent state for the next agent: no partial migrations or ambiguous temporary patterns.
