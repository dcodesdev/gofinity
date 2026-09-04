import { afterEach, describe, expect, test } from "bun:test"
import { cpSync, mkdirSync, mkdtempSync, readFileSync, rmSync, writeFileSync } from "node:fs"
import { tmpdir } from "node:os"
import { join } from "node:path"
import { ContentError, contentRoot, loadContent } from "./load.ts"

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

  test("documents the whole curriculum in the roadmap, with one challenge written", () => {
    const track = tracks.find((t) => t.slug === "go-basics")
    expect(track?.roadmap.length).toBeGreaterThanOrEqual(10)
    expect(track?.roadmap.filter((e) => e.status === "available")).toHaveLength(1)
    expect(track?.challenges).toHaveLength(1)
  })

  test("exposes 01-hello-gofinity with an editable starter and read-only tests", () => {
    const challenge = tracks[0]?.challenges[0]
    expect(challenge?.slug).toBe("hello-gofinity")
    expect(challenge?.order).toBe(1)
    expect(challenge?.difficulty).toBe("easy")
    expect(challenge?.descriptionMd).toContain("Hello, Gofinity")

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
    const roadmap = (meta.roadmap as { slug: string; status: string }[]).map((e) => ({
      ...e,
      status: "available",
    }))
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
