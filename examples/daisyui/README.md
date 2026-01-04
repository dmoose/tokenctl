# DaisyUI 5 Theme Generator Example

This example demonstrates how to use tokctl as a **theme generator for DaisyUI 5**. It defines all 26 DaisyUI token variables, allowing you to create custom themes that integrate seamlessly with DaisyUI's component library.

## What This Example Provides

- **18 Color Tokens** - All DaisyUI semantic colors with auto-generated content colors
- **2 Size Tokens** - Field and selector sizes with 5-level scale (xs, sm, md, lg, xl)
- **1 Border Token** - Default border width
- **3 Radius Tokens** - Box, field, and selector border radii
- **2 Effect Tokens** - Depth and noise toggles
- **Light & Dark Themes** - Complete theme variations

## Token Architecture

```
tokens/
├── colors.json      # 18 color tokens (primary, secondary, accent, neutral, base, status)
├── sizes.json       # size-field, size-selector with $scale variants
├── borders.json     # border width and radius tokens
├── effects.json     # depth and noise effect toggles
└── themes/
    ├── light.json   # Light theme (default)
    └── dark.json    # Dark theme (extends light)
```

## Key Features Demonstrated

### Auto-Generated Content Colors

Content colors are automatically calculated for WCAG AA contrast compliance:

```json
{
  "primary": {
    "$value": "oklch(49.12% 0.309 275.75)"
  },
  "primary-content": {
    "$value": "contrast({color.primary})"
  }
}
```

### Auto-Derived Base Shades

Base-200 and base-300 are derived from base-100 using `shade()`:

```json
{
  "base-100": { "$value": "oklch(100% 0 0)" },
  "base-200": { "$value": "shade({color.base-100}, 1)" },
  "base-300": { "$value": "shade({color.base-100}, 2)" }
}
```

### Size Scales

Size tokens automatically generate xs, sm, md, lg, xl variants:

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

Generated output:
- `--size-field: 2.5rem`
- `--size-field-xs: 1.5rem`
- `--size-field-sm: 2rem`
- `--size-field-md: 2.5rem`
- `--size-field-lg: 3rem`
- `--size-field-xl: 3.5rem`

### CSS @property Declarations

Color tokens include `$property: true` for smooth animated theme transitions:

```json
{
  "primary": {
    "$value": "oklch(49.12% 0.309 275.75)",
    "$property": true
  }
}
```

## Usage

### 1. Build the Theme

```bash
tokctl build examples/daisyui --output dist/daisyui
```

### 2. Generated Output

The build produces `dist/daisyui/tokens.css`:

```css
@property --color-primary {
  syntax: '<color>';
  inherits: true;
  initial-value: oklch(49.12% 0.309 275.75);
}
/* ... more @property declarations */

@import "tailwindcss";

@theme {
  --color-primary: oklch(49.12% 0.309 275.75);
  --color-primary-content: oklch(100.00% 0.000 275.75);
  --color-secondary: oklch(69.71% 0.329 342.55);
  /* ... all 26 tokens plus size variants */
}

@layer base {
  :root, [data-theme="light"] {
    /* light theme is default */
  }
  [data-theme="dark"] {
    --color-primary: oklch(65.69% 0.196 275.75);
    --color-base-100: oklch(25.33% 0.016 252.42);
    /* ... dark theme overrides */
  }
}
```

### 3. Use with DaisyUI

In your Tailwind CSS file:

```css
/* Import tokctl-generated tokens BEFORE DaisyUI */
@import "./dist/daisyui/tokens.css";

/* DaisyUI components will use your custom tokens */
@plugin "daisyui";
```

Or in `tailwind.config.js`:

```js
export default {
  plugins: [require("daisyui")],
  daisyui: {
    themes: false, // Disable built-in themes, use tokctl's
  },
}
```

### 4. Theme Switching

Use the standard DaisyUI theme switching mechanism:

```html
<!-- Light theme (default) -->
<html data-theme="light">

<!-- Dark theme -->
<html data-theme="dark">
```

With `$property` declarations, theme transitions animate smoothly:

```css
:root {
  transition: --color-primary 300ms ease,
              --color-base-100 300ms ease;
}
```

