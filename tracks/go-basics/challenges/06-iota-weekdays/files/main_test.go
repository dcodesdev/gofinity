package main

import (
	"fmt"
	"testing"
)

func TestDayNumbers(t *testing.T) {
	tests := []struct {
		day  Weekday
		want int
	}{
		{day: Sunday, want: 0},
		{day: Monday, want: 1},
		{day: Tuesday, want: 2},
		{day: Wednesday, want: 3},
		{day: Thursday, want: 4},
		{day: Friday, want: 5},
		{day: Saturday, want: 6},
	}

	for _, tt := range tests {
		if int(tt.day) != tt.want {
			t.Errorf("int(%s) = %d, want %d", tt.day, int(tt.day), tt.want)
		}
	}
}

func TestString(t *testing.T) {
	tests := []struct {
		day  Weekday
		want string
	}{
		{day: Sunday, want: "Sunday"},
		{day: Wednesday, want: "Wednesday"},
		{day: Saturday, want: "Saturday"},
		{day: Weekday(9), want: "Unknown(9)"},
		{day: Weekday(-1), want: "Unknown(-1)"},
	}

	for _, tt := range tests {
		if got := tt.day.String(); got != tt.want {
			t.Errorf("Weekday(%d).String() = %q, want %q", int(tt.day), got, tt.want)
		}
	}
}

func TestStringerIsUsedByFmt(t *testing.T) {
	if got := fmt.Sprintf("%v", Friday); got != "Friday" {
		t.Errorf(`fmt.Sprintf("%%v", Friday) = %q, want %q`, got, "Friday")
	}
	if got := fmt.Sprint(Weekday(12)); got != "Unknown(12)" {
		t.Errorf("fmt.Sprint(Weekday(12)) = %q, want %q", got, "Unknown(12)")
	}
}

func TestIsWeekend(t *testing.T) {
	tests := []struct {
		day  Weekday
		want bool
	}{
		{day: Sunday, want: true},
		{day: Monday, want: false},
		{day: Friday, want: false},
		{day: Saturday, want: true},
	}

	for _, tt := range tests {
		if got := tt.day.IsWeekend(); got != tt.want {
			t.Errorf("%s.IsWeekend() = %v, want %v", tt.day, got, tt.want)
		}
	}
}

func TestNext(t *testing.T) {
	tests := []struct {
		day  Weekday
		want Weekday
	}{
		{day: Sunday, want: Monday},
		{day: Thursday, want: Friday},
		{day: Friday, want: Saturday},
		{day: Saturday, want: Sunday},
		{day: Weekday(42), want: Weekday(42)},
	}

	for _, tt := range tests {
		if got := tt.day.Next(); got != tt.want {
			t.Errorf("%s.Next() = %s, want %s", tt.day, got, tt.want)
		}
	}
}
