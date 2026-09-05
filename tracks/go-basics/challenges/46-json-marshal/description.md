# Marshalling JSON

[`encoding/json`](https://pkg.go.dev/encoding/json) turns Go values into JSON
with one function:

```go
data, err := json.Marshal(v)   // data is a []byte
```

It works by reflection, walking `v` at run time and writing the bytes as it
goes. That single fact explains almost everything else about the package.

## Only exported fields exist

Reflection cannot read a field it is not allowed to name, so an unexported
field is invisible: it is never written, and never filled in when decoding.
There is no tag that changes it. If a value has to cross the JSON boundary, its
field starts with a capital letter.

With no tags at all, the key is the Go field name verbatim - `"Title"`, not
`"title"`. Renaming is the next challenge; this one is about what happens
before you reach for a tag.

## What each Go type becomes

| Go | JSON |
| --- | --- |
| `string` | string |
| `int`, `float64` | number |
| `bool` | true / false |
| struct | object, fields in declaration order |
| map | object, **keys sorted** |
| slice, array | array |
| `nil` slice or map, `nil` pointer | `null` |

Two of those rows are the ones that bite. A map's keys are sorted by the
encoder, so encoding the same map twice gives byte-identical output even though
ranging over it would not. And a **nil** slice is `null` while an **empty**
slice is `[]` - the same distinction Go draws, carried through to the wire and
all the way to whoever consumes it.

## Indenting

[`json.MarshalIndent`](https://pkg.go.dev/encoding/json#MarshalIndent) takes a
prefix and an indent, and is `Marshal` plus line breaks. The prefix goes in
front of every line **after** the first, and the indent is one level. Two spaces
and no prefix is the usual choice:

```go
json.MarshalIndent(v, "", "  ")
```

Neither function adds a trailing newline.

## Angle brackets come back escaped

`Marshal` escapes `<`, `>` and `&` as `\u003c`, `\u003e` and `\u0026`, so the
output is safe to paste inside a `<script>` tag. It is the same string once
decoded, so this only matters when you are comparing raw bytes - which the tests
here do.

## The error is not decoration

Most values encode, but not all. A channel, a function or a complex number has
no JSON representation, and a value that reaches itself makes the encoder stop
rather than recurse forever. All three come back as an error, which is why
`Marshal` returns one.

## Task

Implement `EncodeTask`, `EncodeIndented`, `EncodeAll`, `EncodeCounts` and
`CanEncode`.

## Hints

- [`json.Marshal`](https://pkg.go.dev/encoding/json#Marshal) gives you a
  `[]byte`; the functions here return a `string`, so convert with
  `string(data)`.
- On an error, return `""` and the error. Do not return the partial bytes.
- `EncodeAll(nil)` should come out as `null` and `EncodeAll([]Task{})` as `[]` -
  you get both for free by passing the slice straight through.
- `CanEncode` is one call and one comparison: marshal, discard the bytes with
  `_`, and report whether the error was nil.
