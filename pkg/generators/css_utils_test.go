// tokctl/pkg/generators/css_utils_test.go
package generators

import (
	"testing"
)

func TestSerializeValue(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected string
	}{
		{
			name:     "String value",
			value:    "10px",
			expected: "10px",
		},
		{
			name:     "Integer value",
			value:    42,
			expected: "42",
		},
		{
			name:     "Float value",
			value:    3.14,
			expected: "3.14",
		},
		{
			name:     "Array with strings",
			value:    []interface{}{"10px", "20px", "30px"},
			expected: "10px, 20px, 30px",
		},
		{
			name:     "Array with mixed types",
			value:    []interface{}{"0px", 1, "2px", 3.5},
			expected: "0px, 1, 2px, 3.5",
		},
		{
			name:     "Empty array",
			value:    []interface{}{},
			expected: "",
		},
		{
			name:     "Single element array",
			value:    []interface{}{"value"},
			expected: "value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SerializeValue(tt.value)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestSerializeValueForProperty_SpaceSeparated(t *testing.T) {
	tests := []struct {
		name     string
		property string
		value    interface{}
		expected string
	}{
		{
			name:     "Margin with 4 values",
			property: "margin",
			value:    []interface{}{"10px", "20px", "10px", "20px"},
			expected: "10px 20px 10px 20px",
		},
		{
			name:     "Padding with 2 values",
			property: "padding",
			value:    []interface{}{"1rem", "2rem"},
			expected: "1rem 2rem",
		},
		{
			name:     "Border-width single value array",
			property: "border-width",
			value:    []interface{}{"2px"},
			expected: "2px",
		},
		{
			name:     "Border-radius with 4 values",
			property: "border-radius",
			value:    []interface{}{"4px", "4px", "0", "0"},
			expected: "4px 4px 0 0",
		},
		{
			name:     "Gap (flexbox/grid)",
			property: "gap",
			value:    []interface{}{"1rem", "2rem"},
			expected: "1rem 2rem",
		},
		{
			name:     "Grid-template-columns",
			property: "grid-template-columns",
			value:    []interface{}{"1fr", "2fr", "1fr"},
			expected: "1fr 2fr 1fr",
		},
		{
			name:     "Border shorthand",
			property: "border",
			value:    []interface{}{"1px", "solid", "#000"},
			expected: "1px solid #000",
		},
		{
			name:     "Background-size",
			property: "background-size",
			value:    []interface{}{"cover", "contain"},
			expected: "cover contain",
		},
		{
			name:     "Background-position",
			property: "background-position",
			value:    []interface{}{"center", "top"},
			expected: "center top",
		},
		{
			name:     "Inset",
			property: "inset",
			value:    []interface{}{"0", "0", "0", "0"},
			expected: "0 0 0 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SerializeValueForProperty(tt.property, tt.value)
			if result != tt.expected {
				t.Errorf("Property '%s': Expected '%s', got '%s'", tt.property, tt.expected, result)
			}
		})
	}
}

func TestSerializeValueForProperty_CommaSeparated(t *testing.T) {
	tests := []struct {
		name     string
		property string
		value    interface{}
		expected string
	}{
		{
			name:     "Font-family",
			property: "font-family",
			value:    []interface{}{"Arial", "sans-serif"},
			expected: "Arial, sans-serif",
		},
		{
			name:     "Box-shadow multiple layers",
			property: "box-shadow",
			value:    []interface{}{"0 1px 2px rgba(0,0,0,0.1)", "0 2px 4px rgba(0,0,0,0.2)"},
			expected: "0 1px 2px rgba(0,0,0,0.1), 0 2px 4px rgba(0,0,0,0.2)",
		},
		{
			name:     "Text-shadow",
			property: "text-shadow",
			value:    []interface{}{"1px 1px 2px black", "0 0 1em red"},
			expected: "1px 1px 2px black, 0 0 1em red",
		},
		{
			name:     "Transform multiple functions",
			property: "transform",
			value:    []interface{}{"rotate(45deg)", "scale(1.5)", "translate(10px, 20px)"},
			expected: "rotate(45deg), scale(1.5), translate(10px, 20px)",
		},
		{
			name:     "Transition multiple properties",
			property: "transition",
			value:    []interface{}{"opacity 0.3s ease", "transform 0.2s linear"},
			expected: "opacity 0.3s ease, transform 0.2s linear",
		},
		{
			name:     "Animation",
			property: "animation",
			value:    []interface{}{"slide 1s ease-in", "fade 0.5s"},
			expected: "slide 1s ease-in, fade 0.5s",
		},
		{
			name:     "Background-image multiple layers",
			property: "background-image",
			value:    []interface{}{"url(image1.png)", "url(image2.png)"},
			expected: "url(image1.png), url(image2.png)",
		},
		{
			name:     "Filter",
			property: "filter",
			value:    []interface{}{"blur(5px)", "brightness(0.8)"},
			expected: "blur(5px), brightness(0.8)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SerializeValueForProperty(tt.property, tt.value)
			if result != tt.expected {
				t.Errorf("Property '%s': Expected '%s', got '%s'", tt.property, tt.expected, result)
			}
		})
	}
}

