package main

import "testing"

func TestHeadline(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"hello go world", "Hello Go World (hello-go-world)"},
		{"Go 1.24 release", "Go 1.24 Release (go-1-24-release)"},
		{"gofinity", "Gofinity (gofinity)"},
	}
	for _, c := range cases {
		if got := Headline(c.in); got != c.want {
			t.Errorf("Headline(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestSummary(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"go is fun", "3 words"},
		{"go", "1 word"},
		{"", "no words"},
		{"   ", "no words"},
	}
	for _, c := range cases {
		if got := Summary(c.in); got != c.want {
			t.Errorf("Summary(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
