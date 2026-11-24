package calc_test

import (
	"testing"

	"example.com/covered/calc"
	"example.com/covered/calc/calc_testing"
)

func TestSum(t *testing.T) {
	if got := calc.Sum(calc_testing.Numbers()...); got != 6 {
		t.Errorf("Sum = %d, want 6", got)
	}
}

func TestClamp(t *testing.T) {
	if calc.Clamp(-1, 0, 9) != 0 || calc.Clamp(5, 0, 9) != 5 {
		t.Error("Clamp moved a number")
	}
}
