import { describe, expect, test } from "bun:test"
import {
  challengeDirNameSchema,
  challengeJsonSchema,
  filePathSchema,
  isEditableByDefault,
  lessonDirNameSchema,
  lessonJsonSchema,
  roadmapEntrySchema,
  slugSchema,
  trackJsonSchema,
} from "./schema.ts"

const validTrack = {
  title: "Go Basics",
  description: "Zero to hero in Go.",
  order: 0,
  published: true,
  roadmap: [
    {
      slug: "hello-gofinity",
      title: "Hello, Gofinity",
      summary: "Your first Go program.",
      status: "available",
    },
  ],
}

const validChallenge = {
  title: "Hello, Gofinity",
  difficulty: "easy",
  order: 1,
  published: true,
  files: [
    { path: "main.go", kind: "starter" },
    { path: "main_test.go", kind: "test" },
  ],
}

/** The first issue message, for asserting on *why* something was rejected. */
function reason(result: { success: boolean; error?: { issues: { message: string }[] } }): string {
  return result.error?.issues.map((i) => i.message).join(" | ") ?? ""
}

describe("slugSchema", () => {
  test("accepts lowercase hyphen-separated slugs", () => {
    for (const slug of ["go", "go-basics", "hello-gofinity", "go2-generics"]) {
      expect(slugSchema.safeParse(slug).success).toBe(true)
    }
  })

  test("rejects uppercase, spaces, underscores, and stray hyphens", () => {
    for (const slug of ["Go-Basics", "go basics", "go_basics", "-go", "go-", "go--basics", ""]) {
      expect(slugSchema.safeParse(slug).success).toBe(false)
    }
  })
})

describe("challengeDirNameSchema", () => {
  test("accepts a two-digit prefix followed by a slug", () => {
    expect(challengeDirNameSchema.safeParse("01-hello-gofinity").success).toBe(true)
    expect(challengeDirNameSchema.safeParse("99-generics").success).toBe(true)
  })

  test("rejects a missing, short, or malformed prefix", () => {
    for (const name of ["hello-gofinity", "1-hello", "001-hello", "01hello", "01-"]) {
      expect(challengeDirNameSchema.safeParse(name).success).toBe(false)
    }
  })
})

describe("filePathSchema", () => {
  test("accepts workspace-relative paths, including nested ones", () => {
    expect(filePathSchema.safeParse("main.go").success).toBe(true)
    expect(filePathSchema.safeParse("internal/greet/greet.go").success).toBe(true)
  })

  test("rejects absolute paths", () => {
    expect(reason(filePathSchema.safeParse("/etc/passwd"))).toContain("absolute")
    expect(reason(filePathSchema.safeParse("C:/windows/system32"))).toContain("absolute")
  })

  test("rejects traversal and dot segments", () => {
    expect(reason(filePathSchema.safeParse("../main.go"))).toContain("`..`")
    expect(reason(filePathSchema.safeParse("a/../../b.go"))).toContain("`..`")
    expect(reason(filePathSchema.safeParse("./main.go"))).toContain("`.`")
    expect(reason(filePathSchema.safeParse("a//b.go"))).toContain("empty")
  })

  test("rejects backslash separators, so a Windows path cannot sneak through", () => {
    expect(reason(filePathSchema.safeParse("a\\..\\b.go"))).toContain("separator")
  })

  test("rejects the empty path and absurdly long ones", () => {
    expect(filePathSchema.safeParse("").success).toBe(false)
    expect(filePathSchema.safeParse(`${"a".repeat(256)}.go`).success).toBe(false)
  })
})

