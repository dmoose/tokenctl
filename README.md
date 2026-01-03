# tokctl

**tokctl** (Token Control) is a W3C Design Tokens manager that acts as the single source of truth for your design system. It allows you to define atomic and semantic tokens in standard JSON files and generate consumption artifacts (CSS, Go, etc.) for your applications.

## Key Features

- **W3C Compliant**: Uses the standard [W3C Design Token Format](https://tr.designtokens.org/format/).
- **Semantic First**: Encourages a layered design system (Brand -> Semantic -> Component).
- **Tailwind 4 Ready**: Generates modern Tailwind CSS `@theme` configurations.
- **Reference Resolution**: Supports deep referencing (`{brand.primary}`) and cycle detection.
- **Theme Inheritance**: Use `$extends` to create theme variations that inherit from parent themes.
- **Detailed Error Messages**: Validation errors show source file names for easy debugging.

## Installation

```bash
go install github.com/dmoose/tokctl/cmd/tokctl@latest
```

## Quick Start

### 1. Initialize a System

```bash
tokctl init my-design-system
```

This creates a standard directory structure:
```
my-design-system/
├── tokens/
│   ├── brand/
│   ├── semantic/
│   ├── spacing/
│   └── themes/  # Optional: theme variations
```

### 2. Define Tokens

Edit `tokens/brand/colors.json`:
```json
{
  "color": {
    "brand": {
      "primary": { "$value": "#3b82f6" },
      "secondary": { "$value": "#8b5cf6" }
    }
  }
}
```

### 3. Create Theme Variations (Optional)

Edit `tokens/themes/dark.json`:
```json
{
  "$extends": "light",
  "$description": "Dark theme extends light theme",
  "color": {
    "brand": {
      "primary": { "$value": "#1e40af" }
    }
  }
}
```

The `$extends` keyword allows themes to inherit from other themes, creating variations efficiently.

### 4. Define Components

Edit `tokens/components/button.json`:
```json
{
  "components": {
    "button": {
      "$type": "component",
      "$class": "btn",
      "variants": {
        "primary": {
          "$class": "btn-primary",
          "background-color": "{color.brand.primary}",
          "color": "#ffffff"
        }
      }
    }
  }
}
```

### 5. Build Artifacts

Generate CSS variables for Tailwind 4:

```bash
tokctl build my-design-system --format=tailwind --output=./dist
```

**Output (`dist/tokens.css`):**
```css
@import "tailwindcss";

@theme {
  --color-brand-primary: #3b82f6;
  --color-brand-secondary: #8b5cf6;
}
```

## Output Formats

### Tailwind CSS 4

```css
@import "tailwindcss";

@theme {
  --color-primary: oklch(0.59 0.21 258.34);
  --color-primary-content: oklch(1 0 0);
  --spacing-sm: 0.5rem;
}

@layer base {
  [data-theme="dark"] {
    --color-primary: oklch(0.69 0.21 258.34);
  }
}

@layer components {
  .btn-primary {
    background-color: var(--color-primary);
    color: var(--color-primary-content);
  }
}
```

### JSON Catalog

A machine-readable dump of the resolved system for tool integration:

```bash
tokctl build my-design-system --format=catalog --output=./dist
```

## Commands

- `tokctl init [dir]`: Scaffold a new token system with standard directory structure.
- `tokctl build [dir]`: Build artifacts (CSS, Catalog).
  - `--format=tailwind`: Generate Tailwind CSS (default)
  - `--format=catalog`: Generate JSON catalog
  - `--output=<dir>`: Output directory (default: `dist`)
- `tokctl validate [dir]`: Check for broken references, circular dependencies, and schema compliance.
  - Shows source file names in error messages for easy debugging
- `make help`: View all available make targets for development

## Examples

Working examples demonstrating tokctl features are in the `examples/` directory:

- **basic/**: Simple token system with brand colors, spacing, and semantic tokens
- **themes/**: Theme inheritance using `$extends` (light/dark theme variations)
- **components/**: Component definitions with variants and states

Build any example:
```bash
tokctl build examples/themes --output=dist/themes
```

See [examples/README.md](examples/README.md) for detailed documentation.

## Architecture

`tokctl` is the **Source of Truth** for your design system. It manages the W3C token graph, resolves references, and handles theme inheritance.

**Core Components**:
- **Loader**: Recursive JSON loading with source file tracking for error reporting.
- **Resolver**: Deep reference resolution (`{brand.primary}`) with cycle detection.
- **Validator**: Validates references, schema compliance, and reports errors with file context.
- **Generators**: Unified generation API producing Tailwind CSS, JSON catalogs, etc.
- **Theme Inheritance**: Resolves `$extends` chains with circular dependency protection.

## Development

Build and test:
```bash
make build          # Build binary
make test           # Run all tests
make coverage       # Generate coverage report
make demo           # Run full workflow demo
make examples       # Build all examples
```

See `make help` for all available targets.

## Documentation

- [README.md](README.md) - Quick start and overview (this file)
- [TOKENS.md](TOKENS.md) - Comprehensive developer guide
- [examples/](examples/) - Working examples with documentation
- [testdata/](testdata/) - Test fixtures and expected outputs
