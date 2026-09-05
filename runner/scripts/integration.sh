#!/usr/bin/env sh
# End-to-end check of the built runner image.
#
# Every case here is something unit tests cannot prove, because it is a property
# of the container rather than of the code: that the image runs at all, that the
# kill actually kills, and that there is no network inside it.
#
#   ./scripts/integration.sh
#   RUNNER_IMAGE=foo:bar ./scripts/integration.sh
#
# Without a Docker daemon the script skips, the same way DB tests skip without
# Postgres. Set REQUIRE_DOCKER=1 (CI does) to make that a failure instead.
set -eu

IMAGE="${RUNNER_IMAGE:-gofinity-runner:latest}"
HERE="$(cd "$(dirname "$0")" && pwd)"
WORK="$(mktemp -d)"
FAILURES=0

cleanup() { rm -rf "$WORK"; }
trap cleanup EXIT

if ! docker info >/dev/null 2>&1; then
  if [ "${REQUIRE_DOCKER:-}" = "1" ]; then
    echo "FAIL: no Docker daemon, and REQUIRE_DOCKER=1" >&2
    exit 1
  fi
  echo "SKIP: no Docker daemon reachable - set REQUIRE_DOCKER=1 to make this a failure"
  exit 0
fi

if ! docker image inspect "$IMAGE" >/dev/null 2>&1; then
  echo "FAIL: image $IMAGE is not built - run ./build.sh first" >&2
  exit 1
fi

# --- fixtures ---------------------------------------------------------------

cat > "$WORK/go.mod" <<'EOF'
module gofinity/hello

go 1.24
EOF

cat > "$WORK/main_test.go" <<'EOF'
package main

import "testing"

func TestGreet(t *testing.T) {
	if got := Greet("Ada"); got != "Hello, Ada!" {
		t.Errorf("Greet(\"Ada\") = %q", got)
	}
}
EOF

cat > "$WORK/pass.go" <<'EOF'
package main

import "fmt"

func Greet(name string) string { return fmt.Sprintf("Hello, %s!", name) }

func main() { fmt.Println(Greet("Gofinity")) }
EOF

cat > "$WORK/fail.go" <<'EOF'
package main

func Greet(name string) string { return "" }

func main() {}
EOF

cat > "$WORK/broken.go" <<'EOF'
package main

func Greet(name string) string { return 1 }

func main() {}
EOF

cat > "$WORK/loop.go" <<'EOF'
package main

func Greet(name string) string {
	for {
	}
}

func main() {}
EOF

# The network case has its own test file: it passes only when the dial fails,
# so `ok: true` is the proof that the container is isolated.
cat > "$WORK/network.go" <<'EOF'
package main

func Greet(name string) string { return "" }

func main() {}
EOF

cat > "$WORK/network_test.go" <<'EOF'
package main

import (
	"net"
	"testing"
	"time"
)

func TestTheContainerHasNoNetwork(t *testing.T) {
	conn, err := net.DialTimeout("tcp", "1.1.1.1:80", 3*time.Second)
	if err == nil {
		conn.Close()
		t.Fatal("the container reached the network")
	}
	t.Logf("dial refused as expected: %v", err)
}
EOF

# --- helpers ----------------------------------------------------------------

# The flags below mirror the HostConfig the API builds in Phase 8. If they
# diverge, this script stops testing what actually runs in production.
run_container() {
  payload="$1"
  timeout_s="$2"
  runner="docker run --rm \
    --network none \
    --memory 512m --memory-swap 512m --cpus 1 --pids-limit 128 \
    --read-only --tmpfs /tmp:rw,exec,size=256m,mode=1777 \
    --cap-drop ALL --security-opt no-new-privileges \
    --user 65532:65532 \
    --label gofinity.runner=integration \
    -e GOFINITY_PAYLOAD=$payload \
    $IMAGE"

  if command -v timeout >/dev/null 2>&1; then
    timeout "$timeout_s" $runner 2>&1 || true
  else
    $runner 2>&1 || true
  fi
}

# A payload near the size limits is far past the kernel's per-argument limit, so
# it goes in over stdin - which is also the entrypoint's documented fallback.
run_container_stdin() {
  timeout_s="$1"
  runner="docker run --rm -i \
    --network none \
    --memory 512m --memory-swap 512m --cpus 1 --pids-limit 128 \
    --read-only --tmpfs /tmp:rw,exec,size=256m,mode=1777 \
    --cap-drop ALL --security-opt no-new-privileges \
    --user 65532:65532 \
    --label gofinity.runner=integration \
    $IMAGE"

  if command -v timeout >/dev/null 2>&1; then
    timeout "$timeout_s" $runner 2>&1 || true
  else
    $runner 2>&1 || true
  fi
}

