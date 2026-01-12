<!-- tokenctl/README.md -->
# tokenctl

**tokenctl** (Token Control) is a W3C Design Tokens manager that acts as the single source of truth for your design system. Define tokens in JSON files and generate CSS artifacts for Tailwind 4 and other consumers.

## Key Features

- **W3C Compliant**: Uses the standard [W3C Design Token Format](https://tr.designtokens.org/format/)
- **Tailwind 4 Ready**: Generates modern `@theme` configurations with `@layer` support
- **Pure CSS Output**: Generate CSS without Tailwind dependency (`--format=css`)
- **Reference Resolution**: Deep referencing (`{color.brand.primary}`) with cycle detection
- **Theme Inheritance**: `$extends` for theme variations that inherit from parent themes
- **Computed Values**: `contrast()`, `darken()`, `lighten()`, `shade()`, `calc()` expressions
- **Scale Expansion**: `$scale` generates size variants automatically (xs, sm, md, lg, xl)
- **CSS @property**: `$property` field generates typed CSS custom properties for animations
- **Responsive Tokens**: `$breakpoints` and `$responsive` for media query generation
- **Layer Validation**: `--strict-layers` enforces brand → semantic → component architecture
- **Token Search**: CLI search by name, type, or category
- **LLM Manifests**: Category-scoped JSON manifests for context-efficient LLM usage
- **Rich Metadata**: `$description`, `$usage`, `$avoid` fields for documentation
- **Component Composition**: `$contains`, `$requires` for component relationships
- **Constraint Validation**: `$min`/`$max` bounds checking on dimension and number tokens
- **Type Validation**: Validates colors, dimensions, numbers, fontFamily, effect, duration
- **Source Tracking**: Validation errors include source file paths

## Installation

```bash
go install github.com/dmoose/tokenctl/cmd/tokenctl@latest
```

## Quick Start

### 1. Initialize a System

```bash
tokenctl init my-design-system
```

Creates:
```
my-design-system/
├── tokens/
│   ├── brand/colors.json
│   ├── semantic/status.json
│   ├── spacing/scale.json
│   └── themes/
```

### 2. Define Tokens

**tokens/brand/colors.json:**
```json
{
  "color": {
    "$type": "color",
    "primary": { "$value": "oklch(49.12% 0.309 275.75)" },
    "primary-content": { "$value": "contrast({color.primary})" },
    "secondary": { "$value": "#8b5cf6" }
  }
}
```

### 3. Create Theme Variations

**tokens/themes/dark.json:**
```json
{
  "$extends": "light",
  "color": {
    "primary": { "$value": "oklch(65% 0.2 275)" }
  }
}
```

### 4. Build

```bash
tokenctl build my-design-system --output=./dist
```

**Output (dist/tokens.css):**
```css
@import "tailwindcss";

@theme {
  --color-primary: oklch(49.12% 0.309 275.75);
  --color-primary-content: oklch(100% 0 0);
  --color-secondary: #8b5cf6;
}

@layer base {
  [data-theme="dark"] {
    --color-primary: oklch(65% 0.2 275);
  }
}
```

## Token Features

### References

```json
{
  "button-bg": { "$value": "{color.primary}" }
}
```

### Computed Colors

```json
{
  "primary-content": { "$value": "contrast({color.primary})" },
  "primary-hover": { "$value": "darken({color.primary}, 10%)" },
  "base-200": { "$value": "shade({color.base-100}, 1)" }
}
```

### Scale Expansion

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

Generates: `--size-field`, `--size-field-xs`, `--size-field-sm`, etc.

### CSS @property

```json
{
  "color": {
    "primary": {
      "$value": "oklch(49% 0.3 275)",
      "$property": true
    }
  }
}
```

Enables animated theme transitions.

### Constraints

```json
{
  "size": {
    "field": {
      "$value": "2.5rem",
      "$min": "1rem",
      "$max": "5rem"
    }
  }
}
```

## Commands

```bash
tokenctl init [dir]                    # Initialize token system

tokenctl build [dir]                   # Build artifacts
  --format=tailwind                  # Tailwind 4 CSS (default)
  --format=css                       # Pure CSS (no Tailwind import)
  --format=catalog                   # Full JSON catalog
  --format=manifest:CATEGORY         # Category-scoped manifest
  --output=<dir>                     # Output directory (default: dist)
  --customizable-only                # Only tokens marked $customizable: true

tokenctl validate [dir]                # Validate tokens
  --strict                           # Fail on warnings
  --strict-layers                    # Enforce layer reference rules

tokenctl search [query]                # Search tokens
  --type=<type>                      # Filter by type (color, dimension, etc.)
  --category=<cat>                   # Filter by category
  --dir=<dir>                        # Token directory (default: .)
```

## Catalog Format (v2.0)

The `--format=catalog` option generates a JSON catalog for external tool integration. The v2.0 schema includes resolved theme data:

```json
{
  "meta": {
    "version": "2.0",
    "generated_at": "2025-01-03T12:00:00Z",
    "tokenctl_version": "1.1.0"
  },
  "tokens": {
    "color.primary": "#3b82f6",
    "spacing.sm": "0.5rem"
  },
  "components": {
    "button": {
      "classes": ["btn", "btn-primary"],
      "definitions": { }
    }
  },
  "themes": {
    "light": {
      "extends": null,
      "tokens": { "color.primary": "#60a5fa" },
      "diff": { "color.primary": "#60a5fa" }
    },
    "dark": {
      "extends": "light",
      "description": "Dark theme extends light",
      "tokens": { "color.primary": "#1e40af" },
      "diff": { "color.primary": "#1e40af" }
    }
  }
}
```

| Field | Description |
|-------|-------------|
| `meta.version` | Catalog schema version (semver) |
| `meta.tokenctl_version` | tokenctl version that generated this catalog |
| `tokens` | Flattened, resolved base tokens |
| `components` | Component definitions with generated class names |
| `themes.<name>.extends` | Parent theme name (null if extends base) |
| `themes.<name>.description` | Theme description from `$description` field |
| `themes.<name>.tokens` | Fully resolved token values for this theme |
| `themes.<name>.diff` | Only tokens that differ from parent/base |

## Examples

```bash
tokenctl build examples/computed --output=dist/computed
tokenctl build examples/themes --output=dist/themes
tokenctl build examples/validation --output=dist/validation
tokenctl build examples/daisyui --output=dist/daisyui
```

See [examples/README.md](examples/README.md) for details.

## Token Types

| Type | Description | Example |
|------|-------------|---------|
| `color` | CSS colors | `#3b82f6`, `oklch(49% 0.3 275)` |
| `dimension` | Length values | `1rem`, `16px` |
| `number` | Numeric values | `400`, `0.5` |
| `fontFamily` | Font stacks | `["Inter", "sans-serif"]` |
| `duration` | Time values | `150ms`, `0.3s` |
| `effect` | Binary toggle | `0` or `1` |
| `component` | Component definition | See TOKENS.md |

## Documentation

- [README.md](README.md) - Quick start (this file)
- [HOWTO.md](HOWTO.md) - Comprehensive design system guide
- [TOKENS.md](TOKENS.md) - Token format, types, expressions, constraints
- [ADVANCED_USAGE.md](ADVANCED_USAGE.md) - CSS composition patterns
- [examples/](examples/) - Working examples
- [testdata/](testdata/) - Test fixtures

## Development

```bash
make build          # Build binary
make test           # Run tests
make coverage       # Coverage report
make demo           # Full workflow demo
make examples       # Build all examples
make help           # All targets
```
