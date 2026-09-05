package main

import (
	"slices"
	"strings"
	"testing"
)

// tracker returns an acquire and a release that record what they were called
// with, plus a log of the calls in order.
func tracker() (acquire, release func(string), log *[]string) {
	var calls []string
	acquire = func(name string) { calls = append(calls, "acquire:"+name) }
	release = func(name string) { calls = append(calls, "release:"+name) }
	return acquire, release, &calls
}

func TestWithResource(t *testing.T) {
	acquire, release, log := tracker()
	ran := false

	WithResource("db", acquire, release, func() {
		ran = true
		if got := *log; !slices.Equal(got, []string{"acquire:db"}) {
			t.Errorf("log during the body = %v, want the resource already acquired", got)
		}
	})

	if !ran {
		t.Fatal("WithResource did not run the body")
	}
	want := []string{"acquire:db", "release:db"}
	if got := *log; !slices.Equal(got, want) {
		t.Errorf("log = %v, want %v", got, want)
	}
}

func TestWithResourceReleasesOnPanic(t *testing.T) {
	acquire, release, log := tracker()

	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Error("the panic did not reach the caller, want WithResource to let it through")
			}
		}()
		WithResource("db", acquire, release, func() { panic("boom") })
	}()

	want := []string{"acquire:db", "release:db"}
	if got := *log; !slices.Equal(got, want) {
		t.Errorf("log = %v, want %v: cleanup has to happen on the panicking path too", got, want)
	}
}

func TestEachResourceHoldsOneAtATime(t *testing.T) {
	acquire, release, log := tracker()
	var seen []string

	EachResource([]string{"a", "b", "c"}, acquire, release, func(name string) {
		seen = append(seen, name)
	})

	if want := []string{"a", "b", "c"}; !slices.Equal(seen, want) {
		t.Errorf("body saw %v, want %v", seen, want)
	}
	want := []string{
		"acquire:a", "release:a",
		"acquire:b", "release:b",
		"acquire:c", "release:c",
	}
	if got := *log; !slices.Equal(got, want) {
		t.Errorf("log = %v, want %v: defer runs at the end of the function, not the end of the loop body", got, want)
	}
}

func TestEachResourceWithNoNames(t *testing.T) {
	acquire, release, log := tracker()
	EachResource(nil, acquire, release, func(string) { t.Error("the body ran with no names") })
	if got := *log; len(got) != 0 {
		t.Errorf("log = %v, want nothing", got)
	}
}

func TestNestedReleasesInReverse(t *testing.T) {
	acquire, release, log := tracker()
	Nested(acquire, release)

	want := []string{"acquire:outer", "acquire:inner", "release:inner", "release:outer"}
	if got := *log; !slices.Equal(got, want) {
		t.Errorf("log = %v, want %v", got, want)
	}
}

func TestTryUseOnANormalReturn(t *testing.T) {
	acquire, release, log := tracker()

	if err := TryUse("file", acquire, release, func() {}); err != nil {
		t.Errorf("TryUse returned %v for a body that did not panic, want nil", err)
	}
	want := []string{"acquire:file", "release:file"}
	if got := *log; !slices.Equal(got, want) {
		t.Errorf("log = %v, want %v", got, want)
	}
}

func TestTryUseTurnsAPanicIntoAnError(t *testing.T) {
	acquire, release, log := tracker()

	err := TryUse("file", acquire, release, func() { panic("boom") })
	if err == nil {
		t.Fatal("TryUse returned nil for a panicking body, want an error")
	}
	if err.Error() != "use file: boom" {
		t.Errorf("TryUse returned %q, want %q", err.Error(), "use file: boom")
	}
	want := []string{"acquire:file", "release:file"}
	if got := *log; !slices.Equal(got, want) {
		t.Errorf("log = %v, want %v: the resource is released whether or not the body panicked", got, want)
	}
}

func TestTryUseNamesTheResource(t *testing.T) {
	acquire, release, _ := tracker()

	err := TryUse("socket", acquire, release, func() { panic(7) })
	if err == nil {
		t.Fatal("TryUse returned nil for a panicking body, want an error")
	}
	if !strings.HasPrefix(err.Error(), "use socket: ") {
		t.Errorf("TryUse returned %q, want it to start with %q", err.Error(), "use socket: ")
	}
	if !strings.Contains(err.Error(), "7") {
		t.Errorf("TryUse returned %q, want it to carry the panic value", err.Error())
	}
}
