/**
 * Encode a runner payload from files on disk.
 *
 * Usage: bun run scripts/payload.ts [--timeout-ms N] <path=source> ...
 *
 * Each argument maps a path inside the runner workspace to a file on disk:
 *
 *   bun run scripts/payload.ts main.go=/tmp/solution.go go.mod=/tmp/go.mod
 *
 * It exists so `scripts/integration.sh` can stay a shell script without having
 * to escape Go source into JSON by hand.
 */
import { basename } from "node:path"

const args = process.argv.slice(2)
const files: { path: string; content: string }[] = []
let timeoutMs: number | undefined

for (let i = 0; i < args.length; i++) {
  const arg = args[i]
  if (arg === undefined) continue

  if (arg === "--timeout-ms") {
    const value = args[++i]
    if (value === undefined) throw new Error("--timeout-ms needs a value")
    timeoutMs = Number(value)
    continue
  }

  const separator = arg.indexOf("=")
  const [target, source] =
    separator === -1 ? [basename(arg), arg] : [arg.slice(0, separator), arg.slice(separator + 1)]
  files.push({ path: target, content: await Bun.file(source).text() })
}

if (files.length === 0) throw new Error("no files given")

const payload = timeoutMs === undefined ? { files } : { files, timeoutMs }
process.stdout.write(Buffer.from(JSON.stringify(payload)).toString("base64"))
