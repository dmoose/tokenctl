// tokenctl/pkg/tokens/dimension.go

package tokens

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

// Dimension represents a CSS dimension value (number + unit)
type Dimension struct {
	Value float64
	Unit  string
}

// Common CSS units
var validUnits = map[string]bool{
	// Absolute lengths
	"px": true, "cm": true, "mm": true, "in": true, "pt": true, "pc": true,
	// Relative lengths
	"em": true, "rem": true, "ex": true, "ch": true,
	"vw": true, "vh": true, "vmin": true, "vmax": true,
	"%": true,
	// Time
	"s": true, "ms": true,
	// Angle
	"deg": true, "rad": true, "turn": true,
	// No unit (pure number)
	"": true,
}

// dimensionRegex matches a number followed by an optional unit
var dimensionRegex = regexp.MustCompile(`^(-?[0-9]*\.?[0-9]+)(px|cm|mm|in|pt|pc|em|rem|ex|ch|vw|vh|vmin|vmax|%|s|ms|deg|rad|turn)?$`)

// ParseDimension parses a CSS dimension string into a Dimension struct
// Examples: "10px", "1.5rem", "100%", "0", "2.5"
func ParseDimension(s string) (Dimension, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return Dimension{}, fmt.Errorf("empty dimension string")
	}

	matches := dimensionRegex.FindStringSubmatch(s)
	if matches == nil {
		return Dimension{}, fmt.Errorf("invalid dimension format: %s", s)
	}

	value, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return Dimension{}, fmt.Errorf("invalid number in dimension: %s", s)
	}

	unit := ""
	if len(matches) > 2 {
		unit = matches[2]
	}

	return Dimension{Value: value, Unit: unit}, nil
}

// String returns the dimension as a CSS string
func (d Dimension) String() string {
	// Round to reasonable precision to avoid floating point artifacts
	rounded := roundTo(d.Value, 4)

	// Format nicely - avoid unnecessary decimals
	if rounded == float64(int(rounded)) {
		return fmt.Sprintf("%d%s", int(rounded), d.Unit)
	}

	// Format with up to 4 decimal places, trimming trailing zeros
	s := fmt.Sprintf("%.4f", rounded)
	s = strings.TrimRight(s, "0")
	s = strings.TrimRight(s, ".")
	return s + d.Unit
}

// roundTo rounds a float to n decimal places using math.Round for accuracy
func roundTo(val float64, places int) float64 {
	pow := math.Pow(10, float64(places))
	return math.Round(val*pow) / pow
}

// Add adds two dimensions (must have same unit)
func (d Dimension) Add(other Dimension) (Dimension, error) {
	if d.Unit != other.Unit {
		return Dimension{}, fmt.Errorf("cannot add dimensions with different units: %s and %s", d.Unit, other.Unit)
	}
	return Dimension{Value: d.Value + other.Value, Unit: d.Unit}, nil
}

// Subtract subtracts another dimension (must have same unit)
func (d Dimension) Subtract(other Dimension) (Dimension, error) {
	if d.Unit != other.Unit {
		return Dimension{}, fmt.Errorf("cannot subtract dimensions with different units: %s and %s", d.Unit, other.Unit)
	}
	return Dimension{Value: d.Value - other.Value, Unit: d.Unit}, nil
}

// Multiply multiplies the dimension by a scalar
func (d Dimension) Multiply(scalar float64) Dimension {
	result := d.Value * scalar
	// Round to 4 decimal places to avoid floating point artifacts
	result = roundTo(result, 4)
	return Dimension{Value: result, Unit: d.Unit}
}

// Divide divides the dimension by a scalar
func (d Dimension) Divide(scalar float64) (Dimension, error) {
	if scalar == 0 {
		return Dimension{}, fmt.Errorf("division by zero")
	}
	result := d.Value / scalar
	// Round to 4 decimal places to avoid floating point artifacts
	result = roundTo(result, 4)
	return Dimension{Value: result, Unit: d.Unit}, nil
}

// IsZero returns true if the dimension value is zero
func (d Dimension) IsZero() bool {
	return d.Value == 0
}

// IsDimension checks if a string looks like a dimension value
func IsDimension(s string) bool {
	_, err := ParseDimension(s)
	return err == nil
}

// MustParseDimension parses a dimension and panics on error (for tests/known values)
func MustParseDimension(s string) Dimension {
	d, err := ParseDimension(s)
	if err != nil {
		panic(err)
	}
	return d
}