# Extract the framed result. The contract is in README.md: the last begin
# sentinel, up to the next end sentinel.
extract_result() {
  awk '
    /^<<<GOFINITY:RESULT>>>$/ { body = ""; capture = 1; next }
    /^<<<GOFINITY:END>>>$/    { capture = 0; next }
    capture                   { body = body $0 }
    END                       { print body }
  '
}

expect() {
  label="$1"
  needle="$2"
  body="$3"
  if printf '%s' "$body" | grep -q -- "$needle"; then
    echo "  ok: $label"
  else
    echo "  FAIL: $label - expected $needle in:" >&2
    printf '    %s\n' "$body" >&2
    FAILURES=$((FAILURES + 1))
  fi
}

refute() {
  label="$1"
  needle="$2"
  body="$3"
  if printf '%s' "$body" | grep -q -- "$needle"; then
    echo "  FAIL: $label - did not expect $needle in:" >&2
    printf '    %s\n' "$body" >&2
    FAILURES=$((FAILURES + 1))
  else
    echo "  ok: $label"
  fi
}

encode() {
  bun run "$HERE/payload.ts" "$@"
}

# --- cases ------------------------------------------------------------------

echo "1. a passing solution"
BODY="$(run_container "$(encode go.mod="$WORK/go.mod" main.go="$WORK/pass.go" main_test.go="$WORK/main_test.go")" 60 | extract_result)"
expect "ok is true" '"ok":true' "$BODY"
expect "a test passed" '"status":"passed"' "$BODY"
expect "nothing timed out" '"timedOut":false' "$BODY"

echo "2. a failing solution"
BODY="$(run_container "$(encode go.mod="$WORK/go.mod" main.go="$WORK/fail.go" main_test.go="$WORK/main_test.go")" 60 | extract_result)"
expect "ok is false" '"ok":false' "$BODY"
expect "a test failed" '"status":"failed"' "$BODY"

echo "3. a compile error"
BODY="$(run_container "$(encode go.mod="$WORK/go.mod" main.go="$WORK/broken.go" main_test.go="$WORK/main_test.go")" 60 | extract_result)"
expect "ok is false" '"ok":false' "$BODY"
expect "the compiler error is reported" 'cannot use 1' "$BODY"
refute "it is not a runner error" '"error":' "$BODY"

echo "4. an infinite loop is killed"
STARTED=$(date +%s)
BODY="$(run_container "$(encode --timeout-ms 6000 go.mod="$WORK/go.mod" main.go="$WORK/loop.go" main_test.go="$WORK/main_test.go")" 60 | extract_result)"
ELAPSED=$(( $(date +%s) - STARTED ))
expect "it timed out" '"timedOut":true' "$BODY"
expect "ok is false" '"ok":false' "$BODY"
if [ "$ELAPSED" -gt 20 ]; then
  echo "  FAIL: the kill took ${ELAPSED}s, well past the 6s budget" >&2
  FAILURES=$((FAILURES + 1))
else
  echo "  ok: killed within ${ELAPSED}s"
fi

echo "5. there is no network"
BODY="$(run_container "$(encode go.mod="$WORK/go.mod" main.go="$WORK/network.go" main_test.go="$WORK/network_test.go")" 60 | extract_result)"
expect "the dial failed, so the test passed" '"ok":true' "$BODY"

echo "6. an oversized payload is rejected (over stdin)"
head -c 2000000 /dev/zero | tr '\0' 'x' > "$WORK/huge.txt"
BODY="$(encode go.mod="$WORK/go.mod" main.go="$WORK/pass.go" huge.txt="$WORK/huge.txt" | run_container_stdin 60 | extract_result)"
expect "it is a runner error" '"error":' "$BODY"
expect "the limit is named" 'limit' "$BODY"
refute "nothing ran" '"status":' "$BODY"

echo "7. no containers were left behind"
LEFTOVER="$(docker ps -aq --filter label=gofinity.runner=integration | wc -l | tr -d ' ')"
if [ "$LEFTOVER" != "0" ]; then
  echo "  FAIL: $LEFTOVER container(s) survived" >&2
  FAILURES=$((FAILURES + 1))
else
  echo "  ok: none"
fi

if [ "$FAILURES" -ne 0 ]; then
  echo "$FAILURES check(s) failed" >&2
  exit 1
fi
echo "all runner integration checks passed"
