# JSON and Encoding

Go's `encoding/json` is the package you will reach for on your first day of
real work and never quite stop using. It has four functions worth memorising
and about six rules, and every one of the rules follows from a single fact:
**it works by reflection**. It walks a value at run time, asking what it is and
what fields it has.

Hold on to that and the package stops being a list of quirks.

## Encoding

```go
data, err := json.Marshal(v)                  // []byte
data, err := json.MarshalIndent(v, "", "  ")  // the same, with line breaks
```

Neither adds a trailing newline. The prefix argument to `MarshalIndent` goes in
front of every line after the first, and is almost always `""`.

What each Go type becomes:

| Go | JSON |
| --- | --- |
| `string` | string |
| `int`, `float64` | number |
| `bool` | true / false |
| struct | object, fields in **declaration order** |
| map | object, keys **sorted** |
| slice, array | array |
| nil slice, nil map, nil pointer | `null` |

Two rows there are worth pausing on.

**Maps are sorted.** Ranging over a map gives you a random order, but the
encoder sorts the keys, so marshalling the same map twice gives byte-identical
output. That is why you can compare JSON in a test without a normalisation
step.

**A nil slice is `null`, an empty slice is `[]`.** Go's own distinction,
carried through to the wire and on to whoever consumes it. If a client
distinguishes "no items" from "field missing", this is where it comes from, and
`make([]T, 0)` rather than `var s []T` is the fix.

### Only exported fields exist

Reflection cannot read a field it is not allowed to name, so an unexported
field is invisible: never written, never filled in. There is no tag that
changes that. If a value has to cross the JSON boundary, its field starts with
a capital letter, and that constraint shapes a lot of Go API structs.

### Marshal returns an error for a reason

Most values encode. Channels, functions and complex numbers have no JSON
representation, and a value that reaches itself makes the encoder stop rather
than recurse forever. Each comes back as an error.

Also: `Marshal` escapes `<`, `>` and `&` as `\u003c`, `\u003e` and
`\u0026`, so the output is safe inside a `<script>` tag. Same string once decoded - it only
shows up when you compare raw bytes.

## Decoding

```go
var task Task
err := json.Unmarshal(data, &task)
```

The `&` is not optional. `Unmarshal` writes through that pointer, and passing
the value itself is an `InvalidUnmarshalError`. A decode that leaves you with a
zero value and no error is almost always a missing `&` or an unexported field.

Decoding is a **fill, not a check**:

- A field the input does not mention is left at its zero value. A plain field
  cannot tell "absent" from "present and zero".
- A field the input mentions that the struct does not have is ignored. A typo
  in a key is silence.
- Key matching is case-insensitive. `"title"`, `"Title"` and `"TITLE"` all land
  in `Title`, though an exact tag match wins over a case-insensitive one.

Malformed JSON is an error, and so is a value of the wrong type:
`{"ID":"one"}` into an `int` field is an `UnmarshalTypeError`.

### When you do not know the shape

Decode into `map[string]any` and every value arrives as one of exactly six Go
types: `map[string]any`, `[]any`, `string`, **`float64`**, `bool`, `nil`.

There are no integers. `42` comes back as a `float64` holding 42, and getting
an `int` out is an assertion then a conversion:

```go
f, ok := m["count"].(float64)
if !ok {
	return 0, false
}
return int(f), true
```

Two comma-ok forms stacked - one for the map lookup, one for the assertion.
Skipping either is a panic waiting for the first input you did not write
yourself. When integer precision genuinely matters, `Decoder.UseNumber` gives
you a `json.Number` instead, which is a string you can parse either way.

Reaching for `map[string]any` when a struct would do is the most common
beginner smell in Go JSON code. Define the struct; you get names, types and a
compiler.

## The Decoder and the Encoder

`Unmarshal` takes a `[]byte` holding exactly one value. `json.NewDecoder(r)`
reads from an `io.Reader` and buys you two things it cannot do.

**Strict fields**, turning that silence into an error:

```go
decoder := json.NewDecoder(r)
decoder.DisallowUnknownFields()
```

Worth switching on for anything a human hand-edits, like a config file, where a
misspelled key should stop the program.

**Streams.** `Decode` reads one value per call and leaves the reader positioned
after it, so concatenated values are a loop. `io.EOF` is the only error that
means "done":

