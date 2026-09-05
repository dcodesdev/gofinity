# Map Basics

A **map** is Go's hash table: an unordered collection of key/value pairs where
the key type has to be comparable and the value type can be anything.

```go
scores := map[string]int{"go": 3, "rust": 2}
scores["zig"] = 1
fmt.Println(scores["go"])   // 3
```

Three things about that snippet are worth stopping on.

**A missing key is not an error.** Reading one gives the zero value of the value
type, so `scores["haskell"]` is `0` and nothing complains. When the difference
matters, ask for it with the comma-ok form:

```go
v, ok := scores["haskell"]   // 0, false
```

**A map has to be made before it can be written to.** The zero value of a map is
`nil`, and a nil map reads fine but panics on assignment:

```go
var m map[string]int
fmt.Println(m["a"])   // 0 - fine
m["a"] = 1            // panic: assignment to entry in nil map
```

`make(map[string]int)` or a literal gives you a real one.

**Range order is random.** Go deliberately shuffles it, so any output you want
to be stable has to be sorted first.

## Task

Fill in the seven functions in `main.go`.

1. `Lookup` and `Get` answer the "is it there" question two ways: comma-ok, and
   a caller-supplied fallback.
2. `Add(m, key, n)` adds `n` to whatever is under `key`, treating a new key as
   zero, and the caller sees the change.
3. `Remove(m, key)` deletes and reports whether the key existed.
4. `SortedKeys(m)` returns the keys in ascending order, and `Total(m)` sums the
   values.
5. `FromPairs(keys, values)` builds a map from two parallel slices, returning
   `nil` when the lengths disagree.

## Hints

- `v, ok := m[key]` is the only form that distinguishes a stored `0` from a
  missing key. `Lookup` is one line.
- `m[key] += n` already starts from zero for a new key, so `Add` needs no
  branch.
- `delete(m, key)` is a builtin, and it is a no-op on a key that is not there,
  so `Remove` has to check before deleting to know what to report.
- Build the key slice with `make([]string, 0, len(m))`, append in a `for k :=
  range m` loop, then [`sort.Strings`](https://pkg.go.dev/sort#Strings).
- `FromPairs` must return a usable empty map for two empty inputs, so reach for
  `make` rather than declaring a `var`.
