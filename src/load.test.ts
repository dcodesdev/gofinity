import { afterEach, describe, expect, test } from "bun:test"
import {
  cpSync,
  mkdirSync,
  mkdtempSync,
  readdirSync,
  readFileSync,
  rmSync,
  writeFileSync,
} from "node:fs"
import { tmpdir } from "node:os"
import { join } from "node:path"
import { ContentError, contentRoot, loadAll, loadContent, markdownLinkUrls } from "./load.ts"

const temps: string[] = []

afterEach(() => {
  for (const dir of temps.splice(0)) rmSync(dir, { recursive: true, force: true })
})

/** A copy of the real content tree, so a test can corrupt one file safely. */
function fixture(): string {
  const dir = mkdtempSync(join(tmpdir(), "gofinity-content-"))
  temps.push(dir)
  cpSync(contentRoot, join(dir, "tracks"), { recursive: true })
  return join(dir, "tracks")
}

const challengeDir = (root: string) => join(root, "go-basics", "challenges", "01-hello-gofinity")

function writeJson(path: string, value: unknown): void {
  writeFileSync(path, JSON.stringify(value, null, 2))
}

function readJson(path: string): Record<string, unknown> {
  return JSON.parse(readFileSync(path, "utf8"))
}

// This is the guard against a broken contribution: the shipped content must
// always load, so `bun test` fails the moment a track or challenge is invalid.
describe("the real content tree", () => {
  const tracks = loadContent()

  test("validates end to end", () => {
    expect(tracks.length).toBeGreaterThan(0)
  })

  test("ships the published go-basics track", () => {
    const track = tracks.find((t) => t.slug === "go-basics")
    expect(track).toBeDefined()
    expect(track?.published).toBe(true)
    expect(track?.title).toBe("Go Basics")
  })

  test("documents the whole curriculum in the roadmap, with the written ones available", () => {
    const track = tracks.find((t) => t.slug === "go-basics")
    expect(track?.roadmap.length).toBeGreaterThanOrEqual(10)
    expect(track?.challenges.length).toBeGreaterThanOrEqual(1)
    expect(track?.roadmap.filter((e) => e.status === "available")).toHaveLength(
      track?.challenges.length ?? 0,
    )
  })

  test("exposes 01-hello-gofinity with an editable starter and read-only tests", () => {
    const challenge = tracks[0]?.challenges[0]
    expect(challenge?.slug).toBe("hello-gofinity")
    expect(challenge?.order).toBe(1)
    expect(challenge?.difficulty).toBe("easy")
    expect(challenge?.descriptionMd).toContain("Hello, Gofinity")
    expect(challenge?.tags).toEqual(["fmt", "strings", "functions"])
    expect(challenge?.estimatedMinutes).toBe(10)

    const byPath = new Map(challenge?.files.map((f) => [f.path, f]))
    expect(byPath.get("main.go")?.kind).toBe("starter")
    expect(byPath.get("main.go")?.editable).toBe(true)
    expect(byPath.get("main_test.go")?.kind).toBe("test")
    expect(byPath.get("main_test.go")?.editable).toBe(false)
    expect(byPath.get("go.mod")?.kind).toBe("readonly")
    expect(byPath.get("go.mod")?.editable).toBe(false)
    expect(challenge?.files.map((f) => f.order)).toEqual([0, 1, 2])
  })

  test("keeps the reference solution out of the served files", () => {
    const challenge = tracks[0]?.challenges[0]
    expect(challenge?.solutionFiles.map((f) => f.path)).toEqual(["main.go"])
    expect(challenge?.solutionFiles[0]?.content).toContain("Hello, %s!")
    expect(challenge?.files.some((f) => f.content.includes("Hello, %s!"))).toBe(false)
  })

  test("only fills the starter with a TODO, and the tests exercise the empty-name case", () => {
    const files = new Map(tracks[0]?.challenges[0]?.files.map((f) => [f.path, f.content]))
    expect(files.get("main.go")).toContain("TODO")
    expect(files.get("main_test.go")).toContain("Hello, World!")
    expect(files.get("go.mod")).toContain("module ")
  })
})

