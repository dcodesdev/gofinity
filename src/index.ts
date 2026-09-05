/**
 * `@gofinity/content` - the open-source challenge content for Gofinity.
 *
 * This package is the source of truth in git; `bun run seed` reads it with
 * {@link loadContent} and upserts it into Postgres. It has no dependencies on
 * any other workspace package, because it is destined to be extracted into a
 * standalone public repository.
 */
export {
  ALLOWED_LINK_HOSTS,
  ContentError,
  contentRoot,
  hasAllowedLink,
  type LoadedChallenge,
  type LoadedChallengeFile,
  type LoadedContent,
  type LoadedLesson,
  type LoadedTrack,
  lessonsRoot,
  loadAll,
  loadContent,
  loadLessons,
  markdownLinkUrls,
  packageRoot,
  unpractisedChallenges,
} from "./load.ts"
export {
  type ChallengeFileEntry,
  type ChallengeFileKind,
  type ChallengeJson,
  challengeDirNameSchema,
  challengeFileEntrySchema,
  challengeFileKindSchema,
  challengeJsonSchema,
  challengeTagSchema,
  type Difficulty,
  difficultySchema,
  filePathSchema,
  isEditableByDefault,
  type LessonChallengeRef,
  type LessonJson,
  lessonChallengeRefSchema,
  lessonDirNameSchema,
  lessonJsonSchema,
  type RoadmapEntry,
  roadmapEntrySchema,
  slugSchema,
  type TrackJson,
  trackJsonSchema,
} from "./schema.ts"
