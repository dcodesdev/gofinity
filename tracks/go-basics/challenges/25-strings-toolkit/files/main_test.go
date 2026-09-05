package main

import (
	"slices"
	"testing"
)

func TestNormalize(t *testing.T) {
	tests := []struct{ in, want string }{
		{"  Hello   WORLD \n", "hello world"},
		{"go", "go"},
		{"", ""},
		{"   \t\n ", ""},
		{"One\tTwo\nThree", "one two three"},
		{"already normal", "already normal"},
	}
	for _, tt := range tests {
		if got := Normalize(tt.in); got != tt.want {
			t.Errorf("Normalize(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestCountWords(t *testing.T) {
	tests := []struct {
		in   string
		want int
	}{
		{"", 0},
		{"   ", 0},
		{"one", 1},
		{"one two  three", 3},
		{"\tone\ntwo ", 2},
	}
	for _, tt := range tests {
		if got := CountWords(tt.in); got != tt.want {
			t.Errorf("CountWords(%q) = %d, want %d", tt.in, got, tt.want)
		}
	}
}

func TestTitleWords(t *testing.T) {
	tests := []struct{ in, want string }{
		{"hello wide world", "Hello Wide World"},
		{"  go   is fun ", "Go Is Fun"},
		{"gRPC over http", "GRPC Over Http"},
		{"élodie writes go", "Élodie Writes Go"},
		{"a", "A"},
		{"", ""},
	}
	for _, tt := range tests {
		if got := TitleWords(tt.in); got != tt.want {
			t.Errorf("TitleWords(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestInitials(t *testing.T) {
	tests := []struct{ in, want string }{
		{"ada lovelace", "AL"},
		{"  grace   brewster  murray hopper ", "GBMH"},
		{"élodie king", "ÉK"},
		{"cher", "C"},
		{"", ""},
	}
	for _, tt := range tests {
		if got := Initials(tt.in); got != tt.want {
			t.Errorf("Initials(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestRedact(t *testing.T) {
	tests := []struct{ s, secret, want string }{
		{"token=abc123 ok", "abc123", "token=****** ok"},
		{"aa bb aa", "aa", "** bb **"},
		{"nothing here", "zzz", "nothing here"},
		{"leave me alone", "", "leave me alone"},
		{"", "abc", ""},
	}
	for _, tt := range tests {
		if got := Redact(tt.s, tt.secret); got != tt.want {
			t.Errorf("Redact(%q, %q) = %q, want %q", tt.s, tt.secret, got, tt.want)
		}
	}
}

func TestSplitTrim(t *testing.T) {
	tests := []struct {
		s, sep string
		want   []string
	}{
		{"a, b ,, c", ",", []string{"a", "b", "c"}},
		{" one ", ",", []string{"one"}},
		{"", ",", []string{}},
		{"  ,  , ", ",", []string{}},
		{"a::b:: :c", "::", []string{"a", "b", ":c"}},
	}
	for _, tt := range tests {
		got := SplitTrim(tt.s, tt.sep)
		if got == nil {
			t.Errorf("SplitTrim(%q, %q) = nil, want a non-nil slice", tt.s, tt.sep)
			continue
		}
		if !slices.Equal(got, tt.want) {
			t.Errorf("SplitTrim(%q, %q) = %q, want %q", tt.s, tt.sep, got, tt.want)
		}
	}
}
