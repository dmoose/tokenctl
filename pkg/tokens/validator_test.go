package tokens

import (
	"strings"
	"testing"
)

func TestValidator_BrokenReferences(t *testing.T) {
	tests := []struct {
		name           string
		tokens         map[string]any
		expectErrors   bool
		expectedErrMsg string
	}{
		{
			name: "Valid References",
			tokens: map[string]any{
				"color": map[string]any{
					"primary": map[string]any{
						"$value": "#3b82f6",
					},
					"secondary": map[string]any{
						"$value": "{color.primary}",
					},
				},
			},
			expectErrors: false,
		},
		{
			name: "Missing Reference",
			tokens: map[string]any{
				"color": map[string]any{
					"primary": map[string]any{
						"$value": "{color.nonexistent}",
					},
				},
			},
			expectErrors:   true,
			expectedErrMsg: "reference not found",
		},
		{
			name: "Deep Chain Valid",
			tokens: map[string]any{
				"a": map[string]any{
					"$value": "{b}",
				},
				"b": map[string]any{
					"$value": "{c}",
				},
				"c": map[string]any{
					"$value": "final-value",
				},
			},
			expectErrors: false,
		},
		{
			name: "Deep Chain Broken",
			tokens: map[string]any{
				"a": map[string]any{
					"$value": "{b}",
				},
				"b": map[string]any{
					"$value": "{c}",
				},
				"c": map[string]any{
					"$value": "{missing}",
				},
			},
			expectErrors:   true,
			expectedErrMsg: "reference not found",
		},
		{
			name: "Multiple Broken References",
			tokens: map[string]any{
				"a": map[string]any{
					"$value": "{missing1}",
				},
				"b": map[string]any{
					"$value": "{missing2}",
				},
			},
			expectErrors: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dict := &Dictionary{Root: tt.tokens}
			validator := NewValidator()

			errors, err := validator.Validate(dict)
			if err != nil {
				t.Fatalf("Validation failed to run: %v", err)
			}

			hasErrors := len(errors) > 0

			if tt.expectErrors && !hasErrors {
				t.Errorf("Expected validation errors, got none")
			}

			if !tt.expectErrors && hasErrors {
				t.Errorf("Expected no validation errors, got: %v", errors)
			}

			if tt.expectedErrMsg != "" {
				found := false
				for _, verr := range errors {
					if strings.Contains(verr.Message, tt.expectedErrMsg) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected error message containing '%s', but not found in: %v", tt.expectedErrMsg, errors)
				}
			}
		})
	}
}

func TestValidator_CircularDependencies(t *testing.T) {
	tests := []struct {
		name           string
		tokens         map[string]any
		expectedErrMsg string
	}{
		{
			name: "Direct Cycle",
			tokens: map[string]any{
				"a": map[string]any{
					"$value": "{a}",
				},
			},
			expectedErrMsg: "circular dependency",
		},
		{
			name: "Two-Node Cycle",
			tokens: map[string]any{
				"a": map[string]any{
					"$value": "{b}",
				},
				"b": map[string]any{
					"$value": "{a}",
				},
			},
			expectedErrMsg: "circular dependency",
		},
		{
			name: "Three-Node Cycle",
			tokens: map[string]any{
				"a": map[string]any{
					"$value": "{b}",
				},
				"b": map[string]any{
					"$value": "{c}",
				},
				"c": map[string]any{
					"$value": "{a}",
				},
			},
			expectedErrMsg: "circular dependency",
		},
		{
			name: "Nested Path Cycle",
			tokens: map[string]any{
				"color": map[string]any{
					"primary": map[string]any{
						"$value": "{color.secondary}",
					},
					"secondary": map[string]any{
						"$value": "{color.primary}",
					},
				},
			},
			expectedErrMsg: "circular dependency",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dict := &Dictionary{Root: tt.tokens}
			validator := NewValidator()

			errors, err := validator.Validate(dict)
			if err != nil {
				t.Fatalf("Validation failed to run: %v", err)
			}

			if len(errors) == 0 {
				t.Fatalf("Expected circular dependency error, got none")
			}

			found := false
			for _, verr := range errors {
				if strings.Contains(verr.Message, tt.expectedErrMsg) {
					found = true
					break
				}
			}

			if !found {
				t.Errorf("Expected error containing '%s', got: %v", tt.expectedErrMsg, errors)
			}
		})
	}
}

