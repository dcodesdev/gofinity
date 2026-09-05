package main

import (
	"slices"
	"testing"
)

func TestFizzBuzz(t *testing.T) {
	tests := []struct {
		n    int
		want string
	}{
		{n: 1, want: "1"},
		{n: 2, want: "2"},
		{n: 3, want: "Fizz"},
		{n: 5, want: "Buzz"},
		{n: 6, want: "Fizz"},
		{n: 9, want: "Fizz"},
		{n: 10, want: "Buzz"},
		{n: 15, want: "FizzBuzz"},
		{n: 30, want: "FizzBuzz"},
		{n: 7, want: "7"},
		{n: 98, want: "98"},
		{n: 0, want: "FizzBuzz"},
		{n: -3, want: "Fizz"},
		{n: -7, want: "-7"},
	}

	for _, tt := range tests {
		if got := FizzBuzz(tt.n); got != tt.want {
			t.Errorf("FizzBuzz(%d) = %q, want %q", tt.n, got, tt.want)
		}
	}
}

func TestFizzBuzzUpTo(t *testing.T) {
	want := []string{
		"1", "2", "Fizz", "4", "Buzz", "Fizz", "7", "8",
		"Fizz", "Buzz", "11", "Fizz", "13", "14", "FizzBuzz",
	}

	got := FizzBuzzUpTo(15)
	if !slices.Equal(got, want) {
		t.Errorf("FizzBuzzUpTo(15) = %q, want %q", got, want)
	}

	if got := FizzBuzzUpTo(1); !slices.Equal(got, []string{"1"}) {
		t.Errorf("FizzBuzzUpTo(1) = %q, want [1]", got)
	}

	for _, n := range []int{0, -1} {
		if got := FizzBuzzUpTo(n); len(got) != 0 {
			t.Errorf("FizzBuzzUpTo(%d) = %q, want no elements", n, got)
		}
	}
}

func TestIsLeapYear(t *testing.T) {
	tests := []struct {
		y    int
		want bool
	}{
		{y: 2024, want: true},
		{y: 2023, want: false},
		{y: 2000, want: true},
		{y: 1900, want: false},
		{y: 2100, want: false},
		{y: 1600, want: true},
		{y: 1996, want: true},
	}

	for _, tt := range tests {
		if got := IsLeapYear(tt.y); got != tt.want {
			t.Errorf("IsLeapYear(%d) = %v, want %v", tt.y, got, tt.want)
		}
	}
}
