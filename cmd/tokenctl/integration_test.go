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

	// Should have catalog structure (v3.0)
	// Note: components is omitted when empty (correct behavior)
	expectedStrings := []string{
		"\"meta\":",
		"\"tokens\":",
		"\"version\": \"3.0\"",
		"\"tokenctl_version\":",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(contentStr, expected) {
			t.Errorf("Expected catalog to contain '%s', but it didn't.\nOutput:\n%s", expected, contentStr)
		}
	}

	// generated_at is opt-in. A default export carries no timestamp so
	// the same tokens produce the same bytes.
	if strings.Contains(contentStr, "\"generated_at\"") {
		t.Errorf("catalog carries generated_at without --generated-at:\n%s", contentStr)
	}
}

// TestIntegration_Build_CatalogDeterministic runs the real binary twice
// over the same tokens and compares bytes. Determinism is a property of
// the whole path, not just the generator: a flag default or a stray
// timestamp anywhere in the command would break it.
func TestIntegration_Build_CatalogDeterministic(t *testing.T) {
	t.Parallel()
	fixtureDir := "../../examples/baseline"

	read := func() string {
		outputDir := t.TempDir()
		cmd := exec.Command(getTokenctlPath(), "build", fixtureDir, "--format", "catalog", "--output", outputDir)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("build catalog failed: %v\nOutput: %s", err, out)
		}
		data, err := os.ReadFile(filepath.Join(outputDir, "catalog.json"))
		if err != nil {
			t.Fatalf("reading catalog: %v", err)
		}
		return string(data)
	}

	first := read()
	for i := range 3 {
		if again := read(); again != first {
			t.Fatalf("run %d produced different bytes than run 0", i+1)
		}
	}
}

// TestIntegration_Build_CatalogGeneratedAt: the stamp is available to a
// caller that wants one, and is exactly what they asked for.
func TestIntegration_Build_CatalogGeneratedAt(t *testing.T) {
	t.Parallel()
	outputDir := t.TempDir()
	cmd := exec.Command(getTokenctlPath(), "build", "../../testdata/fixtures/valid",
		"--format", "catalog", "--output", outputDir, "--generated-at", "2026-08-10T00:00:00Z")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build catalog failed: %v\nOutput: %s", err, out)
	}
	data, err := os.ReadFile(filepath.Join(outputDir, "catalog.json"))
	if err != nil {
		t.Fatalf("reading catalog: %v", err)
	}
	if !strings.Contains(string(data), `"generated_at": "2026-08-10T00:00:00Z"`) {
		t.Errorf("injected generated_at missing:\n%s", data)
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

// A misnamed component sub-block used to build clean and ship classes
// with no declarations. The build must now say so, and --strict-unknown-keys
// must turn that into a failure.
func TestIntegration_Build_UnknownKeysAreLoud(t *testing.T) {
	t.Parallel()
	tokensDir := t.TempDir()

	// "ratios" where the schema says "variants" — the production incident.
	tokenFile := filepath.Join(tokensDir, "aspectratio.json")
	if err := os.WriteFile(tokenFile, []byte(`{
      "$layer": "component",
      "components": {
        "aspect-ratio": {
          "$type": "component",
          "$class": "aspect-ratio",
          "base": { "position": "relative" },
          "ratios": {
            "square": { "$class": "aspect-ratio-square", "aspect-ratio": "1 / 1" }
          }
        }
      }
    }`), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	t.Run("warns by default and still builds", func(t *testing.T) {
		t.Parallel()
		outputDir := t.TempDir()
		cmd := exec.Command(getTokenctlPath(), "build", tokensDir, "--output", outputDir)
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("build should still succeed: %v\nOutput: %s", err, output)
		}
		out := string(output)
		if !strings.Contains(out, "ratios") {
			t.Errorf("expected the dropped key to be named in output:\n%s", out)
		}
		if !strings.Contains(out, "aspectratio.json") {
			t.Errorf("expected the source file to be named in output:\n%s", out)
		}
		if !strings.Contains(out, "Warning") {
			t.Errorf("expected a warning label in output:\n%s", out)
		}
	})

	t.Run("fails under --strict-unknown-keys", func(t *testing.T) {
		t.Parallel()
		outputDir := t.TempDir()
		cmd := exec.Command(getTokenctlPath(), "build", tokensDir,
			"--output", outputDir, "--strict-unknown-keys")
		output, err := cmd.CombinedOutput()
		if err == nil {
			t.Fatalf("strict build should fail, got success:\n%s", output)
		}
		if !strings.Contains(string(output), "ratios") {
			t.Errorf("expected the dropped key to be named:\n%s", output)
		}
	})

	t.Run("clean input stays silent", func(t *testing.T) {
		t.Parallel()
		cleanDir := t.TempDir()
		if err := os.WriteFile(filepath.Join(cleanDir, "aspectratio.json"), []byte(`{
          "$layer": "component",
          "components": {
            "aspect-ratio": {
              "$type": "component",
              "$class": "aspect-ratio",
              "// why": "fixed ratio box",
              "base": { "position": "relative" },
              "variants": {
                "square": { "$class": "aspect-ratio-square", "aspect-ratio": "1 / 1" }
              }
            }
          }
        }`), 0o644); err != nil {
			t.Fatalf("write fixture: %v", err)
		}

		cmd := exec.Command(getTokenctlPath(), "build", cleanDir,
			"--output", t.TempDir(), "--strict-unknown-keys")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("clean input must pass strict mode: %v\nOutput: %s", err, output)
		}
		if strings.Contains(string(output), "dropped input") {
			t.Errorf("clean input must not produce findings:\n%s", output)
		}
	})
}

