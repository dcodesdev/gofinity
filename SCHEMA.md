# Content schema

The on-disk format for tracks and challenges. Both JSON files are validated with
Zod (`src/schema.ts`) and **reject unknown keys** - an unrecognised field is a
typo, not an extension point.

## Directory names carry the slugs

| Thing | Slug source | Rule |
| --- | --- | --- |
| Track | `tracks/<track-slug>/` | lowercase, digits, single hyphens: `^[a-z0-9]+(-[a-z0-9]+)*$` |
| Challenge | `tracks/<t>/challenges/<NN-challenge-slug>/` | `NN` is two digits and must equal `order` |

Neither JSON file contains a `slug` field, so the URL and the metadata can never
disagree.

## `track.json`

```json
{
  "title": "Go Basics",
  "description": "Go from zero to comfortable in Go.",
  "order": 0,
  "published": true,
  "roadmap": [
    {
      "slug": "hello-gofinity",
      "title": "Hello, Gofinity",
      "summary": "Your first Go program.",
      "status": "available"
    }
  ]
}
```

| Field | Type | Required | Notes |
| --- | --- | --- | --- |
| `title` | string, 1-120 | yes | |
| `description` | string, 1-2000 | yes | Shown on the track page and in metadata. |
| `order` | integer ≥ 0 | no (0) | Display order among tracks; must be unique. |
| `published` | boolean | no (`false`) | Unpublished tracks are seeded but hidden. |
| `roadmap` | array, ≥ 1 | yes | The full curriculum, including unwritten challenges. |

Roadmap entry: `slug` (slug), `title` (1-120), `summary` (1-400), `status` -
`available` or `planned`.

The roadmap is checked against the directory tree: an entry marked `available`
must have a challenge directory, an entry marked `planned` must not, and every
implemented challenge must appear in the roadmap. That is what keeps the
published curriculum honest as challenges land.

## `challenge.json`

```json
{
  "title": "Hello, Gofinity",
  "difficulty": "easy",
  "order": 1,
  "published": true,
  "tags": ["fmt", "strings", "functions"],
  "estimatedMinutes": 10,
  "files": [
    { "path": "main.go", "kind": "starter" },
    { "path": "main_test.go", "kind": "test" },
    { "path": "go.mod", "kind": "readonly" }
  ]
}
```

| Field | Type | Required | Notes |
| --- | --- | --- | --- |
| `title` | string, 1-120 | yes | |
| `difficulty` | `easy` \| `medium` \| `hard` | yes | Badge only; ordering uses `order`. |
| `order` | integer 0-99 | yes | Must equal the directory's `NN` prefix. |
| `published` | boolean | no (`false`) | |
| `tags` | array of slugs, ≤ 5 | no (`[]`) | Topic chips. Lowercase hyphenated, 1-24 chars, no duplicates. |
| `estimatedMinutes` | integer 1-600 | no | Rough time to solve. Omit it rather than guessing. |
| `files` | array, ≥ 1 | yes | Declares every file in `files/`, in tab order. |

File entry:

| Field | Type | Required | Notes |
| --- | --- | --- | --- |
| `path` | string, 1-255 | yes | Relative to `files/`, `/`-separated. No absolute paths, no `.`/`..`/empty segments, no backslashes. |
| `kind` | `starter` \| `test` \| `hidden` \| `readonly` | yes | See below. |
| `editable` | boolean | no | Defaults from `kind`: `starter` is editable, everything else is not. |

`kind` decides what the API serves and what the editor allows:

- **`starter`** - the user's editable starting point. At least one is required.
- **`test`** - the visible test file, read-only in the editor. At least one is
  required, or nothing can be graded.
- **`hidden`** - never sent to the browser; injected at run time only.
- **`readonly`** - shown but not editable, e.g. `go.mod`.

A challenge must declare each path once, every declared path must exist under
`files/`, and every file under `files/` must be declared.

`tags` and `estimatedMinutes` are display metadata only - nothing grades or
orders on them. Tags name the Go topics the challenge exercises (`slices`,
`error-handling`); keep them few and reuse existing ones rather than inventing a
synonym.

## `description.md`

