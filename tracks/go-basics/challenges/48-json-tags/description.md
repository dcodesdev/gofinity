# JSON Tags

A struct tag is a string literal after a field, and
[`encoding/json`](https://pkg.go.dev/encoding/json) reads the part keyed
`json:`:

```go
type Profile struct {
	Username string `json:"username"`
}
```

Backticks, no spaces around the colon, and the whole tag is one literal even
when several packages share it. A tag is not checked by the compiler - it is
just a string - so a typo is silence, which is the main reason to pin the
expected bytes in a test.

## The three things a tag does

```go
Email    string `json:"email,omitempty"`
Password string `json:"-"`
Address  Address `json:"address"`
```

**Rename.** The name before the comma is the key. Renaming is a two-way
setting: it changes what is written *and* what is read.

**`omitempty`.** Drop the field when its value is *empty*, which means the zero
number, the empty string, `false`, `nil`, or a map, slice or array of length 0.
It does **not** mean an empty struct - `omitempty` on a struct field does
nothing at all, which is a well-known sharp edge.

**`-`.** Never encode, never decode. A field tagged `-` is invisible in both
directions, so nothing in the input can reach it. (The key literally named `-`
is spelled `json:"-,"`, with the trailing comma. You will not need it often,
but that is why it exists.)

## Absent or empty?

`omitempty` on a `string` cannot tell "not set" from "set to nothing", because
both are `""`. A pointer can:

```go
Nickname *string `json:"nickname,omitempty"`
```

Now `nil` is dropped and a pointer to `""` is written as `"nickname":""`. On the
way back, an absent key leaves the pointer nil while `"nickname":""` gives you a
pointer to the empty string. `omitempty` on a pointer tests the pointer, not
what it points at - that is the whole trick.

## Nesting

Nothing special: the encoder recurses, and the tags live on the nested type, so
they apply wherever a value of it appears. An `Address` writes the same keys
whether it is a field, an element of a slice, or encoded on its own.

## When a tag is not enough

A tag renames a field. It cannot change how a *value* is written. For that, the
type implements the two interfaces the package checks for before it falls back
to reflection,
[`json.Marshaler`](https://pkg.go.dev/encoding/json#Marshaler) and
[`json.Unmarshaler`](https://pkg.go.dev/encoding/json#Unmarshaler):

```go
func (c Celsius) MarshalJSON() ([]byte, error)
func (c *Celsius) UnmarshalJSON(data []byte) error
```

Note the receivers. `MarshalJSON` reads, so a value receiver is enough and both
`Celsius` and `*Celsius` get the method. `UnmarshalJSON` writes, so it **must**
be a pointer receiver - there is nowhere else for the result to go. Writing it
on a value receiver compiles, does nothing, and is one of the more annoying
afternoons in Go.

`MarshalJSON` returns a **complete JSON value**, quotes included, not the text
inside them. The easy way to get that right is to build the Go value you want
and marshal it, rather than assembling bytes by hand:

```go
return json.Marshal(text)   // handles the quoting and the escaping
```

## Task

Tag every field of `Profile` and `Address` so the encoding matches the shape
documented on `Profile`, then implement `Celsius`'s two methods and `Encode`,
`Decode` and `RoundTrip`.

## Hints

- Field order in the struct is key order in the output, and the tests pin the
  exact bytes. Do not reorder the fields.
- [`strconv.FormatFloat`](https://pkg.go.dev/strconv#FormatFloat) with
  `(f, 'f', -1, 64)` is the shortest form that reads back as the same
  `float64`: `0` rather than `0.000000`.
- [`strings.CutSuffix(text, "C")`](https://pkg.go.dev/strings#CutSuffix) gives
  you the number and a bool saying whether the suffix was there. A missing
  suffix is an error, not a fallback.
- In `UnmarshalJSON`, decode into a `string` first. If that fails, the input
  was not a JSON string and you are done.
- Assign the result with `*c = Celsius(value)`. Assigning to `c` replaces a
  local pointer and changes nothing.
- `RoundTrip` is `Encode` then `Decode`. Do not reimplement either.
