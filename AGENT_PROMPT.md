# Agent Prompt: Make `confluence` an Ideal Agent-First CLI

You are working in the `Prisma-Labs-Dev/confluence-cli` repository.

Your job is to make this CLI an excellent tool for coding agents that need reliable, compact, read-oriented access to Confluence Cloud.

Do **not** optimize for human-first terminal habits. Optimize for agents.

## Mission

Build and refine `confluence` as a clean, explicit, non-interactive, low-noise Confluence API CLI that agents can understand from `--help`, automate safely, and trust in end-to-end use.

## Core mandate

Treat this as a **clean-slate** agentic tool.

You do **not** need to preserve backwards compatibility.

You may:

- remove commands or flags that do not fit the agent-first model
- rename or restructure commands if the result is clearer
- delete interactive or ambiguous flows
- replace weak output contracts with cleaner ones
- simplify the repo if old structure only preserves historical baggage

If preserving an older interface conflicts with building the right agent tool, prefer the better tool.

## Required principles

- no interactive prompts in agent paths
- excellent root and subcommand `--help`
- structured command output must be machine-readable and predictable
- primary results on stdout, diagnostics/errors on stderr
- stable exit codes
- bounded output by default
- no raw verbose payload dumps unless explicitly requested
- latest supported Confluence Cloud API only; no deprecated API design
- end-to-end validation of the real CLI contract

## Product framing

This CLI should feel like:

- a **Confluence API access layer for agents**
- a **research/discovery CLI**
- a **bounded-output summarizer** over verbose Confluence responses

It should not feel like:

- a human-only browsing tool
- a wrapper around interactive auth/setup
- a raw API dump utility

## Current strengths to preserve or strengthen

From the current tool and docs, these are already good directions:

- JSON-oriented contract
- compact list results with pagination
- page metadata retrieval without body by default
- page tree navigation with bounded depth
- readable `--plain` output for human inspection
- `view` body conversion to Markdown rather than raw HTML in plain mode
- JSON errors on stderr

Do not regress these qualities while improving the CLI.

## Current gaps / improvement targets

Based on the current workspace assessment, focus on these areas:

1. **Remove or de-emphasize interactive auth**
   - Interactive `auth login --prompt` is not aligned with strict agent-first design.
   - Prefer explicit non-interactive auth only.

2. **Make the help surface stronger**
   - Root help and subcommand help should be sufficient for agent use.
   - Move important contract details from README into `--help`.
   - Include examples, defaults, pagination behavior, and output-shape hints in help.

3. **Improve output shape discoverability**
   - Agents should not have to infer JSON shapes by trial and error.
   - Provide stable envelopes and/or schema notes in help.

4. **Audit API freshness**
   - Ensure the implementation uses current supported Confluence Cloud API surfaces only.
   - Remove stale or deprecated API assumptions if any exist.

5. **Strengthen end-to-end and golden validation**
   - The CLI contract must be proven against the real tool behavior, not assumed from unit tests alone.

## What useful looks like for agents

An agent should be able to:

- list spaces with predictable pagination
- list pages in a space with bounded output
- search pages with explicit, safe search inputs
- fetch a page with compact metadata by default
- request body content in explicit formats only when needed
- traverse page trees without blowing up context windows
- understand the shape of each command's output from help text alone

## Command surface priorities

Keep the tool focused and read-oriented.

Priority commands:

- `confluence spaces list`
- `confluence pages list`
- `confluence pages search`
- `confluence pages get`
- `confluence pages tree`
- `confluence version`

Authentication should support non-interactive setup only, or be redesigned so agent usage does not depend on prompt-based flows.

## Output contract requirements

### Default output

Choose and document one clean contract:

- either JSON-by-default with explicit human/plain fallback
- or explicit `--json` everywhere with a clearly documented default

Either is acceptable if the contract is consistent, explicit, and easy for agents to discover.

Do not keep confusing mixed behavior just for compatibility.

### JSON shape

JSON should be:

- stable
- compact
- pre-parsed
- owned by the CLI contract rather than mirroring upstream responses blindly

Prefer response envelopes that make pagination and shape obvious, for example:

```json
{
  "results": [...],
  "page": {
    "limit": 25,
    "nextCursor": "..."
  },
  "schema": {
    "itemType": "page-summary",
    "fields": ["id", "title", "spaceId", "status"]
  }
}
```

For single-object reads:

```json
{
  "item": {...},
  "schema": {
    "itemType": "page-detail",
    "fields": [...]
  }
}
```

If you choose a different envelope, keep it equally explicit.

### Text / plain output

If plain output remains:

- keep it compact
- keep it readable
- keep it bounded
- avoid decorative noise

Plain output should help a human glance at data, not become the main contract agents depend on.

## Content-size discipline

Confluence content can get huge quickly.

Defaults must protect agent context windows:

- page metadata only unless body is explicitly requested
- bounded list/search results by default
- explicit pagination/cursor support
- explicit body format selection
- no raw HTML or storage XML by default

If body content is requested, prefer the most agent-usable representation for the chosen mode.

## Help requirements

Every command help should state:

- what it does
- required flags
- optional flags
- output behavior
- pagination behavior
- examples
- default output shape or default field set

The help text should be the first place an agent can learn the contract.

## API discipline

Use the latest supported Confluence Cloud APIs.

- do not add new code against deprecated API paths
- if the repo already uses something stale, replace it rather than papering over it
- prefer one clear modern API path over compatibility branching

If a migration is needed, do the migration and delete the old path in the same task.

## Error handling

- errors must be explicit and actionable
- keep structured error output
- preserve or improve stable exit code semantics
- fail fast on invalid input
- do not silently fall back to hidden auth or hidden content expansion

## Auth expectations

Agent usage must remain non-interactive.

Good patterns:

- env vars
- explicit flags
- stdin-based credential input
- secure local storage used through explicit, non-interactive flows

Bad patterns:

- prompt-driven setup as a normal path
- hidden state assumptions that agents cannot discover

## Tests and validation

Do not stop at unit tests.

Add or strengthen:

- unit tests for parsing, config, rendering, and edge cases
- golden tests for CLI text and JSON contracts
- end-to-end tests for representative live CLI behavior

The golden/e2e layer should verify:

- root help and subcommand help
- success output shape
- error output shape
- pagination contracts
- bounded list/search defaults
- body-format behavior
- body rendering behavior for plain mode

The purpose is to verify the **contract agents consume**, not just the internal implementation.

## Suggested implementation approach

1. Audit the existing command and output contract against the agent-first principles above.
2. Remove or redesign any interactive or ambiguous flows.
3. Tighten root and subcommand help so it fully explains the live contract.
4. Normalize JSON envelopes and response-shape discoverability.
5. Audit API usage to ensure only current supported Confluence Cloud APIs remain.
6. Add or strengthen golden and end-to-end validation.
7. Simplify or delete stale codepaths instead of preserving them for compatibility.

## Deliverable expectation

The final CLI should be:

- agent-first
- non-interactive
- explicit
- compact
- current
- cleanly tested end-to-end

If you have to choose between preserving historical behavior and shipping a cleaner agent contract, choose the cleaner contract.