describe("trackJsonSchema", () => {
  test("accepts a valid track and defaults order and published", () => {
    const parsed = trackJsonSchema.parse({
      title: validTrack.title,
      description: validTrack.description,
      roadmap: validTrack.roadmap,
    })
    expect(parsed.order).toBe(0)
    expect(parsed.published).toBe(false)
  })

  test("rejects unknown keys, so a typo never silently does nothing", () => {
    expect(trackJsonSchema.safeParse({ ...validTrack, slug: "go-basics" }).success).toBe(false)
  })

  test("rejects a missing title, an empty description, and a non-integer order", () => {
    expect(trackJsonSchema.safeParse({ ...validTrack, title: undefined }).success).toBe(false)
    expect(trackJsonSchema.safeParse({ ...validTrack, description: "" }).success).toBe(false)
    expect(trackJsonSchema.safeParse({ ...validTrack, order: 1.5 }).success).toBe(false)
    expect(trackJsonSchema.safeParse({ ...validTrack, order: -1 }).success).toBe(false)
  })

  test("requires a non-empty roadmap with a known status", () => {
    expect(trackJsonSchema.safeParse({ ...validTrack, roadmap: [] }).success).toBe(false)
    expect(
      roadmapEntrySchema.safeParse({ ...validTrack.roadmap[0], status: "someday" }).success,
    ).toBe(false)
  })
})

describe("challengeJsonSchema", () => {
  test("accepts a valid challenge and defaults published", () => {
    const parsed = challengeJsonSchema.parse({
      title: validChallenge.title,
      difficulty: "easy",
      order: 1,
      files: validChallenge.files,
    })
    expect(parsed.published).toBe(false)
    expect(parsed.files[0]?.editable).toBeUndefined()
    expect(parsed.tags).toEqual([])
    expect(parsed.estimatedMinutes).toBeUndefined()
  })

  test("accepts tags and an estimated time", () => {
    const parsed = challengeJsonSchema.parse({
      ...validChallenge,
      tags: ["strings", "error-handling"],
      estimatedMinutes: 15,
    })
    expect(parsed.tags).toEqual(["strings", "error-handling"])
    expect(parsed.estimatedMinutes).toBe(15)
  })

  test("rejects malformed, duplicated or excessive tags", () => {
    expect(challengeJsonSchema.safeParse({ ...validChallenge, tags: ["Strings"] }).success).toBe(
      false,
    )
    expect(challengeJsonSchema.safeParse({ ...validChallenge, tags: [""] }).success).toBe(false)
    expect(
      reason(challengeJsonSchema.safeParse({ ...validChallenge, tags: ["go", "go"] })),
    ).toContain("same tag twice")
    expect(
      challengeJsonSchema.safeParse({ ...validChallenge, tags: ["a", "b", "c", "d", "e", "f"] })
        .success,
    ).toBe(false)
  })

  test("rejects a non-positive or fractional estimated time", () => {
    expect(challengeJsonSchema.safeParse({ ...validChallenge, estimatedMinutes: 0 }).success).toBe(
      false,
    )
    expect(
      challengeJsonSchema.safeParse({ ...validChallenge, estimatedMinutes: 12.5 }).success,
    ).toBe(false)
    expect(
      challengeJsonSchema.safeParse({ ...validChallenge, estimatedMinutes: 601 }).success,
    ).toBe(false)
  })

  test("rejects an unknown difficulty and an unknown file kind", () => {
    expect(challengeJsonSchema.safeParse({ ...validChallenge, difficulty: "brutal" }).success).toBe(
      false,
    )
    expect(
      challengeJsonSchema.safeParse({
        ...validChallenge,
        files: [{ path: "main.go", kind: "solution" }],
      }).success,
    ).toBe(false)
  })

  test("rejects an order outside the two-digit directory prefix range", () => {
    expect(challengeJsonSchema.safeParse({ ...validChallenge, order: 100 }).success).toBe(false)
    expect(challengeJsonSchema.safeParse({ ...validChallenge, order: -1 }).success).toBe(false)
  })

  test("rejects a duplicated path", () => {
    const result = challengeJsonSchema.safeParse({
      ...validChallenge,
      files: [...validChallenge.files, { path: "main.go", kind: "readonly" }],
    })
    expect(reason(result)).toContain("same path twice")
  })

  test("requires at least one starter file and at least one test file", () => {
    expect(
      reason(
        challengeJsonSchema.safeParse({
          ...validChallenge,
          files: [{ path: "main_test.go", kind: "test" }],
        }),
      ),
    ).toContain("starter")
    expect(
      reason(
        challengeJsonSchema.safeParse({
          ...validChallenge,
          files: [{ path: "main.go", kind: "starter" }],
        }),
      ),
    ).toContain("test")
  })

  test("rejects an empty file list and unknown keys", () => {
    expect(challengeJsonSchema.safeParse({ ...validChallenge, files: [] }).success).toBe(false)
    expect(challengeJsonSchema.safeParse({ ...validChallenge, slug: "hello" }).success).toBe(false)
  })

  test("rejects a traversing file path", () => {
    expect(
      challengeJsonSchema.safeParse({
        ...validChallenge,
        files: [
          { path: "../../../etc/passwd", kind: "starter" },
          { path: "main_test.go", kind: "test" },
        ],
      }).success,
    ).toBe(false)
  })
})

