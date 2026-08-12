// tokenctl/pkg/derive/format.go
//
// Number formatting that agrees with the TypeScript engine exactly.
//
// The engine's output strings come from Number.prototype.toFixed and
// from Number→String coercion. Go's strconv rounds ties to even; JS
// rounds ties away from zero, on the *exact* binary value of the double.
// Those two rules disagree on any value that lands exactly on a rounding
// boundary, which is precisely the kind of difference that would force a
// float tolerance into the golden comparison. Doing the rounding on an
// exact rational instead keeps the comparison a string equality.
package derive

import (
	"math"
	"math/big"
	"strconv"
	"strings"
)

// toFixed formats v with exactly digits decimal places, matching
// JavaScript's Number.prototype.toFixed: round-half-away-from-zero
// applied to the exact value of the float64, not to its shortest decimal
// representation.
func toFixed(v float64, digits int) string {
	if math.IsNaN(v) {
		return "NaN"
	}
	if math.IsInf(v, 0) {
		if v > 0 {
			return "Infinity"
		}
		return "-Infinity"
	}

	// The spec prepends the sign when x < 0 — strictly less, so negative
	// zero prints unsigned while a small negative that rounds to zero
	// keeps its sign: (-0).toFixed(1) is "0.0" but (-0.04).toFixed(1) is
	// "-0.0".
	neg := v < 0
	// big.Rat holds the double's exact value, so scaling and rounding
	// below are decisions about the real number the double denotes.
	r := new(big.Rat).SetFloat64(math.Abs(v))

	scale := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(digits)), nil)
	r.Mul(r, new(big.Rat).SetInt(scale))

	// floor(r) and remainder, then round halves up.
	num, den := r.Num(), r.Denom()
	q, rem := new(big.Int).QuoRem(num, den, new(big.Int))
	twice := new(big.Int).Lsh(rem, 1) // 2*rem vs den decides the half
	if twice.Cmp(den) >= 0 {
		q.Add(q, big.NewInt(1))
	}

	s := q.String()
	if digits > 0 {
		for len(s) <= digits {
			s = "0" + s
		}
		s = s[:len(s)-digits] + "." + s[len(s)-digits:]
	}
	if neg {
		s = "-" + s
	}
	return s
}

// trimFixed formats v to digits decimals and then drops trailing zeros,
// matching the engine's parseFloat(x.toFixed(n)) followed by string
// interpolation: 1.000 renders as "1", 0.500 as "0.5".
func trimFixed(v float64, digits int) string {
	s := toFixed(v, digits)
	if !strings.Contains(s, ".") {
		return s
	}
	s = strings.TrimRight(s, "0")
	s = strings.TrimSuffix(s, ".")
	if s == "" || s == "-" {
		return "0"
	}
	return s
}

// numberToString renders v the way JavaScript's String(number) does for
// the magnitudes this package produces: shortest representation that
// round-trips, no exponent.
func numberToString(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}
