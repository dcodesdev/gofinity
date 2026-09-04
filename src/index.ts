/**
 * `@gofinity/content` — the open-source challenge content for Gofinity.
 *
 * This package is the source of truth in git; `bun run seed` reads it with
 * {@link loadContent} and upserts it into Postgres. It has no dependencies on
 * any other workspace package, because it is destined to be extracted into a
 * standalone public repository.
 */
export {
  ContentError,
  contentRoot,
  type LoadedChallenge,
  type LoadedChallengeFile,
  type LoadedTrack,
  loadContent,
} from "./load.ts"
export {
  type ChallengeFileEntry,
  type ChallengeFileKind,
  type ChallengeJson,
  challengeDirNameSchema,
  challengeFileEntrySchema,
  challengeFileKindSchema,
  challengeJsonSchema,
  type Difficulty,
  difficultySchema,
  filePathSchema,
  isEditableByDefault,
  type RoadmapEntry,
  roadmapEntrySchema,
  slugSchema,
  type TrackJson,
  trackJsonSchema,
} from "./schema.ts"
