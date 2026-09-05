#!/usr/bin/env bun
/**
 * Validate the whole content tree and print a summary.
 *
 * Standalone on purpose: CI can run this alone to fail a bad challenge
 * contribution without booting the database or the rest of the suite.
 */
import { ContentError, loadAll } from "../src/load.ts"

try {
  const { tracks, lessons, warnings } = loadAll()
  for (const track of tracks) {
    const state = track.published ? "published" : "unpublished"
    console.log(`${track.slug} (${state}) - ${track.challenges.length} challenge(s)`)
    for (const challenge of track.challenges) {
      console.log(
        `  ${challenge.dirName}  ${challenge.difficulty}  ${challenge.files.length} file(s)`,
      )
    }
  }
  for (const lesson of lessons) {
    const state = lesson.published ? "published" : "unpublished"
    console.log(
      `${lesson.dirName} (${state}) practises ${lesson.challenges
        .map((c) => `${c.track}/${c.challenge}`)
        .join(", ")}`,
    )
  }
  for (const warning of warnings) console.warn(`warning: ${warning}`)
  const total = tracks.reduce((n, t) => n + t.challenges.length, 0)
  console.log(`ok - ${tracks.length} track(s), ${total} challenge(s), ${lessons.length} lesson(s)`)
} catch (error) {
  console.error(error instanceof ContentError ? `content error: ${error.message}` : error)
  process.exit(1)
}
