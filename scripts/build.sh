#!/usr/bin/env bash
# build.sh — compile stream-assistant
#
# Usage:
#   ./scripts/build.sh [options]
#
# Options:
#   --embedded        Embed yt-dlp binaries into the output  (requires third_party/bin/)
#   --output <path>   Output binary path         (default: ./bin/stream-assistant)
#   --os <goos>       Target GOOS                (default: current OS)
#   --arch <goarch>   Target GOARCH              (default: current arch)
#   --version <ver>   Embed version string in binary via ldflags
#   --race            Enable race detector
#   --help            Show this message

set -euo pipefail
cd "$(dirname "$0")/.."

OUTPUT="./bin/stream-assistant"
EMBED=0
GOOS="${GOOS:-}"
GOARCH="${GOARCH:-}"
VERSION=""
RACE=0

while [[ $# -gt 0 ]]; do
  case "$1" in
    --embedded)  EMBED=1 ;;
    --output)    OUTPUT="$2"; shift ;;
    --os)        GOOS="$2"; shift ;;
    --arch)      GOARCH="$2"; shift ;;
    --version)   VERSION="$2"; shift ;;
    --race)      RACE=1 ;;
    --help)
      sed -n '/^# Usage/,/^[^#]/{ /^[^#]/d; p }' "$0" | sed 's/^# \?//'
      exit 0 ;;
    *) echo "Unknown option: $1" >&2; exit 1 ;;
  esac
  shift
done

TAGS=""
if [[ $EMBED -eq 1 ]]; then
  if [[ ! -d "third_party/bin" ]]; then
    echo "error: third_party/bin/ not found — run ./scripts/fetch-ytdlp.sh first" >&2
    exit 1
  fi
  TAGS="-tags embed_ytdlp"
fi

LDFLAGS=""
if [[ -n "$VERSION" ]]; then
  LDFLAGS="-ldflags=-X main.version=${VERSION}"
fi

RACE_FLAG=""
if [[ $RACE -eq 1 ]]; then
  RACE_FLAG="-race"
fi

export GOOS GOARCH

echo "Building stream-assistant..."
echo "  embedded : ${EMBED}"
echo "  output   : ${OUTPUT}"
echo "  GOOS     : ${GOOS:-$(go env GOOS)}"
echo "  GOARCH   : ${GOARCH:-$(go env GOARCH)}"
[[ -n "$VERSION" ]] && echo "  version  : ${VERSION}"

mkdir -p "$(dirname "$OUTPUT")"

# shellcheck disable=SC2086
go build $RACE_FLAG $TAGS $LDFLAGS -o "$OUTPUT" ./cmd/stream-assistant/

echo "Done → ${OUTPUT}"
