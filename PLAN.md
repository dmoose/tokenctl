# tokenctl Development Plan

## Overview

tokenctl is the style spec engine for design systems. It validates, builds, and manages design token specifications. This document covers planned features for merge and diff capabilities to support multi-team design system workflows.

## Context: The Larger Vision

tokenctl fits into a component library workflow:

```
Component Libraries (templ, React, DaisyUI, etc.)
                    ↓
            scanners (templscan, etc.)
                    ↓
            STYLE SPEC (tokenctl format)
                    ↓
            generators (tokengen-templ, etc.)
                    ↓
        New components with proper token references
```

The merge/diff features support the middle layer: teams extending a base design system with project-specific tokens and components.

## Feature: `tokenctl merge`

### Purpose

Combine a base spec with a local spec, producing a merged output. Enforces that local specs extend (not override) the base.

### Usage

```bash
tokenctl merge --base=./company-tokens --local=./project-tokens --output=./merged
```

Or with shorthand:

```bash
tokenctl merge ./company-tokens ./project-tokens -o ./merged
```

### Behavior

1. Load base spec (all files in directory)
2. Load local spec (all files in directory)
3. Check for name collisions
4. If no collisions, write merged output
5. If collisions, error with details

### Collision Handling

Name collision = same token path in both specs (e.g., `color.primary` defined in both).

```
Error: Merge blocked - 2 name collisions detected

  color.primary
    base:  tokens/brand/colors.json:12
    local: tokens/overrides.json:5

  spacing.md
    base:  tokens/spacing/scale.json:8
    local: tokens/spacing.json:3

Resolve collisions before merging:
  - Remove duplicate definitions from local spec
  - Use references like {color.primary} instead of redefining
  - Run 'tokenctl diff' for analysis of potential duplicates
```

### Output Structure

Merged spec preserves directory structure from both sources:

```
merged/
├── base/           # or inline if flat preferred
│   └── ...
└── local/
    └── ...
```

Or flattened with source comments:

```json
{
  "color": {
    "primary": { "$value": "...", "$source": "base:brand/colors.json" },
    "widget-accent": { "$value": "...", "$source": "local:components/widget.json" }
  }
}
```

### Flags

| Flag | Description |
|------|-------------|
| `--base` | Path to base spec directory |
| `--local` | Path to local spec directory |
| `--output`, `-o` | Output directory for merged spec |
| `--flatten` | Merge into single file (default: preserve structure) |
| `--dry-run` | Check for collisions without writing output |

## Feature: `tokenctl diff`

### Purpose

Analyze two specs and surface potential issues for human/LLM review. Informational only - does not block operations.

### Usage

```bash
tokenctl diff ./company-tokens ./project-tokens
tokenctl diff ./company-tokens ./project-tokens --output=diff-report.txt
tokenctl diff ./company-tokens ./project-tokens --format=json
```

### Output (Default)

```
=== Token Diff Report ===

Base spec: ./company-tokens (47 tokens)
Local spec: ./project-tokens (12 tokens)

--- Name Collisions (blocking for merge) ---

  color.primary
    base:  oklch(49% 0.309 275)
    local: oklch(52% 0.3 275)

--- Identical Values (likely duplicates) ---

  base:spacing.md = local:spacing.widget-gap
    Both: 1rem
    Suggestion: local should reference {spacing.md}

  base:color.base-100 = local:color.widget-surface
    Both: #ffffff
    Suggestion: local should reference {color.base-100}

--- Similar Values (review recommended) ---

  Colors within 5% lightness:
    base:color.primary      oklch(49% 0.309 275)
    local:color.widget-accent oklch(52% 0.3 275)

  Dimensions within 4px:
    base:spacing.lg    1.5rem (24px)
    local:spacing.widget-padding  1.625rem (26px)

--- New in Local (no issues) ---

  component.widget (new component)
  color.widget-border (new token)
  spacing.widget-internal (new token)

--- Summary ---

  Name collisions: 1 (must resolve before merge)
  Identical values: 2 (likely duplicates)
  Similar values: 2 (review recommended)
  New tokens: 3 (no issues)
```

### Analysis Performed

| Check | Method | Purpose |
|-------|--------|---------|
| Name collision | Exact path match | Blocks merge |
| Identical values | String comparison after normalization | Obvious duplicates |
| Identical colors | Normalize hex/rgb/oklch, compare | Catches format differences |
| Similar colors | Compare oklch lightness ±5% | Surfaces potential duplicates |
| Identical dimensions | Compare normalized values | Obvious duplicates |
| Similar dimensions | Compare within ±4px | Surfaces potential duplicates |

