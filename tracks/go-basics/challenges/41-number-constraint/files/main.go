package main

import "fmt"

// Signed is the set of signed integer types. Every member needs a ~ so named
// types like `type ID int` are accepted too.
//
// TODO: add int8, int16, int32 and int64.
type Signed interface {
	~int
}

// Unsigned is the set of unsigned integer types.
//
// TODO: add uint8, uint16, uint32, uint64 and uintptr.
type Unsigned interface {
	~uint
}

// Integer is every integer type. Constraints embed each other, so this one is a
// union of the two above rather than a fresh list.
//
// TODO: embed Signed and Unsigned.
type Integer interface {
	Signed
}

// Float is the set of floating point types.
//
// TODO: add float64.
type Float interface {
	~float32
}

// Number is every numeric type: the integers and the floats.
//
// TODO: embed Integer and Float.
type Number interface {
	Integer
}

// Ordered is every type the < operator works on. That is the numbers plus
// string, and it is why Max and Sum take different constraints.
//
// TODO: embed Integer and Float, and add ~string.
type Ordered interface {
	Integer
}

// Sum adds every element of s. The zero value of T is the identity, and it is
// spelled with a var declaration because you cannot write 0 for a T.
//
// The sum of an empty slice is the zero value, not an error.
func Sum[T Number](s []T) T {
	// TODO
	var total T
	return total
}

// Average returns the mean of s as a float64, and whether there was anything to
// average. An empty slice is (0, false).
//
// Convert each element as you add it rather than converting the sum: a []int8
// of a hundred 100s overflows long before the division.
func Average[T Number](s []T) (float64, bool) {
	// TODO
	return 0, false
}

// Min returns the smallest element of s, and whether there was one. Seed the
// answer with s[0]; the zero value is wrong for a slice of negatives.
func Min[T Ordered](s []T) (T, bool) {
	// TODO
	var zero T
	return zero, false
}

// Max returns the largest element of s, and whether there was one.
func Max[T Ordered](s []T) (T, bool) {
	// TODO
	var zero T
	return zero, false
}

// Clamp returns v limited to the range [lo, hi]. When hi is below lo the range
// is empty and there is no sensible answer, so it returns lo.
func Clamp[T Ordered](v, lo, hi T) T {
	// TODO
	return v
}

// Abs returns the absolute value of v. Its constraint is an inline union of two
// named constraints - unsigned types are deliberately excluded, because
// negating one of those is never what anybody meant.
func Abs[T Signed | Float](v T) T {
	// TODO
	return v
}

// SumValues adds every value in m. Two type parameters with two different
// constraints: a map key is comparable, and the values are numbers.
func SumValues[K comparable, V Number](m map[K]V) V {
	// TODO
	var total V
	return total
}

func main() {
	fmt.Println(Sum([]int{1, 2, 3}))
	fmt.Println(Max([]string{"go", "rust", "c"}))
}
