// tokenctl/cmd/tokenctl/integration_test.go
package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestMain ensures the tokenctl binary is built before running tests
func TestMain(m *testing.M) {
	// Build the binary
	cmd := exec.Command("go", "build", "-o", "../../.build/tokenctl-test", ".")
	if err := cmd.Run(); err != nil {
		panic("failed to build tokenctl binary: " + err.Error())
	}

	// Run tests
	code := m.Run()

	// Cleanup
	_ = os.RemoveAll("../../.build")

	os.Exit(code)
}

func getTokenctlPath() string {
	return "../../.build/tokenctl-test"
}

func TestIntegration_Init(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	cmd := exec.Command(getTokenctlPath(), "init", tmpDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("init command failed: %v\nOutput: %s", err, output)
	}

	// Verify expected files were created
	expectedFiles := []string{
		filepath.Join(tmpDir, "tokens/brand/colors.json"),
		filepath.Join(tmpDir, "tokens/semantic/status.json"),
		filepath.Join(tmpDir, "tokens/spacing/scale.json"),
	}

	for _, file := range expectedFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			t.Errorf("Expected file not created: %s", file)
		}
	}

	// Verify output message
	if !strings.Contains(string(output), "Initializing new semantic token system") {
		t.Errorf("Expected initialization message in output: %s", output)
	}
}

func TestIntegration_Validate_Valid(t *testing.T) {
	t.Parallel()
	fixtureDir := "../../testdata/fixtures/valid"

	cmd := exec.Command(getTokenctlPath(), "validate", fixtureDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("validate command failed on valid input: %v\nOutput: %s", err, output)
	}

	if !strings.Contains(string(output), "Validation Passed") {
		t.Errorf("Expected validation success message, got: %s", output)
	}
}

func TestIntegration_Validate_BrokenReference(t *testing.T) {
	t.Parallel()
	fixtureDir := "../../testdata/fixtures/invalid"

	cmd := exec.Command(getTokenctlPath(), "validate", fixtureDir)
	output, err := cmd.CombinedOutput()

	// Should fail validation
	if err == nil {
		t.Fatalf("Expected validation to fail on broken reference, but it passed")
	}

	// Should contain error about reference
	if !strings.Contains(string(output), "reference not found") &&
		!strings.Contains(string(output), "circular dependency") {
		t.Errorf("Expected reference error in output, got: %s", output)
	}
}

func TestIntegration_Build_Valid(t *testing.T) {
	t.Parallel()
	fixtureDir := "../../testdata/fixtures/valid"
	outputDir := t.TempDir()

	cmd := exec.Command(getTokenctlPath(), "build", fixtureDir, "--output", outputDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("build command failed: %v\nOutput: %s", err, output)
	}

	// Verify output file was created
	outputFile := filepath.Join(outputDir, "tokens.css")
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Fatalf("Expected output file not created: %s", outputFile)
	}

	// Read output and verify it contains expected content
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	contentStr := string(content)
	expectedStrings := []string{
		"@import \"tailwindcss\"",
		"@theme {",
		"--color-brand-primary:",
		"--spacing-",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(contentStr, expected) {
			t.Errorf("Expected output to contain '%s', but it didn't.\nOutput:\n%s", expected, contentStr)
		}
	}
}

func TestIntegration_Build_WithThemes(t *testing.T) {
	t.Parallel()
	fixtureDir := "../../testdata/fixtures/extends"
	outputDir := t.TempDir()

	cmd := exec.Command(getTokenctlPath(), "build", fixtureDir, "--output", outputDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("build command failed: %v\nOutput: %s", err, output)
	}

	// Verify output file was created
	outputFile := filepath.Join(outputDir, "tokens.css")
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	contentStr := string(content)

	// Should have theme sections
	expectedStrings := []string{
		"@layer base {",
		"[data-theme=\"dark\"]",
		":root, [data-theme=\"light\"]",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(contentStr, expected) {
			t.Errorf("Expected output to contain '%s', but it didn't.\nOutput:\n%s", expected, contentStr)
		}
	}
}

