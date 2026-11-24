// Package calc_testing is the test kit of calc. Most of it is unused, so counting it would sink the total.
package calc_testing

// Numbers returns the numbers the tests sum.
func Numbers() []int { return []int{1, 2, 3} }

// Unused stands in for the part of a kit a run leaves alone.
func Unused(n int) int {
	switch {
	case n < 0:
		return -n
	case n == 0:
		return 1
	case n < 10:
		return n * 2
	case n < 100:
		return n * 3
	}
	return n
}