### What Diff Does NOT Do

- Does not detect semantic duplicates (e.g., "brand-blue" vs "primary" with different values)
- Does not automatically suggest which token to keep
- Does not block any operations (informational only)
- Does not understand designer intent

These are left to LLM-enhanced workflows or human judgment.

### JSON Output Format

```json
{
  "base": { "path": "./company-tokens", "token_count": 47 },
  "local": { "path": "./project-tokens", "token_count": 12 },
  "collisions": [
    {
      "path": "color.primary",
      "base_value": "oklch(49% 0.309 275)",
      "base_file": "brand/colors.json",
      "local_value": "oklch(52% 0.3 275)",
      "local_file": "overrides.json"
    }
  ],
  "identical": [
    {
      "base_path": "spacing.md",
      "local_path": "spacing.widget-gap",
      "value": "1rem"
    }
  ],
  "similar": [
    {
      "type": "color",
      "base_path": "color.primary",
      "local_path": "color.widget-accent",
      "base_value": "oklch(49% 0.309 275)",
      "local_value": "oklch(52% 0.3 275)",
      "difference": "3% lightness"
    }
  ],
  "new": [
    { "path": "component.widget", "file": "components/widget.json" }
  ]
}
```

### Flags

| Flag | Description |
|------|-------------|
| `--output`, `-o` | Write report to file |
| `--format` | Output format: `text` (default), `json`, `markdown` |
| `--no-similar` | Skip similar value detection (faster) |
| `--threshold` | Similarity threshold for colors/dimensions |

## LLM-Enhanced Workflow

The diff output is designed as input for LLM analysis:

```bash
# Generate diff report
tokenctl diff ./base ./local --format=json > diff.json

# LLM prompt (example)
cat << 'EOF'
Here is a diff report between base and local design specs:
$(cat diff.json)

Base spec: $(cat base/tokens/**/*.json)
Local spec: $(cat local/tokens/**/*.json)

For each item in "identical" and "similar":
1. Determine if local should reference the base token
2. If yes, provide the edit to local spec
3. If no, explain why they should remain separate

Output as actionable edits.
EOF
```

The human reviews LLM suggestions, applies edits, then runs merge.

## Implementation Notes

### Merge Implementation

1. Reuse existing `loader.go` for reading specs
2. Build flattened token map from each spec with source tracking
3. Compare keys for collisions
4. On success, write using existing serialization

Estimated complexity: Low (mostly reusing existing code)

### Diff Implementation

1. Load both specs as flattened token maps
2. Name collision: set intersection on keys
3. Identical values: compare normalized string values
4. Similar colors: parse to oklch, compare L channel ±threshold
5. Similar dimensions: parse to px, compare ±threshold
6. Generate report in requested format

Estimated complexity: Medium (color/dimension parsing)

### Color Normalization

For identical detection:
- Normalize hex to lowercase 6-digit (#fff → #ffffff)
- Parse rgb() to hex
- Compare as strings

For similar detection:
- Parse to oklch (or convert via intermediate)
- Compare L (lightness) component
- Skip if format is unparseable (report as "unknown")

### Dimension Normalization

- Parse value and unit
- Convert rem → px (assume 16px base)
- Compare numeric values

## Non-Goals

These are explicitly out of scope:

1. **Automatic resolution**: Tool surfaces issues, humans decide
2. **Semantic analysis**: Can't know if "brand-blue" and "primary" should be the same
3. **Version management**: Specs are directories, not packages
4. **Remote fetching**: Base must be local (use git/npm to fetch first)
5. **Complex merge strategies**: No "theirs/ours/union" - just fail on collision

## Testing Strategy

### Merge Tests

- Merge two non-overlapping specs → success
- Merge with name collision → error with details
- Merge with `--dry-run` → report only, no output
- Merge with `--flatten` → single file output

### Diff Tests

- Diff identical specs → no issues
- Diff with name collisions → reported
- Diff with identical values → detected
- Diff with similar colors → detected (within threshold)
- Diff with similar dimensions → detected (within threshold)
- Diff output formats (text, json, markdown)

## Success Criteria

1. `tokenctl merge` blocks on name collision with clear error
2. `tokenctl diff` surfaces obvious duplicates (identical values)
3. JSON diff output works as LLM input
4. Workflow: diff → LLM review → edit → merge works smoothly
5. No false positives blocking legitimate merges