func TestValidator_SchemaValidation(t *testing.T) {
	tests := []struct {
		name         string
		tokens       map[string]any
		expectErrors bool
		errPath      string
	}{
		{
			name: "Valid Schema",
			tokens: map[string]any{
				"color": map[string]any{
					"primary": map[string]any{
						"$value": "#3b82f6",
						"$type":  "color",
					},
				},
			},
			expectErrors: false,
		},
		{
			name: "Invalid Child - Primitive Value",
			tokens: map[string]any{
				"color": map[string]any{
					"invalid": "should-be-object",
				},
			},
			expectErrors: true,
			errPath:      "color.invalid",
		},
		{
			name: "Invalid Child - Array",
			tokens: map[string]any{
				"spacing": map[string]any{
					"values": []any{1, 2, 3},
				},
			},
			expectErrors: true,
			errPath:      "spacing.values",
		},
		{
			name: "Mixed Valid and Invalid",
			tokens: map[string]any{
				"tokens": map[string]any{
					"valid": map[string]any{
						"$value": "ok",
					},
					"invalid": 123,
				},
			},
			expectErrors: true,
			errPath:      "tokens.invalid",
		},
		{
			name: "Deeply Nested Invalid",
			tokens: map[string]any{
				"level1": map[string]any{
					"level2": map[string]any{
						"level3": map[string]any{
							"bad": []string{"array", "not", "allowed"},
						},
					},
				},
			},
			expectErrors: true,
			errPath:      "level1.level2.level3.bad",
		},
		{
			name: "Metadata Keys Ignored",
			tokens: map[string]any{
				"$schema":  "https://example.com/schema",
				"$version": "1.0.0",
				"color": map[string]any{
					"primary": map[string]any{
						"$value":       "#fff",
						"$type":        "color",
						"$description": "Primary color",
					},
				},
			},
			expectErrors: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dict := &Dictionary{Root: tt.tokens}
			validator := NewValidator()

			errors, err := validator.Validate(dict)
			if err != nil {
				t.Fatalf("Validation failed to run: %v", err)
			}

			hasErrors := len(errors) > 0

			if tt.expectErrors && !hasErrors {
				t.Errorf("Expected validation errors, got none")
			}

			if !tt.expectErrors && hasErrors {
				t.Errorf("Expected no validation errors, got: %v", errors)
			}

			if tt.errPath != "" {
				found := false
				for _, verr := range errors {
					if verr.Path == tt.errPath {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected error at path '%s', but not found. Got errors: %v", tt.errPath, errors)
				}
			}
		})
	}
}

func TestValidator_MultipleErrors(t *testing.T) {
	// Test that validator collects ALL errors, not just the first one
	tokens := map[string]any{
		"broken1": map[string]any{
			"$value": "{missing1}",
		},
		"broken2": map[string]any{
			"$value": "{missing2}",
		},
		"invalid": 123, // Schema error
		"cycle1": map[string]any{
			"$value": "{cycle2}",
		},
		"cycle2": map[string]any{
			"$value": "{cycle1}",
		},
	}

	dict := &Dictionary{Root: tokens}
	validator := NewValidator()

	errors, err := validator.Validate(dict)
	if err != nil {
		t.Fatalf("Validation failed to run: %v", err)
	}

	if len(errors) < 3 {
		t.Errorf("Expected at least 3 errors (2 broken refs, 1 schema, cycles), got %d: %v", len(errors), errors)
	}

	// Verify we have different types of errors
	hasRefError := false
	hasSchemaError := false
	hasCycleError := false

	for _, verr := range errors {
		if strings.Contains(verr.Message, "reference not found") {
			hasRefError = true
		}
		if strings.Contains(verr.Message, "expected object") {
			hasSchemaError = true
		}
		if strings.Contains(verr.Message, "circular dependency") {
			hasCycleError = true
		}
	}

	if !hasRefError {
		t.Error("Expected at least one reference error")
	}
	if !hasSchemaError {
		t.Error("Expected at least one schema error")
	}
	if !hasCycleError {
		t.Error("Expected at least one cycle error")
	}
}

func TestValidator_EmptyDictionary(t *testing.T) {
	dict := NewDictionary()
	validator := NewValidator()

	errors, err := validator.Validate(dict)
	if err != nil {
		t.Fatalf("Validation failed to run: %v", err)
	}

	if len(errors) != 0 {
		t.Errorf("Expected no errors for empty dictionary, got: %v", errors)
	}
}

func TestValidationError_Error(t *testing.T) {
	verr := ValidationError{
		Path:    "color.primary",
		Message: "test error message",
	}

	expected := "color.primary: test error message"
	if verr.Error() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, verr.Error())
	}
}

func TestValidationError_WithSourceFile(t *testing.T) {
	verr := ValidationError{
		Path:       "color.primary",
		Message:    "test error message",
		SourceFile: "tokens/colors.json",
	}

	expected := "color.primary [tokens/colors.json]: test error message"
	if verr.Error() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, verr.Error())
	}
}

