import { z } from "zod"

/**
 * Zod schemas for the on-disk content format.
 *
 * These are the contract external contributors write against. They are
 * deliberately strict — an unknown key is a typo, not an extension point, and
 * catching it here is far cheaper than discovering it after a seed run.
 *
 * Slugs are *not* part of either JSON file: a track's slug is its directory
 * name and a challenge's slug is its directory name minus the `NN-` ordering
 * prefix, so the two can never disagree.
 */

/** URL-safe, lowercase, hyphen-separated. Used verbatim in `/tracks/<slug>`. */
export const slugSchema = z
  .string()
  .min(1)
  .max(64)
  .regex(/^[a-z0-9]+(?:-[a-z0-9]+)*$/, "must be a lowercase hyphen-separated slug")

/** `01-hello-gofinity` → order `1`, slug `hello-gofinity`. */
export const challengeDirNameSchema = z
  .string()
  .regex(/^\d{2}-[a-z0-9]+(?:-[a-z0-9]+)*$/, "must look like `NN-challenge-slug`")

export const difficultySchema = z.enum(["easy", "medium", "hard"])
export type Difficulty = z.infer<typeof difficultySchema>

/**
 * A challenge tag: a short lowercase slug like `strings` or `error-handling`.
 * Same shape as a slug but capped shorter, because tags are rendered as chips
 * and a long one breaks the row.
 */
export const challengeTagSchema = z
  .string()
  .min(1)
  .max(24)
  .regex(/^[a-z0-9]+(?:-[a-z0-9]+)*$/, "must be a lowercase hyphen-separated slug")

export const challengeFileKindSchema = z.enum(["starter", "test", "hidden", "readonly"])
export type ChallengeFileKind = z.infer<typeof challengeFileKindSchema>

/**
 * A workspace-relative path. Anything that could escape the runner's scratch
 * directory is rejected here, before it ever reaches a container.
 */
export const filePathSchema = z
  .string()
  .min(1)
  .max(255)
  .refine((p) => !p.startsWith("/"), "must be relative, not absolute")
  .refine((p) => !/^[a-zA-Z]:[\\/]/.test(p), "must be relative, not absolute")
  .refine((p) => !p.includes("\\"), "must use `/` as the path separator")
  .refine(
    (p) => !p.split("/").some((segment) => segment === "" || segment === "." || segment === ".."),
    "must not contain empty, `.` or `..` segments",
  )

/** One entry in a track's roadmap: the curriculum, including unwritten parts. */
export const roadmapEntrySchema = z.strictObject({
  slug: slugSchema,
  title: z.string().min(1).max(120),
  summary: z.string().min(1).max(400),
  /** `available` means a challenge directory exists; `planned` means it does not. */
  status: z.enum(["available", "planned"]),
})
export type RoadmapEntry = z.infer<typeof roadmapEntrySchema>

/** `tracks/<track-slug>/track.json`. */
export const trackJsonSchema = z.strictObject({
  title: z.string().min(1).max(120),
  description: z.string().min(1).max(2000),
  order: z.int().min(0).default(0),
  published: z.boolean().default(false),
  roadmap: z.array(roadmapEntrySchema).min(1),
})
export type TrackJson = z.infer<typeof trackJsonSchema>

/** One declared file of a challenge workspace, relative to the `files/` dir. */
export const challengeFileEntrySchema = z.strictObject({
  path: filePathSchema,
  kind: challengeFileKindSchema,
  /**
   * Whether the editor lets the user change it. Omitted means "whatever the
   * kind implies": `starter` files are editable, everything else is not.
   */
  editable: z.boolean().optional(),
})
export type ChallengeFileEntry = z.infer<typeof challengeFileEntrySchema>

/** `tracks/<track-slug>/challenges/<NN-slug>/challenge.json`. */
export const challengeJsonSchema = z
  .strictObject({
    title: z.string().min(1).max(120),
    difficulty: difficultySchema,
    /** Must match the `NN-` prefix of the directory name. */
    order: z.int().min(0).max(99),
    published: z.boolean().default(false),
    /** Topic chips, shown on the track rows and in the workspace. */
    tags: z.array(challengeTagSchema).max(5).default([]),
    /** Rough time to solve, in minutes. Omitted means "no estimate". */
    estimatedMinutes: z.int().min(1).max(600).optional(),
    files: z.array(challengeFileEntrySchema).min(1),
  })
  .refine((c) => new Set(c.tags).size === c.tags.length, "declares the same tag twice")
  .refine(
    (c) => new Set(c.files.map((f) => f.path)).size === c.files.length,
    "declares the same path twice",
  )
  .refine(
    (c) => c.files.some((f) => f.kind === "starter"),
    "must declare at least one `starter` file for the user to edit",
  )
  .refine(
    (c) => c.files.some((f) => f.kind === "test"),
    "must declare at least one `test` file, or nothing can be graded",
  )
export type ChallengeJson = z.infer<typeof challengeJsonSchema>

/** `starter` files are the only ones a user may change unless told otherwise. */
export function isEditableByDefault(kind: ChallengeFileKind): boolean {
  return kind === "starter"
}
