package main

import "testing"

func TestGreet(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "a name", input: "Gofinity", want: "Hello, Gofinity!"},
		{name: "another name", input: "Ada", want: "Hello, Ada!"},
		{name: "empty falls back to World", input: "", want: "Hello, World!"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Greet(tt.input); got != tt.want {
				t.Errorf("Greet(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
