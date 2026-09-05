package main

import (
	"encoding/json"
	"fmt"
)

// Task is the value the whole file encodes. Note what is and is not here:
// four exported fields and one unexported one. encoding/json can only see the
// exported four, because it works through reflection and reflection cannot
// read a field it is not allowed to name.
type Task struct {
	ID    int
	Title string
	Done  bool
	Tags  []string

	// owner is unexported, so it never appears in the output and is never
	// filled in by a decoder. There is no tag that changes that.
	owner string
}

// EncodeTask returns the JSON encoding of t as a string.
//
// With no struct tags anywhere, each key is the Go field name verbatim,
// capital included: {"ID":1,"Title":"ship it","Done":false,"Tags":["work"]}.
// Renaming them is the next challenge.
func EncodeTask(t Task) (string, error) {
	// Marshal returns a []byte; the conversion to string is a copy, and the
	// only reason this function exists rather than being inlined everywhere.
	data, err := json.Marshal(t)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// EncodeIndented returns the JSON encoding of v, indented for a human to read:
// no prefix, and two spaces per level.
//
//	{
//	  "ID": 1,
//	  "Title": "ship it"
//	}
//
// It takes an any rather than a Task, because indentation has nothing to do
// with the type being encoded.
func EncodeIndented(v any) (string, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// EncodeAll returns the JSON encoding of a slice of tasks.
//
// The distinction the tests care about: a nil slice encodes as null, and an
// empty non-nil slice encodes as []. That difference survives the whole way to
// a browser, so it is worth knowing before you have to debug it.
func EncodeAll(tasks []Task) (string, error) {
	data, err := json.Marshal(tasks)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// EncodeCounts returns the JSON encoding of m.
//
// A Go map has no order, but the encoder does: it sorts the keys, so encoding
// the same map twice gives byte-identical output.
func EncodeCounts(m map[string]int) (string, error) {
	data, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// CanEncode reports whether v can be encoded at all.
//
// Not everything can. Channels, functions and complex numbers have no JSON
// representation, and a cycle - a value that reaches itself - makes the
// encoder give up rather than recurse forever. Each of those comes back as an
// error, so this is a one-line question about one.
func CanEncode(v any) bool {
	_, err := json.Marshal(v)
	return err == nil
}

func main() {
	out, err := EncodeTask(Task{ID: 1, Title: "ship it", Tags: []string{"work"}})
	if err != nil {
		fmt.Println("encode failed:", err)
		return
	}
	fmt.Println(out)
}
