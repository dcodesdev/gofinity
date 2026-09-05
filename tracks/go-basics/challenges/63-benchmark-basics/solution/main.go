package main

import (
	"errors"
	"fmt"
	"strings"
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
	if len(parts) == 0 {
		return ""
	}
	size := len(sep) * (len(parts) - 1)
	for _, p := range parts {
		size += len(p)
	}
	var b strings.Builder
	b.Grow(size)
	for i, p := range parts {
		if i > 0 {
			b.WriteString(sep)
		}
		b.WriteString(p)
	}
	return b.String()
}

// FilterInPlace returns the elements of xs that keep reports true for, in
// order, without allocating: the result reuses xs's own backing array, so xs
// is scrambled afterwards and the caller must use the returned slice.
func FilterInPlace(xs []int, keep func(int) bool) []int {
	out := xs[:0]
	for _, x := range xs {
		if keep(x) {
			out = append(out, x)
		}
	}
	return out
}

// ParseSum returns the sum of the comma-separated decimal integers in data:
// ParseSum([]byte("1,2,-3")) is 0. Empty data is 0 with no error. A field that
// is not an integer - empty, "-", "12x" - is an error, and the sum is then 0.
//
// It must not allocate on the success path, which rules out string(data),
// strings.Split and bytes.Split. Walk the bytes.
func ParseSum(data []byte) (int, error) {
	if len(data) == 0 {
		return 0, nil
	}
	sum, start := 0, 0
	for i := 0; i <= len(data); i++ {
		if i < len(data) && data[i] != ',' {
			continue
		}
		n, err := parseField(data[start:i])
		if err != nil {
			return 0, err
		}
		sum += n
		start = i + 1
	}
	return sum, nil
}

// errBadField is a package-level value, not a fmt.Errorf at the point of
// failure: building a message would allocate, and although the success path is
// the one being measured, an error that carries no detail is also honest here.
var errBadField = errors.New("not a comma-separated list of integers")

// parseField turns one field into an int without allocating.
func parseField(field []byte) (int, error) {
	negative := false
	if len(field) > 0 && field[0] == '-' {
		negative, field = true, field[1:]
	}
	if len(field) == 0 {
		return 0, errBadField
	}
	n := 0
	for _, c := range field {
		if c < '0' || c > '9' {
			return 0, errBadField
		}
		n = n*10 + int(c-'0')
	}
	if negative {
		n = -n
	}
	return n, nil
}

// BenchJoin benchmarks JoinFields over BuildParts(benchSize).
//
// The setup belongs outside the loop and outside the measurement: build the
// parts, call b.ReportAllocs, reset the clock, and only then loop b.N times.
// Assign the result to SinkString.
func BenchJoin(b *testing.B) {
	parts := BuildParts(benchSize)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SinkString = JoinFields(parts, ",")
	}
}

// BenchFilter benchmarks FilterInPlace over BuildNumbers(benchSize).
//
// FilterInPlace destroys its input, so each iteration needs a fresh slice -
// and building it is not what is being measured. Stop the clock around the
// rebuild and start it again before the call. Assign the result to SinkInts.
func BenchFilter(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		nums := BuildNumbers(benchSize)
		b.StartTimer()
		SinkInts = FilterInPlace(nums, func(n int) bool { return n%3 == 0 })
	}
}

func main() {
	fmt.Println(JoinFields([]string{"go", "test", "-bench=."}, " "))
	fmt.Println(FilterInPlace([]int{1, 2, 3, 4, 5, 6}, func(n int) bool { return n%2 == 0 }))
	fmt.Println(ParseSum([]byte("1,2,-3,400,5")))
}
