# Mutating Methods

A type whose methods change its own state is the most ordinary thing in Go, and
it comes down to two decisions: which receiver each method takes, and whether
the zero value works without a constructor.

## The zero value first

`Stack` holds a `[]int`. Appending to a nil slice allocates one, so the zero
`Stack` is already an empty stack:

```go
var s Stack
s.Push(1)      // fine, no constructor
```

That is worth designing for. `bytes.Buffer`, `strings.Builder` and `sync.Mutex`
all work this way, and it is why you rarely see a `New` function in the standard
library.

`Tally` holds a `map[string]int`, and a map does not have that property.
**Reading** a nil map is fine and gives the zero value; **writing** to one
panics with "assignment to entry in nil map". So a type with a map field either
needs a constructor, or needs its writing methods to make the map first. This
challenge does both: `NewTally` exists, and `Add` still copes with a `Tally`
someone built as `var t Tally`.

## Which receiver

Anything that assigns to a field needs a pointer receiver. `Push` assigns to
`s.items` - `append` may return a different slice header - so `Push` is
`func (s *Stack) Push(...)`. `Len` only reads, so it can be a value receiver,
and the copy that makes is cheap: copying a struct with a slice field copies the
three-word header, not the elements.

The usual advice is to pick one receiver kind per type and stay with it, so a
reader does not have to check each method. The mixture here is deliberate, to
make the difference visible.

## The slice trap, again

`DrainAll` takes a `[]Stack` and has to empty each one. Ranging over values
hands out copies, and draining a copy leaves the original full:

```go
for _, s := range stacks { s.Drain() }    // drains nothing
for i := range stacks    { stacks[i].Drain() }  // drains the slice
```

## Task

Fill in `main.go`: the `Stack` methods, `DrainAll`, and the `Tally` methods.

## Hints

- `Pop` removes by reslicing: read `s.items[len(s.items)-1]`, then assign
  `s.items = s.items[:len(s.items)-1]`. Check the length first, or the index
  panics.
- `Drain` can loop on `Pop` until it returns false. Start from `[]int{}` so a
  drained empty stack gives `[]` rather than nil.
- `PushAll` is variadic: `vs` is a `[]int`, and `PushAll()` gives it length 0.
- `Add` and `Merge` write to the map, so both need the
  `if t.counts == nil { t.counts = map[string]int{} }` guard. `Count` and
  `Total` only read, and reading a nil map needs no guard at all.
- `Merge` must not touch `other`, so read from `other.counts` and write only to
  `t.counts`. It is also called with nil, so check for that first.
