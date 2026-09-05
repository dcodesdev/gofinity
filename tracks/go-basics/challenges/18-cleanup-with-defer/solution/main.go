package main

import "fmt"

// WithResource acquires name, runs body, and releases name afterwards - on
// every exit path, including a panic in body. A panic in body still reaches
// the caller once the release has happened.
func WithResource(name string, acquire, release func(string), body func()) {
	acquire(name)
	// The defer goes on the line after the acquire, before anything can go
	// wrong in between. It runs while the panic is still unwinding, which is
	// the whole reason cleanup lives here rather than at the bottom.
	defer release(name)
	body()
}

// EachResource acquires each name, runs body on it, and releases it *before*
// moving on to the next one, so at most one resource is held at a time.
func EachResource(names []string, acquire, release func(string), body func(string)) {
	for _, name := range names {
		// defer is scoped to the *function*, not to the loop body, so deferring
		// the release here would hold every resource until EachResource
		// returned. A function per iteration gives each one its own exit.
		WithResource(name, acquire, release, func() { body(name) })
	}
}

// Nested acquires "outer" then "inner" and releases them in the reverse order:
// inner first, then outer.
func Nested(acquire, release func(string)) {
	acquire("outer")
	defer release("outer")
	acquire("inner")
	defer release("inner")
}

// TryUse acquires name, runs body, and always releases name. A panic in body
// is turned into an error of the form "use <name>: <panic value>" instead of
// escaping. TryUse returns nil when body returns normally.
func TryUse(name string, acquire, release func(string), body func()) (err error) {
	acquire(name)
	defer release(name)
	// Deferred second, so it runs first, while the panic is still unwinding.
	// By the time release runs, err has already been set.
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("use %s: %v", name, r)
		}
	}()
	body()
	return nil
}

func main() {
	acquire := func(name string) { fmt.Println("acquire", name) }
	release := func(name string) { fmt.Println("release", name) }
	EachResource([]string{"a", "b"}, acquire, release, func(string) {})
	fmt.Println(TryUse("c", acquire, release, func() { panic("boom") }))
}
