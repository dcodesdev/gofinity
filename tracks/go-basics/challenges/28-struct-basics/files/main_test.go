package main

import "testing"

var shelf = []Book{
	{Title: "The Go Programming Language", Author: "Donovan", Pages: 380},
	{Title: "Learning Go", Author: "Bodner", Pages: 376},
	{Title: "Go in Action", Author: "Kennedy", Pages: 264},
	{Title: "100 Go Mistakes", Author: "Bodner", Pages: 0},
}

func TestNewBook(t *testing.T) {
	got := NewBook("Learning Go", "Bodner", 376)
	want := Book{Title: "Learning Go", Author: "Bodner", Pages: 376}
	if got != want {
		t.Errorf("NewBook = %+v, want %+v", got, want)
	}

	// A struct's zero value is every field's zero value, and it needs no
	// constructor to exist.
	if got := NewBook("", "", 0); got != (Book{}) {
		t.Errorf("NewBook with empty parts = %+v, want the zero Book", got)
	}
}

func TestDescribe(t *testing.T) {
	tests := []struct {
		in   Book
		want string
	}{
		{Book{Title: "Learning Go", Author: "Bodner", Pages: 376}, "Learning Go by Bodner (376 pages)"},
		{Book{Title: "Solo", Pages: 1}, "Solo by  (1 pages)"},
		{Book{}, " by  (0 pages)"},
	}
	for _, tt := range tests {
		if got := Describe(tt.in); got != tt.want {
			t.Errorf("Describe(%+v) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestAddPages(t *testing.T) {
	original := Book{Title: "Learning Go", Author: "Bodner", Pages: 376}
	got := AddPages(original, 24)
	if want := (Book{Title: "Learning Go", Author: "Bodner", Pages: 400}); got != want {
		t.Errorf("AddPages = %+v, want %+v", got, want)
	}

	// The whole point of value semantics: the caller's book is untouched.
	if original.Pages != 376 {
		t.Errorf("AddPages changed the caller's book to %d pages - a struct argument is a copy", original.Pages)
	}

	if got := AddPages(original, -376); got.Pages != 0 {
		t.Errorf("AddPages(-376) left %d pages, want 0", got.Pages)
	}
}

func TestTotalPages(t *testing.T) {
	if got := TotalPages(shelf); got != 1020 {
		t.Errorf("TotalPages(shelf) = %d, want 1020", got)
	}
	if got := TotalPages([]Book{}); got != 0 {
		t.Errorf("TotalPages(empty) = %d, want 0", got)
	}
	if got := TotalPages(nil); got != 0 {
		t.Errorf("TotalPages(nil) = %d, want 0", got)
	}
}

func TestLongest(t *testing.T) {
	got, ok := Longest(shelf)
	if !ok {
		t.Fatal("Longest(shelf) reported no book, want ok")
	}
	if got.Title != "The Go Programming Language" {
		t.Errorf("Longest(shelf) = %q, want %q", got.Title, "The Go Programming Language")
	}

	// A tie keeps the earlier book.
	tie := []Book{{Title: "first", Pages: 10}, {Title: "second", Pages: 10}}
	if got, _ := Longest(tie); got.Title != "first" {
		t.Errorf("Longest(tie) = %q, want the earlier book %q", got.Title, "first")
	}

	got, ok = Longest(nil)
	if ok {
		t.Errorf("Longest(nil) reported ok with %+v, want false", got)
	}
	if got != (Book{}) {
		t.Errorf("Longest(nil) = %+v, want the zero Book", got)
	}
}

func TestByAuthor(t *testing.T) {
	got := ByAuthor(shelf, "Bodner")
	if len(got) != 2 {
		t.Fatalf("ByAuthor(Bodner) returned %d books, want 2", len(got))
	}
	if got[0].Title != "Learning Go" || got[1].Title != "100 Go Mistakes" {
		t.Errorf("ByAuthor(Bodner) = %+v, want them in shelf order", got)
	}

	if got := ByAuthor(shelf, "bodner"); len(got) != 0 {
		t.Errorf("ByAuthor is case-insensitive: %+v, want an exact match only", got)
	}
	if got := ByAuthor(nil, "Bodner"); len(got) != 0 {
		t.Errorf("ByAuthor(nil) = %+v, want nothing", got)
	}
}