func TestSerializeValueForProperty_VendorPrefixes(t *testing.T) {
	tests := []struct {
		name     string
		property string
		value    interface{}
		expected string
	}{
		{
			name:     "Webkit-transform",
			property: "-webkit-transform",
			value:    []interface{}{"rotate(45deg)", "scale(1.5)"},
			expected: "rotate(45deg), scale(1.5)",
		},
		{
			name:     "Moz-box-shadow",
			property: "-moz-box-shadow",
			value:    []interface{}{"0 1px 2px black", "0 2px 4px red"},
			expected: "0 1px 2px black, 0 2px 4px red",
		},
		{
			name:     "Webkit-border-radius (space-separated)",
			property: "-webkit-border-radius",
			value:    []interface{}{"4px", "4px", "0", "0"},
			expected: "4px 4px 0 0",
		},
		{
			name:     "Ms-flex (space-separated)",
			property: "-ms-flex",
			value:    []interface{}{"1", "1", "auto"},
			expected: "1 1 auto",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SerializeValueForProperty(tt.property, tt.value)
			if result != tt.expected {
				t.Errorf("Property '%s': Expected '%s', got '%s'", tt.property, tt.expected, result)
			}
		})
	}
}

func TestSerializeValueForProperty_CaseInsensitive(t *testing.T) {
	tests := []struct {
		name     string
		property string
		value    interface{}
		expected string
	}{
		{
			name:     "MARGIN uppercase",
			property: "MARGIN",
			value:    []interface{}{"10px", "20px"},
			expected: "10px 20px",
		},
		{
			name:     "Padding mixed case",
			property: "PaDDinG",
			value:    []interface{}{"1rem", "2rem"},
			expected: "1rem 2rem",
		},
		{
			name:     "Font-Family mixed case",
			property: "Font-Family",
			value:    []interface{}{"Arial", "sans-serif"},
			expected: "Arial, sans-serif",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SerializeValueForProperty(tt.property, tt.value)
			if result != tt.expected {
				t.Errorf("Property '%s': Expected '%s', got '%s'", tt.property, tt.expected, result)
			}
		})
	}
}

func TestSerializeValueForProperty_NonArrayValues(t *testing.T) {
	tests := []struct {
		name     string
		property string
		value    interface{}
		expected string
	}{
		{
			name:     "String value for margin",
			property: "margin",
			value:    "10px 20px",
			expected: "10px 20px",
		},
		{
			name:     "String value for font-family",
			property: "font-family",
			value:    "Arial, sans-serif",
			expected: "Arial, sans-serif",
		},
		{
			name:     "Integer value",
			property: "z-index",
			value:    100,
			expected: "100",
		},
		{
			name:     "Float value",
			property: "opacity",
			value:    0.75,
			expected: "0.75",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SerializeValueForProperty(tt.property, tt.value)
			if result != tt.expected {
				t.Errorf("Property '%s': Expected '%s', got '%s'", tt.property, tt.expected, result)
			}
		})
	}
}

func TestSerializeValueForProperty_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		property string
		value    interface{}
		expected string
	}{
		{
			name:     "Empty array",
			property: "margin",
			value:    []interface{}{},
			expected: "",
		},
		{
			name:     "Single element array - space-separated property",
			property: "padding",
			value:    []interface{}{"1rem"},
			expected: "1rem",
		},
		{
			name:     "Single element array - comma-separated property",
			property: "font-family",
			value:    []interface{}{"monospace"},
			expected: "monospace",
		},
		{
			name:     "Unknown property defaults to comma separation",
			property: "custom-property",
			value:    []interface{}{"value1", "value2"},
			expected: "value1, value2",
		},
		{
			name:     "Nil-like behavior with empty string",
			property: "color",
			value:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SerializeValueForProperty(tt.property, tt.value)
			if result != tt.expected {
				t.Errorf("Property '%s': Expected '%s', got '%s'", tt.property, tt.expected, result)
			}
		})
	}
}

func TestGetArraySeparator(t *testing.T) {
	spaceSeparated := []string{
		"margin",
		"padding",
		"border-width",
		"border-radius",
		"gap",
		"grid-template-columns",
		"-webkit-border-radius",
	}

	commaSeparated := []string{
		"font-family",
		"box-shadow",
		"text-shadow",
		"transform",
		"transition",
		"animation",
		"filter",
		"background-image",
		"unknown-property",
	}

	for _, prop := range spaceSeparated {
		t.Run(prop+" should be space-separated", func(t *testing.T) {
			sep := getArraySeparator(prop)
			if sep != " " {
				t.Errorf("Expected space separator for '%s', got '%s'", prop, sep)
			}
		})
	}

	for _, prop := range commaSeparated {
		t.Run(prop+" should be comma-separated", func(t *testing.T) {
			sep := getArraySeparator(prop)
			if sep != ", " {
				t.Errorf("Expected comma separator for '%s', got '%s'", prop, sep)
			}
		})
	}
}
