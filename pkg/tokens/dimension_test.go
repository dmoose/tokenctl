// tokenctl/pkg/tokens/dimension_test.go

package tokens

import (
	"testing"
)

func TestParseDimension(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		input     string
		wantValue float64
		wantUnit  string
		wantErr   bool
	}{
		// Pixels
		{"pixels", "10px", 10, "px", false},
		{"pixels decimal", "10.5px", 10.5, "px", false},
		{"pixels zero", "0px", 0, "px", false},

		// Rem
		{"rem", "2.5rem", 2.5, "rem", false},
		{"rem integer", "1rem", 1, "rem", false},

		// Em
		{"em", "1.5em", 1.5, "em", false},

		// Percentage
		{"percent", "50%", 50, "%", false},
		{"percent decimal", "33.33%", 33.33, "%", false},

		// Viewport units
		{"vw", "100vw", 100, "vw", false},
		{"vh", "50vh", 50, "vh", false},
		{"vmin", "25vmin", 25, "vmin", false},
		{"vmax", "75vmax", 75, "vmax", false},

		// Time units
		{"seconds", "0.3s", 0.3, "s", false},
		{"milliseconds", "300ms", 300, "ms", false},

		// Angle units
		{"degrees", "45deg", 45, "deg", false},
		{"radians", "3.14rad", 3.14, "rad", false},
		{"turns", "0.5turn", 0.5, "turn", false},

		// Unitless
		{"unitless integer", "42", 42, "", false},
		{"unitless decimal", "1.5", 1.5, "", false},
		{"unitless zero", "0", 0, "", false},

		// Negative values
		{"negative px", "-10px", -10, "px", false},
		{"negative decimal", "-2.5rem", -2.5, "rem", false},

		// Edge cases
		{"leading decimal", ".5rem", 0.5, "rem", false},

		// Errors
		{"empty string", "", 0, "", true},
		{"just unit", "px", 0, "", true},
		{"invalid unit", "10xyz", 0, "", true},
		{"letters", "abc", 0, "", true},
		{"space in middle", "10 px", 0, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dim, err := ParseDimension(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseDimension(%q) expected error, got nil", tt.input)
				}
				return
			}

			if err != nil {
				t.Fatalf("ParseDimension(%q) unexpected error: %v", tt.input, err)
			}

			if dim.Value != tt.wantValue {
				t.Errorf("ParseDimension(%q).Value = %v, want %v", tt.input, dim.Value, tt.wantValue)
			}

			if dim.Unit != tt.wantUnit {
				t.Errorf("ParseDimension(%q).Unit = %q, want %q", tt.input, dim.Unit, tt.wantUnit)
			}
		})
	}
}

func TestDimension_String(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		dim  Dimension
		want string
	}{
		{"integer px", Dimension{10, "px"}, "10px"},
		{"decimal rem", Dimension{2.5, "rem"}, "2.5rem"},
		{"zero", Dimension{0, "px"}, "0px"},
		{"unitless", Dimension{42, ""}, "42"},
		{"percentage", Dimension{50, "%"}, "50%"},
		{"negative", Dimension{-10, "px"}, "-10px"},
		{"whole number stored as float", Dimension{16.0, "px"}, "16px"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.dim.String()
			if got != tt.want {
				t.Errorf("Dimension{%v, %q}.String() = %q, want %q", tt.dim.Value, tt.dim.Unit, got, tt.want)
			}
		})
	}
}

func TestDimension_Add(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		d1      Dimension
		d2      Dimension
		want    Dimension
		wantErr bool
	}{
		{"same unit", Dimension{10, "px"}, Dimension{5, "px"}, Dimension{15, "px"}, false},
		{"same unit rem", Dimension{1.5, "rem"}, Dimension{0.5, "rem"}, Dimension{2, "rem"}, false},
		{"different units", Dimension{10, "px"}, Dimension{1, "rem"}, Dimension{}, true},
		{"unitless", Dimension{10, ""}, Dimension{5, ""}, Dimension{15, ""}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := tt.d1.Add(tt.d2)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Add() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Add() unexpected error: %v", err)
			}

			if got != tt.want {
				t.Errorf("Add() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDimension_Subtract(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		d1      Dimension
		d2      Dimension
		want    Dimension
		wantErr bool
	}{
		{"same unit", Dimension{10, "px"}, Dimension{3, "px"}, Dimension{7, "px"}, false},
		{"result negative", Dimension{5, "px"}, Dimension{10, "px"}, Dimension{-5, "px"}, false},
		{"different units", Dimension{10, "px"}, Dimension{1, "rem"}, Dimension{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := tt.d1.Subtract(tt.d2)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Subtract() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Subtract() unexpected error: %v", err)
			}

			if got != tt.want {
				t.Errorf("Subtract() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDimension_Multiply(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		dim    Dimension
		scalar float64
		want   Dimension
	}{
		{"double", Dimension{10, "px"}, 2, Dimension{20, "px"}},
		{"half", Dimension{10, "px"}, 0.5, Dimension{5, "px"}},
		{"by zero", Dimension{10, "px"}, 0, Dimension{0, "px"}},
		{"scale factor", Dimension{2.5, "rem"}, 0.6, Dimension{1.5, "rem"}},
		{"negative", Dimension{10, "px"}, -1, Dimension{-10, "px"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.dim.Multiply(tt.scalar)
			if got != tt.want {
				t.Errorf("Multiply(%v) = %v, want %v", tt.scalar, got, tt.want)
			}
		})
	}
}

func TestDimension_Divide(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		dim     Dimension
		scalar  float64
		want    Dimension
		wantErr bool
	}{
		{"divide by 2", Dimension{10, "px"}, 2, Dimension{5, "px"}, false},
		{"divide by 0.5", Dimension{10, "px"}, 0.5, Dimension{20, "px"}, false},
		{"divide by zero", Dimension{10, "px"}, 0, Dimension{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := tt.dim.Divide(tt.scalar)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Divide() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Divide() unexpected error: %v", err)
			}

			if got != tt.want {
				t.Errorf("Divide(%v) = %v, want %v", tt.scalar, got, tt.want)
			}
		})
	}
}

func TestDimension_IsZero(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		dim  Dimension
		want bool
	}{
		{"zero px", Dimension{0, "px"}, true},
		{"non-zero", Dimension{10, "px"}, false},
		{"zero unitless", Dimension{0, ""}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.dim.IsZero(); got != tt.want {
				t.Errorf("IsZero() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsDimension(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input string
		want  bool
	}{
		{"10px", true},
		{"2.5rem", true},
		{"100%", true},
		{"0", true},
		{"abc", false},
		{"", false},
		{"10xyz", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			if got := IsDimension(tt.input); got != tt.want {
				t.Errorf("IsDimension(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestMustParseDimension(t *testing.T) {
	t.Parallel()
	// Test successful parse
	dim := MustParseDimension("10px")
	if dim.Value != 10 || dim.Unit != "px" {
		t.Errorf("MustParseDimension(\"10px\") = %v, want {10, px}", dim)
	}

	// Test panic on invalid input
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustParseDimension with invalid input should panic")
		}
	}()
	MustParseDimension("invalid")
}
