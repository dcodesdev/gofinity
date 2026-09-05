package main

import (
	"errors"
	"strings"
	"testing"
)

func TestSafeDivideOnGoodInput(t *testing.T) {
	cases := []struct{ a, b, want int }{
		{a: 7, b: 2, want: 3},
		{a: -9, b: 3, want: -3},
		{a: 0, b: 5, want: 0},
	}

	for _, tc := range cases {
		got, err := SafeDivide(tc.a, tc.b)
		if err != nil {
			t.Errorf("SafeDivide(%d, %d) returned error %v, want nil", tc.a, tc.b, err)
		}
		if got != tc.want {
			t.Errorf("SafeDivide(%d, %d) = %d, want %d", tc.a, tc.b, got, tc.want)
		}
	}
}

func TestSafeDivideRecoversTheRuntimePanic(t *testing.T) {
	got, err := SafeDivide(7, 0)
	if err == nil {
		t.Fatal("SafeDivide(7, 0) returned no error, want one")
	}
	if got != 0 {
		t.Errorf("SafeDivide(7, 0) = %d alongside an error, want 0", got)
	}
	if !strings.HasPrefix(err.Error(), "safe divide: ") {
		t.Errorf("error = %q, want it to start with %q", err.Error(), "safe divide: ")
	}
	if !strings.Contains(err.Error(), "divide by zero") {
		t.Errorf("error = %q, want it to carry the recovered runtime error", err.Error())
	}
}

func TestMustPositiveReturnsPositives(t *testing.T) {
	for _, n := range []int{1, 42} {
		if got := MustPositive(n); got != n {
			t.Errorf("MustPositive(%d) = %d, want %d", n, got, n)
		}
	}
}

func TestMustPositivePanics(t *testing.T) {
	for _, n := range []int{0, -1} {
		func() {
			defer func() {
				r := recover()
				if r == nil {
					t.Errorf("MustPositive(%d) returned normally, want a panic", n)
					return
				}
				err, ok := r.(error)
				if !ok || !errors.Is(err, ErrNotPositive) {
					t.Errorf("MustPositive(%d) panicked with %v, want ErrNotPositive", n, r)
				}
			}()
			MustPositive(n)
		}()
	}
}

func TestSafeIndex(t *testing.T) {
	nums := []int{10, 20, 30}

	if got, err := SafeIndex(nums, 1); err != nil || got != 20 {
		t.Errorf("SafeIndex(%v, 1) = %d, %v, want 20, <nil>", nums, got, err)
	}

	for _, i := range []int{-1, 3} {
		got, err := SafeIndex(nums, i)
		if err == nil {
			t.Errorf("SafeIndex(%v, %d) returned no error, want one", nums, i)
			continue
		}
		if got != 0 {
			t.Errorf("SafeIndex(%v, %d) = %d alongside an error, want 0", nums, i, got)
		}
		if !strings.HasPrefix(err.Error(), "safe index: ") {
			t.Errorf("error = %q, want it to start with %q", err.Error(), "safe index: ")
		}
		if !strings.Contains(err.Error(), "index out of range") {
			t.Errorf("error = %q, want it to carry the recovered runtime error", err.Error())
		}
	}
}

func TestGuardOnANormalReturn(t *testing.T) {
	ran := false
	if err := Guard(func() { ran = true }); err != nil {
		t.Errorf("Guard returned %v for a body that did not panic, want nil", err)
	}
	if !ran {
		t.Error("Guard did not run the body")
	}
}

func TestGuardKeepsAnErrorPanicValue(t *testing.T) {
	err := Guard(func() { MustPositive(-1) })
	if err == nil {
		t.Fatal("Guard returned nil for a panicking body, want an error")
	}
	if !errors.Is(err, ErrNotPositive) {
		t.Errorf("Guard returned %v, want the panicked error itself", err)
	}
}

func TestGuardFormatsOtherPanicValues(t *testing.T) {
	err := Guard(func() { panic("boom") })
	if err == nil {
		t.Fatal("Guard returned nil for a panicking body, want an error")
	}
	if err.Error() != "panic: boom" {
		t.Errorf("Guard returned %q, want %q", err.Error(), "panic: boom")
	}

	if err := Guard(func() { panic(42) }); err == nil || err.Error() != "panic: 42" {
		t.Errorf("Guard for panic(42) returned %v, want %q", err, "panic: 42")
	}
}
