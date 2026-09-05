package main

import (
	"fmt"
	"io"
	"strings"
)

// Task is the shape we decode into. As in the previous challenge there are no
// struct tags yet, so the decoder matches on the Go field names - though it
// matches them case-insensitively, which is why "title" in the input still
// lands in Title.
type Task struct {
	ID    int
	Title string
	Done  bool
	Tags  []string
}

// DecodeTask decodes a single JSON object into a Task.
//
// Fields the input does not mention are left at their zero value, and fields
// the input mentions but Task does not have are ignored. Both are deliberate:
// decoding is a best-effort fill, not a schema check.
func DecodeTask(data []byte) (Task, error) {
	// TODO
	return Task{}, nil
}

// DecodeAll decodes a JSON array of objects into a slice of tasks.
//
// The JSON literal null decodes into a nil slice, and [] into an empty non-nil
// one, the mirror image of what encoding does.
func DecodeAll(data []byte) ([]Task, error) {
	// TODO
	return nil, nil
}

// DecodeAny decodes an arbitrary JSON object into a map, for when the shape is
// not known ahead of time.
//
// Every value comes back as one of six Go types: bool, float64, string, nil,
// []any, or map[string]any. Note the third of those - **every** JSON number
// becomes a float64, however integral it looked in the input.
func DecodeAny(data []byte) (map[string]any, error) {
	// TODO
	return nil, nil
}

// LookupInt returns m[key] as an int, and whether that was possible.
//
// It is false when the key is absent and false when the value is there but is
// not a number - a string "3" is not a 3. Remember that the value under the
// key is a float64, so getting to an int is a type assertion followed by a
// conversion.
func LookupInt(m map[string]any, key string) (int, bool) {
	// TODO
	return 0, false
}

// DecodeStrict decodes a single object like DecodeTask, but fails when the
// input carries a field Task does not have.
//
// That is not the default and cannot be switched on through json.Unmarshal: it
// is a setting on a json.Decoder, so this function builds one over the bytes.
// Use it when a typo in a config file should be an error rather than silence.
func DecodeStrict(data []byte) (Task, error) {
	// TODO
	return Task{}, nil
}

// DecodeStream reads a sequence of JSON objects from r - concatenated, with
// nothing between them but optional whitespace - and returns them in order.
//
//	{"ID":1}{"ID":2}
//	{"ID":3}
//
// json.Unmarshal cannot do this: it wants the whole input to be one value. A
// json.Decoder reads one value per call to Decode and returns io.EOF when the
// stream is spent, which is the loop this function is.
//
// An empty stream gives an empty, non-nil slice. A malformed value is an
// error, not a short read.
func DecodeStream(r io.Reader) ([]Task, error) {
	// TODO
	return nil, nil
}

func main() {
	task, err := DecodeTask([]byte(`{"id":1,"title":"ship it","tags":["work"]}`))
	if err != nil {
		fmt.Println("decode failed:", err)
		return
	}
	fmt.Printf("%+v\n", task)

	stream, err := DecodeStream(strings.NewReader(`{"ID":1}{"ID":2}`))
	if err != nil {
		fmt.Println("stream failed:", err)
		return
	}
	fmt.Println(len(stream), "task(s) in the stream")
}
