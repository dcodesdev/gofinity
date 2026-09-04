# Content schema

The on-disk format for tracks and challenges. Both JSON files are validated with
Zod (`src/schema.ts`) and **reject unknown keys** — an unrecognised field is a
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
| `title` | string, 1–120 | yes | |
| `description` | string, 1–2000 | yes | Shown on the track page and in metadata. |
| `order` | integer ≥ 0 | no (0) | Display order among tracks; must be unique. |
| `published` | boolean | no (`false`) | Unpublished tracks are seeded but hidden. |
| `roadmap` | array, ≥ 1 | yes | The full curriculum, including unwritten challenges. |

Roadmap entry: `slug` (slug), `title` (1–120), `summary` (1–400), `status` —
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
| `title` | string, 1–120 | yes | |
| `difficulty` | `easy` \| `medium` \| `hard` | yes | Badge only; ordering uses `order`. |
| `order` | integer 0–99 | yes | Must equal the directory's `NN` prefix. |
| `published` | boolean | no (`false`) | |
| `tags` | array of slugs, ≤ 5 | no (`[]`) | Topic chips. Lowercase hyphenated, 1–24 chars, no duplicates. |
| `estimatedMinutes` | integer 1–600 | no | Rough time to solve. Omit it rather than guessing. |
| `files` | array, ≥ 1 | yes | Declares every file in `files/`, in tab order. |

File entry:

| Field | Type | Required | Notes |
| --- | --- | --- | --- |
| `path` | string, 1–255 | yes | Relative to `files/`, `/`-separated. No absolute paths, no `.`/`..`/empty segments, no backslashes. |
| `kind` | `starter` \| `test` \| `hidden` \| `readonly` | yes | See below. |
| `editable` | boolean | no | Defaults from `kind`: `starter` is editable, everything else is not. |

`kind` decides what the API serves and what the editor allows:

- **`starter`** — the user's editable starting point. At least one is required.
- **`test`** — the visible test file, read-only in the editor. At least one is
  required, or nothing can be graded.
- **`hidden`** — never sent to the browser; injected at run time only.
- **`readonly`** — shown but not editable, e.g. `go.mod`.

A challenge must declare each path once, every declared path must exist under
`files/`, and every file under `files/` must be declared.

`tags` and `estimatedMinutes` are display metadata only — nothing grades or
orders on them. Tags name the Go topics the challenge exercises (`slices`,
`error-handling`); keep them few and reuse existing ones rather than inventing a
synonym.

## `description.md`

Plain Markdown, rendered on the challenge page. Must not be empty. Start with an
`#` heading, then the task, then optional hints.

## `solution/`

At least one file. Loaded as `solutionFiles` and used as a known-passing
submission by the runner and API tests. **Never served to the frontend.**
