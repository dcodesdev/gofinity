package main

import (
	"strings"
	"testing"
)

func TestDecodeTask(t *testing.T) {
	got, err := DecodeTask([]byte(`{"ID":1,"Title":"ship it","Done":true,"Tags":["work","go"]}`))
	if err != nil {
		t.Fatalf("DecodeTask returned an error: %v", err)
	}
	if got.ID != 1 || got.Title != "ship it" || !got.Done {
		t.Errorf("DecodeTask = %+v, want {ID:1 Title:ship it Done:true}", got)
	}
	if len(got.Tags) != 2 || got.Tags[0] != "work" || got.Tags[1] != "go" {
		t.Errorf("DecodeTask Tags = %v, want [work go]", got.Tags)
	}
}

func TestDecodeTaskMatchesKeysCaseInsensitively(t *testing.T) {
	got, err := DecodeTask([]byte(`{"id":7,"title":"lowercase keys","done":true}`))
	if err != nil {
		t.Fatalf("DecodeTask returned an error: %v", err)
	}
	if got.ID != 7 || got.Title != "lowercase keys" || !got.Done {
		t.Errorf("DecodeTask = %+v, want the lowercase keys to match the fields", got)
	}
}

func TestDecodeTaskLeavesMissingFieldsAtZero(t *testing.T) {
	got, err := DecodeTask([]byte(`{"Title":"only a title"}`))
	if err != nil {
		t.Fatalf("DecodeTask returned an error: %v", err)
	}
	if got.ID != 0 || got.Done || got.Tags != nil {
		t.Errorf("DecodeTask = %+v, want everything but Title left at its zero value", got)
	}
}

func TestDecodeTaskIgnoresUnknownFields(t *testing.T) {
	got, err := DecodeTask([]byte(`{"ID":2,"Titel":"typo","priority":"high"}`))
	if err != nil {
		t.Fatalf("DecodeTask returned an error: %v", err)
	}
	if got.ID != 2 {
		t.Errorf("DecodeTask ID = %d, want 2", got.ID)
	}
	if got.Title != "" {
		t.Errorf("DecodeTask Title = %q, want %q - Titel is not Title", got.Title, "")
	}
}

func TestDecodeTaskRejectsMalformedInput(t *testing.T) {
	if _, err := DecodeTask([]byte(`{"ID":`)); err == nil {
		t.Error("DecodeTask on truncated JSON returned nil error, want an error")
	}
	if _, err := DecodeTask([]byte(`not json at all`)); err == nil {
		t.Error("DecodeTask on garbage returned nil error, want an error")
	}
}

func TestDecodeTaskRejectsTheWrongType(t *testing.T) {
	// ID is an int and the input says it is a string. That is an
	// UnmarshalTypeError, not something to paper over.
	if _, err := DecodeTask([]byte(`{"ID":"one"}`)); err == nil {
		t.Error(`DecodeTask on {"ID":"one"} returned nil error, want a type error`)
	}
}

