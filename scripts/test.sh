#!/usr/bin/env bash
# test.sh — run tests for stream-assistant
#
# Usage:
#   ./scripts/test.sh [options]
#
# Options:
#   --unit            Run unit tests only           (default: all non-integration)
#   --integration     Run integration tests only    (requires network access)
#   --all             Run both unit and integration tests
#   --coverage        Generate coverage report      (outputs coverage/coverage.html)
#   --race            Enable race detector
#   --run <pattern>   Filter tests by name pattern  (passed to -run)
#   --timeout <dur>   Test timeout                  (default: 2m; integration: 5m)
#   --verbose         Enable verbose output (-v)
#   --help            Show this message

set -euo pipefail
cd "$(dirname "$0")/.."

RUN_UNIT=1
RUN_INTEGRATION=0
COVERAGE=0
RACE=0
RUN_FILTER=""
TIMEOUT=""
VERBOSE=0

while [[ $# -gt 0 ]]; do
  case "$1" in
    --unit)          RUN_UNIT=1; RUN_INTEGRATION=0 ;;
    --integration)   RUN_UNIT=0; RUN_INTEGRATION=1 ;;
    --all)           RUN_UNIT=1; RUN_INTEGRATION=1 ;;
    --coverage)      COVERAGE=1 ;;
    --race)          RACE=1 ;;
    --run)           RUN_FILTER="$2"; shift ;;
    --timeout)       TIMEOUT="$2"; shift ;;
    --verbose)       VERBOSE=1 ;;
    --help)
      sed -n '/^# Usage/,/^[^#]/{ /^[^#]/d; p }' "$0" | sed 's/^# \?//'
      exit 0 ;;
    *) echo "Unknown option: $1" >&2; exit 1 ;;
  esac
  shift
done

RACE_FLAG=""
[[ $RACE -eq 1 ]] && RACE_FLAG="-race"

VERBOSE_FLAG=""
[[ $VERBOSE -eq 1 ]] && VERBOSE_FLAG="-v"

FILTER_FLAG=""
[[ -n "$RUN_FILTER" ]] && FILTER_FLAG="-run ${RUN_FILTER}"

# ── unit tests ──────────────────────────────────────────────────────────────
if [[ $RUN_UNIT -eq 1 ]]; then
  UNIT_TIMEOUT="${TIMEOUT:-2m}"
  UNIT_PKGS=$(go list ./... | grep -v "test/integration")
  echo "Running unit tests (timeout: ${UNIT_TIMEOUT})..."

  COVER_FLAGS=""
  if [[ $COVERAGE -eq 1 ]]; then
    mkdir -p coverage
    COVER_FLAGS="-coverprofile=coverage/unit.out -covermode=atomic"
  fi

  # shellcheck disable=SC2086
  go test $RACE_FLAG $VERBOSE_FLAG -timeout "$UNIT_TIMEOUT" $COVER_FLAGS $FILTER_FLAG $UNIT_PKGS
  echo "Unit tests passed."
fi

# ── integration tests ────────────────────────────────────────────────────────
if [[ $RUN_INTEGRATION -eq 1 ]]; then
  INT_TIMEOUT="${TIMEOUT:-5m}"
  echo "Running integration tests (timeout: ${INT_TIMEOUT})..."
  echo "Warning: integration tests require network access."

  COVER_FLAGS=""
  if [[ $COVERAGE -eq 1 ]]; then
    mkdir -p coverage
    COVER_FLAGS="-coverprofile=coverage/integration.out -covermode=atomic"
  fi

  # shellcheck disable=SC2086
  go test -tags integration $RACE_FLAG $VERBOSE_FLAG -timeout "$INT_TIMEOUT" $COVER_FLAGS $FILTER_FLAG ./test/integration/
  echo "Integration tests passed."
fi

# ── coverage report ──────────────────────────────────────────────────────────
if [[ $COVERAGE -eq 1 ]]; then
  PROFILES=()
  [[ -f coverage/unit.out ]]        && PROFILES+=(coverage/unit.out)
  [[ -f coverage/integration.out ]] && PROFILES+=(coverage/integration.out)

  if [[ ${#PROFILES[@]} -gt 0 ]]; then
    # Merge profiles if both exist
    if [[ ${#PROFILES[@]} -gt 1 ]]; then
      echo "mode: atomic" > coverage/merged.out
      for p in "${PROFILES[@]}"; do
        tail -n +2 "$p" >> coverage/merged.out
      done
      FINAL="coverage/merged.out"
    else
      FINAL="${PROFILES[0]}"
    fi
    go tool cover -html="$FINAL" -o coverage/coverage.html
    echo "Coverage report → coverage/coverage.html"
    go tool cover -func="$FINAL" | tail -1
  fi
fi
