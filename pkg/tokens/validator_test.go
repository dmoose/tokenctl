package tokens

import (
	"strings"
	"testing"
)

func TestValidator_BrokenReferences(t *testing.T) {
	tests := []struct {
		name           string
		tokens         map[string]interface{}
		expectErrors   bool
		expectedErrMsg string
	}{
		{
			name: "Valid References",
			tokens: map[string]interface{}{
				"color": map[string]interface{}{
					"primary": map[string]interface{}{
						"$value": "#3b82f6",
					},
					"secondary": map[string]interface{}{
						"$value": "{color.primary}",
					},
				},
			},
			expectErrors: false,
		},
		{
			name: "Missing Reference",
			tokens: map[string]interface{}{
				"color": map[string]interface{}{
					"primary": map[string]interface{}{
						"$value": "{color.nonexistent}",
					},
				},
			},
			expectErrors:   true,
			expectedErrMsg: "reference not found",
		},
		{
			name: "Deep Chain Valid",
			tokens: map[string]interface{}{
				"a": map[string]interface{}{
					"$value": "{b}",
				},
				"b": map[string]interface{}{
					"$value": "{c}",
				},
				"c": map[string]interface{}{
					"$value": "final-value",
				},
			},
			expectErrors: false,
		},
		{
			name: "Deep Chain Broken",
			tokens: map[string]interface{}{
				"a": map[string]interface{}{
					"$value": "{b}",
				},
				"b": map[string]interface{}{
					"$value": "{c}",
				},
				"c": map[string]interface{}{
					"$value": "{missing}",
				},
			},
			expectErrors:   true,
			expectedErrMsg: "reference not found",
		},
		{
			name: "Multiple Broken References",
			tokens: map[string]interface{}{
				"a": map[string]interface{}{
					"$value": "{missing1}",
				},
				"b": map[string]interface{}{
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
		tokens         map[string]interface{}
		expectedErrMsg string
	}{
		{
			name: "Direct Cycle",
			tokens: map[string]interface{}{
				"a": map[string]interface{}{
					"$value": "{a}",
				},
			},
			expectedErrMsg: "circular dependency",
		},
		{
			name: "Two-Node Cycle",
			tokens: map[string]interface{}{
				"a": map[string]interface{}{
					"$value": "{b}",
				},
				"b": map[string]interface{}{
					"$value": "{a}",
				},
			},
			expectedErrMsg: "circular dependency",
		},
		{
			name: "Three-Node Cycle",
			tokens: map[string]interface{}{
				"a": map[string]interface{}{
					"$value": "{b}",
				},
				"b": map[string]interface{}{
					"$value": "{c}",
				},
				"c": map[string]interface{}{
					"$value": "{a}",
				},
			},
			expectedErrMsg: "circular dependency",
		},
		{
			name: "Nested Path Cycle",
			tokens: map[string]interface{}{
				"color": map[string]interface{}{
					"primary": map[string]interface{}{
						"$value": "{color.secondary}",
					},
					"secondary": map[string]interface{}{
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
		tokens       map[string]interface{}
		expectErrors bool
		errPath      string
	}{
		{
			name: "Valid Schema",
			tokens: map[string]interface{}{
				"color": map[string]interface{}{
					"primary": map[string]interface{}{
						"$value": "#3b82f6",
						"$type":  "color",
					},
				},
			},
			expectErrors: false,
		},
		{
			name: "Invalid Child - Primitive Value",
			tokens: map[string]interface{}{
				"color": map[string]interface{}{
					"invalid": "should-be-object",
				},
			},
			expectErrors: true,
			errPath:      "color.invalid",
		},
		{
			name: "Invalid Child - Array",
			tokens: map[string]interface{}{
				"spacing": map[string]interface{}{
					"values": []interface{}{1, 2, 3},
				},
			},
			expectErrors: true,
			errPath:      "spacing.values",
		},
		{
			name: "Mixed Valid and Invalid",
			tokens: map[string]interface{}{
				"tokens": map[string]interface{}{
					"valid": map[string]interface{}{
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
			tokens: map[string]interface{}{
				"level1": map[string]interface{}{
					"level2": map[string]interface{}{
						"level3": map[string]interface{}{
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
			tokens: map[string]interface{}{
				"$schema":  "https://example.com/schema",
				"$version": "1.0.0",
				"color": map[string]interface{}{
					"primary": map[string]interface{}{
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
	tokens := map[string]interface{}{
		"broken1": map[string]interface{}{
			"$value": "{missing1}",
		},
		"broken2": map[string]interface{}{
			"$value": "{missing2}",
		},
		"invalid": 123, // Schema error
		"cycle1": map[string]interface{}{
			"$value": "{cycle2}",
		},
		"cycle2": map[string]interface{}{
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
