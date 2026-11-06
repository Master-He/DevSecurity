#!/usr/bin/env sh
set -eu

ROOT="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
ARCHIVE="$ROOT/SAST/joern-cli-windows-x86_64.zip"
RUNTIME_ROOT="$ROOT/SAST/.runtime"
JOERN_HOME="$RUNTIME_ROOT/joern-cli"
MARKER="$JOERN_HOME/lib/io.joern.joern-cli-4.0.674.jar"

if [ -f "$MARKER" ]; then
  printf 'Joern 4.0.674 runtime is ready: %s\n' "$JOERN_HOME"
  exit 0
fi
if [ ! -f "$ARCHIVE" ]; then
  printf 'Joern archive not found: %s\n' "$ARCHIVE" >&2
  exit 1
fi

mkdir -p "$RUNTIME_ROOT"
STAGING="$RUNTIME_ROOT/.extract-$$"
mkdir -p "$STAGING"
trap 'rm -rf "$STAGING"' EXIT INT TERM
if command -v unzip >/dev/null 2>&1; then
  unzip -q "$ARCHIVE" -d "$STAGING"
else
  python3 -m zipfile -e "$ARCHIVE" "$STAGING"
fi
[ -f "$STAGING/joern-cli/lib/io.joern.joern-cli-4.0.674.jar" ]
rm -rf "$JOERN_HOME"
mv "$STAGING/joern-cli" "$JOERN_HOME"
printf 'Joern 4.0.674 runtime prepared at %s\n' "$JOERN_HOME"
