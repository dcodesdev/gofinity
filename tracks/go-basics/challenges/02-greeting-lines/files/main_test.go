package main

import "testing"

func TestGreetAll(t *testing.T) {
	tests := []struct {
		name  string
		input []string
		want  string
	}{
		{
			name:  "one name",
			input: []string{"Gofinity"},
			want:  "1. Hello, Gofinity!",
		},
		{
			name:  "three names are numbered from one",
			input: []string{"Ada", "Grace", "Alan"},
			want:  "1. Hello, Ada!\n2. Hello, Grace!\n3. Hello, Alan!",
		},
		{
			name:  "an empty name is greeted as World",
			input: []string{"Ada", "", "Alan"},
			want:  "1. Hello, Ada!\n2. Hello, World!\n3. Hello, Alan!",
		},
		{
			name:  "no names at all",
			input: []string{},
			want:  "",
		},
		{
			name:  "a nil slice behaves like an empty one",
			input: nil,
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GreetAll(tt.input); got != tt.want {
				t.Errorf("GreetAll(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestGreetAllHasNoTrailingNewline(t *testing.T) {
	got := GreetAll([]string{"Ada", "Grace"})
	if len(got) > 0 && got[len(got)-1] == '\n' {
		t.Errorf("GreetAll(...) = %q, want no trailing newline", got)
	}
}
