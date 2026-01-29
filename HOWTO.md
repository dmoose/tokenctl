# tokenctl: Design System Guide

This guide explains the philosophy, architecture, and best practices for building maintainable design systems with tokenctl. For the token format reference (types, fields, expressions, components), see [TOKENS.md](TOKENS.md).

## Table of Contents

1. [Philosophy](#philosophy)
2. [Architecture](#architecture)
3. [Getting Started](#getting-started)
4. [Token Organization](#token-organization)
5. [Responsive Design](#responsive-design)
6. [Component Patterns](#component-patterns)
7. [Theming](#theming)
8. [LLM Integration](#llm-integration)
9. [Validation](#validation)
10. [Extending Design Systems](#extending-design-systems)
11. [Migration Guide](#migration-guide)

---

## Philosophy

### The Problem

Large web projects often struggle with CSS maintainability:
- Specificity conflicts requiring cascading overrides
- Inconsistent spacing, colors, and typography
- No single source of truth for design decisions
- Generated code (including LLM output) that drifts from the system

### The Solution

tokenctl uses a **tokens-first** approach:

1. **All styling decisions become tokens** - Colors, spacing, typography, effects
2. **Tokens reference other tokens** - Building a semantic hierarchy
3. **Components consume tokens** - Never arbitrary values
4. **Validation enforces rules** - Catch violations at build time

This creates a vocabulary that both humans and LLMs can use consistently.

### Core Principles

| Principle | What it Means |
|-----------|---------------|
| **Single Source of Truth** | All design values live in token files |
| **Semantic Layering** | Raw values → semantic names → component usage |
| **Reference-Only Components** | Components use `var(--token)`, never raw values |
| **Validated Architecture** | Layer rules enforced via `--strict-layers` |
| **Context-Efficient Manifests** | LLMs get exactly the tokens they need |

---

## Architecture

### Three-Layer Design

tokenctl supports a three-layer architecture that enforces clean separation:

```
┌─────────────────────────────────────────────────────┐
│                   COMPONENT LAYER                   │
│  btn-bg, card-padding, input-border-radius          │
│  Can only reference: semantic tokens                │
└───────────────────────┬─────────────────────────────┘
                        │ references
┌───────────────────────▼─────────────────────────────┐
│                   SEMANTIC LAYER                    │
│  primary, success, error, spacing-md                │
│  Can reference: brand tokens                        │
└───────────────────────┬─────────────────────────────┘
                        │ references
┌───────────────────────▼─────────────────────────────┐
│                    BRAND LAYER                      │
│  blue-500, gray-100, 1rem, 400                      │
│  Raw values only (no references)                    │
└─────────────────────────────────────────────────────┘
```

### Layer Definitions

```json
{
  "brand": {
    "$layer": "brand",
    "$type": "color",
    "blue-500": { "$value": "#3b82f6" },
    "blue-600": { "$value": "#2563eb" },
    "purple-500": { "$value": "#8b5cf6" }
  },
  "semantic": {
    "$layer": "semantic",
    "$type": "color",
    "primary": { "$value": "{brand.blue-500}" },
    "primary-hover": { "$value": "{brand.blue-600}" },
    "accent": { "$value": "{brand.purple-500}" }
  },
  "component": {
    "$layer": "component",
    "$type": "color",
    "btn-bg": { "$value": "{semantic.primary}" },
    "btn-bg-hover": { "$value": "{semantic.primary-hover}" }
  }
}
```

### Why Layers Matter

**Without layers:**
```css
/* Component directly references raw value - tight coupling */
.btn { background: #3b82f6; }
```

**With layers:**
```css
/* Component references semantic token - loose coupling */
.btn { background: var(--component-btn-bg); }
```

When you change your brand color from blue to purple:
- **Without layers:** Find and replace across entire codebase
- **With layers:** Change `semantic.primary` reference, everything updates

---

## Getting Started

### 1. Initialize

```bash
tokenctl init my-system
```

### 2. Define Brand Tokens

**tokens/brand/colors.json:**
```json
{
  "$layer": "brand",
  "color": {
    "$type": "color",
    "blue-500": { "$value": "#3b82f6" },
    "blue-600": { "$value": "#2563eb" },
    "gray-50": { "$value": "#f9fafb" },
    "gray-900": { "$value": "#111827" }
  }
}
```

### 3. Create Semantic Layer

**tokens/semantic/colors.json:**
```json
{
  "$layer": "semantic",
  "color": {
    "$type": "color",
    "primary": {
      "$value": "{color.blue-500}",
      "$description": "Primary brand color",
      "$usage": ["buttons", "links", "focus rings"]
    },
    "primary-hover": { "$value": "{color.blue-600}" },
    "surface": { "$value": "{color.gray-50}" },
    "text": { "$value": "{color.gray-900}" }
  }
}
```

### 4. Build

```bash
# Tailwind 4 output
tokenctl build my-system --format=tailwind

# Pure CSS output (no Tailwind dependency)
tokenctl build my-system --format=css

# Validate with layer rules
tokenctl validate my-system --strict-layers
```

---

## Token Organization

### Recommended Directory Structure

```
tokens/
├── brand/
│   ├── colors.json       # Raw color palette
│   ├── spacing.json      # Base spacing scale
│   └── typography.json   # Font families, weights
├── semantic/
│   ├── colors.json       # primary, success, error, etc.
│   ├── spacing.json      # spacing-sm, spacing-md, etc.
│   └── typography.json   # font-heading, font-body
├── components/
│   ├── button.json       # .btn component tokens
│   ├── card.json         # .card component tokens
│   └── input.json        # Form input tokens
└── themes/
    ├── light.json        # Light theme overrides
    └── dark.json         # Dark theme (extends light)
```

### Rich Metadata

Every token can carry metadata for documentation and LLM comprehension: `$description`, `$usage` (array of intended uses), `$avoid` (anti-patterns), `$deprecated` (migration guidance), and `$customizable` (safe to override). See [Token Structure](TOKENS.md#token-structure) for the full field reference.

```json
{
  "color": {
    "primary": {
      "$value": "#3b82f6",
      "$type": "color",
      "$description": "Primary brand color for key actions",
      "$usage": ["Primary button backgrounds", "Link text color"],
      "$avoid": "Don't use for large background areas",
      "$customizable": true
    }
  }
}
```

---

## Responsive Design

### Strategy: Fluid-First, Overrides When Needed

Modern responsive design combines two approaches:

1. **Fluid values** - Use `clamp()` for smooth scaling
2. **Breakpoint overrides** - Discrete changes at specific widths

### Fluid Tokens

For continuous scaling, use CSS `clamp()`:

```json
{
  "spacing": {
    "$type": "dimension",
    "section": {
      "$value": "clamp(2rem, 5vw, 6rem)",
      "$description": "Fluid section padding"
    }
  },
  "font": {
    "size": {
      "heading": {
        "$value": "clamp(1.5rem, 4vw, 3rem)",
        "$description": "Fluid heading size"
      }
    }
  }
}
```

These scale smoothly without media queries.

### Responsive Overrides

For discrete breakpoint changes, use `$responsive`:

```json
{
  "$breakpoints": {
    "sm": "640px",
    "md": "768px",
    "lg": "1024px",
    "xl": "1280px"
  },
  "spacing": {
    "$type": "dimension",
    "md": {
      "$value": "1rem",
      "$responsive": {
        "md": "1.25rem",
        "lg": "1.5rem"
      }
    }
  },
  "font": {
    "size": {
      "body": {
        "$value": "1rem",
        "$responsive": {
          "md": "1.125rem",
          "lg": "1.25rem"
        }
      }
    }
  }
}
```

**Generated CSS:**
```css
:root {
  --spacing-md: 1rem;
  --font-size-body: 1rem;
}

@media (min-width: 768px) {
  :root {
    --spacing-md: 1.25rem;
    --font-size-body: 1.125rem;
  }
}

@media (min-width: 1024px) {
  :root {
    --spacing-md: 1.5rem;
    --font-size-body: 1.25rem;
  }
}
```

### When to Use Each

| Approach | Best For | Example |
|----------|----------|---------|
| **Fluid (clamp)** | Continuous scaling | Section padding, heading sizes |
| **Breakpoint overrides** | Discrete changes | Grid columns, layout shifts |
| **Both** | Complex responsive needs | Combine fluid base with overrides |

---

## Component Patterns

Components use `$type: "component"` and generate CSS classes with `base`, `variants`, `sizes`, and `states`. See [Components](TOKENS.md#components) for the full schema including states and container queries.

### Composition Metadata

Document component relationships for LLMs:

```json
{
  "card": {
    "$type": "component",
    "$class": "card",
    "$description": "Container for card content",
    "$contains": ["card-body", "card-title", "card-actions", "card-image"]
  },
  "card-body": {
    "$type": "component",
    "$class": "card-body",
    "$description": "Main content area inside a card",
    "$requires": "card"
  },
  "card-title": {
    "$type": "component",
    "$class": "card-title",
    "$description": "Title text inside a card",
    "$requires": "card"
  }
}
```

**Manifest output:**
```json
{
  "components.card": {
    "description": "Container for card content",
    "contains": ["card-body", "card-title", "card-actions", "card-image"],
    "classes": ["card"]
  },
  "components.card-body": {
    "description": "Main content area inside a card",
    "requires": "card",
    "classes": ["card-body"]
  }
}
```

---

## Theming

### Theme Inheritance

Themes can extend other themes:

```json
// themes/light.json
{
  "$description": "Default light theme",
  "color": {
    "surface": { "$value": "#ffffff" },
    "text": { "$value": "#1f2937" }
  }
}
```

```json
// themes/dark.json
{
  "$extends": "light",
  "$description": "Dark theme",
  "color": {
    "surface": { "$value": "#1f2937" },
    "text": { "$value": "#f9fafb" }
  }
}
```

### Theme Switching

Generated CSS uses `data-theme` attributes:

```css
:root, [data-theme="light"] {
  --color-surface: #ffffff;
  --color-text: #1f2937;
}

[data-theme="dark"] {
  --color-surface: #1f2937;
  --color-text: #f9fafb;
}
```

**HTML:**
```html
<html data-theme="dark">
```

**JavaScript:**
```javascript
document.documentElement.setAttribute('data-theme', 'dark');
```

### Animated Transitions

Add `$property: true` to color tokens to generate CSS `@property` declarations, enabling smooth animated transitions between themes instead of instant snaps. See [CSS @property Declarations](TOKENS.md#css-property-declarations) for the full specification.

---

## LLM Integration

### Context-Efficient Manifests

Generate category-specific manifests to minimize LLM context usage:

```bash
# All tokens (may be large)
tokenctl build --format=catalog

# Just colors for a color-related task
tokenctl build --format=manifest:color

# Just components for UI work
tokenctl build --format=manifest:components

# Just spacing
tokenctl build --format=manifest:spacing
```

### Token Search

LLMs can search tokens without loading entire files:

```bash
# Find tokens by name
tokenctl search "primary"

# Filter by type
tokenctl search --type=color

# Filter by category
tokenctl search --category=spacing
```

**Example output:**
```
color.primary: #3b82f6
  Primary brand color for key actions
  Usage: Primary button backgrounds, Link text color

color.primary-hover: #2563eb
  Darker primary for hover states
```

### Manifest Schema

Manifests include rich metadata for LLM comprehension:

```json
{
  "meta": {
    "version": "2.1",
    "category": "color",
    "tokenctl_version": "1.2.0"
  },
  "tokens": {
    "color.primary": {
      "value": "#3b82f6",
      "type": "color",
      "description": "Primary brand color",
      "usage": ["buttons", "links", "focus rings"],
      "avoid": "Don't use for large backgrounds"
    }
  }
}
```

### Component Relationships

Component manifests include composition metadata:

```json
{
  "components": {
    "card": {
      "description": "Container for card content",
      "contains": ["card-body", "card-title", "card-actions"],
      "classes": ["card"]
    }
  }
}
```

This tells LLMs which components can be nested together.

---

## Validation

### Basic Validation

```bash
tokenctl validate ./my-tokens
```

Checks:
- Token syntax
- Reference resolution (no broken references)
- Type validation (colors are valid colors, etc.)
- Cycle detection (no circular references)

### Strict Layer Validation

```bash
tokenctl validate ./my-tokens --strict-layers
```

Enforces the layer hierarchy:
- **Brand layer**: Can only contain raw values
- **Semantic layer**: Can reference brand tokens
- **Component layer**: Can only reference semantic tokens

**Violation example:**
```
[Error] component.btn-bg [component] cannot reference brand.blue-500 [brand]: layer violation
```

Fix by routing through semantic layer:
```json
{
  "semantic": {
    "primary": { "$value": "{brand.blue-500}" }
  },
  "component": {
    "btn-bg": { "$value": "{semantic.primary}" }
  }
}
```

### Constraint Validation

Dimension and number tokens support `$min`/`$max` bounds checking. See [Constraints](TOKENS.md#constraints) for details.

---

## Migration Guide

### From Raw CSS

1. **Inventory existing values** - List all colors, spacing values, font sizes
2. **Create brand tokens** - Raw values only
3. **Create semantic layer** - Map brand to purpose
4. **Update components** - Replace values with `var(--token)`
5. **Validate** - Run `tokenctl validate --strict-layers`

**Before:**
```css
.btn {
  background: #3b82f6;
  padding: 0.5rem 1rem;
}
```

**After:**
```css
.btn {
  background: var(--component-btn-bg);
  padding: var(--spacing-sm) var(--spacing-md);
}
```

### From Tailwind 3

Tailwind 3 uses `tailwind.config.js`. Migrate to token files:

**tailwind.config.js (before):**
```js
module.exports = {
  theme: {
    colors: {
      primary: '#3b82f6',
    }
  }
}
```

**tokens/semantic/colors.json (after):**
```json
{
  "$layer": "semantic",
  "color": {
    "$type": "color",
    "primary": { "$value": "#3b82f6" }
  }
}
```

Then: `tokenctl build --format=tailwind`

### From Design Tool Export

Many design tools export W3C tokens. Import directly:

```bash
# Figma Tokens export
cp figma-export.json tokens/brand/colors.json

# Add layer annotations
# Add $layer field to each token group
```

---

## Extending Design Systems

Packaged design systems can be extended using CSS `@layer` without needing a build step for simple overrides.

### CSS Layer Architecture

Design systems built with tokenctl use this layer order:

```css
@layer tokens, components, themes, user;
```

Later layers automatically override earlier ones—no `!important` needed.

### Level 1: Use As-Is

Just import the base system:

```html
<link rel="stylesheet" href="@acme/design-system/dist/base.css">
```

### Level 2: Override Semantic Tokens

Create a simple CSS file with your brand values:

**my-brand.css:**
```css
@layer user {
  :root {
    --color-primary: #10b981;        /* Your brand green */
    --color-secondary: #6366f1;      /* Your brand purple */
    --font-family-base: "Outfit", sans-serif;
  }
}
```

```html
<link rel="stylesheet" href="@acme/design-system/dist/base.css">
<link rel="stylesheet" href="my-brand.css">
```

All components automatically use your colors—no build step required.

### Customizable Tokens

Design systems should mark which tokens are safe to override:

```json
{
  "color": {
    "primary": {
      "$value": "#3b82f6",
      "$customizable": true,
      "$description": "Override with your brand color"
    },
    "primary-hover": {
      "$value": "darken({color.primary}, 10%)",
      "$description": "Computed - do not override directly"
    }
  }
}
```

Generate a manifest of just the customization points:

```bash
tokenctl build --format=manifest:color --customizable-only
```

**Output (for LLMs):**
```json
{
  "tokens": {
    "color.primary": {
      "value": "#3b82f6",
      "description": "Override with your brand color",
      "customizable": true
    },
    "color.secondary": {
      "value": "#8b5cf6",
      "description": "Secondary brand color",
      "customizable": true
    }
  }
}
```

Non-customizable tokens (computed values, internal tokens) are excluded.

### LLM Customization Pattern

Prompt pattern for LLM-assisted customization:

```
You are customizing a design system.
You can ONLY modify tokens marked "customizable": true.

Available customization points:
{manifest.json contents}

The user wants: "Make it feel more playful with rounded corners"

Generate CSS overrides for @layer user.
```

**LLM output:**
```css
@layer user {
  :root {
    --color-primary: oklch(70% 0.25 330);
    --radius-btn: 9999px;
    --radius-card: 1.5rem;
  }
}
```

### When to Use Token-Level Merge

CSS layer overrides work for 80% of cases. You need token-level merge only when:

1. **Computed values must recalculate** - If you override `primary` and need `primary-hover` to recompute via `darken()`
2. **Manifest accuracy matters** - LLMs need final resolved values including your overrides
3. **Validation of extensions** - Check your overrides against layer rules

For these cases, use multi-directory merge: `tokenctl build ./base ./overrides`. See [MERGE.md](MERGE.md) for details.

---

## Best Practices

### Do

- Define all values as tokens
- Use semantic names (`primary`, not `blue-500` in components)
- Add descriptions and usage hints
- Validate with `--strict-layers`
- Generate category manifests for LLM efficiency
- Use fluid tokens (`clamp()`) for smooth responsive scaling

### Don't

- Use raw values in component definitions
- Skip the semantic layer
- Create tokens without descriptions
- Let components reference brand tokens directly
- Generate full catalogs when a category manifest suffices

### Token Naming

| Layer | Naming Convention | Example |
|-------|-------------------|---------|
| Brand | Descriptive of the value | `blue-500`, `gray-100` |
| Semantic | Descriptive of purpose | `primary`, `error`, `surface` |
| Component | Descriptive of usage | `btn-bg`, `card-shadow` |

---

## Output Formats

| Format | Use Case | Command |
|--------|----------|---------|
| `tailwind` | Tailwind 4 projects | `--format=tailwind` |
| `css` | Non-Tailwind projects | `--format=css` |
| `catalog` | Full export for tools | `--format=catalog` |
| `manifest:CATEGORY` | LLM context efficiency | `--format=manifest:color` |

---

## Summary

tokenctl transforms design system management from chaotic CSS to structured tokens:

1. **Define tokens** in JSON with rich metadata
2. **Organize by layer** (brand → semantic → component)
3. **Add responsive support** with fluid values and breakpoint overrides
4. **Validate architecture** with `--strict-layers`
5. **Generate output** for Tailwind, pure CSS, or JSON manifests
6. **Enable LLMs** with searchable, context-efficient token access

The result: consistent styling that humans and LLMs can both understand and use correctly.
