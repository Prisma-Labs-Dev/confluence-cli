# AGENTS.md

Repository rules for agents working in `confluence-cli`.

## Mission

This CLI is built by agents for agents.

Human-friendly output is secondary.

## Required Behavior

1. Non-interactive agent paths must never hang.
2. JSON is the default machine contract.
3. Errors must be JSON on stderr with stable code mapping.
4. Exit codes must remain consistent with `contract/error-codes.md`.

## Work Tracking

1. Read `docs/work/INDEX.md` before making changes.
2. Continue active packet if one exists.
3. Update finding status/result fields as work progresses.
4. Archive completed packets under `docs/work/archive/<year>/`.

## Existing Local Context

This repo also includes `agents.mt`.

Treat `AGENTS.md` as the primary Codex entrypoint and keep it aligned with `agents.mt` intent.