describe("optional challenge metadata", () => {
  test("defaults to no tags and no estimate when the JSON omits them", () => {
    const root = fixture()
    const path = join(challengeDir(root), "challenge.json")
    const { tags: _tags, estimatedMinutes: _estimate, ...rest } = readJson(path)
    writeJson(path, rest)
    const challenge = loadContent(root)[0]?.challenges[0]
    expect(challenge?.tags).toEqual([])
    expect(challenge?.estimatedMinutes).toBeNull()
  })

  test("rejects a tag that is not a lowercase slug", () => {
    const root = fixture()
    const path = join(challengeDir(root), "challenge.json")
    writeJson(path, { ...readJson(path), tags: ["Not A Slug"] })
    expect(() => loadContent(root)).toThrow(/tags\.0/)
  })
})

describe("loadContent rejections", () => {
  test("a missing root", () => {
    expect(() => loadContent(join(tmpdir(), "definitely-not-here-gofinity"))).toThrow(ContentError)
  })

  test("malformed JSON names the file", () => {
    const root = fixture()
    writeFileSync(join(root, "go-basics", "track.json"), "{ nope")
    expect(() => loadContent(root)).toThrow(/track\.json: is not valid JSON/)
  })

  test("an unknown key in track.json", () => {
    const root = fixture()
    const path = join(root, "go-basics", "track.json")
    writeJson(path, { ...readJson(path), extra: true })
    expect(() => loadContent(root)).toThrow(/track\.json/)
  })

  test("a challenge directory without the NN- prefix", () => {
    const root = fixture()
    const dir = challengeDir(root)
    cpSync(dir, join(root, "go-basics", "challenges", "hello-gofinity"), { recursive: true })
    rmSync(dir, { recursive: true })
    expect(() => loadContent(root)).toThrow(/NN-challenge-slug/)
  })

  test("an order that disagrees with the directory prefix", () => {
    const root = fixture()
    const path = join(challengeDir(root), "challenge.json")
    writeJson(path, { ...readJson(path), order: 7 })
    expect(() => loadContent(root)).toThrow(/declares order 7 but the directory prefix is `01`/)
  })

  test("a declared file that is not on disk", () => {
    const root = fixture()
    const path = join(challengeDir(root), "challenge.json")
    const meta = readJson(path)
    writeJson(path, {
      ...meta,
      files: [...(meta.files as unknown[]), { path: "missing.go", kind: "hidden" }],
    })
    expect(() => loadContent(root)).toThrow(/declares `missing\.go`, which is not in `files\/`/)
  })

  test("a file on disk that nothing declares", () => {
    const root = fixture()
    writeFileSync(join(challengeDir(root), "files", "stray.go"), "package main\n")
    expect(() => loadContent(root)).toThrow(/undeclared file\(s\): stray\.go/)
  })

  test("a missing description.md and a missing solution", () => {
    const withoutDescription = fixture()
    rmSync(join(challengeDir(withoutDescription), "description.md"))
    expect(() => loadContent(withoutDescription)).toThrow(/description\.md: is missing/)

    const withoutSolution = fixture()
    rmSync(join(challengeDir(withoutSolution), "solution"), { recursive: true })
    expect(() => loadContent(withoutSolution)).toThrow(/solution: is missing/)
  })

  test("an empty description.md", () => {
    const root = fixture()
    writeFileSync(join(challengeDir(root), "description.md"), "   \n")
    expect(() => loadContent(root)).toThrow(/description\.md: is empty/)
  })

  test("a roadmap that claims an unwritten challenge is available", () => {
    const root = fixture()
    const path = join(root, "go-basics", "track.json")
    const meta = readJson(path)
    const roadmap = [
      ...(meta.roadmap as { slug: string; status: string }[]),
      {
        slug: "not-written-yet",
        title: "Not Written Yet",
        summary: "A roadmap entry with no challenge directory behind it.",
        status: "available",
      },
    ]
    writeJson(path, { ...meta, roadmap })
    expect(() => loadContent(root)).toThrow(/available but it has no directory/)
  })

  test("a roadmap missing an implemented challenge", () => {
    const root = fixture()
    const path = join(root, "go-basics", "track.json")
    const meta = readJson(path)
    const roadmap = (meta.roadmap as { slug: string }[]).filter((e) => e.slug !== "hello-gofinity")
    writeJson(path, { ...meta, roadmap })
    expect(() => loadContent(root)).toThrow(/missing the implemented challenge `hello-gofinity`/)
  })

  test("a roadmap that calls an implemented challenge planned", () => {
    const root = fixture()
    const path = join(root, "go-basics", "track.json")
    const meta = readJson(path)
    const roadmap = (meta.roadmap as { slug: string; status: string }[]).map((e) =>
      e.slug === "hello-gofinity" ? { ...e, status: "planned" } : e,
    )
    writeJson(path, { ...meta, roadmap })
    expect(() => loadContent(root)).toThrow(/planned but it is implemented/)
  })

  test("two tracks sharing an order", () => {
    const root = fixture()
    cpSync(join(root, "go-basics"), join(root, "go-basics-copy"), { recursive: true })
    expect(() => loadContent(root)).toThrow(/two tracks with the same order/)
  })

  test("a track directory whose name is not a slug", () => {
    const root = fixture()
    cpSync(join(root, "go-basics"), join(root, "Go_Basics"), { recursive: true })
    expect(() => loadContent(root)).toThrow(/lowercase hyphen-separated slug/)
  })

  test("a track with no challenges directory still loads", () => {
    const root = mkdtempSync(join(tmpdir(), "gofinity-empty-"))
    temps.push(root)
    mkdirSync(join(root, "empty-track"))
    writeJson(join(root, "empty-track", "track.json"), {
      title: "Empty",
      description: "Nothing here yet.",
      roadmap: [{ slug: "later", title: "Later", summary: "Not written.", status: "planned" }],
    })
    const tracks = loadContent(root)
    expect(tracks[0]?.challenges).toEqual([])
    expect(tracks[0]?.published).toBe(false)
  })
})

