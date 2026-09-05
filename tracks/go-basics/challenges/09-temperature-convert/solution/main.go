package main

import (
	"fmt"
	"math"
)

// CToF converts degrees Celsius to degrees Fahrenheit.
func CToF(c float64) float64 {
	// 9.0/5.0 stays a float. 9/5 would be integer division and give 1.
	return c*9.0/5.0 + 32
}

// FToC converts degrees Fahrenheit to degrees Celsius.
func FToC(f float64) float64 {
	return (f - 32) * 5.0 / 9.0
}

// RoundTenth rounds v to one decimal place, halves away from zero.
// RoundTenth(0.25) is 0.3 and RoundTenth(-0.25) is -0.3.
func RoundTenth(v float64) float64 {
	return math.Round(v*10) / 10
}

// FahrenheitWhole converts c to Fahrenheit and rounds to the nearest whole
// degree, halves away from zero.
func FahrenheitWhole(c float64) int {
	// math.Round first, then convert: int() on its own truncates.
	return int(math.Round(CToF(c)))
}

// Report renders one line, both scales to one decimal place:
//
//	Report(21.5) == "21.5°C = 70.7°F"
func Report(c float64) string {
	return fmt.Sprintf("%.1f°C = %.1f°F", c, CToF(c))
}

func main() {
	fmt.Println(Report(21.5))
}