```go
for {
	var task Task
	err := decoder.Decode(&task)
	if errors.Is(err, io.EOF) {
		return tasks, nil
	}
	if err != nil {
		return nil, err
	}
	tasks = append(tasks, task)
}
```

`json.NewEncoder(w).Encode(v)` is the mirror image, and it writes straight to
the writer rather than building a `[]byte` first. It **does** append a newline,
unlike `Marshal`, and it is where `SetEscapeHTML(false)` and `SetIndent` live.
Decoding an `http.Response.Body` and encoding to an `http.ResponseWriter` are
the reason you see the streaming forms more often than the byte ones in real
code.

## Tags

A struct tag is a string literal after a field. The `json:` key holds a name
and options:

```go
type Profile struct {
	Username string   `json:"username"`
	Email    string   `json:"email,omitempty"`
	Password string   `json:"-"`
	Nickname *string  `json:"nickname,omitempty"`
	Roles    []string `json:"roles,omitempty"`
}
```

**Rename** with the name before the comma. It is two-way: it changes what is
written and what is read.

**`omitempty`** drops the field when the value is the zero number, the empty
string, `false`, `nil`, or a map, slice or array of length 0. It does **not**
cover an empty struct - `omitempty` on a struct field does nothing at all, and
that is the package's most reported non-bug.

**`-`** means never encode and never decode. Nothing in the input can reach the
field, which is what makes it the right way to keep a secret in the same struct
as the data. (The key literally named `-` is `json:"-,"`, with the trailing
comma.)

Tags are just strings. The compiler does not check them, so a typo is silence -
pin the expected bytes in a test.

### Absent or empty

`omitempty` on a `string` cannot tell "not set" from "set to nothing". A
pointer can: `nil` is dropped, and a pointer to `""` is written as `""`. On the
way back, an absent key leaves the pointer nil. `omitempty` on a pointer tests
the pointer, not what it points at.

Pay the pointer only where the difference matters. Most fields do not need it.

## When a tag is not enough

A tag renames a field; it cannot change how a *value* is written. For that, the
type implements the interfaces the package checks for before falling back to
reflection:

```go
func (c Celsius) MarshalJSON() ([]byte, error)
func (c *Celsius) UnmarshalJSON(data []byte) error
```

Look at the receivers. `MarshalJSON` reads, so a value receiver is enough and
both `Celsius` and `*Celsius` have it. `UnmarshalJSON` writes, so it **must**
be a pointer receiver - there is nowhere else for the result to go. On a value
receiver it compiles, silently does nothing, and costs you an afternoon. This
is the value-versus-pointer-receiver rule from earlier, with teeth.

`MarshalJSON` returns a complete JSON value, quotes included. Build the Go
value you want and marshal that rather than assembling bytes:

```go
return json.Marshal(text)   // quoting and escaping handled
```

The same pair is how `time.Time` becomes an RFC 3339 string, and how any domain
type with its own wire format - a money amount, an ID, a unit - keeps that
format in one place instead of at every call site.

## Further reading

- [encoding/json](https://pkg.go.dev/encoding/json) - `Marshal`, `Unmarshal`,
  the `Decoder` and `Encoder`, and the tag grammar spelled out in full.
- [JSON and Go](https://go.dev/blog/json) - the blog post that walks encoding,
  decoding and `map[string]any` in the order this lesson does.
- [encoding](https://pkg.go.dev/encoding) - `TextMarshaler` and friends, the
  interfaces `encoding/json` falls back to when there is no JSON pair.
- [Struct types](https://go.dev/ref/spec#Struct_types) - what a struct tag is
  to the language: a string literal the compiler stores and never reads.
- [time](https://pkg.go.dev/time) - `Time` already carries its own JSON pair,
  and its layout constants are the model for a custom one.

## Practise

Three challenges. The first marshals: field names without tags, sorted map
keys, nil versus empty, and the values that cannot be encoded at all. The
second decodes: partial fills, ignored fields, `float64` out of
`map[string]any`, `DisallowUnknownFields`, and a `Decoder` loop over a stream.
The third is the tag exercise - rename, `omitempty`, `-`, a `*string` for
absent-versus-empty - plus a `Celsius` that carries its own unit through
`MarshalJSON` and `UnmarshalJSON`, with the tests pinning the exact bytes.
