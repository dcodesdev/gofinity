import { readdirSync, readFileSync, statSync } from "node:fs"
import { join, relative, resolve, sep } from "node:path"
import { fileURLToPath } from "node:url"
import {
  type ChallengeFileKind,
  challengeDirNameSchema,
  challengeJsonSchema,
  type Difficulty,
  isEditableByDefault,
  type LessonChallengeRef,
  lessonDirNameSchema,
  lessonJsonSchema,
  type RoadmapEntry,
  slugSchema,
  trackJsonSchema,
} from "./schema.ts"

/**
 * Absolute path of this package, the parent of `tracks/` and `lessons/`.
 *
 * `import.meta.url` rather than Bun's `import.meta.dir`, because this module is
 * also bundled into a Node build by consumers; `import.meta.dir` is `undefined`
 * there and the path resolution would throw at import time. A bundled consumer
 * cannot rely on this default pointing anywhere useful and passes an explicit
 * root to `loadAll` instead, but it must still be able to import the module.
 */
export const packageRoot = resolve(fileURLToPath(import.meta.url), "../..")

/** Absolute path of the `tracks/` directory shipped with this package. */
export const contentRoot = join(packageRoot, "tracks")

/** Absolute path of the `lessons/` directory shipped with this package. */
export const lessonsRoot = join(packageRoot, "lessons")

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
  /** Topic chips, in the order the content declares them. */
  tags: string[]
  /** Rough time to solve, in minutes, or `null` when the content omits it. */
  estimatedMinutes: number | null
  descriptionMd: string
  files: LoadedChallengeFile[]
  /** Reference solution files. Fixtures for tests - never served to a browser. */
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
  assertLinksAllowed(descriptionPath, descriptionMd)

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
      `contains undeclared file(s): ${[...onDisk].sort().join(", ")} - add them to challenge.json or delete them`,
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
    tags: meta.tags,
    estimatedMinutes: meta.estimatedMinutes ?? null,
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
 * first problem - the seed script and CI both rely on that being loud.
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

export interface LoadedLesson {
  /** Directory name minus the `NN-` prefix. Used verbatim in `/learn/<slug>`. */
  slug: string
  /** The directory name itself, e.g. `01-what-go-is`. */
  dirName: string
  title: string
  summary: string
  /** The `NN-` prefix, as a number. Unique across the sequence. */
  order: number
  published: boolean
  /** Rough time to read, in minutes, or `null` when the content omits it. */
  estimatedMinutes: number | null
  bodyMd: string
  /** The challenges this lesson drills, in attempt order. */
  challenges: LessonChallengeRef[]
}

/** Everything the tree declares: the tracks, the lesson sequence, and the soft warnings. */
export interface LoadedContent {
  tracks: LoadedTrack[]
  lessons: LoadedLesson[]
  /**
   * Problems that are worth saying out loud but must not fail a build, e.g. a
   * published challenge no lesson practises. `content:check` prints them.
   */
  warnings: string[]
}

function loadLesson(lessonDir: string, dirName: string): LoadedLesson {
  const parsedName = lessonDirNameSchema.safeParse(dirName)
  if (!parsedName.success) throw new ContentError(lessonDir, formatIssues(parsedName.error))
  const order = Number(dirName.slice(0, 2))
  const slug = dirName.slice(3)

  const jsonPath = join(lessonDir, "lesson.json")
  const parsed = lessonJsonSchema.safeParse(readJson(jsonPath))
  if (!parsed.success) throw new ContentError(jsonPath, formatIssues(parsed.error))
  const meta = parsed.data

  const bodyPath = join(lessonDir, "lesson.md")
  if (!exists(bodyPath)) throw new ContentError(bodyPath, "is missing")
  const bodyMd = readFileSync(bodyPath, "utf8")
  if (bodyMd.trim() === "") throw new ContentError(bodyPath, "is empty")
  const firstLine = bodyMd.trim().split("\n", 1)[0] ?? ""
  if (!/^#\s+\S/.test(firstLine)) {
    throw new ContentError(bodyPath, "must start with an `# ` heading")
  }
  assertLinksAllowed(bodyPath, bodyMd)

  return {
    slug,
    dirName,
    title: meta.title,
    summary: meta.summary,
    order,
    published: meta.published,
    estimatedMinutes: meta.estimatedMinutes ?? null,
    bodyMd,
    challenges: meta.challenges,
  }
}

