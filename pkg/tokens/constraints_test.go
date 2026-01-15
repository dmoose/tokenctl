// tokenctl/pkg/tokens/constraints_test.go

package tokens

import (
	"testing"
)

func TestParseConstraints_NoConstraints(t *testing.T) {
	token := map[string]any{
		"$value": "10px",
		"$type":  "dimension",
	}

	constraint, err := ParseConstraints(token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if constraint != nil {
		t.Errorf("expected nil constraint, got %v", constraint)
	}
}

func TestParseConstraints_DimensionMinOnly(t *testing.T) {
	token := map[string]any{
		"$value": "10px",
		"$type":  "dimension",
		"$min":   "5px",
	}

	constraint, err := ParseConstraints(token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if constraint == nil {
		t.Fatal("expected constraint, got nil")
		return // unreachable, but satisfies staticcheck
	}
	if constraint.IsNumber {
		t.Error("expected dimension constraint, got number")
	}
	if constraint.Min == nil {
		t.Fatal("expected min constraint")
		return // unreachable, but satisfies staticcheck
	}
	if constraint.Min.Value != 5 || constraint.Min.Unit != "px" {
		t.Errorf("expected min 5px, got %s", constraint.Min.String())
	}
	if constraint.Max != nil {
		t.Error("expected nil max constraint")
	}
}

func TestParseConstraints_DimensionMaxOnly(t *testing.T) {
	token := map[string]any{
		"$value": "10px",
		"$type":  "dimension",
		"$max":   "20px",
	}

	constraint, err := ParseConstraints(token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if constraint == nil {
		t.Fatal("expected constraint, got nil")
		return // unreachable, but satisfies staticcheck
	}
	if constraint.Max == nil {
		t.Fatal("expected max constraint")
		return // unreachable, but satisfies staticcheck
	}
	if constraint.Max.Value != 20 || constraint.Max.Unit != "px" {
		t.Errorf("expected max 20px, got %s", constraint.Max.String())
	}
}

func TestParseConstraints_DimensionMinAndMax(t *testing.T) {
	token := map[string]any{
		"$value": "2.5rem",
		"$type":  "dimension",
		"$min":   "1rem",
		"$max":   "5rem",
	}

	constraint, err := ParseConstraints(token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if constraint == nil {
		t.Fatal("expected constraint, got nil")
		return // unreachable, but satisfies staticcheck
	}
	if constraint.Min == nil || constraint.Max == nil {
		t.Fatal("expected both min and max constraints")
		return // unreachable, but satisfies staticcheck
	}
	if constraint.Min.Value != 1 || constraint.Min.Unit != "rem" {
		t.Errorf("expected min 1rem, got %s", constraint.Min.String())
	}
	if constraint.Max.Value != 5 || constraint.Max.Unit != "rem" {
		t.Errorf("expected max 5rem, got %s", constraint.Max.String())
	}
}

func TestParseConstraints_NumberConstraints(t *testing.T) {
	token := map[string]any{
		"$value": 0.5,
		"$type":  "number",
		"$min":   0.0,
		"$max":   1.0,
	}

	constraint, err := ParseConstraints(token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if constraint == nil {
		t.Fatal("expected constraint, got nil")
	}
	if !constraint.IsNumber {
		t.Error("expected number constraint")
	}
	if constraint.MinNum == nil || *constraint.MinNum != 0 {
		t.Error("expected min 0")
	}
	if constraint.MaxNum == nil || *constraint.MaxNum != 1 {
		t.Error("expected max 1")
	}
}

func TestParseConstraints_IntConstraints(t *testing.T) {
	token := map[string]any{
		"$value": 5,
		"$type":  "number",
		"$min":   1,
		"$max":   10,
	}

	constraint, err := ParseConstraints(token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if constraint == nil {
		t.Fatal("expected constraint, got nil")
	}
	if !constraint.IsNumber {
		t.Error("expected number constraint")
	}
}

func TestParseConstraints_InvalidMinGreaterThanMax(t *testing.T) {
	token := map[string]any{
		"$value": "10px",
		"$type":  "dimension",
		"$min":   "20px",
		"$max":   "10px",
	}

	_, err := ParseConstraints(token)
	if err == nil {
		t.Error("expected error for min > max")
	}
}

func TestParseConstraints_MismatchedUnits(t *testing.T) {
	token := map[string]any{
		"$value": "10px",
		"$type":  "dimension",
		"$min":   "1rem",
		"$max":   "20px",
	}

	_, err := ParseConstraints(token)
	if err == nil {
		t.Error("expected error for mismatched units")
	}
}

func TestConstraint_CheckValue_DimensionInRange(t *testing.T) {
	token := map[string]any{
		"$min": "5px",
		"$max": "20px",
	}

	constraint, err := ParseConstraints(token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tests := []struct {
		value   string
		wantErr bool
	}{
		{"10px", false},
		{"5px", false},  // exactly min
		{"20px", false}, // exactly max
		{"4px", true},   // below min
		{"21px", true},  // above max
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			err := constraint.CheckValue(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckValue(%s) error = %v, wantErr %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestConstraint_CheckValue_NumberInRange(t *testing.T) {
	min := 0.0
	max := 1.0
	constraint := &Constraint{
		IsNumber: true,
		MinNum:   &min,
		MaxNum:   &max,
	}

	tests := []struct {
		name    string
		value   any
		wantErr bool
	}{
		{"middle", 0.5, false},
		{"min", 0.0, false},
		{"max", 1.0, false},
		{"below min", -0.1, true},
		{"above max", 1.1, true},
		{"int in range", 1, false},
		{"string number in range", "0.5", false},
		{"string number below", "-0.5", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := constraint.CheckValue(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckValue(%v) error = %v, wantErr %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestConstraint_CheckValue_UnitMismatch(t *testing.T) {
	token := map[string]any{
		"$min": "5px",
		"$max": "20px",
	}

	constraint, err := ParseConstraints(token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = constraint.CheckValue("10rem")
	if err == nil {
		t.Error("expected error for unit mismatch")
	}
}

func TestConstraint_CheckValue_InvalidDimension(t *testing.T) {
	token := map[string]any{
		"$min": "5px",
		"$max": "20px",
	}

	constraint, err := ParseConstraints(token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = constraint.CheckValue("invalid")
	if err == nil {
		t.Error("expected error for invalid dimension")
	}
}

func TestConstraint_CheckValue_NilConstraint(t *testing.T) {
	var constraint *Constraint = nil
	err := constraint.CheckValue("anything")
	if err != nil {
		t.Errorf("nil constraint should not return error, got: %v", err)
	}
}

func TestConstraint_String(t *testing.T) {
	tests := []struct {
		name       string
		constraint *Constraint
		want       string
	}{
		{
			name:       "nil constraint",
			constraint: nil,
			want:       "no constraints",
		},
		{
			name: "dimension min only",
			constraint: &Constraint{
				Min: &Dimension{Value: 5, Unit: "px"},
			},
			want: "min: 5px",
		},
		{
			name: "dimension min and max",
			constraint: &Constraint{
				Min: &Dimension{Value: 5, Unit: "px"},
				Max: &Dimension{Value: 20, Unit: "px"},
			},
			want: "min: 5px, max: 20px",
		},
		{
			name: "number min and max",
			constraint: func() *Constraint {
				min, max := 0.0, 1.0
				return &Constraint{
					IsNumber: true,
					MinNum:   &min,
					MaxNum:   &max,
				}
			}(),
			want: "min: 0, max: 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.constraint.String()
			if got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseConstraints_StringNumber(t *testing.T) {
	token := map[string]any{
		"$value": "5",
		"$min":   "0",
		"$max":   "10",
	}

	constraint, err := ParseConstraints(token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if constraint == nil {
		t.Fatal("expected constraint, got nil")
	}
	if !constraint.IsNumber {
		t.Error("expected number constraint for string numbers")
	}
}

func TestParseConstraints_InvalidConstraintValue(t *testing.T) {
	token := map[string]any{
		"$value": "10px",
		"$min":   "invalid",
	}

	_, err := ParseConstraints(token)
	if err == nil {
		t.Error("expected error for invalid constraint value")
	}
}

func TestParseConstraints_UnsupportedType(t *testing.T) {
	token := map[string]any{
		"$value": "10px",
		"$min":   []string{"invalid"},
	}

	_, err := ParseConstraints(token)
	if err == nil {
		t.Error("expected error for unsupported constraint type")
	}
}