/** A fixture root holding `tracks/` and a written `lessons/`, for `loadAll`. */
/**
 * A copy of the real tracks with every challenge but `hello-gofinity` removed,
 * so the lesson tests read the same however large the curriculum grows. The
 * roadmap entries of the pruned challenges go back to `planned`, because
 * `available` promises a directory.
 */
function oneChallengeTracks(root: string): void {
  const tracksRoot = join(root, "tracks")
  cpSync(contentRoot, tracksRoot, { recursive: true })
  for (const track of readdirSync(tracksRoot)) {
    const challenges = join(tracksRoot, track, "challenges")
    const kept: string[] = []
    for (const dirName of readdirSync(challenges)) {
      if (dirName.replace(/^\d{2}-/, "") === "hello-gofinity") {
        kept.push("hello-gofinity")
        continue
      }
      rmSync(join(challenges, dirName), { recursive: true, force: true })
    }
    const trackPath = join(tracksRoot, track, "track.json")
    const meta = readJson(trackPath)
    const roadmap = (meta.roadmap as { slug: string; status: string }[]).map((entry) => ({
      ...entry,
      status: kept.includes(entry.slug) ? entry.status : "planned",
    }))
    writeJson(trackPath, { ...meta, roadmap })
  }
}

function lessonFixture(challenges?: unknown[]): string {
  const dir = mkdtempSync(join(tmpdir(), "gofinity-lessons-"))
  temps.push(dir)
  oneChallengeTracks(dir)
  writeLesson(dir, "01-what-go-is", {
    title: "What Go is",
    summary: "Why Go exists and what a Go program looks like.",
    published: true,
    estimatedMinutes: 8,
    challenges: challenges ?? [{ track: "go-basics", challenge: "hello-gofinity" }],
  })
  return dir
}

function writeLesson(root: string, dirName: string, meta: unknown, body?: string): string {
  const dir = join(root, "lessons", dirName)
  mkdirSync(dir, { recursive: true })
  writeJson(join(dir, "lesson.json"), meta)
  writeFileSync(join(dir, "lesson.md"), body ?? "# What Go is\n\nGo is a small language.\n")
  return dir
}

