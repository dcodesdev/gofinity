package main

import (
	"fmt"
	"testing"
)

// The sinks. A benchmark's whole body can be deleted by the compiler if its
// result is never used, and then you have measured an empty loop. Assigning to
// a package-level variable keeps the work alive.
//
// They are typed rather than `any` on purpose: assigning a slice or a string to
// an `any` allocates, and that allocation would show up in every allocation
// count in this file.
var (
	SinkString string
	SinkInts   []int
	SinkInt    int
)

// benchSize is the input size both benchmarks use. It is large enough that one
// iteration is worth timing.
const benchSize = 20000

// BuildParts returns n strings to join. It is the benchmark's setup, and it is
// deliberately slower than the work being measured, which is exactly the
// situation b.ResetTimer exists for.
func BuildParts(n int) []string {
	parts := make([]string, 0, n)
	for i := 0; i < n; i++ {
		part := make([]byte, 1+i%20)
		for j := range part {
			part[j] = byte('a' + (i+j)%26)
		}
		parts = append(parts, string(part))
	}
	return parts
}

// BuildNumbers returns n numbers to filter. It costs several times what
// filtering them costs, which is what makes it worth stopping the clock for.
func BuildNumbers(n int) []int {
	xs := make([]int, n)
	for i := range xs {
		v := uint64(i)*2654435761 + 1
		for r := 0; r < 20; r++ {
			v ^= v >> 13
			v *= 0x9E3779B97F4A7C15
		}
		xs[i] = int(v % 1000)
	}
	return xs
}

// JoinFields returns the parts joined with sep, like strings.Join.
//
// It must allocate at most twice however long parts is: build the result in a
// strings.Builder that has been told the final length up front, not with `+=`.
func JoinFields(parts []string, sep string) string {
	// TODO
	return ""
}

// FilterInPlace returns the elements of xs that keep reports true for, in
// order, without allocating: the result reuses xs's own backing array, so xs
// is scrambled afterwards and the caller must use the returned slice.
func FilterInPlace(xs []int, keep func(int) bool) []int {
	// TODO
	return nil
}

// ParseSum returns the sum of the comma-separated decimal integers in data:
// ParseSum([]byte("1,2,-3")) is 0. Empty data is 0 with no error. A field that
// is not an integer - empty, "-", "12x" - is an error, and the sum is then 0.
//
// It must not allocate on the success path, which rules out string(data),
// strings.Split and bytes.Split. Walk the bytes.
func ParseSum(data []byte) (int, error) {
	// TODO
	return 0, nil
}

// BenchJoin benchmarks JoinFields over BuildParts(benchSize).
//
// The setup belongs outside the loop and outside the measurement: build the
// parts, call b.ReportAllocs, reset the clock, and only then loop b.N times.
// Assign the result to SinkString.
func BenchJoin(b *testing.B) {
	// TODO
}

// BenchFilter benchmarks FilterInPlace over BuildNumbers(benchSize).
//
// FilterInPlace destroys its input, so each iteration needs a fresh slice -
// and building it is not what is being measured. Stop the clock around the
// rebuild and start it again before the call. Assign the result to SinkInts.
func BenchFilter(b *testing.B) {
	// TODO
}

func main() {
	fmt.Println(JoinFields([]string{"go", "test", "-bench=."}, " "))
	fmt.Println(FilterInPlace([]int{1, 2, 3, 4, 5, 6}, func(n int) bool { return n%2 == 0 }))
	fmt.Println(ParseSum([]byte("1,2,-3,400,5")))
}