/**
 * Read and validate the lesson sequence, checking every `{ track, challenge }`
 * reference against `tracks`. A missing `lessons/` directory is zero lessons,
 * not an error: the tree is valid before any lesson has been written.
 */
/** The `<track>/<challenge>` key a lesson reference and a challenge share. */
function refKey(track: string, challenge: string): string {
  return `${track}/${challenge}`
}

export function loadLessons(tracks: LoadedTrack[], root: string = lessonsRoot): LoadedLesson[] {
  if (!exists(root)) return []
  const lessons = listDirs(root).map((dirName) => loadLesson(join(root, dirName), dirName))

  const orders = lessons.map((l) => l.order)
  if (new Set(orders).size !== orders.length) {
    throw new ContentError(root, "has two lessons with the same order")
  }
  const slugs = lessons.map((l) => l.slug)
  if (new Set(slugs).size !== slugs.length) {
    throw new ContentError(root, "has two lessons with the same slug")
  }

  const challengesByRef = new Map(
    tracks.flatMap((track) =>
      track.challenges.map(
        (challenge) => [refKey(track.slug, challenge.slug), { track, challenge }] as const,
      ),
    ),
  )
  for (const lesson of lessons) {
    const jsonPath = join(root, lesson.dirName, "lesson.json")
    for (const ref of lesson.challenges) {
      const key = refKey(ref.track, ref.challenge)
      const found = challengesByRef.get(key)
      if (!found) {
        throw new ContentError(jsonPath, `references \`${key}\`, which does not exist`)
      }
      // A published lesson is a promise the reader can act on, so it may only
      // send them somewhere they can actually go.
      if (lesson.published && !(found.track.published && found.challenge.published)) {
        throw new ContentError(jsonPath, `is published but references \`${key}\`, which is not`)
      }
    }
  }

  return lessons
}

/**
 * Published challenges no published lesson practises. Not an error, since a
 * challenge may legitimately land before its lesson does, but at this scale
 * it is the only thing that keeps the two halves of the curriculum in step.
 */
export function unpractisedChallenges(tracks: LoadedTrack[], lessons: LoadedLesson[]): string[] {
  const practised = new Set(
    lessons
      .filter((l) => l.published)
      .flatMap((l) => l.challenges.map((c) => `${c.track}/${c.challenge}`)),
  )
  return tracks
    .filter((track) => track.published)
    .flatMap((track) =>
      track.challenges.filter((c) => c.published).map((c) => refKey(track.slug, c.slug)),
    )
    .filter((key) => !practised.has(key))
}

/**
 * Load the whole tree, tracks *and* lessons, from a directory holding
 * `tracks/` and, optionally, `lessons/`. {@link loadContent} stays the
 * tracks-only entry point so existing callers are unaffected.
 */
export function loadAll(root: string = packageRoot): LoadedContent {
  const tracks = loadContent(join(root, "tracks"))
  const lessons = loadLessons(tracks, join(root, "lessons"))
  const warnings = unpractisedChallenges(tracks, lessons).map(
    (key) => `no published lesson practises the published challenge \`${key}\``,
  )
  // A lesson that teaches something should say where the authoritative wording
  // is. A warning, not an error, so the sequence keeps loading while the
  // `## Further reading` sections are still being written.
  for (const lesson of lessons) {
    if (lesson.published && !hasAllowedLink(lesson.bodyMd)) {
      warnings.push(`lesson \`${lesson.dirName}\` links no official Go documentation`)
    }
  }
  return { tracks, lessons, warnings }
}

