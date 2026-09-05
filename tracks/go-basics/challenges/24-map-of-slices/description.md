# Map of Slices

A map value can be anything, and `map[K][]V` - a key with a list under it - is
the shape you reach for whenever you are grouping. Go makes the common case
short:

```go
groups[key] = append(groups[key], value)
```

That works even when `key` is not in the map yet. The missing value reads as a
**nil slice**, `append` on a nil slice allocates a new one, and the assignment
puts the result back. Three separate rules line up so that grouping needs no
"does this key exist" branch.

The assignment is not optional, though. `append` returns a header that may point
at a different array, and the map holds a *copy* of the header, so
`append(groups[key], value)` on its own throws the result away.

The other half of this challenge is the flip side of the same fact: **copying a
map copies slice headers, not arrays.** Two maps built that way share their
elements, and a write through one is visible through the other. A copy is only
deep if each slice is copied too.

```go
for k, v := range src {
	dst[k] = append([]string(nil), v...)   // its own array
}
```

## Task

Fill in the seven functions in `main.go`.

1. `Append` adds one member to a group, creating the group when it is new and
   doing nothing at all to a nil map.
2. `GroupByFirstLetter` buckets words by their lowercased first letter, in input
   order.
3. `Keys` returns the group names sorted, and `Count` totals the members.
4. `Clone` is a deep copy, and `Merge` combines two maps into a third that
   borrows nothing from either.
5. `Invert` turns groups-of-members into members-of-groups, each list sorted.

## Hints

- Writing to a nil map panics, so `Append` needs an explicit `if groups == nil`
  guard before it touches anything.
- `word[:1]` is the first byte of a string, which is enough here because every
  test word starts with an ASCII letter.
- `append([]string(nil), members...)` is the shortest deep copy of one group.
  [`slices.Clone`](https://pkg.go.dev/slices#Clone) does the same thing if you
  would rather import it.
- `Merge` is `Clone` plus a loop. Cloning first is what keeps it from appending
  into a caller's spare capacity - and there is a test for exactly that.
- `Invert` gets sorted lists for free if you range over `Keys(groups)` rather
  than over the map.
