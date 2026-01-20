<!-- tokenctl/TOKENS.md -->
# Token-Based Design Systems with tokenctl

This guide covers token-based design system concepts and how tokenctl implements them. It serves as both a learning resource for the token approach and a reference for tokenctl's specific features.

## Table of Contents

1. [Token Architecture](#token-architecture)
2. [Token Structure](#token-structure)
3. [Token Types](#token-types)
4. [References](#references)
5. [Expressions & Computed Values](#expressions--computed-values)
6. [Scale Expansion](#scale-expansion)
7. [Constraints](#constraints)
8. [CSS @property Declarations](#css-property-declarations)
9. [CSS @keyframes Animations](#css-keyframes-animations)
10. [Theme System](#theme-system)
11. [Components](#components)
12. [Best Practices](#best-practices)
13. [Troubleshooting](#troubleshooting)

---

## Token Architecture

Token-based design systems organize styling decisions into layers of abstraction. This creates a single source of truth that can generate multiple output formats.

### The Layered Approach

```
┌─────────────────────────────────────┐
│           Components                │  → Button, Card, Input tokens
│    (reference semantic tokens)      │
├─────────────────────────────────────┤
│            Semantic                 │  → primary, success, error, base-100
│    (reference brand tokens)         │
├─────────────────────────────────────┤
│              Brand                  │  → Hex values, OKLCH values
│       (concrete values)             │
└─────────────────────────────────────┘
```

**Why layers matter:**
- Change a brand color once, it propagates everywhere
- Semantic names (`primary`, `success`) work across themes
- Component tokens reference semantic tokens, not raw values

### Directory Structure

```
my-design-system/
├── tokens/
│   ├── brand/
│   │   └── colors.json       # Base color values
│   ├── semantic/
│   │   └── status.json       # success, error, warning
│   ├── spacing/
│   │   └── scale.json        # Spacing scale
│   ├── typography/
│   │   └── fonts.json        # Font families, sizes
│   └── themes/
│       ├── light.json        # Light theme overrides
│       └── dark.json         # Dark theme (can extend light)
└── dist/
    └── tokens.css            # Generated output
```

---

## Token Structure

tokenctl uses the W3C Design Tokens format. Each token is defined with a `$value` and optional metadata.

### Basic Token

```json
{
  "color": {
    "primary": {
      "$value": "#3b82f6",
      "$type": "color",
      "$description": "Primary brand color"
    }
  }
}
```

### Token Fields

| Field | Required | Description |
|-------|----------|-------------|
| `$value` | Yes | The token's value |
| `$type` | No | Type hint for validation and generation |
| `$description` | No | Documentation for the token |
| `$deprecated` | No | Mark token as deprecated (bool or string reason) |

### Type Inheritance

When `$type` is set on a group, all child tokens inherit it:

```json
{
  "color": {
    "$type": "color",
    "primary": { "$value": "#3b82f6" },
    "secondary": { "$value": "#8b5cf6" }
  }
}
```

Both `color.primary` and `color.secondary` inherit `$type: "color"`.

### Nesting

Tokens can be nested to create logical groupings:

```json
{
  "color": {
    "brand": {
      "$type": "color",
      "primary": { "$value": "#3b82f6" },
      "secondary": { "$value": "#8b5cf6" }
    },
    "status": {
      "$type": "color",
      "success": { "$value": "#10b981" },
      "error": { "$value": "#ef4444" }
    }
  }
}
```

Generated CSS variables: `--color-brand-primary`, `--color-brand-secondary`, `--color-status-success`, `--color-status-error`

---

## Token Types

tokenctl validates values based on their `$type`. Supported types:

### color

CSS color values. Accepts hex, rgb, hsl, oklch, or named colors.

```json
{
  "color": {
    "$type": "color",
    "hex": { "$value": "#3b82f6" },
    "oklch": { "$value": "oklch(49.12% 0.309 275.75)" },
    "rgb": { "$value": "rgb(59, 130, 246)" },
    "named": { "$value": "rebeccapurple" }
  }
}
```

OKLCH is recommended for perceptual uniformity when manipulating colors.

### dimension

CSS length values with units.

```json
{
  "spacing": {
    "$type": "dimension",
    "sm": { "$value": "0.5rem" },
    "md": { "$value": "1rem" },
    "lg": { "$value": "1.5rem" }
  }
}
```

Valid units: `px`, `rem`, `em`, `%`, `vw`, `vh`, etc.

### number

Numeric values without units.

```json
{
  "opacity": {
    "$type": "number",
    "disabled": { "$value": 0.5 },
    "hover": { "$value": 0.8 }
  },
  "font": {
    "weight": {
      "$type": "number",
      "normal": { "$value": 400 },
      "bold": { "$value": 700 }
    }
  }
}
```

### fontFamily

Font stack as string or array.

```json
{
  "font": {
    "family": {
      "sans": {
        "$type": "fontFamily",
        "$value": ["Inter", "ui-sans-serif", "system-ui", "sans-serif"]
      },
      "mono": {
        "$type": "fontFamily",
        "$value": ["JetBrains Mono", "ui-monospace", "monospace"]
      }
    }
  }
}
```

Arrays are joined with commas in CSS output.

### duration

Time values for transitions and animations.

```json
{
  "timing": {
    "$type": "duration",
    "fast": { "$value": "150ms" },
    "normal": { "$value": "250ms" },
    "slow": { "$value": "400ms" }
  }
}
```

### effect

Binary toggle values (0 or 1). Used for feature flags like DaisyUI's depth/noise effects.

```json
{
  "effect": {
    "$type": "effect",
    "depth": {
      "$value": 1,
      "$description": "Enable depth shadows"
    },
    "noise": {
      "$value": 0,
      "$description": "Enable noise texture"
    }
  }
}
```

Validation fails if value is anything other than 0 or 1.

---

## References

Tokens can reference other tokens using `{token.path}` syntax.

### Basic Reference

```json
{
  "color": {
    "brand": {
      "primary": { "$value": "#3b82f6" }
    }
  },
  "components": {
    "button": {
      "background": { "$value": "{color.brand.primary}" }
    }
  }
}
```

`components.button.background` resolves to `#3b82f6`.

### Reference Resolution

References are resolved recursively. Given:

```json
{
  "color": {
    "brand": { "$value": "#3b82f6" },
    "semantic": { "$value": "{color.brand}" },
    "button": { "$value": "{color.semantic}" }
  }
}
```

`color.button` resolves to `#3b82f6` through the chain.

### Cycle Detection

Circular references are detected and reported:

```json
{
  "a": { "$value": "{b}" },
  "b": { "$value": "{a}" }
}
```

```
[Error] a: cycle detected: a -> b -> a
```

### Cross-File References

References work across files. A token in `semantic/status.json` can reference a token defined in `brand/colors.json`.

---

## Expressions & Computed Values

tokenctl supports expressions for computed token values.

### calc()

Arithmetic with dimensions:

```json
{
  "spacing": {
    "$type": "dimension",
    "base": { "$value": "1rem" },
    "xs": { "$value": "calc({spacing.base} * 0.25)" },
    "sm": { "$value": "calc({spacing.base} * 0.5)" },
    "lg": { "$value": "calc({spacing.base} * 1.5)" },
    "xl": { "$value": "calc({spacing.base} * 2)" }
  }
}
```

Supports `+`, `-`, `*`, `/` operations.

### contrast()

Generates a WCAG AA compliant content color for a background:

```json
{
  "color": {
    "$type": "color",
    "primary": { "$value": "oklch(49.12% 0.309 275.75)" },
    "primary-content": { "$value": "contrast({color.primary})" }
  }
}
```

`contrast()` returns white or black (in matching format) based on which provides better contrast.

### darken() and lighten()

Adjust color lightness by percentage:

```json
{
  "color": {
    "$type": "color",
    "neutral": { "$value": "#1f2937" },
    "neutral-light": { "$value": "lighten({color.neutral}, 30%)" },
    "neutral-dark": { "$value": "darken({color.neutral}, 20%)" }
  }
}
```

Operations are performed in OKLCH color space for perceptual uniformity.

### shade()

Derive surface shades from a base color. Each level reduces lightness by ~4%:

```json
{
  "color": {
    "$type": "color",
    "base-100": { "$value": "oklch(100% 0 0)" },
    "base-200": { "$value": "shade({color.base-100}, 1)" },
    "base-300": { "$value": "shade({color.base-100}, 2)" }
  }
}
```

Matches DaisyUI's base color progression pattern.

### scale()

Multiply a dimension by a factor:

```json
{
  "size": {
    "field": { "$value": "2.5rem" },
    "field-lg": { "$value": "scale({size.field}, 1.2)" }
  }
}
```

---

## Scale Expansion

The `$scale` field automatically generates variant tokens:

```json
{
  "size": {
    "$type": "dimension",
    "field": {
      "$value": "2.5rem",
      "$description": "Base field size",
      "$scale": {
        "xs": 0.6,
        "sm": 0.8,
        "md": 1.0,
        "lg": 1.2,
        "xl": 1.4
      }
    }
  }
}
```

This expands to:

| Token | Value | Generated Expression |
|-------|-------|---------------------|
| `size.field` | 2.5rem | (base value) |
| `size.field-xs` | 1.5rem | `calc({size.field} * 0.6)` |
| `size.field-sm` | 2rem | `calc({size.field} * 0.8)` |
| `size.field-md` | 2.5rem | `{size.field}` |
| `size.field-lg` | 3rem | `calc({size.field} * 1.2)` |
| `size.field-xl` | 3.5rem | `calc({size.field} * 1.4)` |

For factor `1.0`, a direct reference is used instead of calc.

---

## Constraints

Dimension and number tokens support `$min` and `$max` constraints:

```json
{
  "size": {
    "$type": "dimension",
    "field": {
      "$value": "2.5rem",
      "$min": "1rem",
      "$max": "5rem"
    }
  },
  "opacity": {
    "$type": "number",
    "disabled": {
      "$value": 0.5,
      "$min": 0,
      "$max": 1
    }
  }
}
```

Validation fails if `$value` is outside the specified range:

```
[Error] size.field [tokens/sizes.json]: constraint violation: value 0.5rem is less than min 1rem
```

---

## CSS @property Declarations

Add `$property: true` to generate CSS `@property` declarations for type-safe custom properties:

```json
{
  "color": {
    "$type": "color",
    "primary": {
      "$value": "oklch(49.12% 0.309 275.75)",
      "$property": true
    }
  }
}
```

Generated CSS:

```css
@property --color-primary {
  syntax: '<color>';
  inherits: true;
  initial-value: oklch(49.12% 0.309 275.75);
}

@theme {
  --color-primary: oklch(49.12% 0.309 275.75);
}
```

### Custom Inheritance

Disable inheritance for properties that shouldn't cascade:

```json
{
  "timing": {
    "$type": "duration",
    "fast": {
      "$value": "150ms",
      "$property": { "inherits": false }
    }
  }
}
```

### Type to Syntax Mapping

| Token `$type` | CSS `syntax` |
|---------------|--------------|
| `color` | `<color>` |
| `dimension` | `<length>` |
| `number` | `<number>` |
| `duration` | `<time>` |
| `effect` | `<integer>` |

Types without a mapping (like `fontFamily`) are skipped.

### Animated Theme Transitions

With `@property` declarations, theme switches can animate:

```css
:root {
  transition: --color-primary 300ms ease;
}
```

Without `@property`, custom properties change instantly. With it, they transition.

---

## CSS @keyframes Animations

Define CSS animations as tokens in a `keyframes` section at the root level of any token file.

### Basic Keyframes

```json
{
  "keyframes": {
    "skeleton-pulse": {
      "0%, 100%": { "opacity": "1" },
      "50%": { "opacity": "0.5" }
    }
  }
}
```

**Generated CSS:**
```css
@keyframes skeleton-pulse {
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: 0.5;
  }
}
```

### Using from/to Keywords

```json
{
  "keyframes": {
    "slide-in": {
      "from": { "transform": "translateX(-100%)" },
      "to": { "transform": "translateX(0)" }
    },
    "fade-in": {
      "from": { "opacity": "0" },
      "to": { "opacity": "1" }
    }
  }
}
```

### Complex Animations

```json
{
  "keyframes": {
    "bounce": {
      "0%, 100%": {
        "transform": "translateY(0)",
        "animation-timing-function": "ease-out"
      },
      "50%": {
        "transform": "translateY(-25%)",
        "animation-timing-function": "ease-in"
      }
    }
  }
}
```

### Referencing in Components

Keyframes are referenced by name in component animation properties:

```json
{
  "components": {
    "skeleton": {
      "$class": "skeleton",
      "base": {
        "animation": "skeleton-pulse 2s cubic-bezier(0.4, 0, 0.6, 1) infinite"
      }
    }
  },
  "keyframes": {
    "skeleton-pulse": {
      "0%, 100%": { "opacity": "1" },
      "50%": { "opacity": "0.5" }
    }
  }
}
```

### Keyframe Ordering

Frames are automatically sorted by percentage in the generated CSS:
- `from` is treated as 0%
- `to` is treated as 100%
- Percentage values are sorted numerically

---

## Theme System

Themes are defined in `tokens/themes/` and override base tokens.

### Theme Files

**tokens/themes/light.json:**
```json
{
  "color": {
    "brand": {
      "primary": { "$value": "#60a5fa" }
    }
  }
}
```

**tokens/themes/dark.json:**
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

### Theme Inheritance

The `$extends` field creates theme inheritance chains:

- `dark` extends `light`
- Dark theme inherits all light theme values
- Only overridden tokens need to be specified

Circular inheritance is detected and reported.

### Generated Output

```css
@import "tailwindcss";

@theme {
  --color-brand-primary: #3b82f6;
  /* ... base tokens */
}

@layer base {
  :root, [data-theme="light"] {
    --color-brand-primary: #60a5fa;
  }
  [data-theme="dark"] {
    --color-brand-primary: #1e40af;
  }
}
```

Only tokens that differ from the base are output in theme blocks.

### Theme Switching

```html
<html data-theme="light">
  <!-- or -->
<html data-theme="dark">
```

```javascript
function toggleTheme() {
  const html = document.documentElement;
  const current = html.getAttribute('data-theme');
  html.setAttribute('data-theme', current === 'light' ? 'dark' : 'light');
}
```

---

## Components

Components are defined with `$type: "component"` and generate CSS classes.

### Component Structure

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
          "color": "{color.brand.primary-content}"
        },
        "secondary": {
          "$class": "btn-secondary",
          "background-color": "transparent",
          "border": "1px solid {color.brand.primary}",
          "color": "{color.brand.primary}"
        }
      },
      "sizes": {
        "sm": {
          "$class": "btn-sm",
          "padding": "{spacing.xs} {spacing.sm}",
          "font-size": "0.875rem"
        },
        "lg": {
          "$class": "btn-lg",
          "padding": "{spacing.md} {spacing.lg}",
          "font-size": "1.125rem"
        }
      }
    }
  }
}
```

### State Selectors

Variants can include pseudo-class states:

```json
{
  "primary": {
    "$class": "btn-primary",
    "background-color": "{color.brand.primary}",
    ":hover": {
      "background-color": "{color.brand.primary-hover}"
    },
    ":active": {
      "transform": "scale(0.98)"
    }
  }
}
```

### Generated CSS

```css
@layer components {
  .btn {
    /* base styles */
  }
  .btn-primary {
    background-color: var(--color-brand-primary);
    color: var(--color-brand-primary-content);
  }
  .btn-primary:hover {
    background-color: var(--color-brand-primary-hover);
  }
  .btn-secondary {
    background-color: transparent;
    border: 1px solid var(--color-brand-primary);
    color: var(--color-brand-primary);
  }
  .btn-sm {
    padding: var(--spacing-xs) var(--spacing-sm);
    font-size: 0.875rem;
  }
  .btn-lg {
    padding: var(--spacing-md) var(--spacing-lg);
    font-size: 1.125rem;
  }
}
```

Token references in component properties are converted to `var(--token-path)`.

---

## Best Practices

### Naming Conventions

**Do:**
- Use semantic names: `primary`, `success`, `base-content`
- Pair colors with content variants: `primary` + `primary-content`
- Use consistent hierarchy: `color.brand.primary`, `spacing.md`

**Don't:**
- Use appearance-based names: `blue-500`, `dark-gray`
- Skip content colors for backgrounds

### Content Color Pairing

Every background color should have a matching content color:

```json
{
  "color": {
    "$type": "color",
    "primary": { "$value": "oklch(49.12% 0.309 275.75)" },
    "primary-content": { "$value": "contrast({color.primary})" },
    "success": { "$value": "#10b981" },
    "success-content": { "$value": "contrast({color.success})" }
  }
}
```

### File Organization

Group tokens by purpose:

```
tokens/
├── brand/colors.json       # Core brand colors
├── semantic/status.json    # success, error, warning, info
├── spacing/scale.json      # Spacing scale
├── typography/fonts.json   # Font families
├── sizing/fields.json      # Component sizes
└── themes/
    ├── light.json
    └── dark.json
```

### Type Annotations

Always annotate groups with `$type` for validation:

```json
{
  "color": {
    "$type": "color",
    "$description": "Color tokens",
    "primary": { "$value": "#3b82f6" }
  }
}
```

---

## Troubleshooting

### Reference Not Found

```
[Error] components.button.bg [tokens/components.json]: reference not found: color.prinary
```

**Cause:** Typo in reference path or referenced token doesn't exist.

**Fix:** Check the token path. References are case-sensitive.

### Circular Reference Detected

```
[Error] color.a: cycle detected: color.a -> color.b -> color.a
```

**Cause:** Two or more tokens reference each other in a loop.

**Fix:** Break the cycle by using a concrete value for one token:

```json
{
  "color": {
    "a": { "$value": "#3b82f6" },
    "b": { "$value": "{color.a}" }
  }
}
```

### Invalid Color Format

```
[Error] color.primary [tokens/brand.json]: invalid color: unable to parse "notacolor"
```

**Cause:** Color value isn't a valid CSS color.

**Fix:** Use hex, rgb, hsl, oklch, or named colors:

```json
{
  "primary": { "$value": "#3b82f6" }
}
```

### Constraint Violation

```
[Error] size.field [tokens/sizes.json]: constraint violation: value 0.5rem is less than min 1rem
```

**Cause:** Token value is outside `$min`/`$max` bounds.

**Fix:** Adjust the value or constraints.

### Invalid Effect Value

```
[Error] effect.depth [tokens/effects.json]: invalid effect: effect must be 0 or 1, got 2
```

**Cause:** Effect tokens only accept 0 or 1.

**Fix:** Use `0` to disable, `1` to enable.

### Expected Object Error

```
[Error] spacing.md: expected object, got string
```

**Cause:** A token path points to a primitive instead of a token object.

**Fix:** Ensure all tokens have the `$value` wrapper:

```json
{
  "spacing": {
    "md": { "$value": "1rem" }
  }
}
```

Not:

```json
{
  "spacing": {
    "md": "1rem"
  }
}
```

### Debugging

1. **Validate frequently:**
   ```bash
   tokenctl validate
   ```
   Errors include source file paths for easy location.

2. **Check generated output:**
   ```bash
   tokenctl build --output=./debug
   cat ./debug/tokens.css
   ```

3. **Use the examples:**
   ```bash
   tokenctl build examples/computed --output=dist/computed
   tokenctl build examples/themes --output=dist/themes
   tokenctl build examples/validation --output=dist/validation
   ```

4. **Run the demo workflow:**
   ```bash
   make demo
   ```

---

## Quick Reference

### CLI Commands

```bash
tokenctl init [dir]           # Initialize token system
tokenctl validate [dir]       # Validate tokens
tokenctl build [dir]          # Build artifacts
  --format=tailwind         # CSS output (default)
  --format=catalog          # JSON catalog
  --output=<dir>            # Output directory (default: dist)
```

### Token Syntax

```json
{
  "group": {
    "$type": "color",
    "$description": "Group description",
    "token": {
      "$value": "#3b82f6",
      "$description": "Token description",
      "$property": true,
      "$min": "...",
      "$max": "...",
      "$scale": { "xs": 0.6, "md": 1.0, "xl": 1.4 }
    }
  }
}
```

### Expression Functions

| Function | Example | Description |
|----------|---------|-------------|
| Reference | `{color.primary}` | Reference another token |
| calc | `calc({spacing.base} * 2)` | Arithmetic with dimensions |
| contrast | `contrast({color.primary})` | WCAG AA content color |
| darken | `darken({color.neutral}, 20%)` | Reduce lightness |
| lighten | `lighten({color.neutral}, 30%)` | Increase lightness |
| shade | `shade({color.base}, 1)` | Derive surface shade |
| scale | `scale({size.field}, 1.2)` | Multiply dimension |

### Token Types

| Type | Validation |
|------|------------|
| `color` | Valid CSS color |
| `dimension` | Number with unit |
| `number` | Numeric value |
| `fontFamily` | String or array |
| `duration` | Time value |
| `effect` | 0 or 1 |
| `component` | Component definition |