describe("loadLessons", () => {
  test("a missing lessons directory is zero lessons, not an error", () => {
    const root = mkdtempSync(join(tmpdir(), "gofinity-nolessons-"))
    temps.push(root)
    cpSync(contentRoot, join(root, "tracks"), { recursive: true })
    const loaded = loadAll(root)
    expect(loaded.lessons).toEqual([])
    expect(loaded.tracks.length).toBeGreaterThan(0)
  })

  test("takes the slug and order from the directory name", () => {
    const lesson = loadAll(lessonFixture()).lessons[0]
    expect(lesson?.slug).toBe("what-go-is")
    expect(lesson?.dirName).toBe("01-what-go-is")
    expect(lesson?.order).toBe(1)
    expect(lesson?.title).toBe("What Go is")
    expect(lesson?.published).toBe(true)
    expect(lesson?.estimatedMinutes).toBe(8)
    expect(lesson?.bodyMd).toContain("Go is a small language.")
    expect(lesson?.challenges).toEqual([{ track: "go-basics", challenge: "hello-gofinity" }])
  })

  test("returns lessons in directory order", () => {
    const root = lessonFixture()
    writeLesson(root, "02-values-and-variables", {
      title: "Values and variables",
      summary: "Declaring, assigning and the zero value.",
      challenges: [{ track: "go-basics", challenge: "hello-gofinity" }],
    })
    expect(loadAll(root).lessons.map((l) => l.order)).toEqual([1, 2])
  })

  test("rejects a lesson directory without the NN- prefix", () => {
    const root = lessonFixture()
    writeLesson(root, "what-go-is", {
      title: "What Go is",
      summary: "Why Go exists.",
      challenges: [{ track: "go-basics", challenge: "hello-gofinity" }],
    })
    expect(() => loadAll(root)).toThrow(/NN-lesson-slug/)
  })

  test("rejects an unknown key in lesson.json", () => {
    const root = lessonFixture()
    const path = join(root, "lessons", "01-what-go-is", "lesson.json")
    writeJson(path, { ...readJson(path), order: 1 })
    expect(() => loadAll(root)).toThrow(ContentError)
  })

  test("rejects a missing, empty or headingless lesson.md", () => {
    const missing = lessonFixture()
    rmSync(join(missing, "lessons", "01-what-go-is", "lesson.md"))
    expect(() => loadAll(missing)).toThrow(/lesson\.md: is missing/)

    const empty = lessonFixture()
    writeFileSync(join(empty, "lessons", "01-what-go-is", "lesson.md"), "  \n")
    expect(() => loadAll(empty)).toThrow(/lesson\.md: is empty/)

    const headingless = lessonFixture()
    writeFileSync(join(headingless, "lessons", "01-what-go-is", "lesson.md"), "Go is small.\n")
    expect(() => loadAll(headingless)).toThrow(/must start with an `# ` heading/)
  })

  test("rejects two lessons sharing an order", () => {
    const root = lessonFixture()
    cpSync(join(root, "lessons", "01-what-go-is"), join(root, "lessons", "01-what-go-is-again"), {
      recursive: true,
    })
    expect(() => loadAll(root)).toThrow(/two lessons with the same order/)
  })

  test("rejects two lessons sharing a slug", () => {
    const root = lessonFixture()
    cpSync(join(root, "lessons", "01-what-go-is"), join(root, "lessons", "02-what-go-is"), {
      recursive: true,
    })
    expect(() => loadAll(root)).toThrow(/two lessons with the same slug/)
  })

  test("rejects a reference to a challenge that does not exist", () => {
    const root = lessonFixture([{ track: "go-basics", challenge: "nope" }])
    expect(() => loadAll(root)).toThrow(/references `go-basics\/nope`, which does not exist/)
  })

  test("rejects a reference to a track that does not exist", () => {
    const root = lessonFixture([{ track: "rust-basics", challenge: "hello-gofinity" }])
    expect(() => loadAll(root)).toThrow(/references `rust-basics\/hello-gofinity`/)
  })

  test("rejects a published lesson pointing at an unpublished challenge", () => {
    const root = lessonFixture()
    const path = join(
      root,
      "tracks",
      "go-basics",
      "challenges",
      "01-hello-gofinity",
      "challenge.json",
    )
    writeJson(path, { ...readJson(path), published: false })
    expect(() => loadAll(root)).toThrow(/is published but references `go-basics\/hello-gofinity`/)
  })

  test("allows an unpublished lesson to point at an unpublished challenge", () => {
    const root = lessonFixture()
    const lessonPath = join(root, "lessons", "01-what-go-is", "lesson.json")
    writeJson(lessonPath, { ...readJson(lessonPath), published: false })
    const challengePath = join(
      root,
      "tracks",
      "go-basics",
      "challenges",
      "01-hello-gofinity",
      "challenge.json",
    )
    writeJson(challengePath, { ...readJson(challengePath), published: false })
    expect(loadAll(root).lessons).toHaveLength(1)
  })
})