/**
 * Hosts a content link may point at. Official Go documentation only: the spec,
 * the docs, the blog and the package index. A tutorial on someone else's site
 * rots, and `blog.golang.org` only redirects, so neither is on the list.
 *
 * This is the single place the allowlist lives - `content:check` reads it.
 */
export const ALLOWED_LINK_HOSTS = ["go.dev", "pkg.go.dev"]

/** Drop fenced code blocks, so a URL inside a Go snippet is not a link. */
function stripFencedBlocks(md: string): string {
  const kept: string[] = []
  let fence: string | null = null
  for (const line of md.split("\n")) {
    const opener = /^ {0,3}(`{3,}|~{3,})/.exec(line)
    if (fence !== null) {
      if (opener && line.trim().startsWith(fence)) fence = null
      continue
    }
    if (opener?.[1]) {
      fence = opener[1]
      continue
    }
    kept.push(line)
  }
  return kept.join("\n")
}

/** Drop inline code spans, for the same reason as {@link stripFencedBlocks}. */
function stripInlineCode(md: string): string {
  return md.replace(/(`+)[\s\S]*?\1/g, " ")
}

const INLINE_LINK = /\[[^\]]*\]\(\s*<?([^)\s>]+)>?[^)]*\)/g
const REFERENCE_LINK = /^ {0,3}\[[^\]]+\]:\s*<?([^\s>]+)>?/gm
const BARE_URL = /[a-z][a-z0-9+.-]*:\/\/[^\s<>)\]"'`]+/gi

/**
 * Every URL a markdown body links to: inline links, reference definitions,
 * angle autolinks and the bare URLs GFM turns into links. Code, fenced or
 * inline, is ignored - `strings.TrimPrefix(s, "http://")` is a snippet, not a
 * link.
 */
export function markdownLinkUrls(md: string): string[] {
  const prose = stripInlineCode(stripFencedBlocks(md))
  const urls: string[] = []
  let rest = prose.replace(INLINE_LINK, (_match, url: string) => {
    urls.push(url)
    return " "
  })
  rest = rest.replace(REFERENCE_LINK, (_match, url: string) => {
    urls.push(url)
    return " "
  })
  for (const match of rest.matchAll(BARE_URL)) {
    urls.push(match[0].replace(/[.,;:!?]+$/, ""))
  }
  return urls
}

/** True for a link that stays on this site: `/tracks/...`, `#anchor`, `./x.md`. */
function isRelativeLink(url: string): boolean {
  return !/^[a-z][a-z0-9+.-]*:/i.test(url) && !url.startsWith("//")
}

/**
 * Fail the load for any off-site link that is not official Go documentation
 * over https. Relative links are in-site navigation and stay unchecked.
 */
function assertLinksAllowed(path: string, md: string): void {
  for (const url of markdownLinkUrls(md)) {
    if (isRelativeLink(url)) continue
    let parsed: URL
    try {
      parsed = new URL(url)
    } catch {
      throw new ContentError(path, `links \`${url}\`, which is not a valid URL`)
    }
    if (parsed.protocol !== "https:") {
      throw new ContentError(path, `links \`${url}\` - only https links are allowed`)
    }
    if (!ALLOWED_LINK_HOSTS.includes(parsed.hostname)) {
      throw new ContentError(
        path,
        `links \`${url}\` - only ${ALLOWED_LINK_HOSTS.join(" and ")} are allowed`,
      )
    }
  }
}

/** True when the body links at least one allow-listed documentation page. */
export function hasAllowedLink(md: string): boolean {
  return markdownLinkUrls(md).some((url) => {
    if (isRelativeLink(url)) return false
    try {
      return ALLOWED_LINK_HOSTS.includes(new URL(url).hostname)
    } catch {
      return false
    }
  })
}
