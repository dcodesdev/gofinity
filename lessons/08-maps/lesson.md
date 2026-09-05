# Maps

A map is Go's hash table, and it is the second of the two collection types you
will write every day. A slice indexes by position; a map indexes by anything
comparable.

```go
scores := map[string]int{"go": 3, "rust": 2}
scores["zig"] = 1
delete(scores, "rust")
fmt.Println(len(scores))   // 2
```

The type is `map[K]V`. `K` must be **comparable** - it has to support `==` - so
strings, numbers, booleans, pointers, channels, interfaces and structs or arrays
built only from those are all legal keys. Slices, maps and functions are not, so
`map[[]string]int` does not compile. `V` has any type at all, including another
map or a slice.

## Making one

The zero value of a map is `nil`, and a nil map is not an empty map you can
grow. It is readable and nothing else:

```go
var m map[string]int
fmt.Println(m["a"])   // 0
fmt.Println(len(m))   // 0
for k := range m {}   // zero iterations
delete(m, "a")        // no-op
m["a"] = 1            // panic: assignment to entry in nil map
```

Every operation is defined and harmless except the one that writes. That
asymmetry is deliberate: it means a struct with a nil map field can be read
without a nil check, and it means the panic arrives the first time someone
actually tries to store something.

To get a usable map, use a literal or `make`:

```go
a := map[string]int{}                 // empty literal
b := make(map[string]int)             // the same thing
c := make(map[string]int, 100)        // room for ~100, as a hint
```

The size argument to `make` is a capacity hint only. There is no `cap` for maps,
and the hint changes nothing you can observe except how often the map has to
grow while you fill it.

## Reading

A read never fails. A key that is not there gives the **zero value** of the
value type:

```go
counts := map[string]int{}
counts["nothing"]      // 0, and the map is still empty
```

Nothing was inserted by that read. This is why counting is so short in Go:

```go
counts[word]++         // a new word starts at zero
```

When you need to tell "absent" from "present and zero" - and with a value type
like `int` or `bool` or `string` you often do - use the two-result form:

```go
v, ok := counts["nothing"]   // 0, false
```

`ok` is the only thing that distinguishes them. That comma-ok shape is the same
one you have already seen from a type assertion and a channel receive; Go reuses
it wherever a value might not be there.

Combined with an `if` statement's initialiser it reads as one thought:

```go
if v, ok := counts[word]; ok {
	fmt.Println(word, "appeared", v, "times")
}
```

## Writing and deleting

`m[k] = v` inserts or overwrites - there is no separate insert. `delete(m, k)`
removes the entry and does nothing at all if the key was absent, so it returns
nothing; if you want to know whether the key was there, look before you delete.

You cannot take the address of a map element. `&m[k]` does not compile, and
neither does assigning to a field of a struct value stored in a map:

```go
type point struct{ x, y int }
m := map[string]point{"a": {1, 2}}
m["a"].x = 3        // does not compile
```

The reason is that the map may move its entries when it grows, so a pointer into
one could not stay valid. The two ways round it are to read, change and write
back, or to store `map[string]*point` and mutate through the pointer.

## Maps are references

A map value is a pointer to the hash table. Assigning one, or passing it to a
function, copies that reference and not the contents:

```go
func add(m map[string]int, k string) { m[k]++ }

counts := map[string]int{}
add(counts, "go")
fmt.Println(counts["go"])   // 1
```

No pointer parameter is needed, and no return value. The flip side is that
"copying" a map does not copy anything: two variables naming the same map see
each other's writes. `maps.Clone` gives you a shallow copy, and if the values
are slices or maps themselves, shallow is not enough - each one has to be copied
too, or the copy and the original still share their contents.

Because a map is a reference and not a value, maps are **not comparable**. `==`
is only defined against `nil`; use `maps.Equal` to compare contents.

## Range order is random

This is the rule that catches everyone:

```go
for k, v := range m {
	fmt.Println(k, v)      // a different order every run
}
```

Go randomises the iteration order on purpose, so that code cannot come to depend
on an order the implementation never promised. Any output that has to be stable
gets sorted on the way out:

```go
keys := make([]string, 0, len(m))
for k := range m {
	keys = append(keys, k)
}
sort.Strings(keys)
for _, k := range keys {
	fmt.Println(k, m[k])
}
```

A map holds the data; a sorted slice is how you present it. Adding a key during
a range is allowed but the new key may or may not be visited; deleting one you
have not reached yet means it will not be.

## Sets, and maps of slices

Go has no set type. A map with an empty value stands in for one, and the two
usual spellings are `map[string]bool` and `map[string]struct{}` - the second
stores nothing at all per entry, at the cost of a slightly noisier literal:

```go
seen := map[string]struct{}{}
seen["go"] = struct{}{}
if _, ok := seen["go"]; ok { /* ... */ }
```

For grouping, `map[K][]V` is the shape, and three rules line up to make it a
one-liner:

```go
groups[key] = append(groups[key], value)
```

The missing key reads as a nil slice, `append` on a nil slice allocates, and the
assignment stores the result. No "does this key exist" branch is needed.

The assignment is the part to be careful about. The map holds a *copy* of the
slice header, and `append` may return a different one, so
`append(groups[key], value)` on its own compiles, does work, and throws the
answer away.

## Further reading

- [Go maps in action](https://go.dev/blog/maps) - the blog post on comma-ok,
  deleting while ranging, and why iteration order is random.
- [Map types](https://go.dev/ref/spec#Map_types) - the spec on key comparability,
  the nil map, and what `len` and `delete` do.
- [Effective Go: maps](https://go.dev/doc/effective_go#maps) - the idioms,
  including sets and the missing-key zero value.
- [maps](https://pkg.go.dev/maps) - the standard library helpers for cloning a
  map and iterating its keys and values.

## Practise

Three challenges. The first is the mechanics: comma-ok against a stored zero, a
fallback lookup, `delete` reporting what it found, sorted keys, and building a
map from parallel slices. The second is the classic word count - tokenise,
normalise, tally, then sort by count and break ties by word, because a top-three
straight out of a range loop is a different top-three each run. The third is
`map[string][]string`: appending into a missing entry, a nil map that must not
panic, and a deep copy and a merge whose tests fail if the result borrows a
backing array from its input.
