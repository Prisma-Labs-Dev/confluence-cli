# CLI Agent Contract

This directory defines the machine-facing contract for `confluence`.

## Success envelopes

### Lists

List commands return:

```json
{
  "results": [...],
  "page": {
    "limit": 10,
    "nextCursor": "..."
  },
  "schema": {
    "itemType": "space-summary",
    "fields": ["id", "key", "name", "type", "status"]
  }
}
```

### Single items

Single-object commands return:

```json
{
  "item": {...},
  "schema": {
    "itemType": "page-detail",
    "fields": ["id", "title", "spaceId", "status", "version"]
  }
}
```

The `schema` block is owned by the CLI contract, not by the upstream Confluence payload shape.

## Error envelope

Errors go to stderr as JSON:

```json
{
  "error": {
    "code": "VALIDATION",
    "message": "human-readable explanation",
    "hint": "optional next step"
  }
}
```

## Output guarantees

1. JSON is the default stdout contract.
2. Plain output is available only when explicitly requested with `--format plain`.
3. Output is bounded by default.
4. Body content is omitted unless explicitly requested.
5. Errors never go to stdout.

## API usage notes

- Core read flows (`spaces`, `pages list`, `pages get`, `pages tree`) use Confluence Cloud REST v2.
- `pages search` uses the current supported Confluence Cloud REST v1 search endpoint because REST v2 does not yet provide equivalent search capability.
- The CLI surface remains agent-first even when an upstream API limitation requires a legacy endpoint internally.

## Help as contract surface

The first place an agent should learn the live contract is:

```sh
confluence <command> --help
```

Help output is golden-tested and should stay aligned with implementation.
