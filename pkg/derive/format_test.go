package derive

import (
	"math"
	"testing"
)

// toFixed has to match JavaScript's Number.prototype.toFixed, not Go's
// strconv. The two disagree on ties: strconv rounds to even, JS rounds
// away from zero. Expectations below are what node prints.
func TestToFixed_MatchesJavaScript(t *testing.T) {
	t.Parallel()

	tests := []struct {
		v      float64
		digits int
		want   string
	}{
		// Ties round away from zero, where Go's strconv would round to
		// even and give "0.2" and "2".
		{0.25, 1, "0.3"},
		{2.5, 0, "3"},
		{1.5, 0, "2"},
		{-0.25, 1, "-0.3"},

		// A tie only in decimal: the double nearest 0.0045 is slightly
		// below it, so the honest answer is 0.004. This is the exact
		// value the engine produces for a tinted neutral at tint 30.
		{0.015 * 0.3, 3, "0.004"},
		{0.010 * 0.3, 3, "0.003"},
		{0.020 * 0.3, 3, "0.006"},

		// Integers and padding.
		{100, 1, "100.0"},
		{0, 3, "0.000"},
		{55, 1, "55.0"},
		{1, 0, "1"},
		{359, 0, "359"},

		// Negative zero prints without a sign, as in JS. Written via
		// Copysign because the Go literal -0.0 is just 0.0.
		{math.Copysign(0, -1), 1, "0.0"},
		// A negative value that rounds to zero keeps its sign, also as
		// in JS: (-0.04).toFixed(1) === "-0.0".
		{-0.04, 1, "-0.0"},

		{0.2, 3, "0.200"},
		{0.1 + 0.2, 3, "0.300"},
	}

	for _, tt := range tests {
		if got := toFixed(tt.v, tt.digits); got != tt.want {
			t.Errorf("toFixed(%v, %d) = %q, want %q", tt.v, tt.digits, got, tt.want)
		}
	}
}

// trimFixed models parseFloat(x.toFixed(n)) followed by interpolation,
// which is how the engine renders every density-scaled dimension.
func TestTrimFixed(t *testing.T) {
	t.Parallel()

	tests := []struct {
		v      float64
		digits int
		want   string
	}{
		{1.0, 3, "1"},
		{0.5, 3, "0.5"},
		{0.25, 3, "0.25"},
		{2.0, 3, "2"},
		{1.375, 3, "1.375"},
		{0.125 * 1.3, 3, "0.163"},
		{0.25 * 0.75, 3, "0.188"},
		{1.0 * 1.025, 3, "1.025"},
		{0, 3, "0"},
	}

	for _, tt := range tests {
		if got := trimFixed(tt.v, tt.digits); got != tt.want {
			t.Errorf("trimFixed(%v, %d) = %q, want %q", tt.v, tt.digits, got, tt.want)
		}
	}
}

func TestNumberToString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		v    float64
		want string
	}{
		{1.25, "1.25"},
		{1.3, "1.3"},
		{1.5, "1.5"},
		{1.65, "1.65"},
		{100, "100"},
		{0.2, "0.2"},
		{102.5, "102.5"},
	}

	for _, tt := range tests {
		if got := numberToString(tt.v); got != tt.want {
			t.Errorf("numberToString(%v) = %q, want %q", tt.v, got, tt.want)
		}
	}
}
