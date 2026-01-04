<!-- tokctl/README.md -->
# tokctl

**tokctl** (Token Control) is a W3C Design Tokens manager that acts as the single source of truth for your design system. Define tokens in JSON files and generate CSS artifacts for Tailwind 4 and other consumers.

## Key Features

- **W3C Compliant**: Uses the standard [W3C Design Token Format](https://tr.designtokens.org/format/)
- **Tailwind 4 Ready**: Generates modern `@theme` configurations with `@layer` support
- **Reference Resolution**: Deep referencing (`{color.brand.primary}`) with cycle detection
- **Theme Inheritance**: `$extends` for theme variations that inherit from parent themes
- **Computed Values**: `contrast()`, `darken()`, `lighten()`, `shade()`, `calc()` expressions
- **Scale Expansion**: `$scale` generates size variants automatically (xs, sm, md, lg, xl)
- **CSS @property**: `$property` field generates typed CSS custom properties for animations
- **Constraint Validation**: `$min`/`$max` bounds checking on dimension and number tokens
- **Type Validation**: Validates colors, dimensions, numbers, fontFamily, effect, duration
- **Source Tracking**: Validation errors include source file paths

## Installation

```bash
go install github.com/dmoose/tokctl/cmd/tokctl@latest
```

## Quick Start

### 1. Initialize a System

```bash
tokctl init my-design-system
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
tokctl build my-design-system --output=./dist
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
tokctl init [dir]              # Initialize token system
tokctl validate [dir]          # Validate tokens
  --strict                     # Fail on warnings
tokctl build [dir]             # Build artifacts
  --format=tailwind            # CSS output (default)
  --format=catalog             # JSON catalog
  --output=<dir>               # Output directory (default: dist)
```

## Examples

```bash
tokctl build examples/computed --output=dist/computed
tokctl build examples/themes --output=dist/themes
tokctl build examples/validation --output=dist/validation
tokctl build examples/daisyui --output=dist/daisyui
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
