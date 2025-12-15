# tokctl

**tokctl** (Token Control) is a W3C Design Tokens manager that acts as the single source of truth for your design system. It allows you to define atomic and semantic tokens in standard JSON files and generate consumption artifacts (CSS, Go, etc.) for your applications.

## Key Features

- **W3C Compliant**: Uses the standard [W3C Design Token Format](https://tr.designtokens.org/format/).
- **Semantic First**: Encourages a layered design system (Brand -> Semantic -> Component).
- **Tailwind 4 Ready**: Generates modern Tailwind CSS `@theme` configurations.
- **Reference Resolution**: Supports deep referencing (`{brand.primary}`) and cycle detection.

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
│   └── ...
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

### 3. Define Components

Edit `tokens/components/button.json`:
```json
{
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
```

### 4. Build Artifacts

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

A machine-readable dump of the resolved system for tools like `templscan`.

## Commands

- `tokctl init [dir]`: Scaffold a new token system.
- `tokctl build [dir]`: Build artifacts (CSS, Catalog).
- `tokctl validate [dir]`: Check for broken references and schema compliance. Use `--strict` to fail on warnings.

## Architecture

`tokctl` is the **Source of Truth** for your design system. It manages the W3C token graph, resolves references, and handles theme inheritance.

**Core Components**:
- **Loader**: Recursive JSON loading with separate Theme/Base handling.
- **Resolver**: Deep reference resolution (`{brand.primary}`) with cycle detection.
- **Validator**: Logic checks for broken refs and schema issues.
- **Generators**: Output engines for Tailwind 4, Catalog, etc.
