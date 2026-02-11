# Work Tracking

Use work packets to track implementation and review progress explicitly.

## Lifecycle

1. Create packet in `active/` from `_WORK_PACKET_TEMPLATE.md`.
2. Keep findings table current while work is in progress.
3. Mark each finding with both:
- `Status`: `todo | in_progress | blocked | done | wontfix`
- `Result`: `pass | fail | n/a`
4. When all exit criteria are met, move packet to `archive/<year>/`.
5. Append one summary row to `INDEX.md`.

## Purpose

This keeps context-window work explicit for the next agent:

1. what problems were identified
2. what was fixed
3. what is still open
4. why decisions were made
