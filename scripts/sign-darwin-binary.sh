#!/usr/bin/env bash
set -euo pipefail

if [[ $# -lt 2 ]]; then
  echo "usage: $0 <binary-path> <target>" >&2
  exit 2
fi

binary_path=$1
target=$2

case "$target" in
  darwin_*) ;;
  *) exit 0 ;;
esac

if [[ "$(uname)" != "Darwin" ]]; then
  echo "Skipping signing for $binary_path: codesign is only available on macOS."
  exit 0
fi

if ! command -v codesign >/dev/null 2>&1; then
  echo "Skipping signing for $binary_path: codesign is not available."
  exit 0
fi

identity=${A365_CODESIGN_IDENTITY:-}
if [[ -z "$identity" ]] && command -v security >/dev/null 2>&1; then
  identity=$(
    security find-identity -v -p codesigning 2>/dev/null |
      awk '/"Developer ID Application:/ && developer == "" { developer = $2 } /"Apple Development:/ && development == "" { development = $2 } END { if (developer != "") print developer; else if (development != "") print development }'
  )
fi

if [[ -n "$identity" ]]; then
  echo "Signing $binary_path with configured codesigning identity."
  codesign --force --sign "$identity" "$binary_path"
else
  echo "Signing $binary_path ad-hoc."
  codesign --force --sign - "$binary_path"
fi

codesign --verify --verbose=2 "$binary_path"
