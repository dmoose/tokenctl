# Multi-Directory Merge

tokenctl can merge multiple token directories into a single design system. This is useful when you have a base component library and one or more extensions that build on top of it.

## Use Case

You maintain a shared component library with brand colors, semantic tokens, and base components. A dashboard project extends this with additional tokens and component variants. Rather than duplicating the base system, you point tokenctl at both directories:

```bash
tokenctl build ./base-components ./dashboard-ext --output=./dist
tokenctl validate ./base-components ./dashboard-ext
```

Directories are merged left-to-right. The first directory is the foundation; each subsequent directory extends or overrides it.

## Directory Layout

Each directory follows the standard tokenctl layout:

```
base-components/
  tokens/
    brand/colors.json
    semantic/status.json
    components/button.json
    themes/dark.json

dashboard-ext/
  tokens/
    brand/colors.json        # overrides blue-500, adds red-500
    semantic/dashboard.json   # adds danger semantic token
    components/button.json    # adds danger button variant
    themes/dark.json          # adds red-500 dark override
```

## Merge Rules

### Tokens (nodes with `$value`)

Tokens are replaced entirely. If the extension defines a token at the same path, its value, description, and all metadata replace the base token completely.

**Base** `brand/colors.json`:
```json
{
  "color": {
    "brand": {
      "blue-500": {
        "$value": "#3b82f6",
        "$description": "Primary brand blue"
      }
    }
  }
}
```

**Extension** `brand/colors.json`:
```json
{
  "color": {
    "brand": {
      "blue-500": {
        "$value": "#2563eb",
        "$description": "Dashboard brand blue"
      }
    }
  }
}
```

**Result**: `color.brand.blue-500` is `#2563eb` with description "Dashboard brand blue".

### Groups and Components (nodes without `$value`)

Groups and components are merged recursively. The extension can add new children without affecting existing ones.

**Base** `components/button.json`:
```json
{
  "button": {
    "$type": "component",
    "primary": {
      "background-color": { "$value": "{color.semantic.primary}" }
    },
    "success": {
      "background-color": { "$value": "{color.semantic.success}" }
    }
  }
}
```

**Extension** `components/button.json`:
```json
{
  "button": {
    "$type": "component",
    "danger": {
      "background-color": { "$value": "{color.semantic.danger}" }
    }
  }
}
```

**Result**: The button component has all three variants: primary, success, and danger.

### Themes

Themes with the same name are merged. Themes that only exist in one directory pass through unchanged.

**Base** `themes/dark.json`:
```json
{
  "color": { "brand": { "blue-500": { "$value": "#60a5fa" } } }
}
```

**Extension** `themes/dark.json`:
```json
{
  "color": { "brand": { "red-500": { "$value": "#f87171" } } }
}
```

**Result**: The dark theme overrides both `blue-500` and `red-500`.

### References

References resolve against the merged result. An extension can define a token that references tokens from the base, or vice versa. All references are resolved after the full merge is complete.

## Common Patterns

### Add new tokens

The extension defines tokens at new paths. These are added alongside the base tokens.

```json
{
  "color": {
    "brand": {
      "red-500": { "$value": "#ef4444" }
    }
  }
}
```

### Add component variants

The extension adds children to an existing component. Base variants are preserved.

```json
{
  "button": {
    "$type": "component",
    "danger": {
      "background-color": { "$value": "{color.semantic.danger}" }
    }
  }
}
```

### Override a token value

The extension redefines a token at the same path. The extension's value wins.

```json
{
  "color": {
    "brand": {
      "blue-500": { "$value": "#2563eb" }
    }
  }
}
```

### Extend a theme

The extension adds overrides to an existing theme.

```json
{
  "color": {
    "brand": {
      "red-500": { "$value": "#f87171" }
    }
  }
}
```

## Gotchas

### 1. Token metadata is replaced, not merged

When you override a token value, all metadata (`$description`, `$type`, `$min`, `$max`, etc.) from the base is lost. The extension must include any metadata it wants to keep.

**Workaround**: Copy the relevant metadata fields into your extension token.

### 2. Merge is additive only

There is no way to remove a token that exists in the base. The extension can override its value but cannot delete it.

**Workaround**: If you need to remove tokens, restructure the base so the tokens you want to exclude are in a separate directory that you don't include.

### 3. Directory order matters

`tokenctl build A B` and `tokenctl build B A` produce different results when both define the same token. The last directory wins.

**Workaround**: Establish a clear convention: base directories first, extensions after.

### 4. Scale replacement leaves orphan steps

If the base defines a `$scale` (e.g., xs through xl) and the extension overrides the base token's `$value`, the scale expansion still uses the base's `$scale` definition because scales are expanded during load. The override only replaces the individual token, not the expanded variants.

**Workaround**: If you need to change a scaled token, override each expanded step individually, or define the complete scale in the extension.

### 5. Validate the merged result, not individual directories

An extension directory may contain references to tokens that only exist in the base. Running `tokenctl validate` on the extension alone will report broken references.

**Workaround**: Always validate with all directories: `tokenctl validate ./base ./ext`.
