# gofinity

The open-source half of [Gofinity](https://github.com/dcodesdev/gofinity-app):

| Directory | What it is |
| --- | --- |
| `tracks/`, `src/` | challenge content — tracks, starter files, tests, reference solutions — and the Zod schemas that validate them |
| `runner/` | the sandbox that executes untrusted Go code: a Docker image and the Go entrypoint inside it |

MIT licensed, and open to contributions. A challenge is a directory of Go files
and one JSON manifest; if you can write a Go test, you can write one.

The platform consumes this repository as a git submodule.

## This repository stands alone

It has **no dependency on any other workspace package** — content's only
dependency is `zod`, and `runner/` is Go standard library only. Nothing here may
import `@gofinity/db`, `@gofinity/api`, or anything else from the platform
monorepo. Keep it that way.

## The runner image

`.github/workflows/runner.yml` builds, tests and publishes `runner/` on every
commit to `main`:

```sh
docker pull ghcr.io/dcodesdev/gofinity-runner:latest
```

`ghcr.io/dcodesdev/gofinity-runner:sha-<commit>` pins one exact commit.
`runner/README.md` is the reference: the payload, the stdout contract, every
limit.

```sh
cd runner
./build.sh                          # build gofinity-runner:latest locally
./scripts/test.sh                   # gofmt, go vet, go test
./scripts/integration.sh            # container-level checks against the image
```

Both scripts skip when their toolchain is missing; `REQUIRE_GO=1` and
`REQUIRE_DOCKER=1` turn a skip into a failure, and CI sets both.

## Prerequisites

- [Bun](https://bun.sh) 1.4+ — validates and tests the tree.
- Go 1.24+ — for running a challenge's tests the way a learner will.
- Docker — only to build or test `runner/`.

You do not need Docker, a database, or the rest of the platform to contribute a
challenge.

## Commands

```sh
bun install
bun run content:check   # validate the whole tree and print a summary
bun test                # the schema and loader tests, over the real tree
```

`content:check` exits non-zero on the first problem and names the offending
path. CI runs both, in a job with no database, no Docker and no browser, so a
content-only contribution gets a verdict in under a minute.

## Layout

```
tracks/
  <track-slug>/
    track.json
    challenges/
      <NN-challenge-slug>/
        challenge.json
        description.md
        files/          # the workspace the learner gets: starter, tests, go.mod
        solution/       # reference solution — a test fixture, never served
```

A track's slug is its directory name. A challenge's slug is its directory name
minus the two-digit ordering prefix (`01-hello-gofinity` → `hello-gofinity`),
and that prefix must equal `order` in `challenge.json`. Slugs come from
directory names, so they cannot drift out of sync with the JSON.

[SCHEMA.md](./SCHEMA.md) is the field-by-field JSON contract.

## Add a challenge

1. Pick an entry from the track's `roadmap` that is still `"status": "planned"`,
   or add a new one.
2. Create `tracks/<track>/challenges/<NN-slug>/` with `challenge.json`,
   `description.md`, `files/` and `solution/`.
3. Declare every file in `files/` in `challenge.json`. An undeclared file on
   disk is an error, and so is a declared file that does not exist. A challenge
   needs at least one `starter` file and at least one `test` file.
4. Check the tests pass against your own solution:

   ```sh
   cd tracks/<track>/challenges/<NN-slug>
   work=$(mktemp -d) && cp files/* solution/* "$work" && (cd "$work" && go test ./...)
   ```

   The copy goes to a scratch directory so the starter files stay unsolved.

5. Flip the roadmap entry to `"status": "available"`.
6. Run `bun run content:check && bun test`. The suite walks the real tree, so a
   broken contribution fails here rather than in a database.

### File kinds

| `kind` | What it is | Editable by default |
| --- | --- | --- |
| `starter` | the code the learner fills in | yes |
| `test` | tests they can see and run | no |
| `hidden` | grading tests, never sent to a browser | no |
| `readonly` | `go.mod` and anything else they should not touch | no |

`editable` overrides the default. Anything not editable is read from the
database on every run, so the browser cannot substitute it.

`solution/` files are returned by the loader as `challenge.solutionFiles`. They
exist so the platform's own tests have a known-passing submission; they are
never served to a learner.

### Go version

A challenge's `go.mod` `go` directive must not exceed the version baked into the
platform's runner image (currently Go 1.25). Prefer the oldest version your
challenge actually needs.

## Using the package

```ts
import { loadContent } from "@gofinity/content"

const tracks = loadContent() // throws ContentError on anything malformed
```

`loadContent()` reads and validates the entire tree and returns tracks with
their challenges and file contents inlined. It throws on the first problem,
naming the path — the seed script and CI both depend on that being loud rather
than skipping bad content.

## How it reaches the platform

Git is the source of truth. Gofinity's `bun run seed` is a **projection** of
this tree, not an import: it upserts on slug, deletes file rows the tree no
longer declares, and never deletes a challenge or track — those cascade to real
learner submissions. Retire a challenge by setting `"published": false`, never
by deleting its directory.

## License

MIT — see [LICENSE](./LICENSE).
