/**
 * The `content:verify` checks as tests, so `bun test` grades the content rather
 * than only validating that it loads.
 *
 * Slow by construction: two `go test` runs per published challenge, each one a
 * fresh temp workspace with a cold build cache. Run
 * `bun run content:verify <track>/<challenge>` while iterating on one challenge
 * and leave the whole sweep to `bun test`.
 *
 * Without a Go toolchain the suite skips, like the other Go-dependent tests.
 * REQUIRE_GO=1 turns the skip into a failure, and CI sets it.
 */
import { describe, expect, test } from "bun:test"
import {
  type Direction,
  expectedOk,
  hasGo,
  runDirection,
  verifyTargets,
} from "./verify-challenge.ts"

const go = hasGo()

// A plain test rather than a skipped suite: under REQUIRE_GO the point is to
// fail loudly, and a suite that skips itself reports success.
test.if(process.env.REQUIRE_GO === "1")("a Go toolchain is available", () => {
  expect(go).toBe(true)
})

const targets = go ? verifyTargets() : []

describe.skipIf(!go)("content verifies against a real Go toolchain", () => {
  test("there is published content to verify", () => {
    expect(targets.length).toBeGreaterThan(0)
  })

  const expectation: Record<Direction, string> = {
    solution: "passes with solution/ overlaid",
    starter: "fails without solution/",
  }

  for (const { name, challenge } of targets) {
    for (const direction of ["solution", "starter"] as const) {
      // The timeout is a cold `go test` on a fresh temp module: seconds, not
      // milliseconds.
      const timeout = 120_000
      test(
        `${name} ${expectation[direction]}`,
        async () => {
          const { ok, output } = await runDirection(challenge, direction)
          if (ok !== expectedOk(direction)) {
            throw new Error(
              direction === "solution"
                ? `tests do not pass against solution/\n${output}`
                : `tests already pass without any work\n${output}`,
            )
          }
        },
        timeout,
      )
    }
  }
})
