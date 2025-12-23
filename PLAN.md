# Implementation Plan: Filling tokctl Gaps for DaisyUI Parity

## Overview

This plan addresses the priority gaps identified for making tokctl capable of generating a DaisyUI-equivalent design system. The focus is on plumbing/infrastructure, not component library content.

**Priority Order:**
1. ✅ Color utilities package (using go-colorful)
2. ✅ Computed value support in token resolution
3. Enhanced validation with range constraints
4. Effect token handling

---

## Phase 1: Color Utilities Package ✅ COMPLETE

**Goal:** Provide color parsing, contrast calculation, and content color auto-generation.

### Completed Deliverables

| File | Status | Description |
|------|--------|-------------|
| `pkg/colors/colors.go` | ✅ | Core color type and multi-format parsing |
| `pkg/colors/contrast.go` | ✅ | WCAG contrast calculations |
| `pkg/colors/content.go` | ✅ | Content color auto-generation |
| `pkg/colors/colors_test.go` | ✅ | Comprehensive tests (64.6% coverage) |
| `go.mod` | ✅ | Added `github.com/lucasb-eyer/go-colorful v1.3.0` |

### Capabilities Implemented

**Color Parsing** - `Parse()` accepts:
- Hex: `#fff`, `#ffffff`, `#ffffffff`
- RGB: `rgb(255, 128, 0)`, `rgba(255, 128, 0, 0.5)`
- HSL: `hsl(180, 50%, 50%)`, `hsla(...)`
- OKLCH: `oklch(50% 0.2 180)` - DaisyUI's preferred format
- Named colors: `red`, `blue`, `black`, etc.

**Color Output:**
- `ToCSS(format)` - Output in any format
- `ToOKLCH()` - DaisyUI-compatible OKLCH format
- `Hex()`, `ToRGB()`, `ToHSL()`
- `ToOriginalFormat()` - Preserve input format

**WCAG Contrast:**
- `ContrastRatio(c1, c2)` - Calculate contrast (1.0 to 21.0)
- `MeetsWCAG(c1, c2, level, largeText)` - Check AA/AAA compliance
- `ContrastLevel(c1, c2)` - Get human-readable level
- `RelativeLuminance(c)` - WCAG luminance formula

**Content Color Generation:**
- `ContentColor(bg)` - Generate WCAG AA compliant content color
- `ContentColorWithRatio(bg, ratio)` - Target specific contrast ratio
- `DaisyContentColor(bg)` - DaisyUI-style content colors
- `OptimalTextColor(bg)` - Simple black/white selection

---

## Phase 2: Computed Values Support ✅ COMPLETE

**Goal:** Enable expressions in token values for scales, calculations, and derived values.

### Completed Deliverables

| File | Status | Description |
|------|--------|-------------|
| `pkg/tokens/dimension.go` | ✅ | CSS dimension parsing and arithmetic |
| `pkg/tokens/dimension_test.go` | ✅ | Dimension tests |
| `pkg/tokens/expressions.go` | ✅ | Expression evaluation engine |
| `pkg/tokens/expressions_test.go` | ✅ | Expression tests |
| `pkg/tokens/scale.go` | ✅ | Scale expansion for `$scale` shorthand |
| `pkg/tokens/scale_test.go` | ✅ | Scale tests |
| `pkg/tokens/resolver.go` | ✅ | Integrated expression evaluation |
| `pkg/tokens/loader.go` | ✅ | Integrated scale expansion during loading |
| `examples/computed/` | ✅ | Example demonstrating all new features |

### Capabilities Implemented

**Dimension Parsing & Arithmetic:**
- Parse CSS dimensions: `10px`, `2.5rem`, `100%`, etc.
- Arithmetic operations: Add, Subtract, Multiply, Divide
- Proper floating-point rounding to avoid artifacts

**Expression Evaluation in `$value`:**

| Expression | Example | Description |
|------------|---------|-------------|
| `calc()` | `calc({size.base} * 0.8)` | Arithmetic with token references |
| `contrast()` | `contrast({color.primary})` | Generate WCAG-compliant content color |
| `darken()` | `darken({color.primary}, 20%)` | Darken a color in OKLCH space |
| `lighten()` | `lighten({color.neutral}, 30%)` | Lighten a color in OKLCH space |
| `scale()` | `scale({size.base}, 1.5)` | Multiply dimension by factor |

**Scale Expansion (`$scale`):**

```json
{
  "size": {
    "field": {
      "$value": "2.5rem",
      "$scale": { "xs": 0.6, "sm": 0.8, "md": 1.0, "lg": 1.2, "xl": 1.4 }
    }
  }
}
```

Automatically expands to:
- `size.field` → `2.5rem`
- `size.field-xs` → `1.5rem`
- `size.field-sm` → `2rem`
- `size.field-md` → `2.5rem`
- `size.field-lg` → `3rem`
- `size.field-xl` → `3.5rem`

### Test Coverage

- `pkg/tokens`: 82.8% coverage
- `pkg/colors`: 64.6% coverage

---

## Phase 3: Enhanced Validation (TODO)

**Goal:** Add range constraints and type-specific validation.

### 3.1 Constraint Syntax

```json
{
  "size": {
    "field": {
      "$value": "2.5rem",
      "$type": "dimension",
      "$min": "1rem",
      "$max": "5rem"
    }
  },
  "border": {
    "$value": "1px",
    "$type": "dimension",
    "$min": "0px",
    "$max": "10px"
  }
}
```

