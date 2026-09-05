#!/usr/bin/env bun
/**
 * Verify challenges by running their tests with a real Go toolchain.
 *
 * Agents have no Docker, so the runner image cannot grade content here. This
 * does the same job one level down: it materialises a challenge's workspace in
 * a temp directory and runs `go test ./...` against it.
 *
 * Two directions, and a challenge is only correct if both hold:
 *
 *   solution  `files/` overlaid with `solution/` must **pass**.
 *   starter   `files/` untouched must **fail**: tests that pass before any
 *             work is done grade nothing, and at 63 challenges that has to be
 *             caught mechanically rather than by eye.
 *
 *   bun run scripts/verify-challenge.ts                  # both, every published challenge
 *   bun run scripts/verify-challenge.ts --starter        # the starter direction only
 *   bun run scripts/verify-challenge.ts --solution       # the solution direction only
 *   bun run scripts/verify-challenge.ts go-basics/hello-gofinity   # one challenge
 *
 * Without a Go toolchain it skips rather than fails, exactly like
 * `runner/scripts/test.sh`. REQUIRE_GO=1 turns that skip into a failure, and
 * CI sets it.
 */
import { mkdirSync, mkdtempSync, rmSync, writeFileSync } from "node:fs"
import { tmpdir } from "node:os"
import { dirname, join } from "node:path"
import { ContentError, type LoadedChallenge, loadContent } from "../src/load.ts"

type Direction = "solution" | "starter"

const args = process.argv.slice(2)
const wantStarter = args.includes("--starter")
const wantSolution = args.includes("--solution")
const filters = args.filter((a) => !a.startsWith("--"))
const directions: Direction[] =
  wantStarter === wantSolution ? ["solution", "starter"] : wantStarter ? ["starter"] : ["solution"]

function hasGo(): boolean {
  return Bun.which("go") !== null
}

if (!hasGo()) {
  if (process.env.REQUIRE_GO === "1") {
    console.error("FAIL: no Go toolchain, and REQUIRE_GO=1")
    process.exit(1)
  }
  console.log("SKIP: no Go toolchain, set REQUIRE_GO=1 to make this a failure")
  process.exit(0)
}

/** Write `files/`, then overlay `solution/` on top when the direction asks for it. */
function materialise(challenge: LoadedChallenge, direction: Direction): string {
  const dir = mkdtempSync(join(tmpdir(), "gofinity-verify-"))
  const write = (path: string, content: string) => {
    const target = join(dir, path)
    mkdirSync(dirname(target), { recursive: true })
    writeFileSync(target, content)
  }
  for (const file of challenge.files) write(file.path, file.content)
  if (direction === "solution") {
    for (const file of challenge.solutionFiles) write(file.path, file.content)
  }
  return dir
}

interface RunResult {
  ok: boolean
  output: string
}

async function goTest(dir: string): Promise<RunResult> {
  const proc = Bun.spawn(["go", "test", "./..."], {
    cwd: dir,
    stdout: "pipe",
    stderr: "pipe",
    // The content is standard library only, so nothing should ever be fetched.
    // Turning the proxy off makes an accidental dependency a loud failure
    // rather than a slow network call.
    env: { ...process.env, GOPROXY: "off", GOFLAGS: "-mod=mod" },
  })
  const [stdout, stderr, code] = await Promise.all([
    new Response(proc.stdout).text(),
    new Response(proc.stderr).text(),
    proc.exited,
  ])
  return { ok: code === 0, output: `${stdout}${stderr}`.trim() }
}

function indent(text: string): string {
  return text
    .split("\n")
    .map((line) => `      ${line}`)
    .join("\n")
}

let tracks: ReturnType<typeof loadContent>
try {
  tracks = loadContent()
} catch (error) {
  console.error(error instanceof ContentError ? `content error: ${error.message}` : error)
  process.exit(1)
}

const targets: { name: string; challenge: LoadedChallenge }[] = []
for (const track of tracks) {
  for (const challenge of track.challenges) {
    if (!challenge.published) continue
    const name = `${track.slug}/${challenge.slug}`
    if (filters.length > 0 && !filters.includes(name) && !filters.includes(challenge.slug)) continue
    targets.push({ name, challenge })
  }
}

if (targets.length === 0) {
  console.error(
    filters.length > 0
      ? `no published challenge matches ${filters.join(", ")}`
      : "no published challenges to verify",
  )
  process.exit(1)
}

let failures = 0
for (const { name, challenge } of targets) {
  for (const direction of directions) {
    const dir = materialise(challenge, direction)
    try {
      const { ok, output } = await goTest(dir)
      // The starter is expected to fail: that is what proves the tests grade.
      const expected = direction === "solution"
      if (ok === expected) {
        console.log(`ok    ${name}  ${direction}`)
      } else {
        failures++
        console.log(
          direction === "solution"
            ? `FAIL  ${name}  solution  - tests do not pass against solution/`
            : `FAIL  ${name}  starter   - tests already pass without any work`,
        )
        if (output !== "") console.log(indent(output))
      }
    } finally {
      rmSync(dir, { recursive: true, force: true })
    }
  }
}

const checks = targets.length * directions.length
if (failures > 0) {
  console.error(`${failures} of ${checks} check(s) failed`)
  process.exit(1)
}
console.log(`ok - ${targets.length} challenge(s), ${checks} check(s)`)
