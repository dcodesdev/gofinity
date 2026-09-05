package main

import (
	"errors"
	"reflect"
	"testing"
)

// The pointer is what satisfies error, not the struct.
var _ error = (*ParseError)(nil)

func TestParseErrorMessage(t *testing.T) {
	withKey := &ParseError{Line: 3, Key: "host", Err: ErrMissingValue}
	if want := `line 3: key "host": missing value`; withKey.Error() != want {
		t.Errorf("Error() = %q, want %q", withKey.Error(), want)
	}
	noKey := &ParseError{Line: 2, Err: ErrMalformed}
	if want := "line 2: missing '='"; noKey.Error() != want {
		t.Errorf("Error() = %q, want %q", noKey.Error(), want)
	}
}

func TestParseErrorUnwraps(t *testing.T) {
	err := error(&ParseError{Line: 1, Key: "n", Err: ErrNotNumeric})
	if !errors.Is(err, ErrNotNumeric) {
		t.Error("errors.Is(err, ErrNotNumeric) = false - Unwrap must return e.Err")
	}
	if errors.Is(err, ErrMalformed) {
		t.Error("errors.Is(err, ErrMalformed) = true, want false")
	}
	var pe *ParseError
	if !errors.As(err, &pe) {
		t.Fatal("errors.As did not find the *ParseError")
	}
	if pe.Line != 1 || pe.Key != "n" {
		t.Errorf("errors.As gave %+v, want Line 1 and Key \"n\"", pe)
	}
}

func TestParseLineSuccess(t *testing.T) {
	key, value, err := ParseLine(1, "port=8080")
	if err != nil {
		t.Fatalf("ParseLine returned %v, want no error", err)
	}
	if key != "port" || value != 8080 {
		t.Errorf(`ParseLine = (%q, %d), want ("port", 8080)`, key, value)
	}
	if key, value, err := ParseLine(4, "zero=0"); err != nil || key != "zero" || value != 0 {
		t.Errorf("ParseLine on zero=0 = (%q, %d, %v), want (\"zero\", 0, nil)", key, value, err)
	}
}

func TestParseLineFailures(t *testing.T) {
	cases := []struct {
		line     string
		wantKey  string
		wantErr  error
		wantText string
	}{
		{"oops", "", ErrMalformed, "line 7: missing '='"},
		{"=8080", "", ErrMalformed, "line 7: missing '='"},
		{"host=", "host", ErrMissingValue, `line 7: key "host": missing value`},
		{"n=abc", "n", ErrNotNumeric, `line 7: key "n": not numeric`},
		{"n=-1", "n", ErrNotNumeric, `line 7: key "n": not numeric`},
	}
	for _, tc := range cases {
		key, value, err := ParseLine(7, tc.line)
		if err == nil {
			t.Errorf("ParseLine(7, %q) returned no error, want one", tc.line)
			continue
		}
		if !errors.Is(err, tc.wantErr) {
			t.Errorf("ParseLine(7, %q) error = %v, want %v in the chain", tc.line, err, tc.wantErr)
		}
		if err.Error() != tc.wantText {
			t.Errorf("ParseLine(7, %q) error = %q, want %q", tc.line, err.Error(), tc.wantText)
		}
		var pe *ParseError
		if !errors.As(err, &pe) {
			t.Errorf("ParseLine(7, %q) is not a *ParseError", tc.line)
			continue
		}
		if pe.Line != 7 || pe.Key != tc.wantKey {
			t.Errorf("ParseLine(7, %q) = %+v, want Line 7 and Key %q", tc.line, pe, tc.wantKey)
		}
		if key != "" || value != 0 {
			t.Errorf("ParseLine(7, %q) = (%q, %d) beside an error, want the zero values", tc.line, key, value)
		}
	}
}

