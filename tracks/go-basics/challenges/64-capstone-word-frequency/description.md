# Capstone: Word Frequency

Everything in this track so far arrived one idea at a time. A real program does
not: it is strings *and* maps *and* sorting *and* errors, and the interesting
part is how they fit together.

This is the classic one. Read some text, count the words, print the most
frequent ones with their share of the total. It is five short functions, and
every one of them has a decision in it that people get wrong the first time.

## The pipeline

```
text  ->  Tokenize  ->  []string
      ->  Tally     ->  map[string]int
      ->  Rank      ->  []Count      (sorted, cut to the top N)
      ->  Report    ->  string
```

`Analyze` is the four of them in a row. Splitting a program along the seams
like that is not ceremony: each stage has a value type on either side of it, so
each stage can be tested with a literal and read on its own.

## Tokenizing is a definition, not a `Split`

[`strings.Fields`](https://pkg.go.dev/strings#Fields) splits on spaces, which
makes `"stop!"` and `"stop"` two different words and `"Go,"` a third.
[`strings.Split(text, " ")`](https://pkg.go.dev/strings#Split) is worse, and
neither survives a newline.

So decide what a word *is*, and write that down:

- A maximal run of letters and digits.
- An apostrophe is part of the word when it sits directly between two of them,
  so `don't` stays whole and `'quoted'` loses its quotes.
- Everything else separates.
- The result is lowercase, so `Go` and `go` are the same word.

Note "letters", not "a-z".
[`unicode.IsLetter`](https://pkg.go.dev/unicode#IsLetter) and
[`unicode.ToLower`](https://pkg.go.dev/unicode#ToLower) know about `é`, `Ï` and
`Å`; a byte range does not, and the bug it produces is invisible in English and
embarrassing everywhere else.

## Length is measured in runes

The same trap, one function later. `len("café")` is 5, because the `é` takes
two bytes, so a minimum-length filter written with `len` keeps different words
depending on the language of the text.
[`utf8.RuneCountInString`](https://pkg.go.dev/unicode/utf8#RuneCountInString)
counts characters, which is what "shorter than four letters" meant.

## Ties have to break somewhere

Ranging over a map gives you the keys in a random order, deliberately, and Go
reshuffles it on every run. Sort by count and stop there, and the words that
tie come out in a different order each time: the report cannot be diffed, and a
test of it fails one run in three.

So the comparison has a second level. Count descending, then the word
ascending, and the output is a function of the input again:

```go
sort.Slice(ranked, func(i, j int) bool {
	if ranked[i].N != ranked[j].N {
		return ranked[i].N > ranked[j].N
	}
	return ranked[i].Word < ranked[j].Word
})
```

## The output format is part of the contract

A column that moves is a column nobody can read, so the line is a fixed
`Printf`:

```go
fmt.Sprintf("%2d. %-12s %4d %5.1f%%\n", rank, c.Word, c.N, share)
```

`%-12s` pads a short word on the right and leaves a long one alone: padding
verbs never truncate. `%%` is how you print a literal percent sign, and every
line ends in `\n` - including the last, so two reports concatenate cleanly.

## "Nothing to count" is not an empty report

If the text has no words, or the stop list eats all of them, an empty string is
a lie: it looks like a successful run of an empty file. That is a sentinel
error, wrapped so the caller can see where it came from:

```go
var ErrNoWords = errors.New("no countable words")

return "", fmt.Errorf("analyze: %w", ErrNoWords)
```

`%w` is what makes [`errors.Is(err, ErrNoWords)`](https://pkg.go.dev/errors#Is)
true two layers up, and the `analyze:` prefix is what makes the message useful
when it is printed.

## Task

Implement `Tokenize`, `Tally`, `Rank`, `Report` and `Analyze` to the contracts
in their doc comments.

The shares in the report are percentages of every word that was counted, not of
the entries that survived `top`: cutting the ranking to five must not turn five
words into 100% of the text.

## Hints

- `Tokenize` is one loop over `[]rune(text)` with a `cur []rune` accumulator
  and a small `flush` closure. The apostrophe case needs to look at the *next*
  rune, which is why the loop indexes the slice rather than ranging the string.
- Build the report with a
  [`strings.Builder`](https://pkg.go.dev/strings#Builder) and
  [`fmt.Fprintf`](https://pkg.go.dev/fmt#Fprintf) rather than `+=` in a loop.
- `Tally` and `Rank` both return a value the caller reads immediately, so
  return an empty map and an empty slice rather than `nil`:
  `make(map[string]int)` and `make([]Count, 0, len(counts))`.
- `top` is a limit, not a promise: `Rank(counts, 99)` on four words returns
  four, and zero or negative means all of them.
- Work the total out from the tally, before `Rank` cuts anything off.
