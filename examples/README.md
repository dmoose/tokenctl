<!-- tokenctl/examples/README.md -->
# Tokenctl Examples

This directory contains working examples demonstrating various tokenctl features. These examples are referenced in the main documentation and can be used as starting points for your own design systems.

## Directory Structure

```
examples/
├── baseline/       # Complete design system with components, theming, responsive tokens
├── basic/          # Simple token system (brand, spacing, semantic colors)
├── components/     # Component definitions with variants and sizes
├── computed/       # Computed values (calc, contrast, darken, lighten, scale)
├── daisyui/        # DaisyUI 5 theme generator example
├── themes/         # Theme inheritance with $extends
└── validation/     # Constraint validation and effect tokens
```

## Running Examples

### Baseline Example

The `baseline/` example is a complete design system demonstrating all tokenctl features:

```bash
# Build the CSS
tokenctl build examples/baseline --format=css --output=examples/baseline/dist

# Open the demo
open examples/baseline/demo.html
```

**Key features:**
- Three-layer architecture (brand → semantic → component)
- Full component library (buttons, cards, forms, alerts, tables)
- Light and dark themes
- Responsive typography
- Built-in CSS reset

See [examples/baseline/README.md](baseline/README.md) for full documentation.

### Basic Example

The `basic/` example shows a minimal token system created with `tokenctl init`:

```bash
# Validate the tokens
tokenctl validate examples/basic

# Build CSS output
tokenctl build examples/basic --output=dist/basic
```

**Generated output includes:**
- CSS custom properties for brand colors
- Semantic status colors (success, error, warning)
- Spacing scale values

### Themes Example

The `themes/` example demonstrates the `$extends` feature for theme inheritance:

```bash
# Validate themes (tests inheritance)
tokenctl validate examples/themes

# Build with themes
tokenctl build examples/themes --output=dist/themes
```

**Key features:**
- `light` theme defines base colors
- `dark` theme extends `light` and overrides specific tokens
- CSS output uses `[data-theme="..."]` selectors

**File structure:**
```
themes/
├── tokens/
│   ├── brand.json          # Base brand colors
│   └── themes/
│       ├── light.json      # Light theme overrides
│       └── dark.json       # Extends light theme
```

### Components Example

The `components/` example shows how to define reusable component tokens:

```bash
# Validate component definitions
tokenctl validate examples/components

# Build component CSS
tokenctl build examples/components --output=dist/components
```

**Generated output includes:**
- Component base styles (`.btn`)
- Variant classes (`.btn-primary`, `.btn-secondary`, etc.)
- Size modifiers (`.btn-sm`, `.btn-lg`)
- Pseudo-state styles (`:hover`, `:active`)
- CSS custom properties for colors and spacing

**Component structure:**
```json
{
  "components": {
    "button": {
      "$type": "component",
      "$class": "btn",
      "variants": {
        "primary": { ... },
        "secondary": { ... }
      },
      "sizes": {
        "small": { ... },
        "large": { ... }
      }
    }
  }
}
```

### Computed Example

The `computed/` example demonstrates expression evaluation and computed values:

```bash
# Build computed values
tokenctl build examples/computed --output=dist/computed
```

**Key features:**
- `calc()` expressions for arithmetic
- `contrast()` for auto-generating content colors
- `darken()` and `lighten()` for color manipulation
- `scale()` for dimension scaling
- `$scale` shorthand for size variants

### Validation Example

The `validation/` example demonstrates enhanced validation features:

```bash
# Validate with constraint checking
tokenctl validate examples/validation

# Build tokens
tokenctl build examples/validation --output=dist/validation
```

**Key features:**
- `$min` and `$max` constraints on dimension tokens
- Numeric range constraints
- Effect tokens (0 or 1) for DaisyUI-style toggles
- Type-specific validation (color, dimension, number, fontFamily, effect)
- `$type` inheritance from parent groups

**Example constraint:**
```json
{
  "size": {
    "$type": "dimension",
    "field": {
      "$value": "2.5rem",
      "$min": "1rem",
      "$max": "5rem"
    }
  }
}
```

### DaisyUI Example

The `daisyui/` example shows how to use tokenctl as a theme generator for DaisyUI 5:

```bash
# Build DaisyUI-compatible tokens
tokenctl build examples/daisyui --output=dist/daisyui
```

**Key features:**
- All 26 DaisyUI token variables
- Auto-generated content colors via `contrast()`
- `$scale` for size variants
- `$property` for animated theme transitions

See [examples/daisyui/README.md](daisyui/README.md) for integration instructions.

## Using Examples as Templates

You can copy any example as a starting point:

```bash
# Copy basic example to start your project
cp -r examples/basic my-design-system

# Customize tokens
vim my-design-system/tokens/brand/colors.json

# Build
tokenctl build my-design-system --output=dist
```

## Generating Catalog

Examples can also be built as JSON catalogs for integration with other tools:

```bash
tokenctl build examples/components --format=catalog --output=dist
```

This creates `catalog.json` with:
- All resolved token values
- Component class names
- Metadata for tooling integration

## See Also

- [Main Documentation](../README.md)
- [Developer Guide](../TOKENS.md)
- [Test Fixtures](../testdata/) - Used by automated tests
