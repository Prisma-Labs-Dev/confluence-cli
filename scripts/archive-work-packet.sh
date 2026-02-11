#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 1 ]]; then
  echo "usage: $0 <packet-file-name.md>" >&2
  exit 2
fi

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
PACKET_NAME="$1"
SRC="$ROOT_DIR/docs/work/active/$PACKET_NAME"

if [[ ! -f "$SRC" ]]; then
  echo "packet not found: $SRC" >&2
  exit 1
fi

YEAR="$(date +%Y)"
DEST_DIR="$ROOT_DIR/docs/work/archive/$YEAR"
mkdir -p "$DEST_DIR"
mv "$SRC" "$DEST_DIR/$PACKET_NAME"

echo "archived: $DEST_DIR/$PACKET_NAME"
echo "next: update docs/work/INDEX.md with completion entry"
