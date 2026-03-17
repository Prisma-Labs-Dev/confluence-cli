#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root"

usage() {
  cat <<'EOF'
Usage: ./scripts/fmt-go.sh [path ...]

Safely format tracked Go files only.

- With no paths, formats all tracked .go files in the repo.
- With paths, formats only matching tracked .go files.
- Non-Go paths are skipped with a warning instead of being passed to gofmt.
EOF
}

if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
  usage
  exit 0
fi

go_files=()
skipped_paths=()

append_unique() {
  local candidate="$1"
  local existing
  for existing in "${go_files[@]:-}"; do
    if [[ "$existing" == "$candidate" ]]; then
      return 0
    fi
  done
  go_files+=("$candidate")
}

collect_from_path() {
  local path="$1"
  local matched=0
  local file=""

  if [[ -f "$path" ]]; then
    if [[ "$path" == *.go ]]; then
      append_unique "$path"
    else
      skipped_paths+=("$path")
    fi
    return 0
  fi

  while IFS= read -r file; do
    matched=1
    if [[ "$file" == *.go ]]; then
      append_unique "$file"
    fi
  done < <(git ls-files -- "$path")

  if [[ "$matched" -eq 0 ]]; then
    skipped_paths+=("$path")
  elif [[ "$path" != *.go && ! -d "$path" ]]; then
    :
  fi
}

if [[ "$#" -eq 0 ]]; then
  while IFS= read -r file; do
    append_unique "$file"
  done < <(git ls-files '*.go')
else
  for path in "$@"; do
    collect_from_path "$path"
  done
fi

if [[ "${#skipped_paths[@]}" -gt 0 ]]; then
  {
    echo "fmt-go: skipped non-Go or unmatched path(s):"
    for path in "${skipped_paths[@]}"; do
      echo "  - $path"
    done
  } >&2
fi

if [[ "${#go_files[@]}" -eq 0 ]]; then
  echo "fmt-go: no Go files selected" >&2
  exit 0
fi

printf '%s\0' "${go_files[@]}" | xargs -0 gofmt -w

echo "fmt-go: formatted ${#go_files[@]} Go file(s)" >&2