func TestValidator_TypeValidation_Color(t *testing.T) {
	tests := []struct {
		name         string
		value        any
		expectErrors bool
	}{
		{"valid hex", "#3b82f6", false},
		{"valid hex short", "#fff", false},
		{"valid rgb", "rgb(255, 128, 0)", false},
		{"valid hsl", "hsl(180, 50%, 50%)", false},
		{"valid oklch", "oklch(50% 0.2 180)", false},
		{"valid named", "red", false},
		{"invalid color", "not-a-color", true},
		{"reference skipped", "{color.base}", false},
		{"contrast expression skipped", "contrast({color.base})", false},
		{"darken expression skipped", "darken({color.base}, 20%)", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dict := &Dictionary{
				Root: map[string]any{
					"color": map[string]any{
						"base": map[string]any{
							"$value": "#3b82f6",
							"$type":  "color",
						},
						"test": map[string]any{
							"$value": tt.value,
							"$type":  "color",
						},
					},
				},
			}

			validator := NewValidator()
			errors, err := validator.Validate(dict)
			if err != nil {
				t.Fatalf("Validation failed to run: %v", err)
			}

			hasErrors := len(errors) > 0
			if tt.expectErrors && !hasErrors {
				t.Error("Expected validation errors, got none")
			}
			if !tt.expectErrors && hasErrors {
				t.Errorf("Expected no validation errors, got: %v", errors)
			}
		})
	}
}

func TestValidator_TypeValidation_Dimension(t *testing.T) {
	tests := []struct {
		name         string
		value        any
		expectErrors bool
	}{
		{"valid px", "10px", false},
		{"valid rem", "1.5rem", false},
		{"valid percent", "100%", false},
		{"valid zero", "0", false},
		{"valid numeric zero", 0, false},
		{"invalid dimension", "invalid", true},
		{"reference skipped", "{size.base}", false},
		{"calc expression skipped", "calc({size.base} * 2)", false},
		{"scale expression skipped", "scale({size.base}, 1.5)", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dict := &Dictionary{
				Root: map[string]any{
					"size": map[string]any{
						"base": map[string]any{
							"$value": "1rem",
							"$type":  "dimension",
						},
						"test": map[string]any{
							"$value": tt.value,
							"$type":  "dimension",
						},
					},
				},
			}

			validator := NewValidator()
			errors, err := validator.Validate(dict)
			if err != nil {
				t.Fatalf("Validation failed to run: %v", err)
			}

			hasErrors := len(errors) > 0
			if tt.expectErrors && !hasErrors {
				t.Error("Expected validation errors, got none")
			}
			if !tt.expectErrors && hasErrors {
				t.Errorf("Expected no validation errors, got: %v", errors)
			}
		})
	}
}

func TestValidator_TypeValidation_Number(t *testing.T) {
	tests := []struct {
		name         string
		value        any
		expectErrors bool
	}{
		{"valid float", 0.5, false},
		{"valid int", 10, false},
		{"valid string number", "123", false},
		{"invalid string", "not-a-number", true},
		{"reference skipped", "{opacity.base}", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dict := &Dictionary{
				Root: map[string]any{
					"opacity": map[string]any{
						"base": map[string]any{
							"$value": 1.0,
							"$type":  "number",
						},
						"test": map[string]any{
							"$value": tt.value,
							"$type":  "number",
						},
					},
				},
			}

			validator := NewValidator()
			errors, err := validator.Validate(dict)
			if err != nil {
				t.Fatalf("Validation failed to run: %v", err)
			}

			hasErrors := len(errors) > 0
			if tt.expectErrors && !hasErrors {
				t.Error("Expected validation errors, got none")
			}
			if !tt.expectErrors && hasErrors {
				t.Errorf("Expected no validation errors, got: %v", errors)
			}
		})
	}
}

func TestValidator_TypeValidation_FontFamily(t *testing.T) {
	tests := []struct {
		name         string
		value        any
		expectErrors bool
	}{
		{"valid string", "Arial, sans-serif", false},
		{"valid array", []any{"Arial", "Helvetica", "sans-serif"}, false},
		{"empty string", "", true},
		{"empty array", []any{}, true},
		{"array with empty string", []any{"Arial", ""}, true},
		{"array with non-string", []any{"Arial", 123}, true},
		{"invalid type", 123, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dict := &Dictionary{
				Root: map[string]any{
					"font": map[string]any{
						"test": map[string]any{
							"$value": tt.value,
							"$type":  "fontFamily",
						},
					},
				},
			}

			validator := NewValidator()
			errors, err := validator.Validate(dict)
			if err != nil {
				t.Fatalf("Validation failed to run: %v", err)
			}

			hasErrors := len(errors) > 0
			if tt.expectErrors && !hasErrors {
				t.Error("Expected validation errors, got none")
			}
			if !tt.expectErrors && hasErrors {
				t.Errorf("Expected no validation errors, got: %v", errors)
			}
		})
	}
}

