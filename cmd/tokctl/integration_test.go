// tokctl/cmd/tokctl/integration_test.go
package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestMain ensures the tokctl binary is built before running tests
func TestMain(m *testing.M) {
	// Build the binary
	cmd := exec.Command("go", "build", "-o", "../../.build/tokctl-test", ".")
	if err := cmd.Run(); err != nil {
		panic("failed to build tokctl binary: " + err.Error())
	}

	// Run tests
	code := m.Run()

	// Cleanup
	os.RemoveAll("../../.build")

	os.Exit(code)
}

func getTokctlPath() string {
	return "../../.build/tokctl-test"
}

func TestIntegration_Init(t *testing.T) {
	tmpDir := t.TempDir()

	cmd := exec.Command(getTokctlPath(), "init", tmpDir)
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
	fixtureDir := "../../testdata/fixtures/valid"

	cmd := exec.Command(getTokctlPath(), "validate", fixtureDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("validate command failed on valid input: %v\nOutput: %s", err, output)
	}

	if !strings.Contains(string(output), "Validation Passed") {
		t.Errorf("Expected validation success message, got: %s", output)
	}
}

func TestIntegration_Validate_BrokenReference(t *testing.T) {
	fixtureDir := "../../testdata/fixtures/invalid"

	cmd := exec.Command(getTokctlPath(), "validate", fixtureDir)
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
	fixtureDir := "../../testdata/fixtures/valid"
	outputDir := t.TempDir()

	cmd := exec.Command(getTokctlPath(), "build", fixtureDir, "--output", outputDir)
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
	fixtureDir := "../../testdata/fixtures/extends"
	outputDir := t.TempDir()

	cmd := exec.Command(getTokctlPath(), "build", fixtureDir, "--output", outputDir)
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
	fixtureDir := "../../testdata/fixtures/valid"
	outputDir := t.TempDir()

	cmd := exec.Command(getTokctlPath(), "build", fixtureDir, "--output", outputDir)
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
	fixtureDir := "../../testdata/fixtures/extends"
	outputDir := t.TempDir()

	cmd := exec.Command(getTokctlPath(), "build", fixtureDir, "--output", outputDir)
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
	fixtureDir := "../../examples/components"
	outputDir := t.TempDir()

	cmd := exec.Command(getTokctlPath(), "build", fixtureDir, "--output", outputDir)
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
	fixtureDir := "../../testdata/fixtures/valid"
	outputDir := t.TempDir()

	cmd := exec.Command(getTokctlPath(), "build", fixtureDir, "--format", "catalog", "--output", outputDir)
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

	// Should have catalog structure
	expectedStrings := []string{
		"\"meta\":",
		"\"tokens\":",
		"\"components\":",
		"\"version\":",
		"\"generated_at\":",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(contentStr, expected) {
			t.Errorf("Expected catalog to contain '%s', but it didn't.\nOutput:\n%s", expected, contentStr)
		}
	}
}

func TestIntegration_Build_InvalidFormat(t *testing.T) {
	fixtureDir := "../../testdata/fixtures/valid"
	outputDir := t.TempDir()

	cmd := exec.Command(getTokctlPath(), "build", fixtureDir, "--format", "invalid-format", "--output", outputDir)
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
	// Test complete workflow: init -> validate -> build
	tmpDir := t.TempDir()

	// Step 1: Init
	cmd := exec.Command(getTokctlPath(), "init", tmpDir)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("init failed: %v\nOutput: %s", err, output)
	}

	// Step 2: Validate
	cmd = exec.Command(getTokctlPath(), "validate", tmpDir)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("validate failed: %v\nOutput: %s", err, output)
	}

	// Step 3: Build
	outputDir := filepath.Join(tmpDir, "dist")
	cmd = exec.Command(getTokctlPath(), "build", tmpDir, "--output", outputDir)
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
	// Verify that $extends actually works correctly
	fixtureDir := "../../testdata/fixtures/extends"

	// First validate
	cmd := exec.Command(getTokctlPath(), "validate", fixtureDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("validate failed on extends fixture: %v\nOutput: %s", err, output)
	}

	// Then build
	outputDir := t.TempDir()
	cmd = exec.Command(getTokctlPath(), "build", fixtureDir, "--output", outputDir)
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
