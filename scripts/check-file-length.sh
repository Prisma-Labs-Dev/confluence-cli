#!/usr/bin/env bash
set -euo pipefail

max_go_lines="${MAX_GO_LINES:-300}"
max_go_test_lines="${MAX_GO_TEST_LINES:-700}"
status=0

while IFS= read -r -d '' file; do
  lines=$(wc -l < "$file" | tr -d ' ')
  limit="$max_go_lines"
  if [[ "$file" == *_test.go ]]; then
    limit="$max_go_test_lines"
  fi

  if (( lines > limit )); then
    printf 'file too large: %s has %s lines (limit %s)\n' "$file" "$lines" "$limit" >&2
    status=1
  fi
done < <(find . -type f -name '*.go' -not -path './.git/*' -print0)

exit "$status"
