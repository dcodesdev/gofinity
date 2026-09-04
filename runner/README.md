# `@gofinity/runner`

The sandbox that executes untrusted Go code: a Docker image, and the Go
entrypoint inside it.

One container per submission. It receives a base64 JSON payload, writes the
files into a scratch directory, runs `go test -json`, and prints exactly one
JSON object to stdout. It never has a network, never runs as root, and never
outlives its budget.

## Commands

```sh
./build.sh                  # build and tag gofinity-runner:latest
./scripts/test.sh           # gofmt + go vet + go test (skips without Go)
./scripts/integration.sh    # end-to-end against the built image (skips without Docker)
```

`RUNNER_IMAGE` overrides the tag in both `build.sh` and `integration.sh`; it is
the same variable the API reads, so a custom tag built here is the one the API
runs. `RUNNER_GO_VERSION` overrides the Go version baked into the image.

`REQUIRE_GO=1` and `REQUIRE_DOCKER=1` turn the two "skip" paths into failures.
CI sets both.

## The payload (stdin → container)

Base64-encoded JSON, in `GOFINITY_PAYLOAD` or on stdin. `GOFINITY_PAYLOAD` is
what the API uses; stdin exists because a payload near the size limit is far
past the kernel's per-argument limit, and because it makes the image usable by
hand.

```jsonc
{
  "files": [                          // 1–64 entries, required
    { "path": "main.go", "content": "package main…" }
  ],
  "command": ["go", "test", "-json", "./..."], // optional, must start with "go"
  "timeoutMs": 10000                  // optional, 500–30000, default 10000
}
```

Rejected before anything is written:

| Rule | Limit |
| --- | --- |
| Files | 1–64 |
| Bytes per file | 256 KiB |
| Bytes total | 1 MiB |
| Path length | 255 |
| Path shape | relative, `/`-separated, no empty / `.` / `..` segment |
| Command | must start with `go` |
| `timeoutMs` | 500–30 000 |

The path rules are the same ones `packages/gofinity/src/schema.ts` enforces on
content. Both ends check: content is validated at seed time, submissions at API
time, and this is the last gate before a write.

## The result (container → stdout)

Exactly one JSON object, wrapped in sentinels:

```
<<<GOFINITY:RESULT>>>
{"ok":true,"passed":4,…}
<<<GOFINITY:END>>>
```

**Contract for the consumer:** find the *last* `<<<GOFINITY:RESULT>>>` line,
take everything up to the next `<<<GOFINITY:END>>>` line, parse it as JSON.

The framing exists because stdout is not exclusively ours. The runner captures
the subprocess's streams, so in practice nothing else is printed — but a panic
in the entrypoint, or a future change, could land on the same stream, and the
consumer must not have to guess which line is the result.

```jsonc
{
  "ok": true,          // exited 0, did not time out, no failed test, built cleanly
  "passed": 4,
  "failed": 0,
  "skipped": 0,
  "tests": [
    {
      "name": "TestGreet/a_name",   // subtests are their own entries
      "package": "gofinity/hello",
      "status": "passed",           // passed | failed | skipped
      "output": "    main_test.go:18: …",
      "elapsedMs": 0
    }
  ],
  "stdout": "…",       // reconstructed human-readable `go test` output
  "stderr": "…",       // the subprocess's stderr, kept separate
  "exitCode": 0,
  "timedOut": false,
  "durationMs": 812,
  "error": "…"         // present ONLY on runner-level failure — see below
}
```

Failure modes are deliberately distinguishable, because they mean different
things to the person reading them:

| What happened | How to tell |
| --- | --- |
| Tests failed | `ok: false`, `failed > 0`, no `error` |
| Code did not compile | `ok: false`, `tests` empty, compiler message in `stdout` |
| Ran too long | `timedOut: true` |
| The payload or the runner was wrong | `error` is set, `exitCode: -1` |

A failing test is **not** an error: `error` is only ever a malformed payload, a
workspace that could not be written, or a missing `go` binary.

Exit code: `0` whenever a result was produced (including a failing one), `1` on
a runner-level error. Read `ok`, not the exit code, to decide whether a
submission passed.

## How the container is run

The API builds this `HostConfig` in `apps/api/src/execution/host-config.ts`;
`scripts/integration.sh` mirrors it, so if the two diverge the script stops
testing what actually ships:

```
--network none                 no DNS, no proxy, no exfiltration
--read-only                    the root filesystem is immutable
--tmpfs /tmp:rw,exec,size=256m everything writable, and `exec` — see below
--user 65532:65532             non-root, fixed uid baked into the image
--cap-drop ALL                 no capabilities at all
--security-opt no-new-privileges
--memory / --cpus / --pids-limit
--label gofinity.runner        so a reaper can find strays
```

`exec` on the tmpfs is **required**: `go test` compiles a test binary into the
scratch directory and then runs it. Mounting `/tmp` `noexec` makes every
submission fail with a permission error.

## Why there is a pre-warmed build cache

Compiling the standard library from cold costs more than a submission's whole
time budget. The image builds a representative module at build time and ships
the resulting `GOCACHE` at `/opt/gocache`; because the root filesystem is
read-only, each run copies it onto the tmpfs before starting (`cache.go`). The
copy is a few hundred milliseconds against a tmpfs.

Warming is best-effort: a failure makes the run slow, not wrong.

## Timeouts

The payload's `timeoutMs` is the *outer* kill — what the API gives the
container. The entrypoint sets its own deadline 750 ms below that, so a run that
overruns still returns a real result (partial output, the tests that did finish)
instead of the API only learning that the container vanished.

The child runs in its own process group and the **group** is signalled, because
`go test` compiles and then execs a test binary: killing only the direct child
would leave a runaway loop alive inside the container.

## Go version

The image is `golang:1.25-alpine`. It must be at least the highest `go`
directive of any challenge's `go.mod` (`01-hello-gofinity` declares `go 1.24`),
or that challenge will not build. Bump `GO_VERSION` in the `Dockerfile` and
`RUNNER_GO_VERSION` in `build.sh` together.

## Layout

```
Dockerfile          three stages: entrypoint, cache pre-warm, runtime
main.go             read payload → warm cache → materialize → run → emit
payload.go          decoding, limits, validation
workspace.go        path guard and file materialization
run.go              process execution, timeout, process-group kill
parse.go            `go test -json` → structured results
result.go           the stdout contract, in types
cache.go            pre-warmed GOCACHE copy
prewarm/            the throwaway module compiled at image build time
testdata/           real recorded `go test -json` output
scripts/            test wrapper, integration checks, payload encoder
```
