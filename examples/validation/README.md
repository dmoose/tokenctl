# Validation Example

This example demonstrates the enhanced validation features in tokctl:

1. **Constraint Validation** - `$min` and `$max` constraints on dimension and number tokens
2. **Type-Specific Validation** - Validation for color, dimension, number, fontFamily, and effect types
3. **Effect Tokens** - DaisyUI-style effect toggles (0 or 1)

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
  }
}
```

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