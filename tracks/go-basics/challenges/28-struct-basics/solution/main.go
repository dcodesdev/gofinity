package main

import "fmt"

// Book is one book on a shelf. Every field is exported, so the tests in this
// package - and any other package - can read and set them by name.
type Book struct {
	Title  string
	Author string
	Pages  int
}

// NewBook builds a Book from its three parts. Use a composite literal with
// field names.
func NewBook(title, author string, pages int) Book {
	// Naming the fields means adding a fourth field later cannot silently
	// shift the arguments of every literal in the program.
	return Book{Title: title, Author: author, Pages: pages}
}

// Describe renders a book as "Title by Author (N pages)". It does no special
// casing: the zero Book renders as " by  (0 pages)", because a struct's zero
// value is every field's zero value.
func Describe(b Book) string {
	return fmt.Sprintf("%s by %s (%d pages)", b.Title, b.Author, b.Pages)
}

// AddPages returns a copy of b with n more pages. The book it was given must
// come back unchanged, because a struct argument is passed by value.
func AddPages(b Book, n int) Book {
	// b is already a copy, so writing to it cannot touch the caller's book.
	b.Pages += n
	return b
}

// TotalPages adds up the pages of every book. An empty or nil slice gives 0.
func TotalPages(books []Book) int {
	total := 0
	for _, b := range books {
		total += b.Pages
	}
	return total
}

// Longest returns the book with the most pages, and false when there are none
// to choose from. On a tie the earliest book wins.
func Longest(books []Book) (Book, bool) {
	if len(books) == 0 {
		return Book{}, false
	}
	best := books[0]
	for _, b := range books[1:] {
		// Strictly greater, so an equal page count leaves the earlier book in
		// place.
		if b.Pages > best.Pages {
			best = b
		}
	}
	return best, true
}

// ByAuthor returns every book whose Author is exactly author, in their original
// order. No match gives an empty slice.
func ByAuthor(books []Book, author string) []Book {
	out := []Book{}
	for _, b := range books {
		if b.Author == author {
			out = append(out, b)
		}
	}
	return out
}

func main() {
	shelf := []Book{
		NewBook("The Go Programming Language", "Donovan", 380),
		NewBook("Learning Go", "Bodner", 376),
	}
	for _, b := range shelf {
		fmt.Println(Describe(b))
	}
	fmt.Println(TotalPages(shelf), "pages in total")
}
