import { readdirSync, readFileSync, statSync } from "node:fs"
import { join, relative, resolve, sep } from "node:path"
import {
  type ChallengeFileKind,
  challengeDirNameSchema,
  challengeJsonSchema,
  type Difficulty,
  isEditableByDefault,
  type RoadmapEntry,
  slugSchema,
  trackJsonSchema,
} from "./schema.ts"

/** Absolute path of the `tracks/` directory shipped with this package. */
export const contentRoot = resolve(import.meta.dir, "..", "tracks")

/** A challenge file with its content read off disk. */
export interface LoadedChallengeFile {
  /** Workspace-relative path, e.g. `main.go`. */
  path: string
  content: string
  kind: ChallengeFileKind
  editable: boolean
  /** Tab order in the editor; the order the file was declared in. */
  order: number
}

export interface LoadedChallenge {
  /** Directory name minus the `NN-` prefix. */
  slug: string
  /** The directory name itself, e.g. `01-hello-gofinity`. */
  dirName: string
  title: string
  difficulty: Difficulty
  order: number
  published: boolean
  descriptionMd: string
  files: LoadedChallengeFile[]
  /** Reference solution files. Fixtures for tests — never served to a browser. */
  solutionFiles: { path: string; content: string }[]
}

export interface LoadedTrack {
  slug: string
  title: string
  description: string
  order: number
  published: boolean
  roadmap: RoadmapEntry[]
  challenges: LoadedChallenge[]
}

/** Thrown for any malformed content, always naming the offending path. */
export class ContentError extends Error {
  constructor(
    readonly path: string,
    message: string,
  ) {
    super(`${path}: ${message}`)
    this.name = "ContentError"
  }
}

function listDirs(dir: string): string[] {
  return readdirSync(dir, { withFileTypes: true })
    .filter((e) => e.isDirectory())
    .map((e) => e.name)
    .sort()
}

/** Every file under `dir`, as `/`-separated paths relative to it. */
function listFilesRecursive(dir: string): string[] {
  const out: string[] = []
  const walk = (current: string) => {
    for (const entry of readdirSync(current, { withFileTypes: true }).sort((a, b) =>
      a.name.localeCompare(b.name),
    )) {
      const full = join(current, entry.name)
      if (entry.isDirectory()) walk(full)
      else if (entry.isFile()) out.push(relative(dir, full).split(sep).join("/"))
    }
  }
  walk(dir)
  return out
}

function exists(path: string): boolean {
  try {
    statSync(path)
    return true
  } catch {
    return false
  }
}

function readJson(path: string): unknown {
  let raw: string
  try {
    raw = readFileSync(path, "utf8")
  } catch {
    throw new ContentError(path, "is missing")
  }
  try {
    return JSON.parse(raw)
  } catch (error) {
    throw new ContentError(path, `is not valid JSON (${(error as Error).message})`)
  }
}

function formatIssues(error: unknown): string {
  const issues = (error as { issues?: { path: PropertyKey[]; message: string }[] }).issues
  if (!issues) return String(error)
  return issues
    .map((i) => `${i.path.length > 0 ? i.path.join(".") : "<root>"}: ${i.message}`)
    .join("; ")
}

