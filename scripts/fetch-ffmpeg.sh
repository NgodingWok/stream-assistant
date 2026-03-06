#!/usr/bin/env bash
# fetch-ffmpeg.sh — download the FFmpeg Windows binary into third_party/bin/
#
# Usage:
#   ./scripts/fetch-ffmpeg.sh [options]
#
# Options:
#   --dir <path>   Output directory  (default: third_party/bin)
#   --help         Show this message
#
# Source: BtbN FFmpeg Builds (https://github.com/BtbN/FFmpeg-Builds)
# Downloads the GPL essentials build (statically linked — no extra DLLs needed).

set -euo pipefail
cd "$(dirname "$0")/.."

OUTPUT_DIR="third_party/bin"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --dir)   OUTPUT_DIR="$2"; shift ;;
    --help)
      sed -n '/^# Usage/,/^[^#]/{ /^[^#]/d; p }' "$0" | sed 's/^# \?//'
      exit 0 ;;
    *) echo "Unknown option: $1" >&2; exit 1 ;;
  esac
  shift
done

ZIP_URL="https://github.com/BtbN/FFmpeg-Builds/releases/download/latest/ffmpeg-master-latest-win64-gpl-essentials.zip"
DEST="${OUTPUT_DIR}/ffmpeg.exe"

mkdir -p "$OUTPUT_DIR"

TMP_ZIP=$(mktemp --suffix=.zip)
trap 'rm -f "$TMP_ZIP"' EXIT

echo "Downloading FFmpeg Windows binary (gpl-essentials)..."
curl -fsSL "$ZIP_URL" -o "$TMP_ZIP"

echo "Extracting ffmpeg.exe..."
# The zip contains a single top-level folder; ffmpeg.exe is in its bin/ sub-directory.
unzip -p "$TMP_ZIP" "*/bin/ffmpeg.exe" > "$DEST"

echo "Done -> ${DEST}"
