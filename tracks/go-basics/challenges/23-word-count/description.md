# Word Count

Counting things is what maps are for, and a word count is the smallest program
that shows the whole pattern: tokenise, normalise, tally, then **sort on the way
out**.

That last step is not decoration. Go randomises the order of `for k := range m`
on purpose, so a function that returns "the top three words" straight out of a
range loop returns a different answer each run. The map holds the counts; a
sorted slice is how you hand them to anyone else.

```go
counts := map[string]int{}
for _, w := range strings.Fields(text) {
	counts[w]++     // a new key starts at zero, so no branch is needed
}
```

Sorting by two keys at once - count descending, then word ascending to break
ties - is the ordinary shape for this, and a
[`sort.Slice`](https://pkg.go.dev/sort#Slice) comparison expresses it directly:

```go
sort.Slice(pairs, func(i, j int) bool {
	if pairs[i].Count != pairs[j].Count {
		return pairs[i].Count > pairs[j].Count
	}
	return pairs[i].Word < pairs[j].Word
})
```

Leave the tie-break out and the equal counts come back in whatever order the map
handed them over, which is to say a different one every time.

## Task

Fill in the five functions in `main.go`. Leave the `Pair` type and the
`Punctuation` constant alone - the tests use both.

1. `Count(text)` splits on whitespace, lowercases each token, trims
   `Punctuation` off both ends, skips anything left empty, and tallies the rest.
2. `Unique(text)` reports how many distinct words there are.
3. `SortedPairs(counts)` returns every entry as a `Pair`, count descending then
   word ascending, without disturbing the map it was given.
4. `TopN(counts, n)` takes the first `n` of those, clamping `n` into range.
5. `Report(counts, n)` renders them as `"word: count\n"` lines.

## Hints

- [`strings.Fields`](https://pkg.go.dev/strings#Fields) splits on runs of any
  whitespace and drops empty tokens, so it is a better fit here than
  [`strings.Split(text, " ")`](https://pkg.go.dev/strings#Split).
- [`strings.Trim(s, Punctuation)`](https://pkg.go.dev/strings#Trim) removes any
  of those characters from both ends and leaves the middle alone, which is why
  `don't` survives intact.
- A token that trims down to `""` was punctuation only. Skip it before counting,
  or you will get an entry under the empty string.
- `Unique` is one line once `Count` works.
- `Report` can build its string with a
  [`strings.Builder`](https://pkg.go.dev/strings#Builder) and
  [`fmt.Fprintf(&b, ...)`](https://pkg.go.dev/fmt#Fprintf), or with `+=` if you
  prefer - the output is short.