func TestIntegration_Build_GoldenFile_Valid(t *testing.T) {
	t.Parallel()
	fixtureDir := "../../testdata/fixtures/valid"
	outputDir := t.TempDir()

	cmd := exec.Command(getTokenctlPath(), "build", fixtureDir, "--output", outputDir)
	_, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("build command failed: %v", err)
	}

	// Read generated output
	outputFile := filepath.Join(outputDir, "tokens.css")
	generated, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	// Read golden file
	goldenFile := "../../testdata/golden/valid.css"
	golden, err := os.ReadFile(goldenFile)
	if err != nil {
		t.Fatalf("Failed to read golden file: %v", err)
	}

	// Compare (normalize whitespace for comparison)
	generatedStr := strings.TrimSpace(string(generated))
	goldenStr := strings.TrimSpace(string(golden))

	if generatedStr != goldenStr {
		t.Errorf("Generated output doesn't match golden file.\n\nGenerated:\n%s\n\nGolden:\n%s", generatedStr, goldenStr)
	}
}

func TestIntegration_Build_GoldenFile_Extends(t *testing.T) {
	t.Parallel()
	fixtureDir := "../../testdata/fixtures/extends"
	outputDir := t.TempDir()

	cmd := exec.Command(getTokenctlPath(), "build", fixtureDir, "--output", outputDir)
	_, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("build command failed: %v", err)
	}

	// Read generated output
	outputFile := filepath.Join(outputDir, "tokens.css")
	generated, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	// Read golden file
	goldenFile := "../../testdata/golden/extends.css"
	golden, err := os.ReadFile(goldenFile)
	if err != nil {
		t.Fatalf("Failed to read golden file: %v", err)
	}

	// Compare
	generatedStr := strings.TrimSpace(string(generated))
	goldenStr := strings.TrimSpace(string(golden))

	if generatedStr != goldenStr {
		t.Errorf("Generated output doesn't match golden file.\n\nGenerated:\n%s\n\nGolden:\n%s", generatedStr, goldenStr)
	}
}

func TestIntegration_Build_Components(t *testing.T) {
	t.Parallel()
	fixtureDir := "../../examples/components"
	outputDir := t.TempDir()

	cmd := exec.Command(getTokenctlPath(), "build", fixtureDir, "--output", outputDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("build command failed: %v\nOutput: %s", err, output)
	}

	// Read output
	outputFile := filepath.Join(outputDir, "tokens.css")
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	contentStr := string(content)

	// Should have component layer
	expectedStrings := []string{
		"@layer components {",
		".btn-primary",
		".btn-secondary",
		".btn-success",
		".btn-error",
		".btn-sm",
		".btn-lg",
		"background-color:",
		":hover",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(contentStr, expected) {
			t.Errorf("Expected output to contain '%s', but it didn't.\nOutput:\n%s", expected, contentStr)
		}
	}
}

func TestIntegration_Build_Catalog(t *testing.T) {
	t.Parallel()
	fixtureDir := "../../testdata/fixtures/valid"
	outputDir := t.TempDir()

	cmd := exec.Command(getTokenctlPath(), "build", fixtureDir, "--format", "catalog", "--output", outputDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("build catalog command failed: %v\nOutput: %s", err, output)
	}

	// Verify catalog.json was created
	outputFile := filepath.Join(outputDir, "catalog.json")
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Fatalf("Expected catalog.json not created: %s", outputFile)
	}

	// Read and verify it's valid JSON
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read catalog file: %v", err)
	}

	contentStr := string(content)

	// Should have catalog structure (v2.1)
	// Note: components is omitted when empty (correct behavior)
	expectedStrings := []string{
		"\"meta\":",
		"\"tokens\":",
		"\"version\": \"2.1\"",
		"\"generated_at\":",
		"\"tokenctl_version\":",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(contentStr, expected) {
			t.Errorf("Expected catalog to contain '%s', but it didn't.\nOutput:\n%s", expected, contentStr)
		}
	}
}

func TestIntegration_Build_CatalogWithThemes(t *testing.T) {
	t.Parallel()
	fixtureDir := "../../testdata/fixtures/extends"
	outputDir := t.TempDir()

	cmd := exec.Command(getTokenctlPath(), "build", fixtureDir, "--format", "catalog", "--output", outputDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("build catalog command failed: %v\nOutput: %s", err, output)
	}

	// Verify catalog.json was created
	outputFile := filepath.Join(outputDir, "catalog.json")
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read catalog file: %v", err)
	}

	contentStr := string(content)

	// Should have themes section
	expectedStrings := []string{
		"\"themes\":",
		"\"light\":",
		"\"dark\":",
		"\"extends\": \"light\"",
		"\"tokens\":",
		"\"diff\":",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(contentStr, expected) {
			t.Errorf("Expected catalog to contain '%s', but it didn't.\nOutput:\n%s", expected, contentStr)
		}
	}

	// Verify dark theme has description from fixture
	if !strings.Contains(contentStr, "Dark theme extends light theme") {
		t.Errorf("Expected dark theme description in catalog output")
	}
}

