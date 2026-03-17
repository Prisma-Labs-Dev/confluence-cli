#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root"

go_files=()
file=""
while IFS= read -r file; do
  go_files+=("$file")
done < <(git ls-files '*.go')

if [[ "${#go_files[@]}" -eq 0 ]]; then
  exit 0
fi

unformatted="$(printf '%s\0' "${go_files[@]}" | xargs -0 gofmt -l)"
if [[ -n "$unformatted" ]]; then
  {
    echo "check-gofmt: unformatted Go files detected:"
    echo "$unformatted"
    echo
    echo "Run 'make fmt' to format tracked Go files safely."
  } >&2
  exit 1
fi
