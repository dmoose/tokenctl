# Tokctl Examples

This directory contains working examples demonstrating various tokctl features. These examples are referenced in the main documentation and can be used as starting points for your own design systems.

## Directory Structure

```
examples/
├── basic/          # Simple token system (brand, spacing, semantic colors)
├── themes/         # Theme inheritance with $extends
├── components/     # Component definitions with variants and sizes
├── computed/       # Computed values (calc, contrast, darken, lighten, scale)
└── validation/     # Constraint validation and effect tokens
```

## Running Examples

### Basic Example

The `basic/` example shows a minimal token system created with `tokctl init`:

```bash
# Validate the tokens
tokctl validate examples/basic

# Build CSS output
tokctl build examples/basic --output=dist/basic
```

**Generated output includes:**
- CSS custom properties for brand colors
- Semantic status colors (success, error, warning)
- Spacing scale values

### Themes Example

The `themes/` example demonstrates the `$extends` feature for theme inheritance:

```bash
# Validate themes (tests inheritance)
tokctl validate examples/themes

# Build with themes
tokctl build examples/themes --output=dist/themes
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
tokctl validate examples/components

# Build component CSS
tokctl build examples/components --output=dist/components
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
tokctl build examples/computed --output=dist/computed
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
tokctl validate examples/validation

# Build tokens
tokctl build examples/validation --output=dist/validation
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

## Using Examples as Templates

You can copy any example as a starting point:

```bash
# Copy basic example to start your project
cp -r examples/basic my-design-system

# Customize tokens
vim my-design-system/tokens/brand/colors.json

# Build
tokctl build my-design-system --output=dist
```

## Generating Catalog

Examples can also be built as JSON catalogs for integration with other tools:

```bash
tokctl build examples/components --format=catalog --output=dist
```

This creates `catalog.json` with:
- All resolved token values
- Component class names
- Metadata for tooling integration

## See Also

- [Main Documentation](../README.md)
- [Developer Guide](../TOKENS.md)
- [Test Fixtures](../testdata/) - Used by automated tests