describe("unpractisedChallenges", () => {
  test("warns rather than fails when no lesson practises a published challenge", () => {
    const root = mkdtempSync(join(tmpdir(), "gofinity-unpractised-"))
    temps.push(root)
    oneChallengeTracks(root)
    const { warnings } = loadAll(root)
    expect(warnings).toHaveLength(1)
    expect(warnings[0]).toContain("go-basics/hello-gofinity")
  })

  test("is silent once a published lesson practises it", () => {
    const warnings = loadAll(lessonFixture()).warnings.filter((w) => w.includes("practises"))
    expect(warnings).toEqual([])
  })

  test("an unpublished lesson does not count as practice", () => {
    const root = lessonFixture()
    const path = join(root, "lessons", "01-what-go-is", "lesson.json")
    writeJson(path, { ...readJson(path), published: false })
    expect(loadAll(root).warnings).toHaveLength(1)
  })
})

describe("markdownLinkUrls", () => {
  test("finds inline, reference, angle and bare links", () => {
    const md = [
      "See [the spec](https://go.dev/ref/spec#Types) and <https://pkg.go.dev/fmt>.",
      "Bare: https://go.dev/doc/effective_go, with a comma after it.",
      "",
      "[ref]: https://go.dev/blog/slices",
    ].join("\n")
    expect(markdownLinkUrls(md)).toEqual([
      "https://go.dev/ref/spec#Types",
      "https://go.dev/blog/slices",
      "https://pkg.go.dev/fmt",
      "https://go.dev/doc/effective_go",
    ])
  })

  test("ignores URLs inside fenced blocks and inline code", () => {
    const md = [
      'Use `strings.TrimPrefix(s, "http://")` for that.',
      "",
      "```go",
      'strings.TrimPrefix(s, "http://")',
      "// https://example.com/not-a-link",
      "```",
      "",
      "Real: [docs](https://go.dev/doc/).",
    ].join("\n")
    expect(markdownLinkUrls(md)).toEqual(["https://go.dev/doc/"])
  })

  test("keeps relative links, which stay unchecked", () => {
    expect(markdownLinkUrls("[here](/learn/what-go-is) and [there](#types)")).toEqual([
      "/learn/what-go-is",
      "#types",
    ])
  })
})

describe("the link rule", () => {
  const descriptionPath = (root: string) => join(challengeDir(root), "description.md")

  test("rejects an off-host link in a description", () => {
    const root = fixture()
    writeFileSync(descriptionPath(root), "# Hello\n\nSee [SO](https://stackoverflow.com/q/1).\n")
    expect(() => loadContent(root)).toThrow(/stackoverflow\.com/)
  })

  test("rejects an http:// link even on an allowed host", () => {
    const root = fixture()
    writeFileSync(descriptionPath(root), "# Hello\n\nSee [spec](http://go.dev/ref/spec).\n")
    expect(() => loadContent(root)).toThrow(/only https links are allowed/)
  })

  test("accepts a go.dev link", () => {
    const root = fixture()
    writeFileSync(descriptionPath(root), "# Hello\n\nSee [fmt](https://pkg.go.dev/fmt).\n")
    expect(loadContent(root)[0]?.challenges[0]?.descriptionMd).toContain("pkg.go.dev/fmt")
  })

  test("ignores a URL inside a fenced block in a description", () => {
    const root = fixture()
    writeFileSync(
      descriptionPath(root),
      '# Hello\n\n```go\nstrings.TrimPrefix(s, "http://evil.example")\n```\n',
    )
    expect(() => loadContent(root)).not.toThrow()
  })

  test("rejects an off-host link in a lesson body", () => {
    const root = lessonFixture()
    writeFileSync(
      join(root, "lessons", "01-what-go-is", "lesson.md"),
      "# What Go is\n\n[video](https://youtube.com/watch?v=1)\n",
    )
    expect(() => loadAll(root)).toThrow(/youtube\.com/)
  })

  test("warns when a published lesson links no official documentation", () => {
    const warnings = loadAll(lessonFixture()).warnings
    expect(warnings).toEqual([expect.stringContaining("01-what-go-is")])
  })

  test("is silent once the lesson links go.dev", () => {
    const root = lessonFixture()
    writeFileSync(
      join(root, "lessons", "01-what-go-is", "lesson.md"),
      "# What Go is\n\nGo is small. See [the docs](https://go.dev/doc/).\n",
    )
    expect(loadAll(root).warnings).toEqual([])
  })

  test("an unpublished lesson without a link does not warn", () => {
    const root = lessonFixture()
    const path = join(root, "lessons", "01-what-go-is", "lesson.json")
    writeJson(path, { ...readJson(path), published: false })
    expect(loadAll(root).warnings.filter((w) => w.includes("official Go"))).toEqual([])
  })
})