function loadChallenge(challengeDir: string, dirName: string): LoadedChallenge {
  const jsonPath = join(challengeDir, "challenge.json")
  const parsedName = challengeDirNameSchema.safeParse(dirName)
  if (!parsedName.success) {
    throw new ContentError(challengeDir, formatIssues(parsedName.error))
  }
  const prefix = dirName.slice(0, 2)
  const slug = dirName.slice(3)

  const parsed = challengeJsonSchema.safeParse(readJson(jsonPath))
  if (!parsed.success) throw new ContentError(jsonPath, formatIssues(parsed.error))
  const meta = parsed.data

  if (meta.order !== Number(prefix)) {
    throw new ContentError(
      jsonPath,
      `declares order ${meta.order} but the directory prefix is \`${prefix}\``,
    )
  }

  const descriptionPath = join(challengeDir, "description.md")
  if (!exists(descriptionPath)) throw new ContentError(descriptionPath, "is missing")
  const descriptionMd = readFileSync(descriptionPath, "utf8")
  if (descriptionMd.trim() === "") throw new ContentError(descriptionPath, "is empty")

  const filesDir = join(challengeDir, "files")
  if (!exists(filesDir)) throw new ContentError(filesDir, "is missing")
  const onDisk = new Set(listFilesRecursive(filesDir))

  const files: LoadedChallengeFile[] = meta.files.map((entry, index) => {
    if (!onDisk.has(entry.path)) {
      throw new ContentError(jsonPath, `declares \`${entry.path}\`, which is not in \`files/\``)
    }
    onDisk.delete(entry.path)
    return {
      path: entry.path,
      content: readFileSync(join(filesDir, entry.path), "utf8"),
      kind: entry.kind,
      editable: entry.editable ?? isEditableByDefault(entry.kind),
      order: index,
    }
  })

  if (onDisk.size > 0) {
    throw new ContentError(
      filesDir,
      `contains undeclared file(s): ${[...onDisk].sort().join(", ")} — add them to challenge.json or delete them`,
    )
  }

  const solutionDir = join(challengeDir, "solution")
  if (!exists(solutionDir)) throw new ContentError(solutionDir, "is missing")
  const solutionFiles = listFilesRecursive(solutionDir).map((path) => ({
    path,
    content: readFileSync(join(solutionDir, path), "utf8"),
  }))
  if (solutionFiles.length === 0) throw new ContentError(solutionDir, "is empty")

  return {
    slug,
    dirName,
    title: meta.title,
    difficulty: meta.difficulty,
    order: meta.order,
    published: meta.published,
    descriptionMd,
    files,
    solutionFiles,
  }
}

function loadTrack(trackDir: string, slug: string): LoadedTrack {
  const parsedSlug = slugSchema.safeParse(slug)
  if (!parsedSlug.success) throw new ContentError(trackDir, formatIssues(parsedSlug.error))

  const jsonPath = join(trackDir, "track.json")
  const parsed = trackJsonSchema.safeParse(readJson(jsonPath))
  if (!parsed.success) throw new ContentError(jsonPath, formatIssues(parsed.error))
  const meta = parsed.data

  const challengesDir = join(trackDir, "challenges")
  const challengeDirs = exists(challengesDir) ? listDirs(challengesDir) : []
  const challenges = challengeDirs.map((dirName) =>
    loadChallenge(join(challengesDir, dirName), dirName),
  )

  const orders = challenges.map((c) => c.order)
  if (new Set(orders).size !== orders.length) {
    throw new ContentError(challengesDir, "has two challenges with the same order")
  }

  // The roadmap is the curriculum, so it must describe every challenge that
  // exists and must not claim one is `available` when it has not been written.
  const bySlug = new Map(challenges.map((c) => [c.slug, c]))
  for (const entry of meta.roadmap) {
    const written = bySlug.has(entry.slug)
    if (entry.status === "available" && !written) {
      throw new ContentError(
        jsonPath,
        `roadmap marks \`${entry.slug}\` available but it has no directory`,
      )
    }
    if (entry.status === "planned" && written) {
      throw new ContentError(
        jsonPath,
        `roadmap marks \`${entry.slug}\` planned but it is implemented`,
      )
    }
  }
  const inRoadmap = new Set(meta.roadmap.map((e) => e.slug))
  for (const challenge of challenges) {
    if (!inRoadmap.has(challenge.slug)) {
      throw new ContentError(
        jsonPath,
        `roadmap is missing the implemented challenge \`${challenge.slug}\``,
      )
    }
  }

  return {
    slug,
    title: meta.title,
    description: meta.description,
    order: meta.order,
    published: meta.published,
    roadmap: meta.roadmap,
    challenges,
  }
}

/**
 * Read and validate the whole content tree. Throws {@link ContentError} on the
 * first problem — the seed script and CI both rely on that being loud.
 */
export function loadContent(root: string = contentRoot): LoadedTrack[] {
  if (!exists(root)) throw new ContentError(root, "is missing")
  const tracks = listDirs(root).map((slug) => loadTrack(join(root, slug), slug))
  const orders = tracks.map((t) => t.order)
  if (new Set(orders).size !== orders.length) {
    throw new ContentError(root, "has two tracks with the same order")
  }
  return tracks
}