## Customizing Your Theme

### Change Primary Color

Edit `tokens/colors.json`:

```json
{
  "primary": {
    "$value": "oklch(55% 0.25 200)",
    "$property": true
  }
}
```

The content color is auto-generated - no need to calculate manually.

### Adjust Base Colors

Change only `base-100`; the shades are derived automatically:

```json
{
  "base-100": {
    "$value": "oklch(98% 0.01 240)"
  }
}
```

`base-200` and `base-300` will be slightly darker versions.

### Modify Size Scale

Adjust the base size or scale multipliers:

```json
{
  "field": {
    "$value": "3rem",
    "$scale": { "xs": 0.5, "sm": 0.75, "md": 1.0, "lg": 1.25, "xl": 1.5 }
  }
}
```

### Enable Effects

Turn on depth shadows in `effects.json` or per-theme:

```json
{
  "depth": { "$value": 1 }
}
```

## DaisyUI Token Reference

| Token | CSS Variable | Default (Light) | Purpose |
|-------|--------------|-----------------|---------|
| **Brand Colors** | | | |
| primary | `--color-primary` | `oklch(49.12% 0.309 275.75)` | Main brand color |
| primary-content | `--color-primary-content` | Auto-generated | Text on primary |
| secondary | `--color-secondary` | `oklch(69.71% 0.329 342.55)` | Secondary brand |
| secondary-content | `--color-secondary-content` | Auto-generated | Text on secondary |
| accent | `--color-accent` | `oklch(76.76% 0.184 183.61)` | Accent color |
| accent-content | `--color-accent-content` | Auto-generated | Text on accent |
| **Neutral** | | | |
| neutral | `--color-neutral` | `oklch(20% 0.024 255.701)` | Dark neutral |
| neutral-content | `--color-neutral-content` | Auto-generated | Text on neutral |
| **Base** | | | |
| base-100 | `--color-base-100` | `oklch(100% 0 0)` | Page background |
| base-200 | `--color-base-200` | `shade(base-100, 1)` | Slight elevation |
| base-300 | `--color-base-300` | `shade(base-100, 2)` | More elevation |
| base-content | `--color-base-content` | Auto-generated | Text on base |
| **Status** | | | |
| info | `--color-info` | `oklch(72.06% 0.191 231.6)` | Info messages |
| success | `--color-success` | `oklch(64.8% 0.15 160)` | Success messages |
| warning | `--color-warning` | `oklch(84.71% 0.199 83.87)` | Warning messages |
| error | `--color-error` | `oklch(71.76% 0.221 22.18)` | Error messages |
| **Sizes** | | | |
| size-field | `--size-field` | `2.5rem` | Button, input height |
| size-selector | `--size-selector` | `1.5rem` | Checkbox, radio size |
| **Border** | | | |
| border | `--border-default` | `1px` | Border width |
| **Radius** | | | |
| radius-box | `--radius-box` | `1rem` | Card, modal radius |
| radius-field | `--radius-field` | `0.5rem` | Button, input radius |
| radius-selector | `--radius-selector` | `9999px` | Checkbox, badge radius |
| **Effects** | | | |
| depth | `--effect-depth` | `0` | Shadow effect (0/1) |
| noise | `--effect-noise` | `0` | Noise texture (0/1) |

## Validation

Validate your theme before building:

```bash
tokctl validate examples/daisyui
```

This checks:
- Color format validity
- Dimension constraints (`$min`/`$max`)
- Effect values (must be 0 or 1)
- Reference resolution

## Notes

- **tokctl generates tokens only** - Component CSS (`.btn`, `.card`, etc.) comes from DaisyUI
- **OKLCH format recommended** - Better perceptual uniformity for color manipulation
- **Content colors are WCAG AA compliant** - The `contrast()` function ensures readable text
- **Animated themes** - `$property` declarations enable smooth color transitions

## See Also

- [DaisyUI Documentation](https://daisyui.com/docs)
- [tokctl ADVANCED_USAGE.md](../../ADVANCED_USAGE.md) - CSS composition patterns
- [Tailwind CSS 4 Theming](https://tailwindcss.com/docs/theme)