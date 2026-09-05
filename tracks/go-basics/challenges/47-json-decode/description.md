# Decoding JSON

Decoding is the other half of
[`encoding/json`](https://pkg.go.dev/encoding/json), and it is the half with
the surprises. The plain form:

```go
var task Task
err := json.Unmarshal(data, &task)
```

The **`&`** is not optional. `Unmarshal` writes through that pointer, so
passing the value itself is an `InvalidUnmarshalError` rather than a silent
no-op. If a decode leaves you with a zero value and no error, check this first.

## Decoding is a fill, not a check

Two behaviours follow from that, and both are on purpose:

- A field the input **does not mention** is left at its zero value. There is no
  way to tell "absent" from "present and zero" with a plain field - a `*string`
  is the usual fix, and the next challenge uses one.
- A field the input mentions that the struct **does not have** is ignored. A
  typo in a key is silence, not an error.

Matching is also case-insensitive: `"title"`, `"Title"` and `"TITLE"` all land
in `Title`. An exact tag match wins over a case-insensitive one, but with no
tags at all, any casing gets in.

What *is* an error: malformed JSON, and a value of the wrong type. `{"ID":"one"}`
into an `int` field is an `UnmarshalTypeError`.

## When you do not know the shape

Decode into `map[string]any` and every value arrives as one of exactly six Go
types:

| JSON | Go |
| --- | --- |
| object | `map[string]any` |
| array | `[]any` |
| string | `string` |
| number | **`float64`** |
| true / false | `bool` |
| null | `nil` |

That fourth row is the one that catches people. There are no integers here.
`42` comes back as a `float64` holding 42, and getting an `int` out of it is an
assertion followed by a conversion:

```go
f, ok := m["count"].(float64)
if !ok {
	return 0, false
}
return int(f), true
```

Two comma-ok forms stacked: one for the map lookup, one for the assertion.
Skipping either is a panic waiting for the first input you did not write
yourself. (When integer precision actually matters -
[`json.Decoder.UseNumber`](https://pkg.go.dev/encoding/json#Decoder.UseNumber) -
there is a [`json.Number`](https://pkg.go.dev/encoding/json#Number) for it.)

## The Decoder

[`json.Unmarshal`](https://pkg.go.dev/encoding/json#Unmarshal) takes a `[]byte`
that must be exactly one JSON value.
[`json.NewDecoder(r)`](https://pkg.go.dev/encoding/json#NewDecoder) reads from
an [`io.Reader`](https://pkg.go.dev/io#Reader) and unlocks two things
`Unmarshal` cannot do.

**Strict fields.** Turn the silence above into an error:

```go
decoder := json.NewDecoder(bytes.NewReader(data))
decoder.DisallowUnknownFields()
err := decoder.Decode(&task)
```

**Streams.** `Decode` reads *one* value per call and leaves the reader
positioned after it, so a sequence of concatenated objects is a loop. It
returns [`io.EOF`](https://pkg.go.dev/io#EOF) when the stream is spent, and that
is the only error that means "done":

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

Decoding straight from an
[`http.Response.Body`](https://pkg.go.dev/net/http#Response) is the same call,
which is why you see it far more often than `Unmarshal` in real code.

## Task

Implement `DecodeTask`, `DecodeAll`, `DecodeAny`, `LookupInt`, `DecodeStrict`
and `DecodeStream`.

## Hints

- `DecodeAll` gets the nil-versus-empty behaviour for free: declare
  `var tasks []Task` and pass `&tasks`. `null` leaves it nil, `[]` makes it
  empty and non-nil.
- `DecodeStrict` needs a reader over the bytes it was handed:
  [`bytes.NewReader(data)`](https://pkg.go.dev/bytes#NewReader).
- `DecodeStream` builds its slice with `make([]Task, 0)` so an empty stream
  returns empty rather than nil.
- Use [`errors.Is(err, io.EOF)`](https://pkg.go.dev/errors#Is) for the end of
  the stream, and return every other error.