func TestIntegration_Build_InvalidFormat(t *testing.T) {
	t.Parallel()
	fixtureDir := "../../testdata/fixtures/valid"
	outputDir := t.TempDir()

	cmd := exec.Command(getTokenctlPath(), "build", fixtureDir, "--format", "invalid-format", "--output", outputDir)
	output, err := cmd.CombinedOutput()

	// Should fail
	if err == nil {
		t.Fatalf("Expected build to fail with invalid format, but it succeeded")
	}

	// Should mention unknown format
	if !strings.Contains(string(output), "unknown format") {
		t.Errorf("Expected error about unknown format, got: %s", output)
	}
}

func TestIntegration_Workflow_InitValidateBuild(t *testing.T) {
	t.Parallel()
	// Test complete workflow: init -> validate -> build
	tmpDir := t.TempDir()

	// Step 1: Init
	cmd := exec.Command(getTokenctlPath(), "init", tmpDir)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("init failed: %v\nOutput: %s", err, output)
	}

	// Step 2: Validate
	cmd = exec.Command(getTokenctlPath(), "validate", tmpDir)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("validate failed: %v\nOutput: %s", err, output)
	}

	// Step 3: Build
	outputDir := filepath.Join(tmpDir, "dist")
	cmd = exec.Command(getTokenctlPath(), "build", tmpDir, "--output", outputDir)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %v\nOutput: %s", err, output)
	}

	// Verify output exists
	outputFile := filepath.Join(outputDir, "tokens.css")
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Errorf("Expected output file not created: %s", outputFile)
	}
}

func TestIntegration_ThemeInheritance_Extends(t *testing.T) {
	t.Parallel()
	// Verify that $extends actually works correctly
	fixtureDir := "../../testdata/fixtures/extends"

	// First validate
	cmd := exec.Command(getTokenctlPath(), "validate", fixtureDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("validate failed on extends fixture: %v\nOutput: %s", err, output)
	}

	// Then build
	outputDir := t.TempDir()
	cmd = exec.Command(getTokenctlPath(), "build", fixtureDir, "--output", outputDir)
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("build failed on extends fixture: %v\nOutput: %s", err, output)
	}

	// Verify dark theme only contains differences
	outputFile := filepath.Join(outputDir, "tokens.css")
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output: %v", err)
	}

	contentStr := string(content)

	// Dark theme should be present
	if !strings.Contains(contentStr, "[data-theme=\"dark\"]") {
		t.Error("Expected dark theme selector in output")
	}

	// Light theme should be present
	if !strings.Contains(contentStr, "[data-theme=\"light\"]") {
		t.Error("Expected light theme selector in output")
	}
}

// Multi-directory merge tests

func TestIntegration_Build_MultiDir(t *testing.T) {
	t.Parallel()
	baseDir := "../../testdata/fixtures/merge-base"
	extDir := "../../testdata/fixtures/merge-ext"
	outputDir := t.TempDir()

	cmd := exec.Command(getTokenctlPath(), "build", baseDir, extDir, "--output", outputDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("multi-dir build failed: %v\nOutput: %s", err, output)
	}

	content, err := os.ReadFile(filepath.Join(outputDir, "tokens.css"))
	if err != nil {
		t.Fatalf("Failed to read output: %v", err)
	}
	css := string(content)

	// Tokens from base
	for _, expected := range []string{
		"--color-brand-green-500:",
		"--color-semantic-success:",
	} {
		if !strings.Contains(css, expected) {
			t.Errorf("Expected base token %q in output", expected)
		}
	}

	// Tokens from extension
	for _, expected := range []string{
		"--color-brand-red-500:",
		"--color-semantic-danger:",
	} {
		if !strings.Contains(css, expected) {
			t.Errorf("Expected extension token %q in output", expected)
		}
	}
}

