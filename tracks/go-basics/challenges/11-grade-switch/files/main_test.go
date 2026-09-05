package main

import (
	"slices"
	"testing"
)

func TestGrade(t *testing.T) {
	tests := []struct {
		score int
		want  string
	}{
		{score: 100, want: "A"},
		{score: 90, want: "A"},
		{score: 89, want: "B"},
		{score: 80, want: "B"},
		{score: 79, want: "C"},
		{score: 70, want: "C"},
		{score: 69, want: "D"},
		{score: 60, want: "D"},
		{score: 59, want: "F"},
		{score: 0, want: "F"},
		{score: 101, want: "?"},
		{score: -1, want: "?"},
	}

	for _, tt := range tests {
		if got := Grade(tt.score); got != tt.want {
			t.Errorf("Grade(%d) = %q, want %q", tt.score, got, tt.want)
		}
	}
}

func TestDayKind(t *testing.T) {
	tests := []struct {
		day  string
		want string
	}{
		{day: "mon", want: "weekday"},
		{day: "tue", want: "weekday"},
		{day: "wed", want: "weekday"},
		{day: "thu", want: "weekday"},
		{day: "fri", want: "weekday"},
		{day: "sat", want: "weekend"},
		{day: "sun", want: "weekend"},
		{day: "Mon", want: "unknown"},
		{day: "", want: "unknown"},
		{day: "caturday", want: "unknown"},
	}

	for _, tt := range tests {
		if got := DayKind(tt.day); got != tt.want {
			t.Errorf("DayKind(%q) = %q, want %q", tt.day, got, tt.want)
		}
	}
}

func TestPerks(t *testing.T) {
	tests := []struct {
		tier string
		want []string
	}{
		{tier: "gold", want: []string{"lounge", "priority", "points"}},
		{tier: "silver", want: []string{"priority", "points"}},
		{tier: "bronze", want: []string{"points"}},
		{tier: "none", want: nil},
		{tier: "", want: nil},
	}

	for _, tt := range tests {
		got := Perks(tt.tier)
		if len(got) != len(tt.want) || !slices.Equal(got, tt.want) {
			t.Errorf("Perks(%q) = %q, want %q", tt.tier, got, tt.want)
		}
	}
}

func TestSeason(t *testing.T) {
	tests := []struct {
		month int
		want  string
	}{
		{month: 12, want: "winter"},
		{month: 1, want: "winter"},
		{month: 2, want: "winter"},
		{month: 3, want: "spring"},
		{month: 5, want: "spring"},
		{month: 6, want: "summer"},
		{month: 8, want: "summer"},
		{month: 9, want: "autumn"},
		{month: 11, want: "autumn"},
		{month: 0, want: "unknown"},
		{month: 13, want: "unknown"},
	}

	for _, tt := range tests {
		if got := Season(tt.month); got != tt.want {
			t.Errorf("Season(%d) = %q, want %q", tt.month, got, tt.want)
		}
	}
}
