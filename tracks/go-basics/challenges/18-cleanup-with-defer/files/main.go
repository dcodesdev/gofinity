package main

import "fmt"

// WithResource acquires name, runs body, and releases name afterwards - on
// every exit path, including a panic in body. A panic in body still reaches
// the caller once the release has happened.
func WithResource(name string, acquire, release func(string), body func()) {
	// TODO
}

// EachResource acquires each name, runs body on it, and releases it *before*
// moving on to the next one, so at most one resource is held at a time.
func EachResource(names []string, acquire, release func(string), body func(string)) {
	// TODO
}

// Nested acquires "outer" then "inner" and releases them in the reverse order:
// inner first, then outer.
func Nested(acquire, release func(string)) {
	// TODO
}

// TryUse acquires name, runs body, and always releases name. A panic in body
// is turned into an error of the form "use <name>: <panic value>" instead of
// escaping. TryUse returns nil when body returns normally.
func TryUse(name string, acquire, release func(string), body func()) (err error) {
	// TODO
	return nil
}

func main() {
	acquire := func(name string) { fmt.Println("acquire", name) }
	release := func(name string) { fmt.Println("release", name) }
	EachResource([]string{"a", "b"}, acquire, release, func(string) {})
	fmt.Println(TryUse("c", acquire, release, func() { panic("boom") }))
}