func TestValidator_TypeValidation_Effect(t *testing.T) {
	tests := []struct {
		name         string
		value        any
		expectErrors bool
	}{
		{"valid 0", 0, false},
		{"valid 1", 1, false},
		{"valid float 0", 0.0, false},
		{"valid float 1", 1.0, false},
		{"valid string 0", "0", false},
		{"valid string 1", "1", false},
		{"invalid number", 2, true},
		{"invalid float", 0.5, true},
		{"invalid string", "yes", true},
		{"reference skipped", "{effect.base}", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dict := &Dictionary{
				Root: map[string]any{
					"effect": map[string]any{
						"base": map[string]any{
							"$value": 1,
							"$type":  "effect",
						},
						"test": map[string]any{
							"$value": tt.value,
							"$type":  "effect",
						},
					},
				},
			}

			validator := NewValidator()
			errors, err := validator.Validate(dict)
			if err != nil {
				t.Fatalf("Validation failed to run: %v", err)
			}

			hasErrors := len(errors) > 0
			if tt.expectErrors && !hasErrors {
				t.Error("Expected validation errors, got none")
			}
			if !tt.expectErrors && hasErrors {
				t.Errorf("Expected no validation errors, got: %v", errors)
			}
		})
	}
}

func TestValidator_ConstraintValidation(t *testing.T) {
	tests := []struct {
		name         string
		token        map[string]any
		expectErrors bool
		errContains  string
	}{
		{
			name: "dimension in range",
			token: map[string]any{
				"$value": "2.5rem",
				"$type":  "dimension",
				"$min":   "1rem",
				"$max":   "5rem",
			},
			expectErrors: false,
		},
		{
			name: "dimension below min",
			token: map[string]any{
				"$value": "0.5rem",
				"$type":  "dimension",
				"$min":   "1rem",
				"$max":   "5rem",
			},
			expectErrors: true,
			errContains:  "less than minimum",
		},
		{
			name: "dimension above max",
			token: map[string]any{
				"$value": "10rem",
				"$type":  "dimension",
				"$min":   "1rem",
				"$max":   "5rem",
			},
			expectErrors: true,
			errContains:  "greater than maximum",
		},
		{
			name: "number in range",
			token: map[string]any{
				"$value": 0.5,
				"$type":  "number",
				"$min":   0.0,
				"$max":   1.0,
			},
			expectErrors: false,
		},
		{
			name: "number below min",
			token: map[string]any{
				"$value": -0.5,
				"$type":  "number",
				"$min":   0.0,
				"$max":   1.0,
			},
			expectErrors: true,
			errContains:  "less than minimum",
		},
		{
			name: "invalid constraint definition",
			token: map[string]any{
				"$value": "10px",
				"$min":   "20px",
				"$max":   "10px",
			},
			expectErrors: true,
			errContains:  "cannot be greater than",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dict := &Dictionary{
				Root: map[string]any{
					"test": map[string]any{
						"token": tt.token,
					},
				},
			}

			validator := NewValidator()
			errors, err := validator.Validate(dict)
			if err != nil {
				t.Fatalf("Validation failed to run: %v", err)
			}

			hasErrors := len(errors) > 0
			if tt.expectErrors && !hasErrors {
				t.Error("Expected validation errors, got none")
			}
			if !tt.expectErrors && hasErrors {
				t.Errorf("Expected no validation errors, got: %v", errors)
			}

			if tt.errContains != "" && hasErrors {
				found := false
				for _, verr := range errors {
					if strings.Contains(verr.Message, tt.errContains) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected error containing '%s', got: %v", tt.errContains, errors)
				}
			}
		})
	}
}

func TestValidator_SourceFileTracking(t *testing.T) {
	// Create a dictionary with source file annotations
	dict := &Dictionary{
		Root: map[string]any{
			"color": map[string]any{
				"primary": map[string]any{
					"$value": "{color.nonexistent}",
				},
				"secondary": map[string]any{
					"$value": "#fff",
				},
			},
		},
		SourceFiles: map[string]string{
			"color.primary":   "tokens/brand/colors.json",
			"color.secondary": "tokens/brand/colors.json",
		},
	}

	validator := NewValidator()
	errors, err := validator.Validate(dict)
	if err != nil {
		t.Fatalf("Validation failed to run: %v", err)
	}

	if len(errors) == 0 {
		t.Fatal("Expected validation errors, got none")
	}

	// Find the error for color.primary
	found := false
	for _, verr := range errors {
		if verr.Path == "color.primary" {
			found = true
			if verr.SourceFile != "tokens/brand/colors.json" {
				t.Errorf("Expected source file 'tokens/brand/colors.json', got '%s'", verr.SourceFile)
			}
			if !strings.Contains(verr.Error(), "[tokens/brand/colors.json]") {
				t.Errorf("Expected error message to contain source file, got: %s", verr.Error())
			}
		}
	}

	if !found {
		t.Error("Expected to find error for color.primary")
	}
}