func TestDecodeAll(t *testing.T) {
	got, err := DecodeAll([]byte(`[{"ID":1},{"ID":2},{"ID":3}]`))
	if err != nil {
		t.Fatalf("DecodeAll returned an error: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("DecodeAll returned %d tasks, want 3", len(got))
	}
	for i, task := range got {
		if task.ID != i+1 {
			t.Errorf("DecodeAll[%d].ID = %d, want %d", i, task.ID, i+1)
		}
	}
}

func TestDecodeAllOfNullIsNil(t *testing.T) {
	got, err := DecodeAll([]byte(`null`))
	if err != nil {
		t.Fatalf("DecodeAll returned an error: %v", err)
	}
	if got != nil {
		t.Errorf("DecodeAll(null) = %v, want a nil slice", got)
	}
}

func TestDecodeAllOfEmptyArrayIsEmptyNotNil(t *testing.T) {
	got, err := DecodeAll([]byte(`[]`))
	if err != nil {
		t.Fatalf("DecodeAll returned an error: %v", err)
	}
	if got == nil {
		t.Fatal("DecodeAll([]) returned nil, want an empty non-nil slice")
	}
	if len(got) != 0 {
		t.Errorf("DecodeAll([]) = %v, want length 0", got)
	}
}

func TestDecodeAnyGivesTheSixDynamicTypes(t *testing.T) {
	got, err := DecodeAny([]byte(`{"n":3,"s":"go","b":true,"nothing":null,"list":[1,2],"nested":{"k":"v"}}`))
	if err != nil {
		t.Fatalf("DecodeAny returned an error: %v", err)
	}

	if _, ok := got["n"].(float64); !ok {
		t.Errorf("DecodeAny n has dynamic type %T, want float64 - every JSON number does", got["n"])
	}
	if _, ok := got["s"].(string); !ok {
		t.Errorf("DecodeAny s has dynamic type %T, want string", got["s"])
	}
	if _, ok := got["b"].(bool); !ok {
		t.Errorf("DecodeAny b has dynamic type %T, want bool", got["b"])
	}
	if got["nothing"] != nil {
		t.Errorf("DecodeAny nothing = %v, want nil", got["nothing"])
	}
	list, ok := got["list"].([]any)
	if !ok {
		t.Fatalf("DecodeAny list has dynamic type %T, want []any", got["list"])
	}
	if len(list) != 2 {
		t.Errorf("DecodeAny list = %v, want 2 elements", list)
	}
	nested, ok := got["nested"].(map[string]any)
	if !ok {
		t.Fatalf("DecodeAny nested has dynamic type %T, want map[string]any", got["nested"])
	}
	if nested["k"] != "v" {
		t.Errorf("DecodeAny nested[k] = %v, want v", nested["k"])
	}
}

func TestDecodeAnyRejectsAnArray(t *testing.T) {
	// The return type is a map, so a top-level array cannot land in it.
	if _, err := DecodeAny([]byte(`[1,2,3]`)); err == nil {
		t.Error("DecodeAny on a top-level array returned nil error, want an error")
	}
}

func TestLookupInt(t *testing.T) {
	m, err := DecodeAny([]byte(`{"count":42,"name":"go","ratio":2.5,"flag":true}`))
	if err != nil {
		t.Fatalf("DecodeAny returned an error: %v", err)
	}

	if got, ok := LookupInt(m, "count"); !ok || got != 42 {
		t.Errorf("LookupInt(count) = (%d, %t), want (42, true)", got, ok)
	}
	if got, ok := LookupInt(m, "ratio"); !ok || got != 2 {
		t.Errorf("LookupInt(ratio) = (%d, %t), want (2, true) - the conversion truncates", got, ok)
	}
	if got, ok := LookupInt(m, "name"); ok || got != 0 {
		t.Errorf(`LookupInt(name) = (%d, %t), want (0, false) - "go" is not a number`, got, ok)
	}
	if got, ok := LookupInt(m, "flag"); ok || got != 0 {
		t.Errorf("LookupInt(flag) = (%d, %t), want (0, false)", got, ok)
	}
	if got, ok := LookupInt(m, "missing"); ok || got != 0 {
		t.Errorf("LookupInt(missing) = (%d, %t), want (0, false)", got, ok)
	}
}

func TestLookupIntOnANumericStringIsAMiss(t *testing.T) {
	m := map[string]any{"count": "3"}
	if got, ok := LookupInt(m, "count"); ok || got != 0 {
		t.Errorf(`LookupInt on the string "3" = (%d, %t), want (0, false)`, got, ok)
	}
}

func TestDecodeStrictAcceptsKnownFields(t *testing.T) {
	got, err := DecodeStrict([]byte(`{"ID":5,"Title":"strict","Done":true}`))
	if err != nil {
		t.Fatalf("DecodeStrict returned an error: %v", err)
	}
	if got.ID != 5 || got.Title != "strict" || !got.Done {
		t.Errorf("DecodeStrict = %+v, want {ID:5 Title:strict Done:true}", got)
	}
}

func TestDecodeStrictRejectsAnUnknownField(t *testing.T) {
	_, err := DecodeStrict([]byte(`{"ID":5,"Titel":"typo"}`))
	if err == nil {
		t.Fatal("DecodeStrict accepted an unknown field, want an error")
	}
	if !strings.Contains(err.Error(), "Titel") {
		t.Errorf("DecodeStrict error = %q, want it to name the offending field", err)
	}
}

func TestDecodeStrictStillFillsPartialInput(t *testing.T) {
	// Strict is about fields that are *there* and unknown, not about fields
	// that are missing.
	got, err := DecodeStrict([]byte(`{"Title":"partial"}`))
	if err != nil {
		t.Fatalf("DecodeStrict returned an error: %v", err)
	}
	if got.Title != "partial" || got.ID != 0 {
		t.Errorf("DecodeStrict = %+v, want {ID:0 Title:partial}", got)
	}
}

func TestDecodeStream(t *testing.T) {
	in := `{"ID":1,"Title":"one"}{"ID":2,"Title":"two"}
	{"ID":3,"Title":"three"}`
	got, err := DecodeStream(strings.NewReader(in))
	if err != nil {
		t.Fatalf("DecodeStream returned an error: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("DecodeStream returned %d tasks, want 3", len(got))
	}
	wantTitles := []string{"one", "two", "three"}
	for i, want := range wantTitles {
		if got[i].ID != i+1 || got[i].Title != want {
			t.Errorf("DecodeStream[%d] = %+v, want ID %d and Title %q", i, got[i], i+1, want)
		}
	}
}

func TestDecodeStreamOfOneValue(t *testing.T) {
	got, err := DecodeStream(strings.NewReader(`{"ID":9}`))
	if err != nil {
		t.Fatalf("DecodeStream returned an error: %v", err)
	}
	if len(got) != 1 || got[0].ID != 9 {
		t.Errorf("DecodeStream = %+v, want one task with ID 9", got)
	}
}

func TestDecodeStreamOfNothingIsEmptyNotNil(t *testing.T) {
	got, err := DecodeStream(strings.NewReader(""))
	if err != nil {
		t.Fatalf("DecodeStream on an empty stream returned an error: %v", err)
	}
	if got == nil {
		t.Fatal("DecodeStream on an empty stream returned nil, want an empty non-nil slice")
	}
	if len(got) != 0 {
		t.Errorf("DecodeStream on an empty stream = %v, want length 0", got)
	}
}

func TestDecodeStreamReportsAMalformedValue(t *testing.T) {
	_, err := DecodeStream(strings.NewReader(`{"ID":1}{"ID":`))
	if err == nil {
		t.Error("DecodeStream on a truncated second value returned nil error, want an error")
	}
}