func TestIntegration_Build_MultiDir_Override(t *testing.T) {
	t.Parallel()
	baseDir := "../../testdata/fixtures/merge-base"
	extDir := "../../testdata/fixtures/merge-ext"
	outputDir := t.TempDir()

	cmd := exec.Command(getTokenctlPath(), "build", baseDir, extDir, "--format", "catalog", "--output", outputDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("multi-dir catalog build failed: %v\nOutput: %s", err, output)
	}

	content, err := os.ReadFile(filepath.Join(outputDir, "catalog.json"))
	if err != nil {
		t.Fatalf("Failed to read catalog: %v", err)
	}
	catalog := string(content)

	// Extension value (#2563eb) should win over base (#3b82f6)
	if !strings.Contains(catalog, "#2563eb") {
		t.Error("Expected extension override value #2563eb in catalog")
	}
	if strings.Contains(catalog, "#3b82f6") {
		t.Error("Base value #3b82f6 should be overridden by extension")
	}
}

func TestIntegration_Build_MultiDir_ComponentExtend(t *testing.T) {
	t.Parallel()
	baseDir := "../../testdata/fixtures/merge-base"
	extDir := "../../testdata/fixtures/merge-ext"
	outputDir := t.TempDir()

	cmd := exec.Command(getTokenctlPath(), "build", baseDir, extDir, "--output", outputDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("multi-dir build failed: %v\nOutput: %s", err, output)
	}

	content, err := os.ReadFile(filepath.Join(outputDir, "tokens.css"))
	if err != nil {
		t.Fatalf("Failed to read output: %v", err)
	}
	css := string(content)

	// All 3 button variants should be present (primary+success from base, danger from ext)
	for _, variant := range []string{"primary", "success", "danger"} {
		if !strings.Contains(css, "--button-"+variant+"-background-color:") {
			t.Errorf("Expected button variant %q in merged output", variant)
		}
	}
}

func TestIntegration_Build_MultiDir_ThemeMerge(t *testing.T) {
	t.Parallel()
	baseDir := "../../testdata/fixtures/merge-base"
	extDir := "../../testdata/fixtures/merge-ext"
	outputDir := t.TempDir()

	cmd := exec.Command(getTokenctlPath(), "build", baseDir, extDir, "--output", outputDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("multi-dir build failed: %v\nOutput: %s", err, output)
	}

	content, err := os.ReadFile(filepath.Join(outputDir, "tokens.css"))
	if err != nil {
		t.Fatalf("Failed to read output: %v", err)
	}
	css := string(content)

	// Dark theme should have overrides from both base and ext
	if !strings.Contains(css, `[data-theme="dark"]`) {
		t.Fatal("Expected dark theme selector")
	}
	// Base dark theme contributes blue-500 override
	if !strings.Contains(css, "--color-brand-blue-500: #60a5fa") {
		t.Error("Expected base dark theme override for blue-500")
	}
	// Ext dark theme contributes red-500 override
	if !strings.Contains(css, "--color-brand-red-500: #f87171") {
		t.Error("Expected ext dark theme override for red-500")
	}
}

func TestIntegration_Validate_MultiDir(t *testing.T) {
	t.Parallel()
	baseDir := "../../testdata/fixtures/merge-base"
	extDir := "../../testdata/fixtures/merge-ext"

	cmd := exec.Command(getTokenctlPath(), "validate", baseDir, extDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("multi-dir validate failed: %v\nOutput: %s", err, output)
	}

	if !strings.Contains(string(output), "Validation Passed") {
		t.Errorf("Expected validation to pass, got: %s", output)
	}
}

func TestIntegration_Build_SingleDir_BackwardCompat(t *testing.T) {
	t.Parallel()
	fixtureDir := "../../testdata/fixtures/valid"
	outputDir := t.TempDir()

	// Single dir should work exactly as before
	cmd := exec.Command(getTokenctlPath(), "build", fixtureDir, "--output", outputDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("single-dir build failed: %v\nOutput: %s", err, output)
	}

	content, err := os.ReadFile(filepath.Join(outputDir, "tokens.css"))
	if err != nil {
		t.Fatalf("Failed to read output: %v", err)
	}

	css := string(content)
	if !strings.Contains(css, "@import \"tailwindcss\"") {
		t.Error("Expected tailwind import in single-dir output")
	}
	if !strings.Contains(css, "--color-brand-primary:") {
		t.Error("Expected tokens in single-dir output")
	}
}
