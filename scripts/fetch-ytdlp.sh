#!/usr/bin/env bash
# fetch-ytdlp.sh — download yt-dlp release binaries into third_party/bin/
#
# Usage:
#   ./scripts/fetch-ytdlp.sh [options]
#
# Options:
#   --version <ver>   yt-dlp release tag  (default: latest)
#   --dir <path>      Output directory    (default: third_party/bin)
#   --platform <p>    linux|linux-musl|macos|windows|all  (default: all)
#   --help            Show this message

set -euo pipefail
cd "$(dirname "$0")/.."

YTDLP_VERSION=""
OUTPUT_DIR="third_party/bin"
PLATFORM="all"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --version)  YTDLP_VERSION="$2"; shift ;;
    --dir)      OUTPUT_DIR="$2"; shift ;;
    --platform) PLATFORM="$2"; shift ;;
    --help)
      sed -n '/^# Usage/,/^[^#]/{ /^[^#]/d; p }' "$0" | sed 's/^# \?//'
      exit 0 ;;
    *) echo "Unknown option: $1" >&2; exit 1 ;;
  esac
  shift
done

# ── resolve version ──────────────────────────────────────────────────────────
if [[ -z "$YTDLP_VERSION" ]]; then
  echo "Fetching latest yt-dlp version..."
  YTDLP_VERSION=$(curl -fsSL "https://api.github.com/repos/yt-dlp/yt-dlp/releases/latest" \
    | grep '"tag_name"' | head -1 | sed 's/.*"tag_name": "\(.*\)".*/\1/')
  echo "Latest version: ${YTDLP_VERSION}"
fi

BASE_URL="https://github.com/yt-dlp/yt-dlp/releases/download/${YTDLP_VERSION}"
mkdir -p "$OUTPUT_DIR"

download() {
  local url="$1"
  local dest="$2"
  echo "  Downloading $(basename "$dest")..."
  curl -fsSL "$url" -o "$dest"
}

# ── platform file map ────────────────────────────────────────────────────────
#   platform key → (remote filename, local filename, executable?)
declare -A REMOTE_NAME=(
  [linux]="yt-dlp_linux"
  [linux-arm64]="yt-dlp_linux_aarch64"
  [linux-musl]="yt-dlp_linux_musl"
  [linux-musl-arm64]="yt-dlp_linux_musl_aarch64"
  [macos]="yt-dlp_macos"
  [windows]="yt-dlp.exe"
  [windows-x86]="yt-dlp_x86.exe"
)
declare -A LOCAL_NAME=(
  [linux]="yt-dlp_linux"
  [linux-arm64]="yt-dlp_linux_aarch64"
  [linux-musl]="yt-dlp_musllinux"
  [linux-musl-arm64]="yt-dlp_musllinux_aarch64"
  [macos]="yt-dlp_macos"
  [windows]="yt-dlp.exe"
  [windows-x86]="yt-dlp_x86.exe"
)
declare -A IS_EXEC=(
  [linux]=1 [linux-arm64]=1 [linux-musl]=1 [linux-musl-arm64]=1 [macos]=1
  [windows]=0 [windows-x86]=0
)

if [[ "$PLATFORM" == "all" ]]; then
  TARGETS=(linux linux-arm64 linux-musl linux-musl-arm64 macos windows windows-x86)
else
  IFS=',' read -ra TARGETS <<< "$PLATFORM"
fi

echo "Downloading yt-dlp ${YTDLP_VERSION} → ${OUTPUT_DIR}/"
for target in "${TARGETS[@]}"; do
  if [[ -z "${REMOTE_NAME[$target]+_}" ]]; then
    echo "Unknown platform: $target (valid: linux, linux-arm64, linux-musl, linux-musl-arm64, macos, windows, windows-x86)" >&2
    exit 1
  fi
  dest="${OUTPUT_DIR}/${LOCAL_NAME[$target]}"
  download "${BASE_URL}/${REMOTE_NAME[$target]}" "$dest"
  if [[ "${IS_EXEC[$target]}" -eq 1 ]]; then
    chmod +x "$dest"
  fi
done

echo "Done. Binaries written to ${OUTPUT_DIR}/"
