package main

import "fmt"

// CToF converts degrees Celsius to degrees Fahrenheit.
func CToF(c float64) float64 {
	// TODO
	return 0
}

// FToC converts degrees Fahrenheit to degrees Celsius.
func FToC(f float64) float64 {
	// TODO
	return 0
}

// RoundTenth rounds v to one decimal place, halves away from zero.
// RoundTenth(0.25) is 0.3 and RoundTenth(-0.25) is -0.3.
func RoundTenth(v float64) float64 {
	// TODO
	return 0
}

// FahrenheitWhole converts c to Fahrenheit and rounds to the nearest whole
// degree, halves away from zero.
func FahrenheitWhole(c float64) int {
	// TODO
	return 0
}

// Report renders one line, both scales to one decimal place:
//
//	Report(21.5) == "21.5°C = 70.7°F"
func Report(c float64) string {
	// TODO
	return ""
}

func main() {
	fmt.Println(Report(21.5))
}
