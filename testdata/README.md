<!-- tokenctl/testdata/README.md -->
# Test Data

This directory contains test fixtures and expected outputs for tokenctl's automated test suite.

## Directory Structure

```
testdata/
├── fixtures/          # Input test cases
│   ├── valid/        # Valid token systems
│   ├── extends/      # Theme inheritance with $extends
│   ├── conflicts/    # Merge conflict scenarios
│   └── invalid/      # Invalid tokens for error testing
└── golden/           # Expected output files
    ├── valid.css     # Expected output for valid fixture
    └── extends.css   # Expected output for extends fixture
```

## Test Fixtures

### fixtures/valid/
Standard token system created with `tokenctl init`. Used to test:
- Basic token loading and resolution
- CSS generation
- Validation of well-formed tokens
- Baseline functionality

### fixtures/extends/
Demonstrates theme inheritance using `$extends`:
- Base brand tokens
- Light theme with overrides
- Dark theme that extends light theme
- Multi-level inheritance resolution
- Circular dependency detection

Used to test:
- Theme inheritance implementation
- `$extends` resolution
- Theme-specific CSS generation
- Diff calculation between themes

### fixtures/conflicts/
Contains intentionally conflicting token definitions across multiple files:
- Same token path defined in multiple files
- Type mismatches during merge
- Token vs group conflicts

Used to test:
- Merge conflict warnings
- Last-write-wins behavior
- Warning message generation

### fixtures/invalid/
Contains intentionally broken token systems:

**broken-ref.json**
- References non-existent token
- Tests missing reference detection

**circular.json**
- Three tokens that reference each other in a cycle
- Tests circular dependency detection

## Golden Files

Golden files contain the expected output for specific inputs. Tests compare generated output against these files to detect regressions.

### Updating Golden Files

When intentionally changing output format:

```bash
# Regenerate golden file for valid fixture
tokenctl build testdata/fixtures/valid --output=/tmp/test
cp /tmp/test/tokens.css testdata/golden/valid.css

# Regenerate golden file for extends fixture
tokenctl build testdata/fixtures/extends --output=/tmp/test
cp /tmp/test/tokens.css testdata/golden/extends.css
```

**Important:** Only update golden files when the change is intentional. Unexpected differences indicate a regression.

## Integration Tests

Integration tests in `cmd/tokenctl/integration_test.go` use these fixtures to test:

1. **Command-line interface** - All commands (init, validate, build)
2. **File I/O** - Loading from disk, writing output
3. **End-to-end workflows** - init → validate → build
4. **Error handling** - Invalid input, missing files, broken references
5. **Output validation** - Generated CSS matches golden files

## Adding New Test Fixtures

To add a new test case:

1. **Create fixture directory:**
   ```bash
   mkdir -p testdata/fixtures/my-test-case/tokens
   ```

2. **Add token files:**
   ```bash
   vim testdata/fixtures/my-test-case/tokens/example.json
   ```

3. **Generate golden file (if needed):**
   ```bash
   tokenctl build testdata/fixtures/my-test-case --output=/tmp/test
   cp /tmp/test/tokens.css testdata/golden/my-test-case.css
   ```

4. **Write test in integration_test.go:**
   ```go
   func TestIntegration_MyTestCase(t *testing.T) {
       fixtureDir := "../../testdata/fixtures/my-test-case"
       // ... test logic
   }
   ```

## Go Testing Conventions

This directory follows Go's standard testing conventions:

- `testdata/` is automatically ignored by `go` tool
- Not included in built binaries
- Fixtures are checked into version control
- Provides deterministic test inputs
- Enables offline testing

## Maintenance

- **Keep fixtures minimal** - Only include what's needed for the test
- **Document edge cases** - Add comments in fixture JSON for clarity
- **Update golden files carefully** - Review diffs before committing
- **Test both success and failure** - Include invalid fixtures for error paths
