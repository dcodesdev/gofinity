package main

import "fmt"

// Failure is an error type. Its methods take a pointer receiver, so *Failure
// is what satisfies error - and a *Failure can be nil.
type Failure struct {
	Msg string
}

// Error returns the message, and survives a nil receiver: a method on a nil
// pointer runs fine as long as it does not dereference. Return
// "<no failure>" when f is nil.
func (f *Failure) Error() string {
	if f == nil {
		return "<no failure>"
	}
	return f.Msg
}

// Broken is already written, and it is wrong on purpose. It declares a nil
// *Failure and returns it as an error, so the interface it hands back has a
// type in it - *Failure - and a nil value. An interface is nil only when both
// halves are nil, so `Broken() != nil` even though nothing failed.
//
// Read it, run the test that pins it down, then never write it again.
func Broken(fail bool) error {
	var f *Failure
	if fail {
		f = &Failure{Msg: "broken"}
	}
	return f
}

// Fixed is the same function done right: return the untyped nil when there is
// nothing to report, and only build an interface value when there is.
func Fixed(fail bool) error {
	if !fail {
		// A bare nil, not a nil *Failure. This is the whole fix.
		return nil
	}
	return &Failure{Msg: "broken"}
}

// Wrap turns a *Failure into an error, mapping a nil pointer to a nil error.
// This is the guard you write when a concrete pointer has to cross into an
// interface.
func Wrap(f *Failure) error {
	if f == nil {
		return nil
	}
	return f
}

// HoldsNilPointer reports whether err is a non-nil interface whose value is a
// nil *Failure - the trap itself. A truly nil error is false, and so is a real
// *Failure.
func HoldsNilPointer(err error) bool {
	if err == nil {
		return false
	}
	// The assertion recovers the pointer, and the pointer can then be
	// compared to nil on its own terms.
	f, ok := err.(*Failure)
	return ok && f == nil
}

// Compact returns the errors that are really errors: it drops the nil ones and
// the ones caught by HoldsNilPointer. It returns an empty (or nil) slice when
// nothing survives, and never returns the input slice itself.
func Compact(errs []error) []error {
	out := make([]error, 0, len(errs))
	for _, err := range errs {
		if err == nil || HoldsNilPointer(err) {
			continue
		}
		out = append(out, err)
	}
	return out
}

// FirstError returns the first real error in errs, or nil when there is none.
// "Real" means the same thing it means in Compact.
func FirstError(errs []error) error {
	for _, err := range errs {
		if err == nil || HoldsNilPointer(err) {
			continue
		}
		return err
	}
	// Returning a bare nil again: returning errs[i] from a loop that found
	// nothing would put the trap right back.
	return nil
}

func main() {
	fmt.Println(Broken(false) == nil) // false, and that is the bug
	fmt.Println(Fixed(false) == nil)  // true
}
