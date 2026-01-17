#!/usr/bin/env bash
set -euo pipefail

# Run Go tests with coverage and summarize results.
# Optional: set COVERAGE_THRESHOLD (e.g., 80) to enforce a minimum percentage.

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

COVER_PROFILE="coverage.out"
COVER_FUNC="coverage.txt"

# Clean previous artifacts
rm -f "$COVER_PROFILE" "$COVER_FUNC"

echo "==> computing package list"
# Exclusions target integration-heavy or process wiring code that is impractical to unit test:
# - cmd: CLI entrypoint is exercised via end-to-end usage, not unit coverage.
# - config: config loading/validation is environment and filesystem dependent; validated via integration smoke tests.
# - server/auth and server/middleware: token verification hits remote endpoints; we validate via integration flows instead.
# - server/server and server (root): HTTP server wiring and lifecycle are covered by integration tests, not unit harnesses.
# - server/handler/common: thin logging/error glue covered indirectly.
# - server/integration: end-to-end focused; leave out of unit coverage denominator.
# - storage/content: git-backed storage depends on external git tooling and repositories; exercised in integration/e2e.
# - storage/media: real media backends (e.g., S3) require external services; verified through deployment tests.
EXCLUDE_REGEX="(^github.com/indieinfra/scribble/cmd($|/))|(^github.com/indieinfra/scribble/config($|/))|(^github.com/indieinfra/scribble/server$)|/server/server|/server/handler/common|/server/integration|/server/auth|/server/middleware|/storage/content|/storage/media"

PKGS=$(go list ./... | grep -Ev "$EXCLUDE_REGEX" || true)
if [[ -z "$PKGS" ]]; then
  echo "No packages matched after exclusions: $EXCLUDE_REGEX" >&2
  exit 1
fi
COVERPKG=$(echo "$PKGS" | paste -sd, -)

echo "Packages under coverage:" $PKGS
echo "==> go test with coverage"
go test $PKGS -coverpkg="$COVERPKG" -coverprofile="$COVER_PROFILE" -covermode=atomic

echo "==> coverage summary"
go tool cover -func="$COVER_PROFILE" | tee "$COVER_FUNC"

TOTAL_LINE=$(grep "^total:" "$COVER_FUNC" || true)
if [[ -z "$TOTAL_LINE" ]]; then
  echo "Failed to parse total coverage from $COVER_FUNC" >&2
  exit 1
fi

TOTAL_PCT=$(echo "$TOTAL_LINE" | awk '{print $3}' | tr -d '%')
if [[ -z "$TOTAL_PCT" ]]; then
  echo "Failed to extract coverage percentage" >&2
  exit 1
fi

echo "Total coverage: ${TOTAL_PCT}%"

if [[ -n "${COVERAGE_THRESHOLD:-}" ]]; then
  # Compare as integer; assumes threshold provided as whole number (e.g., 80)
  THRESH=${COVERAGE_THRESHOLD%.*}
  PCT_INT=${TOTAL_PCT%.*}
  if (( PCT_INT < THRESH )); then
    echo "Coverage ${TOTAL_PCT}% is below threshold ${THRESH}%" >&2
    exit 2
  fi
  echo "Coverage meets threshold ${THRESH}%"
fi

