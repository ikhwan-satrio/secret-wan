#!/usr/bin/env bash
# Recalculate vendorHash in flake.nix.
#
# The release workflow uses nix-update, which updates the version, source hash,
# and vendorHash together. This script is a local fallback for refreshing only
# vendorHash when nix-update is not available.
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
flake="$root/flake.nix"
fake_hash="sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="

if ! command -v nix >/dev/null 2>&1; then
  echo "error: nix is required" >&2
  exit 1
fi

if [ ! -f "$flake" ]; then
  echo "error: $flake not found" >&2
  exit 1
fi

current_hash="$(awk -F'"' '/vendorHash = / { print $2; exit }' "$flake")"
if [ -z "$current_hash" ]; then
  echo "error: could not find vendorHash in $flake" >&2
  exit 1
fi

replace_hash() {
  local hash="$1"
  awk -v hash="$hash" '
    /vendorHash = / && !done {
      sub(/vendorHash = "[^"]*"/, "vendorHash = \"" hash "\"")
      done = 1
    }
    { print }
  ' "$flake" >"$flake.tmp"
  mv "$flake.tmp" "$flake"
}

backup="$(mktemp)"
cp "$flake" "$backup"
restore() {
  cp "$backup" "$flake"
  rm -f "$backup"
}
trap restore EXIT

replace_hash "$fake_hash"

echo "resolving vendorHash with a fake hash..." >&2
# Use the path: fetcher so the script also works before flake.nix is committed.
output="$(nix --extra-experimental-features "nix-command flakes" build "path:$root#default" --no-link --print-build-logs 2>&1 || true)"

new_hash="$(printf '%s\n' "$output" \
  | grep -oE 'got:[[:space:]]+sha256-[A-Za-z0-9+/=]+' \
  | tail -n1 \
  | sed -E 's/.*got:[[:space:]]+//' || true)"

if [ -z "$new_hash" ]; then
  echo "error: could not determine vendorHash; nix build output follows" >&2
  printf '%s\n' "$output" >&2
  exit 1
fi

restore
trap - EXIT

replace_hash "$new_hash"

if [ "$new_hash" = "$current_hash" ]; then
  echo "vendorHash unchanged: $new_hash"
else
  echo "vendorHash updated: $current_hash -> $new_hash"
fi
