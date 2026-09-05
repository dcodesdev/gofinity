package main

import (
	"flag"
	"os"
	"strings"
	"testing"
)

// The benchmarks here are run from tests rather than by `go test -bench`, and a
// benchmark runs for a second by default. Ten milliseconds is plenty to compare
// two shapes, and it keeps the whole file under a second.
func TestMain(m *testing.M) {
	_ = flag.Set("test.benchtime", "10ms")
	os.Exit(m.Run())
}

func TestJoinFields(t *testing.T) {
	cases := []struct {
		name  string
		parts []string
		sep   string
		want  string
	}{
		{"empty", nil, ",", ""},
		{"one part", []string{"solo"}, ",", "solo"},
		{"two parts", []string{"a", "b"}, ",", "a,b"},
		{"multi byte separator", []string{"a", "b", "c"}, " -> ", "a -> b -> c"},
		{"empty separator", []string{"go", "lang"}, "", "golang"},
		{"empty parts", []string{"", "", ""}, "-", "--"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := JoinFields(c.parts, c.sep); got != c.want {
				t.Errorf("JoinFields(%q, %q) = %q, want %q", c.parts, c.sep, got, c.want)
			}
		})
	}
	parts := BuildParts(500)
	if got, want := JoinFields(parts, ","), strings.Join(parts, ","); got != want {
		t.Errorf("JoinFields disagrees with strings.Join on a large input")
	}
}

func TestJoinFieldsBarelyAllocates(t *testing.T) {
	parts := BuildParts(2000)
	allocs := testing.AllocsPerRun(20, func() { SinkString = JoinFields(parts, ",") })
	if allocs > 2 {
		t.Errorf("JoinFields allocated %.0f times for 2000 parts, want at most 2 - size the Builder up front instead of growing it", allocs)
	}
	if SinkString != strings.Join(parts, ",") {
		t.Error("JoinFields produced the wrong string for the allocation check")
	}
}

func TestFilterInPlace(t *testing.T) {
	even := func(n int) bool { return n%2 == 0 }
	cases := []struct {
		name string
		in   []int
		keep func(int) bool
		want []int
	}{
		{"empty", nil, even, nil},
		{"keeps none", []int{1, 3, 5}, even, nil},
		{"keeps all", []int{2, 4}, even, []int{2, 4}},
		{"keeps some", []int{1, 2, 3, 4, 5, 6}, even, []int{2, 4, 6}},
		{"order is preserved", []int{8, 1, 6, 3, 4}, even, []int{8, 6, 4}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := FilterInPlace(c.in, c.keep)
			if len(got) != len(c.want) {
				t.Fatalf("FilterInPlace(%v) = %v, want %v", c.in, got, c.want)
			}
			for i := range got {
				if got[i] != c.want[i] {
					t.Fatalf("FilterInPlace(%v) = %v, want %v", c.in, got, c.want)
				}
			}
		})
	}
}

func TestFilterInPlaceReusesTheBackingArray(t *testing.T) {
	xs := []int{1, 2, 3, 4, 5, 6}
	got := FilterInPlace(xs, func(n int) bool { return n%2 == 0 })
	if len(got) == 0 {
		t.Fatal("FilterInPlace returned nothing")
	}
	if &got[0] != &xs[0] {
		t.Error("FilterInPlace returned a new slice - the result must reuse xs's backing array")
	}
	nums := BuildNumbers(1000)
	allocs := testing.AllocsPerRun(20, func() { SinkInts = FilterInPlace(nums, func(n int) bool { return true }) })
	if allocs != 0 {
		t.Errorf("FilterInPlace allocated %.0f times, want 0", allocs)
	}
}

func TestParseSum(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want int
	}{
		{"empty", "", 0},
		{"one number", "42", 42},
		{"several", "1,2,3", 6},
		{"negatives", "1,2,-3,400,5", 405},
		{"zeroes", "0,0,0", 0},
		{"leading zeroes", "007,3", 10},
		{"only negative", "-19", -19},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := ParseSum([]byte(c.in))
			if err != nil {
				t.Fatalf("ParseSum(%q) returned an error: %v", c.in, err)
			}
			if got != c.want {
				t.Errorf("ParseSum(%q) = %d, want %d", c.in, got, c.want)
			}
		})
	}
}

