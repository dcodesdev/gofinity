#!/usr/bin/env sh
# Run the runner's Go tests.
#
# Without a Go toolchain this skips rather than fails, so `bun test` at the root
# still works on a machine that only has Bun. CI sets REQUIRE_GO=1 so the
# coverage cannot quietly disappear — the same bargain the DB tests make with
# REQUIRE_TEST_DB.
set -eu

cd "$(dirname "$0")/.."

if ! command -v go >/dev/null 2>&1; then
  if [ "${REQUIRE_GO:-}" = "1" ]; then
    echo "FAIL: no Go toolchain, and REQUIRE_GO=1" >&2
    exit 1
  fi
  echo "SKIP: no Go toolchain — set REQUIRE_GO=1 to make this a failure"
  exit 0
fi

gofmt -l . | tee /tmp/gofinity-gofmt.$$ >&2
if [ -s /tmp/gofinity-gofmt.$$ ]; then
  rm -f /tmp/gofinity-gofmt.$$
  echo "FAIL: the files above are not gofmt-clean" >&2
  exit 1
fi
rm -f /tmp/gofinity-gofmt.$$

go vet ./...
go test ./...
