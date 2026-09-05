package main

import (
	"slices"
	"testing"
)

func TestCounter(t *testing.T) {
	next := Counter()
	for want := 1; want <= 4; want++ {
		if got := next(); got != want {
			t.Fatalf("call %d of the counter = %d, want %d", want, got, want)
		}
	}
}

func TestCountersAreIndependent(t *testing.T) {
	a, b := Counter(), Counter()
	a()
	a()
	if got := a(); got != 3 {
		t.Errorf("third call of a = %d, want 3", got)
	}
	if got := b(); got != 1 {
		t.Errorf("first call of b = %d, want 1 (each Counter owns its own state)", got)
	}
}

func TestAccumulator(t *testing.T) {
	add := Accumulator()
	steps := []struct {
		in   int
		want int
	}{
		{in: 10, want: 10},
		{in: 5, want: 15},
		{in: -15, want: 0},
		{in: 0, want: 0},
		{in: 7, want: 7},
	}

	for _, step := range steps {
		if got := add(step.in); got != step.want {
			t.Errorf("add(%d) = %d, want %d", step.in, got, step.want)
		}
	}

	if got := Accumulator()(1); got != 1 {
		t.Errorf("a fresh accumulator started at %d, want it to start at 0", got-1)
	}
}

func TestMultiplier(t *testing.T) {
	triple := Multiplier(3)
	if got := triple(4); got != 12 {
		t.Errorf("Multiplier(3)(4) = %d, want 12", got)
	}
	if got := triple(0); got != 0 {
		t.Errorf("Multiplier(3)(0) = %d, want 0", got)
	}
	if got := Multiplier(-2)(5); got != -10 {
		t.Errorf("Multiplier(-2)(5) = %d, want -10", got)
	}
	if got := Multiplier(0)(9); got != 0 {
		t.Errorf("Multiplier(0)(9) = %d, want 0", got)
	}
}

func TestApply(t *testing.T) {
	nums := []int{1, 2, 3}
	got := Apply(nums, Multiplier(10))
	want := []int{10, 20, 30}
	if !slices.Equal(got, want) {
		t.Errorf("Apply(%v, Multiplier(10)) = %v, want %v", nums, got, want)
	}
	if !slices.Equal(nums, []int{1, 2, 3}) {
		t.Errorf("Apply changed its input: %v, want [1 2 3]", nums)
	}

	if got := Apply(nil, Multiplier(2)); len(got) != 0 {
		t.Errorf("Apply(nil, ...) = %v, want an empty result", got)
	}

	square := func(n int) int { return n * n }
	if got := Apply([]int{-3, 4}, square); !slices.Equal(got, []int{9, 16}) {
		t.Errorf("Apply([-3 4], square) = %v, want [9 16]", got)
	}
}

func TestCompose(t *testing.T) {
	double := Multiplier(2)
	increment := func(n int) int { return n + 1 }

	if got := Compose(double, increment)(3); got != 8 {
		t.Errorf("Compose(double, increment)(3) = %d, want 8 (g runs first)", got)
	}
	if got := Compose(increment, double)(3); got != 7 {
		t.Errorf("Compose(increment, double)(3) = %d, want 7", got)
	}
	if got := Compose(double, double)(5); got != 20 {
		t.Errorf("Compose(double, double)(5) = %d, want 20", got)
	}
}

func TestCountdown(t *testing.T) {
	next := Countdown(3)
	steps := []struct {
		want   int
		wantOk bool
	}{
		{want: 3, wantOk: true},
		{want: 2, wantOk: true},
		{want: 1, wantOk: true},
		{want: 0, wantOk: false},
		{want: 0, wantOk: false},
	}

	for i, step := range steps {
		got, ok := next()
		if got != step.want || ok != step.wantOk {
			t.Errorf("call %d = %d, %t, want %d, %t", i+1, got, ok, step.want, step.wantOk)
		}
	}

	if got, ok := Countdown(0)(); got != 0 || ok {
		t.Errorf("Countdown(0)() = %d, %t, want 0, false", got, ok)
	}
	if got, ok := Countdown(-2)(); got != 0 || ok {
		t.Errorf("Countdown(-2)() = %d, %t, want 0, false", got, ok)
	}
}

func TestMultipliers(t *testing.T) {
	fns := Multipliers([]int{2, 3, 10})
	if len(fns) != 3 {
		t.Fatalf("Multipliers([2 3 10]) returned %d functions, want 3", len(fns))
	}

	// Every closure must have captured its own n. Getting [10 10 10] here means
	// they all share one variable.
	want := []int{2, 3, 10}
	for i, f := range fns {
		if got := f(1); got != want[i] {
			t.Errorf("fns[%d](1) = %d, want %d", i, got, want[i])
		}
	}

	if got := Multipliers(nil); len(got) != 0 {
		t.Errorf("Multipliers(nil) returned %d functions, want 0", len(got))
	}
}
