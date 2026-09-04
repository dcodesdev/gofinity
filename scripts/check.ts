#!/usr/bin/env bun
/**
 * Validate the whole content tree and print a summary.
 *
 * Standalone on purpose: CI can run this alone to fail a bad challenge
 * contribution without booting the database or the rest of the suite.
 */
import { ContentError, loadContent } from "../src/load.ts"

try {
  const tracks = loadContent()
  for (const track of tracks) {
    const state = track.published ? "published" : "unpublished"
    console.log(`${track.slug} (${state}) — ${track.challenges.length} challenge(s)`)
    for (const challenge of track.challenges) {
      console.log(
        `  ${challenge.dirName}  ${challenge.difficulty}  ${challenge.files.length} file(s)`,
      )
    }
  }
  const total = tracks.reduce((n, t) => n + t.challenges.length, 0)
  console.log(`ok — ${tracks.length} track(s), ${total} challenge(s)`)
} catch (error) {
  console.error(error instanceof ContentError ? `content error: ${error.message}` : error)
  process.exit(1)
}
