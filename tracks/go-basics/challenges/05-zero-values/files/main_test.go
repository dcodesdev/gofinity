package main

import "testing"

const wantReport = "int: 0\n" +
	"float64: 0\n" +
	"string: \"\"\n" +
	"bool: false\n" +
	"slice: [] nil=true len=0\n" +
	"map: map[] nil=true len=0\n" +
	"pointer: <nil>"

func TestZeroReport(t *testing.T) {
	if got := ZeroReport(); got != wantReport {
		t.Errorf("ZeroReport() =\n%q\nwant\n%q", got, wantReport)
	}
}

func TestSumOrZero(t *testing.T) {
	tests := []struct {
		name  string
		input []int
		want  int
	}{
		{name: "nil slice", input: nil, want: 0},
		{name: "empty slice", input: []int{}, want: 0},
		{name: "one number", input: []int{7}, want: 7},
		{name: "several numbers", input: []int{1, 2, 3, 4}, want: 10},
		{name: "negatives cancel", input: []int{5, -5, 2}, want: 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SumOrZero(tt.input); got != tt.want {
				t.Errorf("SumOrZero(%v) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestGrowFromNil(t *testing.T) {
	if got := GrowFromNil(); got != nil {
		t.Errorf("GrowFromNil() = %v, want nil", got)
	}

	got := GrowFromNil(3, 1, 2)
	want := []int{3, 1, 2}
	if len(got) != len(want) {
		t.Fatalf("GrowFromNil(3, 1, 2) = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("GrowFromNil(3, 1, 2) = %v, want %v", got, want)
		}
	}
}
