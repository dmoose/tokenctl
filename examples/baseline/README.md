# Baseline Design System

A complete design system example demonstrating tokenctl's capabilities.

## Features Demonstrated

- **Layer Architecture**: Brand -> Semantic -> Component token hierarchy
- **Component System**: Buttons, cards, forms, alerts, tables, badges, layout primitives
- **CSS Reset**: Built-in minimal reset in `@layer reset`
- **Theming**: Light/dark themes with `data-theme` switching
- **Responsive Tokens**: Font sizes that scale with breakpoints
- **Composition Metadata**: `$contains` and `$requires` for component relationships
- **Customizable Tokens**: `$customizable: true` marks override points

## Quick Start

```bash
# Build the CSS
tokenctl build examples/baseline --format=css --output=examples/baseline/dist

# Open the demo
open examples/baseline/demo.html
```

## Structure

```
baseline/
├── tokens/
│   ├── brand/
│   │   ├── colors.json      # Raw color palette (OKLCH)
│   │   └── scales.json      # Spacing, radius, shadows, typography
│   ├── semantic/
│   │   ├── colors.json      # primary, success, error, surface, text
│   │   └── typography.json  # Font stacks, sizes (with responsive)
│   ├── components/
│   │   ├── alert.json       # alert, alert-success, alert-warning, alert-error
│   │   ├── button.json      # btn, btn-primary, btn-ghost, sizes
│   │   ├── card.json        # card, card-body, card-title ($contains)
│   │   ├── input.json       # input, textarea, select, label
│   │   ├── badge.json       # badge variants
│   │   ├── layout.json      # container, stack, grid, section
│   │   ├── prose.json       # link, list, divider, description lists
│   │   └── table.json       # table, th, td, tr with hover
│   └── themes/
│       ├── light.json       # Default light theme
│       └── dark.json        # Dark theme ($extends light)
├── dist/
│   └── tokens.css           # Generated CSS
└── demo.html                # Interactive demo page
```

## Components

| Component | Classes | Description |
|-----------|---------|-------------|
| **Alert** | `.alert`, `.alert-info`, `.alert-success`, `.alert-warning`, `.alert-error` | Contextual feedback messages |
| **Button** | `.btn`, `.btn-primary`, `.btn-secondary`, `.btn-outline`, `.btn-ghost`, `.btn-sm`, `.btn-lg` | Interactive buttons with variants and sizes |
| **Card** | `.card`, `.card-body`, `.card-title`, `.card-text`, `.card-actions`, `.card-image` | Content containers with composition rules |
| **Input** | `.input`, `.textarea`, `.select`, `.label`, `.form-group` | Form elements |
| **Badge** | `.badge`, `.badge-primary`, `.badge-success`, `.badge-warning`, `.badge-error` | Status labels |
| **Table** | `.table`, `.th`, `.td`, `.tr` | Data tables with hover states |
| **Prose** | `.link`, `.list`, `.list-item`, `.divider`, `.dl`, `.dt`, `.dd` | Typography and content elements |
| **Layout** | `.container`, `.stack`, `.row`, `.grid`, `.section`, `.prose` | Layout primitives |

## Customization

### Via CSS Layers (No Build)

```css
@layer user {
  :root {
    --color-primary: #10b981;
    --font-family-sans: "Your Font", sans-serif;
  }
}
```

### View Customizable Tokens

```bash
tokenctl build examples/baseline --format=manifest:color --customizable-only
```

Output shows only tokens marked for customization:

```json
{
  "tokens": {
    "color.primary": {
      "value": "oklch(55% 0.20 250)",
      "customizable": true,
      "description": "Primary brand color for key actions and focus"
    }
  }
}
```

## Theming

Toggle themes via JavaScript:

```javascript
document.documentElement.setAttribute('data-theme', 'dark');
```

Or HTML attribute:

```html
<html data-theme="dark">
```
