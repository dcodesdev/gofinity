# gofinity

The open-source half of [Gofinity](https://github.com/dcodesdev/gofinity-app):

| Directory | What it is |
| --- | --- |
| `lessons/` | Learn mode's reading sequence: one lesson per concept, each pointing at the challenges that drill it |
| `tracks/`, `src/` | challenge content - tracks, starter files, tests, reference solutions - and the Zod schemas that validate them |
| `runner/` | the sandbox that executes untrusted Go code: a Docker image and the Go entrypoint inside it |

MIT licensed, and open to contributions. A challenge is a directory of Go files
and one JSON manifest; if you can write a Go test, you can write one.

The platform consumes this repository as a git submodule.

## This repository stands alone

It has **no dependency on any other workspace package** - content's only
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

- [Bun](https://bun.sh) 1.4+ - validates and tests the tree.
- Go 1.24+ - for running a challenge's tests the way a learner will.
- Docker - only to build or test `runner/`.

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
lessons/
  <NN-lesson-slug>/
    lesson.json
    lesson.md
tracks/
  <track-slug>/
    track.json
    challenges/
      <NN-challenge-slug>/
        challenge.json
        description.md
        files/          # the workspace the learner gets: starter, tests, go.mod
        solution/       # reference solution - a test fixture, never served
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
   Optionally add `tags` (up to five lowercase topic slugs) and
   `estimatedMinutes` (how long the challenge should take); both are display
   metadata, and omitting `estimatedMinutes` is better than guessing.
4. Check the tests both pass against your solution and *fail* against the
   untouched starter:

   ```sh
   bun run content:verify <track>/<challenge>
   ```

   It materialises the workspace in a temp directory and runs `go test ./...`
   there, so your `files/` stay unsolved. Both directions matter: a challenge
   whose tests already pass before any work is done grades nothing. Run it with
   no argument to check every published challenge. Without a Go toolchain it
   skips; `REQUIRE_GO=1` makes that a failure.

5. Flip the roadmap entry to `"status": "available"`.
6. Run `bun run content:check && bun test`. The suite walks the real tree, so a
   broken contribution fails here rather than in a database.

## Add a lesson

A lesson is the reading half of a concept: an article that ends by pointing at
the two or three challenges that drill it.

1. Create `lessons/<NN-slug>/` with `lesson.json` and `lesson.md`. `NN` is the
   lesson's position in the sequence and must be unique.
2. List the challenges it practises in `challenges`, in attempt order, as
   `{ "track": "<track-slug>", "challenge": "<challenge-slug>" }`. Every pair
   must resolve, and a published lesson may only point at published content.
3. Write `lesson.md`, starting with an `# ` heading. Do not hand-write the
   practice links at the end; the site renders that block from `challenges`.
4. Close it with a `## Further reading` section, before the final `## Practise`
   paragraph: 2-5 bullets, each an official Go documentation link plus a short
   clause saying what it is good for. See [Links](#links).
5. Run `bun run content:check && bun test`. `content:check` also warns about any
   published challenge no published lesson practises, which is how the reading
   and practice halves stay in step, and about any published lesson that links
   no documentation.

Set `"published": false` to retire a lesson; never delete its directory.

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

## Links

`description.md` and `lesson.md` may only link **official Go documentation**:
the scheme must be `https`, and the host must be `go.dev` or `pkg.go.dev`.
Anything else, including `http://`, fails `content:check` with the file and the
URL. Relative links (`/tracks/...`, `#anchor`) are in-site navigation and stay
unchecked. `ALLOWED_LINK_HOSTS` in `src/load.ts` is the single place the
allowlist lives.

- A **lesson** ends with `## Further reading`, as above. Point at the spec
  section, the package page, or the Effective Go / blog page that is
  authoritative for what the lesson taught, never at a tutorial.
- A **description** links each standard library identifier the task asks for
  inline, the first time the prose names it:
  `[strings.Fields](https://pkg.go.dev/strings#Fields)`. No references section:
  the instructions pane is narrow, so the link belongs where the name is. A
  challenge that uses nothing beyond `fmt` needs no link.

Links are read out of the prose only, so a URL inside a fenced block or a code
span is not a link. Do not guess a `pkg.go.dev` fragment: a name declared inside
a `const` block has no anchor of its own, and a 404 is worse than linking the
package page. Use `go.dev/blog/...`; `blog.golang.org` only redirects.

`bun run content:links` prints every external link in the tree, grouped by file.

## Using the package

```ts
import { loadAll, loadContent } from "@gofinity/content"

const tracks = loadContent() // throws ContentError on anything malformed
const { tracks: t, lessons, warnings } = loadAll()
```

`loadContent()` reads and validates the track tree and returns tracks with their
challenges and file contents inlined. It throws on the first problem, naming the
path, and the seed script and CI both depend on that being loud rather than
skipping bad content.

`loadAll()` does the same for the whole package: tracks, the lesson sequence with
every `{ track, challenge }` reference resolved, and `warnings`, the soft
problems that must not fail a build (today: a published challenge no lesson
practises, and a published lesson that links no documentation).

## How it reaches the platform

Git is the source of truth. Gofinity's `bun run seed` is a **projection** of
this tree, not an import: it upserts on slug, deletes file rows the tree no
longer declares, and never deletes a challenge or track - those cascade to real
learner submissions. Retire a challenge by setting `"published": false`, never
by deleting its directory.

## License

MIT - see [LICENSE](./LICENSE).
