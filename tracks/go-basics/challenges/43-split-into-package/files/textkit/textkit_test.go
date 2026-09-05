package textkit

import "testing"

// This file is `package textkit`, so it lives inside the package and can see
// everything in it, exported or not. The test file next to main.go is a
// different package and can only reach the capitalised names.

func TestTitle(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"hello go world", "Hello Go World"},
		{"  hello   go world ", "Hello Go World"},
		{"GO IS FUN", "Go Is Fun"},
		{"gofinity", "Gofinity"},
		{"", ""},
		{"   ", ""},
	}
	for _, c := range cases {
		if got := Title(c.in); got != c.want {
			t.Errorf("Title(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestSlug(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"Hello, Go World!", "hello-go-world"},
		{"  spaces   everywhere  ", "spaces-everywhere"},
		{"Go 1.24 release", "go-1-24-release"},
		{"---already---", "already"},
		{"!!!", ""},
		{"", ""},
	}
	for _, c := range cases {
		if got := Slug(c.in); got != c.want {
			t.Errorf("Slug(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestWordCount(t *testing.T) {
	cases := []struct {
		in   string
		want int
	}{
		{"go is fun", 3},
		{"go", 1},
		{"  lots   of   space  ", 3},
		{"", 0},
		{"   ", 0},
	}
	for _, c := range cases {
		if got := WordCount(c.in); got != c.want {
			t.Errorf("WordCount(%q) = %d, want %d", c.in, got, c.want)
		}
	}
}

func TestIsWordCharIsVisibleFromInside(t *testing.T) {
	// An unexported helper is still testable: this file is in the package.
	if !isWordChar('g') || isWordChar('-') {
		t.Error("isWordChar should accept letters and digits and reject anything else")
	}
}
