# CLI Agent Contract

This folder defines the machine-facing contract for this CLI.

## Required Error Envelope

Errors to stderr must be JSON:

```json
{
  "error": "human-readable message",
  "code": "MACHINE_CODE"
}
```

## Required Success Behavior

1. JSON output by default on stdout.
2. Stable field names for equivalent operations.
3. Non-zero exit code on failure.

## Required Exit Code Policy

Document project-specific mapping in `error-codes.md`.
