package main

import "fmt"

// Signed is the set of signed integer types. Every member needs a ~ so named
// types like `type ID int` are accepted too.
type Signed interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

// Unsigned is the set of unsigned integer types.
type Unsigned interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

// Integer is every integer type. Constraints embed each other, so this one is a
// union of the two above rather than a fresh list.
type Integer interface {
	Signed | Unsigned
}

// Float is the set of floating point types.
type Float interface {
	~float32 | ~float64
}

// Number is every numeric type: the integers and the floats.
type Number interface {
	Integer | Float
}

// Ordered is every type the < operator works on. That is the numbers plus
// string, and it is why Max and Sum take different constraints.
type Ordered interface {
	Integer | Float | ~string
}

// Sum adds every element of s. The zero value of T is the identity, and it is
// spelled with a var declaration because you cannot write 0 for a T.
//
// The sum of an empty slice is the zero value, not an error.
func Sum[T Number](s []T) T {
	var total T
	for _, v := range s {
		total += v
	}
	return total
}

// Average returns the mean of s as a float64, and whether there was anything to
// average. An empty slice is (0, false).
//
// Convert each element as you add it rather than converting the sum: a []int8
// of a hundred 100s overflows long before the division.
func Average[T Number](s []T) (float64, bool) {
	if len(s) == 0 {
		return 0, false
	}
	total := 0.0
	for _, v := range s {
		total += float64(v)
	}
	return total / float64(len(s)), true
}

// Min returns the smallest element of s, and whether there was one. Seed the
// answer with s[0]; the zero value is wrong for a slice of negatives.
func Min[T Ordered](s []T) (T, bool) {
	if len(s) == 0 {
		var zero T
		return zero, false
	}
	best := s[0]
	for _, v := range s[1:] {
		if v < best {
			best = v
		}
	}
	return best, true
}

// Max returns the largest element of s, and whether there was one.
func Max[T Ordered](s []T) (T, bool) {
	if len(s) == 0 {
		var zero T
		return zero, false
	}
	best := s[0]
	for _, v := range s[1:] {
		if v > best {
			best = v
		}
	}
	return best, true
}

// Clamp returns v limited to the range [lo, hi]. When hi is below lo the range
// is empty and there is no sensible answer, so it returns lo.
func Clamp[T Ordered](v, lo, hi T) T {
	if hi < lo {
		return lo
	}
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// Abs returns the absolute value of v. Its constraint is an inline union of two
// named constraints - unsigned types are deliberately excluded, because
// negating one of those is never what anybody meant.
func Abs[T Signed | Float](v T) T {
	if v < 0 {
		return -v
	}
	return v
}

// SumValues adds every value in m. Two type parameters with two different
// constraints: a map key is comparable, and the values are numbers.
func SumValues[K comparable, V Number](m map[K]V) V {
	var total V
	for _, v := range m {
		total += v
	}
	return total
}

func main() {
	fmt.Println(Sum([]int{1, 2, 3}))
	fmt.Println(Max([]string{"go", "rust", "c"}))
}