describe("isEditableByDefault", () => {
  test("only starter files are editable by default", () => {
    expect(isEditableByDefault("starter")).toBe(true)
    expect(isEditableByDefault("test")).toBe(false)
    expect(isEditableByDefault("hidden")).toBe(false)
    expect(isEditableByDefault("readonly")).toBe(false)
  })
})

const validLesson = {
  title: "What Go is",
  summary: "Why Go exists and what a Go program looks like.",
  published: true,
  estimatedMinutes: 8,
  challenges: [
    { track: "go-basics", challenge: "hello-gofinity" },
    { track: "go-basics", challenge: "greeting-lines" },
  ],
}

describe("lessonDirNameSchema", () => {
  test("accepts an NN- prefixed slug", () => {
    expect(lessonDirNameSchema.safeParse("01-what-go-is").success).toBe(true)
    expect(lessonDirNameSchema.safeParse("22-capstone").success).toBe(true)
  })

  test("rejects a missing prefix, a one-digit prefix and a non-slug name", () => {
    expect(lessonDirNameSchema.safeParse("what-go-is").success).toBe(false)
    expect(lessonDirNameSchema.safeParse("1-what-go-is").success).toBe(false)
    expect(lessonDirNameSchema.safeParse("01-What_Go_Is").success).toBe(false)
  })
})

describe("lessonJsonSchema", () => {
  test("accepts a well-formed lesson", () => {
    expect(lessonJsonSchema.safeParse(validLesson).success).toBe(true)
  })

  test("defaults published to false and leaves the estimate optional", () => {
    const { published: _p, estimatedMinutes: _e, ...rest } = validLesson
    const parsed = lessonJsonSchema.safeParse(rest)
    expect(parsed.success).toBe(true)
    expect(parsed.data?.published).toBe(false)
    expect(parsed.data?.estimatedMinutes).toBeUndefined()
  })

  test("rejects a slug or an order field, both come from the directory name", () => {
    expect(lessonJsonSchema.safeParse({ ...validLesson, slug: "what-go-is" }).success).toBe(false)
    expect(lessonJsonSchema.safeParse({ ...validLesson, order: 1 }).success).toBe(false)
  })

  test("requires between one and four challenges", () => {
    expect(lessonJsonSchema.safeParse({ ...validLesson, challenges: [] }).success).toBe(false)
    expect(
      lessonJsonSchema.safeParse({
        ...validLesson,
        challenges: [1, 2, 3, 4, 5].map((n) => ({ track: "go-basics", challenge: `c-${n}` })),
      }).success,
    ).toBe(false)
  })

  test("rejects the same challenge twice", () => {
    const result = lessonJsonSchema.safeParse({
      ...validLesson,
      challenges: [validLesson.challenges[0], validLesson.challenges[0]],
    })
    expect(result.success).toBe(false)
    expect(reason(result)).toContain("same challenge twice")
  })

  test("rejects a reference that is not a pair of slugs", () => {
    expect(
      lessonJsonSchema.safeParse({ ...validLesson, challenges: [{ challenge: "hello" }] }).success,
    ).toBe(false)
    expect(
      lessonJsonSchema.safeParse({
        ...validLesson,
        challenges: [{ track: "Go Basics", challenge: "hello" }],
      }).success,
    ).toBe(false)
  })
})
