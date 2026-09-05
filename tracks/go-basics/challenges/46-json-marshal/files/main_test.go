package main

import (
	"strings"
	"testing"
)

func TestEncodeTaskUsesTheGoFieldNames(t *testing.T) {
	got, err := EncodeTask(Task{ID: 1, Title: "ship it", Done: false, Tags: []string{"work"}})
	if err != nil {
		t.Fatalf("EncodeTask returned an error: %v", err)
	}
	want := `{"ID":1,"Title":"ship it","Done":false,"Tags":["work"]}`
	if got != want {
		t.Errorf("EncodeTask =\n  %s\nwant\n  %s", got, want)
	}
}

func TestEncodeTaskKeepsTheFieldOrderOfTheStruct(t *testing.T) {
	got, err := EncodeTask(Task{ID: 2, Title: "read", Done: true, Tags: nil})
	if err != nil {
		t.Fatalf("EncodeTask returned an error: %v", err)
	}
	want := `{"ID":2,"Title":"read","Done":true,"Tags":null}`
	if got != want {
		t.Errorf("EncodeTask =\n  %s\nwant\n  %s", got, want)
	}
}

func TestEncodeTaskSkipsTheUnexportedField(t *testing.T) {
	got, err := EncodeTask(Task{ID: 3, Title: "hidden", owner: "ada"})
	if err != nil {
		t.Fatalf("EncodeTask returned an error: %v", err)
	}
	if strings.Contains(got, "ada") || strings.Contains(strings.ToLower(got), "owner") {
		t.Errorf("EncodeTask leaked the unexported field: %s", got)
	}
}

func TestEncodeTaskEscapesAngleBrackets(t *testing.T) {
	got, err := EncodeTask(Task{ID: 4, Title: "a<b & c"})
	if err != nil {
		t.Fatalf("EncodeTask returned an error: %v", err)
	}
	// json.Marshal escapes <, > and & so the output is safe to drop inside a
	// <script> tag. It is still the same string once decoded.
	if strings.Contains(got, "<") || strings.Contains(got, "&") {
		t.Errorf("EncodeTask = %s, want < and & escaped as \\u003c and \\u0026", got)
	}
	if !strings.Contains(got, `\u003c`) {
		t.Errorf("EncodeTask = %s, want the < written as \\u003c", got)
	}
}

func TestEncodeIndentedUsesTwoSpaces(t *testing.T) {
	got, err := EncodeIndented(Task{ID: 1, Title: "ship it", Tags: []string{"work"}})
	if err != nil {
		t.Fatalf("EncodeIndented returned an error: %v", err)
	}
	want := "{\n" +
		"  \"ID\": 1,\n" +
		"  \"Title\": \"ship it\",\n" +
		"  \"Done\": false,\n" +
		"  \"Tags\": [\n" +
		"    \"work\"\n" +
		"  ]\n" +
		"}"
	if got != want {
		t.Errorf("EncodeIndented =\n%s\nwant\n%s", got, want)
	}
}

func TestEncodeIndentedHasNoPrefixAndNoTrailingNewline(t *testing.T) {
	got, err := EncodeIndented(map[string]int{"a": 1})
	if err != nil {
		t.Fatalf("EncodeIndented returned an error: %v", err)
	}
	if strings.HasPrefix(got, " ") || strings.HasPrefix(got, "\n") {
		t.Errorf("EncodeIndented = %q, want no prefix on the first line", got)
	}
	if strings.HasSuffix(got, "\n") {
		t.Errorf("EncodeIndented = %q, want no trailing newline", got)
	}
}

func TestEncodeAll(t *testing.T) {
	got, err := EncodeAll([]Task{{ID: 1, Title: "a"}, {ID: 2, Title: "b"}})
	if err != nil {
		t.Fatalf("EncodeAll returned an error: %v", err)
	}
	want := `[{"ID":1,"Title":"a","Done":false,"Tags":null},{"ID":2,"Title":"b","Done":false,"Tags":null}]`
	if got != want {
		t.Errorf("EncodeAll =\n  %s\nwant\n  %s", got, want)
	}
}

func TestEncodeAllOfNilIsNull(t *testing.T) {
	got, err := EncodeAll(nil)
	if err != nil {
		t.Fatalf("EncodeAll returned an error: %v", err)
	}
	if got != "null" {
		t.Errorf("EncodeAll(nil) = %s, want null", got)
	}
}

func TestEncodeAllOfEmptyIsAnEmptyArray(t *testing.T) {
	got, err := EncodeAll([]Task{})
	if err != nil {
		t.Fatalf("EncodeAll returned an error: %v", err)
	}
	if got != "[]" {
		t.Errorf("EncodeAll([]Task{}) = %s, want [] - an empty slice is not a nil slice", got)
	}
}

func TestEncodeCountsSortsTheKeys(t *testing.T) {
	got, err := EncodeCounts(map[string]int{"pears": 2, "apples": 5, "cherries": 9})
	if err != nil {
		t.Fatalf("EncodeCounts returned an error: %v", err)
	}
	want := `{"apples":5,"cherries":9,"pears":2}`
	if got != want {
		t.Errorf("EncodeCounts =\n  %s\nwant\n  %s", got, want)
	}
}

func TestEncodeCountsIsStableAcrossRuns(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2, "c": 3, "d": 4, "e": 5, "f": 6}
	first, err := EncodeCounts(m)
	if err != nil {
		t.Fatalf("EncodeCounts returned an error: %v", err)
	}
	for i := 0; i < 20; i++ {
		again, err := EncodeCounts(m)
		if err != nil {
			t.Fatalf("EncodeCounts returned an error: %v", err)
		}
		if again != first {
			t.Fatalf("EncodeCounts is not stable: %s then %s", first, again)
		}
	}
}

func TestEncodeCountsOfNilMapIsNull(t *testing.T) {
	var m map[string]int
	got, err := EncodeCounts(m)
	if err != nil {
		t.Fatalf("EncodeCounts returned an error: %v", err)
	}
	if got != "null" {
		t.Errorf("EncodeCounts(nil map) = %s, want null", got)
	}
}

func TestCanEncodeOrdinaryValues(t *testing.T) {
	values := []any{
		Task{ID: 1},
		[]string{"a", "b"},
		map[string]int{"a": 1},
		42,
		"hello",
		nil,
	}
	for _, v := range values {
		if !CanEncode(v) {
			t.Errorf("CanEncode(%#v) = false, want true", v)
		}
	}
}

func TestCanEncodeRejectsChannelsAndFunctions(t *testing.T) {
	if CanEncode(make(chan int)) {
		t.Error("CanEncode(chan int) = true, want false - a channel has no JSON form")
	}
	if CanEncode(func() {}) {
		t.Error("CanEncode(func) = true, want false - a function has no JSON form")
	}
	if CanEncode(complex(1, 2)) {
		t.Error("CanEncode(complex) = true, want false")
	}
}

func TestCanEncodeRejectsACycle(t *testing.T) {
	type node struct {
		Name string
		Next *node
	}
	n := &node{Name: "loop"}
	n.Next = n
	if CanEncode(n) {
		t.Error("CanEncode(a value that points at itself) = true, want false")
	}
}