func TestParseAllSuccess(t *testing.T) {
	cfg, err := ParseAll([]string{"port=8080", "", "retries=3"})
	if err != nil {
		t.Fatalf("ParseAll returned %v, want no error", err)
	}
	want := map[string]int{"port": 8080, "retries": 3}
	if !reflect.DeepEqual(cfg, want) {
		t.Errorf("ParseAll = %v, want %v", cfg, want)
	}
	if cfg, err := ParseAll(nil); err != nil || len(cfg) != 0 {
		t.Errorf("ParseAll(nil) = (%v, %v), want an empty map and no error", cfg, err)
	}
}

func TestParseAllCollectsEveryError(t *testing.T) {
	cfg, err := ParseAll([]string{"port=8080", "oops", "host=", "retries=abc"})
	if err == nil {
		t.Fatal("ParseAll returned no error, want three joined")
	}
	if cfg != nil {
		t.Errorf("ParseAll returned %v beside an error, want a nil map", cfg)
	}
	want := "line 2: missing '='\nline 3: key \"host\": missing value\nline 4: key \"retries\": not numeric"
	if err.Error() != want {
		t.Errorf("ParseAll error =\n%q\nwant\n%q", err.Error(), want)
	}
	for _, sentinel := range []error{ErrMalformed, ErrMissingValue, ErrNotNumeric} {
		if !errors.Is(err, sentinel) {
			t.Errorf("errors.Is(err, %v) = false - errors.Is searches every branch of a join", sentinel)
		}
	}
}

func TestLineOf(t *testing.T) {
	if got := LineOf(nil); got != 0 {
		t.Errorf("LineOf(nil) = %d, want 0", got)
	}
	if got := LineOf(errors.New("unrelated")); got != 0 {
		t.Errorf("LineOf of an unrelated error = %d, want 0", got)
	}
	_, err := ParseAll([]string{"a=1", "oops", "host="})
	if got := LineOf(err); got != 2 {
		t.Errorf("LineOf = %d, want 2 - the first *ParseError in the join", got)
	}
	_, _, one := ParseLine(9, "oops")
	if got := LineOf(one); got != 9 {
		t.Errorf("LineOf of a bare *ParseError = %d, want 9", got)
	}
}

func TestLines(t *testing.T) {
	if got := Lines(nil); len(got) != 0 {
		t.Errorf("Lines(nil) = %v, want nothing", got)
	}
	if got := Lines(errors.New("unrelated")); len(got) != 0 {
		t.Errorf("Lines of an unrelated error = %v, want nothing", got)
	}

	_, err := ParseAll([]string{"port=8080", "oops", "host=", "retries=abc"})
	if got, want := Lines(err), []int{2, 3, 4}; !reflect.DeepEqual(got, want) {
		t.Errorf("Lines = %v, want %v", got, want)
	}

	_, _, one := ParseLine(9, "oops")
	if got, want := Lines(one), []int{9}; !reflect.DeepEqual(got, want) {
		t.Errorf("Lines of a bare *ParseError = %v, want %v", got, want)
	}

	// A single-child wrapper has to be walked too.
	if got, want := Lines(errors.Join(one)), []int{9}; !reflect.DeepEqual(got, want) {
		t.Errorf("Lines through a join of one = %v, want %v", got, want)
	}
}

func TestExplain(t *testing.T) {
	if got := Explain(nil); got != "ok" {
		t.Errorf("Explain(nil) = %q, want %q", got, "ok")
	}
	if got := Explain(errors.New("unrelated")); got != "unknown" {
		t.Errorf("Explain of an unrelated error = %q, want %q", got, "unknown")
	}
	cases := map[string]string{
		"oops":  "malformed",
		"host=": "missing-value",
		"n=abc": "not-numeric",
	}
	for line, want := range cases {
		_, _, err := ParseLine(1, line)
		if got := Explain(err); got != want {
			t.Errorf("Explain(ParseLine(%q)) = %q, want %q", line, got, want)
		}
	}
	_, err := ParseAll([]string{"host=", "oops"})
	if got := Explain(err); got != "malformed" {
		t.Errorf("Explain of a join = %q, want %q - ErrMalformed is checked first", got, "malformed")
	}
}
