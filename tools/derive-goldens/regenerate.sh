#!/usr/bin/env bash
# Regenerate the `tokenctl derive` golden fixtures.
#
# THIS IS A DELIBERATE ACT. `go test` never runs it. The goldens in
# testdata/derive/goldens are committed fixtures that pin the Go port to
# the behaviour of the original TypeScript engine; regenerating them
# replaces the reference, so only run this when the TypeScript engine
# itself has intentionally changed — and review the resulting diff.
#
# Requirements:
#   - node (>= 18)
#   - the tokenctl-extension checkout beside this repo, with its
#     node_modules installed (colorjs.io + esbuild come from there)
#
# Usage:
#   tools/derive-goldens/regenerate.sh [path/to/tokenctl-extension]
set -euo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo="$(cd "$here/../.." && pwd)"
ext="${1:-$(cd "$repo/../tokenctl-extension" 2>/dev/null && pwd || true)}"

if [[ -z "${ext:-}" || ! -f "$ext/src/lib/theme-engine.ts" ]]; then
  echo "error: cannot find tokenctl-extension/src/lib/theme-engine.ts" >&2
  echo "       pass the extension checkout path as the first argument" >&2
  exit 1
fi

esbuild="$ext/node_modules/.bin/esbuild"
if [[ ! -x "$esbuild" ]]; then
  echo "error: $esbuild not found — run 'npm install' in $ext" >&2
  exit 1
fi

out="$repo/testdata/derive/goldens"
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT

echo "engine:   $ext/src/lib/theme-engine.ts"
echo "goldens:  $out"

# Bundle the harness together with the real engine. The @engine alias
# points at the extension's source so the harness imports the original
# file rather than a copy. Bundling also pulls in colorjs.io, which
# resolves out of the extension's node_modules.
"$esbuild" "$here/harness.ts" \
  --bundle \
  --platform=node \
  --format=esm \
  --target=node18 \
  --alias:@engine="$ext/src/lib/theme-engine.ts" \
  --outfile="$work/harness.mjs" \
  --log-level=warning

rm -rf "$out"
mkdir -p "$out"
node "$work/harness.mjs" "$out"

echo
echo "Done. Review the diff before committing:"
echo "  git -C \"$repo\" diff --stat testdata/derive/goldens"