Plain Markdown, rendered on the challenge page. Must not be empty. Start with an
`#` heading, then the task, then optional hints.

Link every standard library identifier the task asks for, inline, the first time
the prose names it: `[strings.Fields](https://pkg.go.dev/strings#Fields)`. There
is no references section in a description - the instructions pane is narrow, so
the link belongs where the name is. A challenge whose task uses nothing beyond
`fmt` needs no link; leave it alone rather than padding it. See
[Links](#links) for the host rule.

## `solution/`

At least one file. Loaded as `solutionFiles` and used as a known-passing
submission by the runner and API tests. **Never served to the frontend.**

## Lessons

Learn mode's reading half. Lessons are a **top-level sequence**, not a per-track
appendix: a lesson names the challenges it practises by track *and* challenge, so
one lesson can draw on any track.

```
lessons/
  <NN-lesson-slug>/
    lesson.json
    lesson.md
```

A lesson's slug is its directory name minus the two-digit prefix
(`01-what-go-is` → `what-go-is`), and the prefix is its `order`. Neither is a
field in `lesson.json`, so the URL and the sequence cannot drift.

### `lesson.json`

```json
{
  "title": "What Go is",
  "summary": "Why Go exists, and what a Go program looks like.",
  "published": true,
  "estimatedMinutes": 8,
  "challenges": [
    { "track": "go-basics", "challenge": "hello-gofinity" },
    { "track": "go-basics", "challenge": "greeting-lines" }
  ]
}
```

| Field | Type | Required | Notes |
| --- | --- | --- | --- |
| `title` | string, 1-120 | yes | |
| `summary` | string, 1-400 | yes | One line, shown on `/learn`. |
| `published` | boolean | no (`false`) | Unpublished lessons are seeded but hidden. |
| `estimatedMinutes` | integer 1-600 | no | Rough time to read. Omit it rather than guessing. |
| `challenges` | array, 1-4 | yes | `{ track, challenge }` slug pairs, in attempt order, no duplicates. |

Every `{ track, challenge }` pair is checked against the loaded tracks: a pair
that does not resolve is an error, and a **published** lesson may only reference
published content, since it is a promise the reader can act on.

A published challenge that no published lesson practises is a **warning**, not an
error, and `content:check` prints it. That is the coverage check that keeps the
reading and the practice halves in step.

### `lesson.md`

Plain Markdown, rendered as the article. Must not be empty, and must start with
an `# ` heading. Close it by pointing at the challenges in `challenges`; the site
renders that block from the JSON, so the prose does not repeat the links.

Every published lesson ends with a `## Further reading` section of 2-5 bullets,
placed **before** the closing `## Practise` paragraph. Each bullet is one link
plus a short clause saying what it is good for, and it points at the spec
section, the package page, or the Effective Go / blog page that is actually
authoritative for what the lesson taught, never at a tutorial. A published
lesson whose body links no allow-listed page is a **warning**, printed by
`content:check`.

Watch the wrapping: a continuation line that starts with `- ` reads as a nested
list item, so a long link title ends with `:` and the gloss follows on the next
line.

## Links

Both `description.md` and `lesson.md` may only link official Go documentation.
`ALLOWED_LINK_HOSTS` in `src/load.ts` is the single place the allowlist lives:

| Rule | Effect |
| --- | --- |
| Scheme must be `https` | An `http://` link is a `ContentError`. |
| Host must be `go.dev` or `pkg.go.dev` | Anything else is a `ContentError` naming the file and the URL. |
| Relative links (`/tracks/...`, `#anchor`) | Allowed, and unchecked. |

`blog.golang.org` only redirects, so use `go.dev/blog/...`. A tutorial on
someone else's site rots and is not on the list.

Links are read out of the prose only: fenced code blocks and inline code spans
are stripped first, so `strings.TrimPrefix(s, "http://")` in a snippet is not a
link. Inline links, reference definitions, angle autolinks and bare URLs all
count.

Do not guess a `pkg.go.dev` fragment. Anchors exist for exported functions,
types, methods (`encoding/json#Decoder.UseNumber`) and the `#pkg-constants`
section; a name declared inside a `const` block has no anchor of its own. When
unsure, link the package page: a less specific link beats a 404.
