package tokens

import (
	"bytes"
	"log"
	"os"
	"strings"
	"testing"
)

func TestLoader_LoadBase(t *testing.T) {
	// Create a temporary directory structure for testing
	tmpDir := t.TempDir()

	// Create test token files
	tokensDir := tmpDir + "/tokens"
	if err := os.MkdirAll(tokensDir+"/brand", 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(tokensDir+"/themes", 0755); err != nil {
		t.Fatal(err)
	}

	// Write base token file
	baseContent := `{
		"color": {
			"primary": {
				"$value": "#3b82f6"
			}
		}
	}`
	if err := os.WriteFile(tokensDir+"/brand/colors.json", []byte(baseContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Write theme file (should be skipped)
	themeContent := `{
		"color": {
			"primary": {
				"$value": "#000"
			}
		}
	}`
	if err := os.WriteFile(tokensDir+"/themes/dark.json", []byte(themeContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Load base
	loader := NewLoader()
	dict, err := loader.LoadBase(tmpDir)
	if err != nil {
		t.Fatalf("LoadBase failed: %v", err)
	}

	// Verify base contains the brand color
	color, ok := dict.Root["color"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected color group in base")
	}

	primary, ok := color["primary"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected primary token in base")
	}

	if primary["$value"] != "#3b82f6" {
		t.Errorf("Expected #3b82f6, got %v", primary["$value"])
	}
}

func TestLoader_LoadThemes(t *testing.T) {
	tmpDir := t.TempDir()

	// Create themes directory
	themesDir := tmpDir + "/tokens/themes"
	if err := os.MkdirAll(themesDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Write theme files
	lightContent := `{
		"light": {
			"color": {
				"primary": {
					"$value": "#fff"
				}
			}
		}
	}`
	if err := os.WriteFile(themesDir+"/light.json", []byte(lightContent), 0644); err != nil {
		t.Fatal(err)
	}

	darkContent := `{
		"dark": {
			"color": {
				"primary": {
					"$value": "#000"
				}
			}
		}
	}`
	if err := os.WriteFile(themesDir+"/dark.json", []byte(darkContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Load themes
	loader := NewLoader()
	themes, err := loader.LoadThemes(tmpDir)
	if err != nil {
		t.Fatalf("LoadThemes failed: %v", err)
	}

	if len(themes) != 2 {
		t.Fatalf("Expected 2 themes, got %d", len(themes))
	}

	// Check light theme was unwrapped
	lightTheme, ok := themes["light"]
	if !ok {
		t.Fatal("Expected light theme")
	}

	color, ok := lightTheme.Root["color"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected color group in light theme")
	}

	primary, ok := color["primary"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected primary token in light theme")
	}

	if primary["$value"] != "#fff" {
		t.Errorf("Expected #fff, got %v", primary["$value"])
	}

	// Check dark theme
	darkTheme, ok := themes["dark"]
	if !ok {
		t.Fatal("Expected dark theme")
	}

	color, ok = darkTheme.Root["color"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected color group in dark theme")
	}

	primary, ok = color["primary"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected primary token in dark theme")
	}

	if primary["$value"] != "#000" {
		t.Errorf("Expected #000, got %v", primary["$value"])
	}
}

func TestLoader_MergeConflictWarnings(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	dict1 := &Dictionary{
		Root: map[string]interface{}{
			"spacing": map[string]interface{}{
				"base": map[string]interface{}{
					"$value": "1rem",
				},
			},
		},
	}

	dict2 := &Dictionary{
		Root: map[string]interface{}{
			"spacing": map[string]interface{}{
				"base": map[string]interface{}{
					"$value": "2rem",
				},
			},
		},
	}

	// Merge with warnings enabled
	if err := dict1.MergeWithPath(dict2, true, "test-file-2.json"); err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	// Check that warning was logged
	output := buf.String()
	if !strings.Contains(output, "Warning: Token 'spacing.base' redefined") {
		t.Errorf("Expected conflict warning, got: %s", output)
	}

	// Verify the second value won
	spacing := dict1.Root["spacing"].(map[string]interface{})
	base := spacing["base"].(map[string]interface{})
	if base["$value"] != "2rem" {
		t.Errorf("Expected 2rem (second value), got %v", base["$value"])
	}
}

func TestLoader_MergeNoWarnings(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	dict1 := &Dictionary{
		Root: map[string]interface{}{
			"spacing": map[string]interface{}{
				"base": map[string]interface{}{
					"$value": "1rem",
				},
			},
		},
	}

	dict2 := &Dictionary{
		Root: map[string]interface{}{
			"spacing": map[string]interface{}{
				"base": map[string]interface{}{
					"$value": "2rem",
				},
			},
		},
	}

	// Merge with warnings disabled
	if err := dict1.MergeWithPath(dict2, false, "test-file-2.json"); err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	// Check that NO warning was logged
	output := buf.String()
	if strings.Contains(output, "Warning") {
		t.Errorf("Expected no warnings, got: %s", output)
	}
}

func TestLoader_TypeMismatchWarning(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	dict1 := &Dictionary{
		Root: map[string]interface{}{
			"value": map[string]interface{}{
				"item": map[string]interface{}{
					"$value": "original",
				},
			},
		},
	}

	dict2 := &Dictionary{
		Root: map[string]interface{}{
			"value": "string-not-map",
		},
	}

	// Merge with warnings enabled
	if err := dict1.MergeWithPath(dict2, true, "test-file-2.json"); err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	// Check that warning was logged with type information
	output := buf.String()
	if !strings.Contains(output, "Warning: Token 'value' redefined") {
		t.Errorf("Expected type mismatch warning, got: %s", output)
	}
	if !strings.Contains(output, "overwriting") {
		t.Errorf("Expected 'overwriting' in warning, got: %s", output)
	}
}

func TestParseJSON(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expectErr bool
	}{
		{
			name: "Valid JSON",
			input: `{
				"color": {
					"primary": {
						"$value": "#fff"
					}
				}
			}`,
			expectErr: false,
		},
		{
			name:      "Invalid JSON",
			input:     `{"unclosed": `,
			expectErr: true,
		},
		{
			name:      "Empty Object",
			input:     `{}`,
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dict, err := ParseJSON(strings.NewReader(tt.input))

			if tt.expectErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if dict == nil {
					t.Error("Expected dictionary, got nil")
				}
			}
		})
	}
}

func TestDictionary_WriteJSON(t *testing.T) {
	dict := &Dictionary{
		Root: map[string]interface{}{
			"color": map[string]interface{}{
				"primary": map[string]interface{}{
					"$value": "#3b82f6",
					"$type":  "color",
				},
			},
		},
	}

	var buf bytes.Buffer
	if err := dict.WriteJSON(&buf); err != nil {
		t.Fatalf("WriteJSON failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "color") {
		t.Error("Expected 'color' in output")
	}
	if !strings.Contains(output, "#3b82f6") {
		t.Error("Expected '#3b82f6' in output")
	}

	// Should be valid JSON
	_, err := ParseJSON(strings.NewReader(output))
	if err != nil {
		t.Errorf("Output is not valid JSON: %v", err)
	}
}

func TestLoader_NonExistentDirectory(t *testing.T) {
	loader := NewLoader()

	_, err := loader.LoadBase("/nonexistent/path/that/should/not/exist")
	if err == nil {
		t.Error("Expected error for nonexistent directory, got nil")
	}
}

func TestLoader_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()

	tokensDir := tmpDir + "/tokens"
	if err := os.MkdirAll(tokensDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Write invalid JSON
	invalidContent := `{invalid json`
	if err := os.WriteFile(tokensDir+"/bad.json", []byte(invalidContent), 0644); err != nil {
		t.Fatal(err)
	}

	loader := NewLoader()
	_, err := loader.LoadBase(tmpDir)
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestDeepCopy(t *testing.T) {
	original := &Dictionary{
		Root: map[string]interface{}{
			"color": map[string]interface{}{
				"primary": map[string]interface{}{
					"$value": "#fff",
				},
			},
			"array": []interface{}{"a", "b", "c"},
		},
	}

	copy := original.DeepCopy()

	// Modify copy
	color := copy.Root["color"].(map[string]interface{})
	primary := color["primary"].(map[string]interface{})
	primary["$value"] = "#000"

	arr := copy.Root["array"].([]interface{})
	arr[0] = "modified"

	// Verify original is unchanged
	origColor := original.Root["color"].(map[string]interface{})
	origPrimary := origColor["primary"].(map[string]interface{})
	if origPrimary["$value"] != "#fff" {
		t.Error("Deep copy modified original map")
	}

	origArr := original.Root["array"].([]interface{})
	if origArr[0] != "a" {
		t.Error("Deep copy modified original array")
	}
}
