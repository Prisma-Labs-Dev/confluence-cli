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

## Compatibility stance

Historical guidance in older agent docs is obsolete.

If legacy behavior conflicts with the cleaner agent-first design in `AGENT_PROMPT.md`, prefer the new design.
