#!/usr/bin/env bun
/**
 * List every external link in the content tree, grouped by file.
 *
 * `content:check` proves a link is on an allowed host; it cannot prove the page
 * exists, because nothing here reaches the network. This prints the whole set
 * so a human, or a link checker, can open them once.
 */
import { ContentError, loadAll, markdownLinkUrls } from "../src/load.ts"

function externalLinks(md: string): string[] {
  const urls = markdownLinkUrls(md).filter((url) => /^https?:/i.test(url))
  return [...new Set(urls)].sort()
}

try {
  const { tracks, lessons } = loadAll()
  const files: { path: string; urls: string[] }[] = []
  for (const lesson of lessons) {
    files.push({ path: `lessons/${lesson.dirName}/lesson.md`, urls: externalLinks(lesson.bodyMd) })
  }
  for (const track of tracks) {
    for (const challenge of track.challenges) {
      files.push({
        path: `tracks/${track.slug}/challenges/${challenge.dirName}/description.md`,
        urls: externalLinks(challenge.descriptionMd),
      })
    }
  }

  const withLinks = files.filter((f) => f.urls.length > 0)
  for (const file of withLinks) {
    console.log(file.path)
    for (const url of file.urls) console.log(`  ${url}`)
  }
  const all = new Set(withLinks.flatMap((f) => f.urls))
  console.log(
    `\n${all.size} distinct link(s) across ${withLinks.length} of ${files.length} file(s); ${
      files.length - withLinks.length
    } file(s) link nothing`,
  )
} catch (error) {
  console.error(error instanceof ContentError ? `content error: ${error.message}` : error)
  process.exit(1)
}
