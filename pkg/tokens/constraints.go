// tokctl/pkg/tokens/constraints.go

package tokens

import (
	"fmt"
	"strconv"
	"strings"
)

// Constraint represents a min/max constraint on a token value
type Constraint struct {
	Min      *Dimension
	Max      *Dimension
	MinNum   *float64
	MaxNum   *float64
	IsNumber bool // true if constraints are pure numbers, false if dimensions
}

// ParseConstraints extracts $min and $max from a token definition
// Returns nil if no constraints are defined
func ParseConstraints(token map[string]interface{}) (*Constraint, error) {
	minVal, hasMin := token["$min"]
	maxVal, hasMax := token["$max"]

	if !hasMin && !hasMax {
		return nil, nil
	}

	constraint := &Constraint{}

	// Determine if we're dealing with numbers or dimensions
	// by checking the type of the values
	if hasMin {
		if err := parseConstraintValue(minVal, constraint, true); err != nil {
			return nil, fmt.Errorf("invalid $min: %w", err)
		}
	}

	if hasMax {
		if err := parseConstraintValue(maxVal, constraint, false); err != nil {
			return nil, fmt.Errorf("invalid $max: %w", err)
		}
	}

	// Validate that min <= max if both are present
	if err := constraint.validate(); err != nil {
		return nil, err
	}

	return constraint, nil
}

// parseConstraintValue parses a single constraint value (min or max)
func parseConstraintValue(val interface{}, c *Constraint, isMin bool) error {
	switch v := val.(type) {
	case float64:
		c.IsNumber = true
		if isMin {
			c.MinNum = &v
		} else {
			c.MaxNum = &v
		}
		return nil

	case int:
		c.IsNumber = true
		f := float64(v)
		if isMin {
			c.MinNum = &f
		} else {
			c.MaxNum = &f
		}
		return nil

	case string:
		// Try to parse as dimension first
		dim, err := ParseDimension(v)
		if err != nil {
			// Try as number
			f, numErr := strconv.ParseFloat(v, 64)
			if numErr != nil {
				return fmt.Errorf("cannot parse constraint value: %s", v)
			}
			c.IsNumber = true
			if isMin {
				c.MinNum = &f
			} else {
				c.MaxNum = &f
			}
			return nil
		}

		// If dimension has no unit, treat as number
		if dim.Unit == "" {
			c.IsNumber = true
			if isMin {
				c.MinNum = &dim.Value
			} else {
				c.MaxNum = &dim.Value
			}
			return nil
		}

		c.IsNumber = false
		if isMin {
			c.Min = &dim
		} else {
			c.Max = &dim
		}
		return nil

	default:
		return fmt.Errorf("unsupported constraint type: %T", val)
	}
}

// validate ensures the constraint is internally consistent
func (c *Constraint) validate() error {
	if c.IsNumber {
		if c.MinNum != nil && c.MaxNum != nil && *c.MinNum > *c.MaxNum {
			return fmt.Errorf("$min (%v) cannot be greater than $max (%v)", *c.MinNum, *c.MaxNum)
		}
	} else {
		if c.Min != nil && c.Max != nil {
			if c.Min.Unit != c.Max.Unit {
				return fmt.Errorf("$min and $max must have same unit: %s vs %s", c.Min.Unit, c.Max.Unit)
			}
			if c.Min.Value > c.Max.Value {
				return fmt.Errorf("$min (%s) cannot be greater than $max (%s)", c.Min.String(), c.Max.String())
			}
		}
	}
	return nil
}

// CheckValue validates that a value satisfies the constraint
// Returns an error describing the violation, or nil if valid
func (c *Constraint) CheckValue(value interface{}) error {
	if c == nil {
		return nil
	}

	if c.IsNumber {
		return c.checkNumber(value)
	}
	return c.checkDimension(value)
}

// checkNumber validates a numeric value against constraints
func (c *Constraint) checkNumber(value interface{}) error {
	var num float64

	switch v := value.(type) {
	case float64:
		num = v
	case int:
		num = float64(v)
	case string:
		var err error
		num, err = strconv.ParseFloat(strings.TrimSpace(v), 64)
		if err != nil {
			return fmt.Errorf("expected number, got: %s", v)
		}
	default:
		return fmt.Errorf("expected number, got %T", value)
	}

	if c.MinNum != nil && num < *c.MinNum {
		return fmt.Errorf("value %v is less than minimum %v", num, *c.MinNum)
	}

	if c.MaxNum != nil && num > *c.MaxNum {
		return fmt.Errorf("value %v is greater than maximum %v", num, *c.MaxNum)
	}

	return nil
}

// checkDimension validates a dimension value against constraints
func (c *Constraint) checkDimension(value interface{}) error {
	var dim Dimension

	switch v := value.(type) {
	case string:
		var err error
		dim, err = ParseDimension(v)
		if err != nil {
			return fmt.Errorf("invalid dimension: %w", err)
		}
	default:
		return fmt.Errorf("expected dimension string, got %T", value)
	}

	// Check unit compatibility
	if c.Min != nil && dim.Unit != c.Min.Unit {
		return fmt.Errorf("value unit %q doesn't match constraint unit %q", dim.Unit, c.Min.Unit)
	}
	if c.Max != nil && dim.Unit != c.Max.Unit {
		return fmt.Errorf("value unit %q doesn't match constraint unit %q", dim.Unit, c.Max.Unit)
	}

	if c.Min != nil && dim.Value < c.Min.Value {
		return fmt.Errorf("value %s is less than minimum %s", dim.String(), c.Min.String())
	}

	if c.Max != nil && dim.Value > c.Max.Value {
		return fmt.Errorf("value %s is greater than maximum %s", dim.String(), c.Max.String())
	}

	return nil
}

// String returns a human-readable representation of the constraint
func (c *Constraint) String() string {
	if c == nil {
		return "no constraints"
	}

	var parts []string

	if c.IsNumber {
		if c.MinNum != nil {
			parts = append(parts, fmt.Sprintf("min: %v", *c.MinNum))
		}
		if c.MaxNum != nil {
			parts = append(parts, fmt.Sprintf("max: %v", *c.MaxNum))
		}
	} else {
		if c.Min != nil {
			parts = append(parts, fmt.Sprintf("min: %s", c.Min.String()))
		}
		if c.Max != nil {
			parts = append(parts, fmt.Sprintf("max: %s", c.Max.String()))
		}
	}

	if len(parts) == 0 {
		return "no constraints"
	}

	return strings.Join(parts, ", ")
}
