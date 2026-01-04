<!-- tokctl/examples/validation/README.md -->
# Validation Example

This example demonstrates enhanced tokctl features:

1. **Constraint Validation** - `$min` and `$max` constraints on dimension and number tokens
2. **Type-Specific Validation** - Validation for color, dimension, number, fontFamily, and effect types
3. **Effect Tokens** - DaisyUI-style effect toggles (0 or 1)
4. **CSS @property Declarations** - Type-safe custom properties with animation support

## Token File

See `tokens/tokens.json` for the complete example.

## Features Demonstrated

### Constraint Validation

Dimension tokens can have `$min` and `$max` constraints:

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
  "radius": {
    "$type": "dimension",
    "box": {
      "$value": "1rem",
      "$min": "0rem",
      "$max": "3rem"
    }
  }
}
```

Note: `$min` and `$max` must use the same unit as the value.

Number tokens also support constraints:

```json
{
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

### Effect Tokens

Effect tokens are boolean-like values (0 or 1) used to enable/disable CSS effects:

```json
{
  "effect": {
    "$type": "effect",
    "depth": {
      "$value": 1,
      "$description": "Enable depth shadows on components"
    },
    "noise": {
      "$value": 0,
      "$description": "Enable noise texture overlay"
    }
  }
}
```

### Type-Specific Validation

The validator checks values based on their `$type`:

| Type | Validation |
|------|------------|
| `color` | Valid CSS color format (hex, rgb, hsl, oklch, named) |
| `dimension` | Valid CSS dimension (number + unit), range constraints |
| `number` | Numeric value, range constraints |
| `fontFamily` | Non-empty string or array of strings |
| `effect` | Value must be 0 or 1 |

### CSS @property Declarations

Add `$property: true` to generate `@property` declarations for animatable tokens:

```json
{
  "color": {
    "$type": "color",
    "primary": {
      "$value": "oklch(50% 0.2 250)",
      "$property": true
    }
  },
  "timing": {
    "fast": {
      "$value": "150ms",
      "$type": "duration",
      "$property": { "inherits": false }
    }
  }
}
```

This enables smooth CSS transitions on theme changes.

## Usage

### Validate

```bash
tokctl validate examples/validation
```

### Build

```bash
tokctl build examples/validation --output dist/validation
```

### Expected Output

The generated CSS will include:

```css
@import "tailwindcss";

@theme {
  --border-default: 1px;
  --border-thick: 2px;
  --color-primary: oklch(49.12% 0.309 275.75);
  --color-primary-content: oklch(100.00% 0.000 0.00);
  --color-secondary: #8b5cf6;
  --effect-depth: 1;
    --effect-noise: 0;
    --timing-fast: 150ms;
    --timing-normal: 250ms;
    --font-family-mono: JetBrains Mono, ui-monospace, monospace;
  --font-family-sans: Inter, ui-sans-serif, system-ui, sans-serif;
  --opacity-disabled: 0.5;
  --opacity-hover: 0.8;
  --radius-box: 1rem;
  --radius-field: 0.5rem;
  --radius-selector: 9999px;
  --size-field: 2.5rem;
  --size-selector: 1.5rem;
}
```

## Testing Validation Errors

To see validation errors, try modifying `tokens.json`:

1. **Constraint violation**: Change `size.field.$value` to `"0.5rem"` (below min of 1rem)
2. **Invalid effect**: Change `effect.depth.$value` to `2` (must be 0 or 1)
3. **Invalid color**: Change `color.primary.$value` to `"not-a-color"`
4. **Empty fontFamily**: Change `font.family.sans.$value` to `[]`

Run `tokctl validate examples/validation` to see the error messages.

## See Also

- [ADVANCED_USAGE.md](../../ADVANCED_USAGE.md) - Full documentation on `$property` and CSS composition patterns