func TestIntegration_Derive_JSONBuildsBackToTheSameVariables(t *testing.T) {
	t.Parallel()
	tokensDir := t.TempDir()
	outputDir := t.TempDir()

	// derive writes a token document...
	derived := filepath.Join(tokensDir, "derived.json")
	cmd := exec.Command(getTokenctlPath(), "derive", "--preset=teal", "-o", derived)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("derive failed: %v\n%s", err, out)
	}

	// ...which tokenctl build must turn back into exactly the custom
	// properties derive itself emits. If the CSS-variable-to-token-path
	// mapping were lossy, tokens would vanish here rather than at the
	// point they were written.
	cmd = exec.Command(getTokenctlPath(), "build", tokensDir, "--format=css", "--output", outputDir)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}
	built, err := os.ReadFile(filepath.Join(outputDir, "tokens.css"))
	if err != nil {
		t.Fatalf("read built css: %v", err)
	}

	cmd = exec.Command(getTokenctlPath(), "derive", "--preset=teal", "--format=css")
	directCSS, err := cmd.Output()
	if err != nil {
		t.Fatalf("derive --format=css failed: %v", err)
	}

	want := declarations(string(directCSS))
	got := declarations(string(built))
	if len(want) == 0 {
		t.Fatal("derive produced no declarations")
	}
	for name, value := range want {
		builtValue, ok := got[name]
		if !ok {
			t.Errorf("%s survived derive but not the build round trip", name)
			continue
		}
		if builtValue != value {
			t.Errorf("%s = %q after build, want %q", name, builtValue, value)
		}
	}
}

// declarations pulls "--name: value" pairs out of a stylesheet.
func declarations(css string) map[string]string {
	out := map[string]string{}
	for _, line := range strings.Split(css, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "--") {
			continue
		}
		name, value, ok := strings.Cut(strings.TrimSuffix(line, ";"), ":")
		if !ok {
			continue
		}
		out[strings.TrimSpace(name)] = strings.TrimSpace(value)
	}
	return out
}

func TestIntegration_Derive_RejectsOutOfRangeControls(t *testing.T) {
	t.Parallel()

	for _, args := range [][]string{
		{"derive", "--saturation=200"},
		{"derive", "--density=50"},
		{"derive", "--tint=-5"},
		{"derive", "--preset=chartreuse"},
		{"derive", "--type=comic-sans"},
		{"derive", "--from-hex=#nothex"},
		{"derive", "--format=xml"},
	} {
		cmd := exec.Command(getTokenctlPath(), args...)
		if out, err := cmd.CombinedOutput(); err == nil {
			t.Errorf("%v should fail, got success:\n%s", args, out)
		}
	}
}

func TestIntegration_Derive_DarkModeUsesThemeSelector(t *testing.T) {
	t.Parallel()

	cmd := exec.Command(getTokenctlPath(), "derive", "--preset=blue", "--dark", "--format=css")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("derive failed: %v", err)
	}
	// A dark theme written to :root would override the light theme it is
	// meant to sit beside.
	if !strings.Contains(string(out), `[data-theme="dark"] {`) {
		t.Errorf("dark mode should default to the theme attribute selector:\n%s", out)
	}
}
