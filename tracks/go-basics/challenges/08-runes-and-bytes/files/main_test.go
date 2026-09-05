package main

import "testing"

func TestByteLen(t *testing.T) {
	tests := []struct {
		in   string
		want int
	}{
		{in: "", want: 0},
		{in: "go", want: 2},
		{in: "héllo", want: 6},
		{in: "日本語", want: 9},
		{in: "°C", want: 3},
	}

	for _, tt := range tests {
		if got := ByteLen(tt.in); got != tt.want {
			t.Errorf("ByteLen(%q) = %d, want %d", tt.in, got, tt.want)
		}
	}
}

func TestRuneLen(t *testing.T) {
	tests := []struct {
		in   string
		want int
	}{
		{in: "", want: 0},
		{in: "go", want: 2},
		{in: "héllo", want: 5},
		{in: "日本語", want: 3},
		{in: "°C", want: 2},
	}

	for _, tt := range tests {
		if got := RuneLen(tt.in); got != tt.want {
			t.Errorf("RuneLen(%q) = %d, want %d", tt.in, got, tt.want)
		}
	}
}

func TestRuneAt(t *testing.T) {
	tests := []struct {
		s      string
		i      int
		want   rune
		wantOK bool
	}{
		{s: "héllo", i: 0, want: 'h', wantOK: true},
		{s: "héllo", i: 1, want: 'é', wantOK: true},
		{s: "héllo", i: 4, want: 'o', wantOK: true},
		{s: "héllo", i: 5, wantOK: false},
		{s: "héllo", i: -1, wantOK: false},
		{s: "日本語", i: 2, want: '語', wantOK: true},
		{s: "", i: 0, wantOK: false},
	}

	for _, tt := range tests {
		got, ok := RuneAt(tt.s, tt.i)
		if ok != tt.wantOK || (ok && got != tt.want) {
			t.Errorf("RuneAt(%q, %d) = (%q, %v), want (%q, %v)",
				tt.s, tt.i, got, ok, tt.want, tt.wantOK)
		}
	}
}

func TestReverseRunes(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{in: "", want: ""},
		{in: "a", want: "a"},
		{in: "gofinity", want: "ytinifog"},
		{in: "héllo", want: "olléh"},
		{in: "日本語", want: "語本日"},
	}

	for _, tt := range tests {
		if got := ReverseRunes(tt.in); got != tt.want {
			t.Errorf("ReverseRunes(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestReverseRunesTwiceIsIdentity(t *testing.T) {
	for _, s := range []string{"héllo", "日本語 text", "°C and °F"} {
		if got := ReverseRunes(ReverseRunes(s)); got != s {
			t.Errorf("ReverseRunes(ReverseRunes(%q)) = %q, want %q", s, got, s)
		}
	}
}
