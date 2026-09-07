/**
 * Stand-in for the platform monorepo's `test/preload.ts`, which this
 * repository's `bunfig.toml` preloads at `../../test/preload.ts`.
 *
 * When this repository is checked out on its own there is no parent workspace,
 * so `.github/workflows/content.yml` copies this file into place. It keeps the
 * same guard: tests must never run with production config.
 */

if (process.env.NODE_ENV === "production") {
  const message =
    'NODE_ENV is "production". Tests refuse to run against production config. ' +
    'Unset NODE_ENV or set it to "test".'
  console.error(`\n[test-guard] ${message}\n`)
  throw new Error(`[test-guard] ${message}`)
}

// `bun test` sets NODE_ENV=test itself, but a `bun run` wrapper may not.
if (!process.env.NODE_ENV) process.env.NODE_ENV = "test"