func TestParseSumRejectsBadInput(t *testing.T) {
	for _, in := range []string{"1,x", "1,,2", "12x", "-", "1,-", " 1,2", "1 ,2", "1.5", ",", "1,"} {
		got, err := ParseSum([]byte(in))
		if err == nil {
			t.Errorf("ParseSum(%q) = %d, want an error", in, got)
			continue
		}
		if got != 0 {
			t.Errorf("ParseSum(%q) returned the sum %d alongside its error, want 0", in, got)
		}
	}
}

func TestParseSumDoesNotAllocate(t *testing.T) {
	data := []byte(strings.TrimSuffix(strings.Repeat("123,-45,6,", 200), ","))
	allocs := testing.AllocsPerRun(20, func() {
		n, err := ParseSum(data)
		if err != nil {
			t.Fatalf("ParseSum returned an error: %v", err)
		}
		SinkInt = n
	})
	if allocs != 0 {
		t.Errorf("ParseSum allocated %.0f times, want 0 - converting the bytes to a string is the usual reason", allocs)
	}
}

// refJoin is the shape BenchJoin is supposed to have. The user's benchmark is
// compared against it rather than against a fixed number of nanoseconds, so the
// check says the same thing on a slow machine and a fast one.
func refJoin(b *testing.B) {
	parts := BuildParts(benchSize)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SinkString = JoinFields(parts, ",")
	}
}

func refFilter(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		nums := BuildNumbers(benchSize)
		b.StartTimer()
		SinkInts = FilterInPlace(nums, func(n int) bool { return n%3 == 0 })
	}
}

func TestBenchJoinMeasuresTheJoin(t *testing.T) {
	ref := testing.Benchmark(refJoin)
	got := testing.Benchmark(BenchJoin)

	if got.N < 1 {
		t.Fatal("BenchJoin did not run")
	}
	if got.N > 1_000_000 {
		t.Fatalf("BenchJoin ran %d iterations for %d ns each: the body does not loop b.N times, so the framework kept asking for more", got.N, got.NsPerOp())
	}
	if got.NsPerOp() < ref.NsPerOp()/2 {
		t.Errorf("BenchJoin reports %d ns/op where joining %d parts takes about %d - it is not measuring the full input", got.NsPerOp(), benchSize, ref.NsPerOp())
	}
	if got.AllocsPerOp() > 4 {
		t.Errorf("BenchJoin reports %d allocs/op where the join itself does at most 2: the setup is being counted, so b.ResetTimer() is missing or is before the setup", got.AllocsPerOp())
	}
}

func TestBenchFilterExcludesTheRebuild(t *testing.T) {
	ref := testing.Benchmark(refFilter)
	got := testing.Benchmark(BenchFilter)

	if got.N < 1 {
		t.Fatal("BenchFilter did not run")
	}
	if got.N > 1_000_000 {
		t.Fatalf("BenchFilter ran %d iterations for %d ns each: the body does not loop b.N times", got.N, got.NsPerOp())
	}
	if got.AllocsPerOp() != 0 {
		t.Errorf("BenchFilter reports %d allocs/op, want 0: the per-iteration rebuild allocates, so it has to happen with the clock stopped", got.AllocsPerOp())
	}
	if got.NsPerOp() < ref.NsPerOp()/2 {
		t.Errorf("BenchFilter reports %d ns/op where filtering %d numbers takes about %d - a fresh slice per iteration is part of the deal", got.NsPerOp(), benchSize, ref.NsPerOp())
	}
	if got.NsPerOp() > ref.NsPerOp()*4 {
		t.Errorf("BenchFilter reports %d ns/op where the filter alone takes about %d - the clock is still running during the rebuild", got.NsPerOp(), ref.NsPerOp())
	}
}