### 3.2 Validation Rules

Add to `pkg/tokens/validator.go`:

```go
// validateConstraints checks $min/$max on dimension tokens
func (v *Validator) validateConstraints(path string, token map[string]interface{}) []ValidationError

// validateColorFormat ensures color values are valid CSS colors
func (v *Validator) validateColorFormat(path string, value interface{}) []ValidationError

// validateDimension ensures dimension values have valid units
func (v *Validator) validateDimension(path string, value interface{}) []ValidationError
```

### 3.3 Type-Specific Validators

| Type | Validations |
|------|-------------|
| `color` | Valid CSS color format, optional OKLCH check |
| `dimension` | Valid CSS dimension (number + unit), range constraints |
| `number` | Numeric value, range constraints |
| `fontFamily` | Non-empty string or array |

### 3.4 Files to Create/Modify

| File | Action | Description |
|------|--------|-------------|
| `pkg/tokens/validator.go` | Modify | Add constraint validation |
| `pkg/tokens/constraints.go` | Create | Constraint parsing and checking |

**Note:** `pkg/tokens/dimension.go` already exists from Phase 2.

---

## Phase 4: Effect Token Support (TODO)

**Goal:** Handle DaisyUI's `--depth` and `--noise` effect tokens properly.

### 4.1 Effect Token Type

Effects are boolean-like tokens (0 or 1) that enable CSS effects:

```json
{
  "effect": {
    "depth": {
      "$value": 1,
      "$type": "effect",
      "$description": "Enable depth shadows on components"
    },
    "noise": {
      "$value": 0,
      "$type": "effect",
      "$description": "Enable noise texture overlay"
    }
  }
}
```

### 4.2 Generator Support

In `pkg/generators/tailwind.go`, handle effect tokens:

```go
// Effect tokens generate:
// 1. The CSS variable (--depth: 1)
// 2. Conditional CSS based on value (if depth=1, add shadow styles)

// Recommendation: Just output the variable, let CSS handle it
// Component definitions handle the actual effect application
```

### 4.3 Files to Modify

| File | Action | Description |
|------|--------|-------------|
| `pkg/generators/tailwind.go` | Modify | Handle effect type output |
| `pkg/tokens/validator.go` | Modify | Validate effect values (0 or 1) |

---

## Implementation Status

```
Phase 1: Color Utilities        ✅ COMPLETE
Phase 2: Computed Values        ✅ COMPLETE
Phase 3: Enhanced Validation    ⬚ TODO (~1-2 days)
Phase 4: Effect Tokens          ⬚ TODO (~0.5 days)

Total Remaining: ~1.5-2.5 days
```

---

## Example: Working DaisyUI-Style Token File

This example works TODAY with the completed phases:

```json
{
  "color": {
    "$type": "color",
    "primary": {
      "$value": "oklch(49.12% 0.309 275.75)",
      "$description": "Primary brand color"
    },
    "primary-content": {
      "$value": "contrast({color.primary})",
      "$description": "Auto-generated content color"
    },
    "neutral": {
      "$value": "#1f2937"
    },
    "neutral-light": {
      "$value": "lighten({color.neutral}, 30%)"
    },
    "neutral-dark": {
      "$value": "darken({color.neutral}, 20%)"
    }
  },
  "size": {
    "field": {
      "$value": "2.5rem",
      "$scale": { "xs": 0.6, "sm": 0.8, "md": 1.0, "lg": 1.2, "xl": 1.4 }
    },
    "selector": {
      "$value": "1.5rem",
      "$scale": { "xs": 0.6, "sm": 0.8, "md": 1.0, "lg": 1.2, "xl": 1.4 }
    }
  },
  "spacing": {
    "base": { "$value": "1rem" },
    "xs": { "$value": "calc({spacing.base} * 0.25)" },
    "sm": { "$value": "calc({spacing.base} * 0.5)" },
    "lg": { "$value": "calc({spacing.base} * 1.5)" },
    "xl": { "$value": "calc({spacing.base} * 2)" }
  },
  "radius": {
    "box": { "$value": "1rem" },
    "field": { "$value": "0.5rem" },
    "selector": { "$value": "9999px" }
  },
  "border": {
    "default": { "$value": "1px" },
    "thick": { "$value": "calc({border.default} * 2)" }
  }
}
```

Build with: `tokctl build examples/computed --output dist/computed`

---

## Deliverables Checklist

**Phase 1:** ✅
- [x] `pkg/colors` package with go-colorful integration
- [x] OKLCH CSS parsing and output
- [x] WCAG contrast calculation
- [x] Content color auto-generation
- [x] Unit tests with >60% coverage

**Phase 2:** ✅
- [x] Expression evaluator
- [x] `calc()` parser with dimension handling
- [x] `contrast()` function integration
- [x] `darken()` and `lighten()` functions
- [x] `scale()` function
- [x] `$scale` expansion during loading
- [x] Integration with resolver
- [x] Unit tests with >80% coverage

**Phase 3:** ⬚
- [ ] `$min`/`$max` constraint validation
- [ ] Color format validation
- [ ] Dimension validation
- [ ] Unit tests

**Phase 4:** ⬚
- [ ] Effect token type handling
- [ ] Generator output for effects
- [ ] Unit tests