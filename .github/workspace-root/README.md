# workspace-root

This repository is consumed by the private Gofinity platform monorepo as a git
submodule at `packages/gofinity`, and several of its configuration files reach
one level *above* it:

- `bunfig.toml` preloads `../../test/preload.ts`
- `tsconfig.json` extends `../../tsconfig.base.json`
- `bunx biome` and `bunx tsc` resolve from the monorepo's root `node_modules`,
  and Biome reads the monorepo's root `biome.json`

Checked out on its own, none of that exists, so `bun run lint`, `bun run
typecheck` and `bun test` cannot run. The files here are the minimum workspace
root that makes them run: `.github/workflows/content.yml` copies them one level
above the checkout before it does anything else.

| File | Copied to | Why |
| --- | --- | --- |
| `root-package.json` | `package.json` | makes this repository a Bun workspace member, and pins the Biome, TypeScript and `@types/bun` versions the monorepo uses |
| `root-biome.json` | `biome.json` | the formatting and lint rules |
| `root-tsconfig.base.json` | `tsconfig.base.json` | the compiler options `tsconfig.json` extends |
| `root-preload.ts` | `test/preload.ts` | the test guard `bunfig.toml` preloads |

They are copies, not the originals, and they are deliberately not verbatim: the
monorepo's versions name private paths (`apps/api`, `packages/db`, `e2e`) that
must not appear in a public repository. Nothing here is used when this
repository is checked out as a submodule - the monorepo's own root files win.

Keep them in step with the monorepo when its root config changes. The signal
that they have drifted is this repository's CI disagreeing with the monorepo's.
The `root-` prefixes exist so no tool auto-discovers these as configuration in
their real location.
