# tokenctl Enhancement Plan

## Overview

This plan extends tokenctl to be a complete design system tool that enables:
- LLM-friendly token consumption (context-efficient manifests)
- Clean, enforceable design system architecture
- Modern responsive design via fluid tokens and breakpoint overrides
- Standalone operation (Tailwind optional)

## Implementation Priority

| # | Feature | Effort | Status |
|---|---------|--------|--------|
| 1 | Category-scoped manifests | ~2 hrs | ✅ Complete |
| 2 | Rich descriptions in catalog | ~1 hr | ✅ Complete |
| 3 | Pure CSS output | ~2 hrs | ✅ Complete |
| 4 | Token search CLI | ~3 hrs | ✅ Complete |
| 5 | Component composition metadata | ~4 hrs | ✅ Complete |
| 6 | Semantic layer validation | ~4 hrs | ✅ Complete |
| 7 | Responsive tokens (fluid + overrides) | ~6 hrs | ✅ Complete |

**Total estimated effort**: ~22 hours
**Completed**: All features implemented

---

## Feature Specifications

### 1. Category-Scoped Manifests

**Problem**: Full catalog wastes LLM context tokens.

**Solution**: Filter output by category.

```bash
tokenctl build --format=manifest:colors
tokenctl build --format=manifest:spacing
tokenctl build --format=manifest:components
```

**Schema additions**: None required. Categories derived from top-level keys or explicit `$category`.

**Output format**:
```json
{
  "meta": {
    "category": "colors",
    "version": "2.0",
    "generated_at": "...",
    "tokenctl_version": "1.2.0"
  },
  "tokens": {
    "color.primary": { ... }
  }
}
```

**Files to modify**:
- `cmd/tokenctl/build.go` - Parse format flag for category
- `pkg/generators/catalog.go` - Add category filtering

---

### 2. Rich Descriptions in Catalog

**Problem**: Catalog has values but no usage context for LLMs.

**Schema additions**:
```json
{
  "color": {
    "primary": {
      "$value": "#3b82f6",
      "$description": "Primary brand color",
      "$usage": ["btn-primary background", "links", "focus rings"],
      "$avoid": "Don't use for large background areas"
    }
  }
}
```

**Catalog output**:
```json
{
  "color.primary": {
    "value": "#3b82f6",
    "type": "color",
    "description": "Primary brand color",
    "usage": ["btn-primary background", "links", "focus rings"],
    "avoid": "Don't use for large background areas"
  }
}
```

**Files to modify**:
- `pkg/generators/catalog.go` - Include rich metadata in output
- `pkg/tokens/types.go` - Document new fields (optional, already flexible)

---

### 3. Pure CSS Output

**Problem**: Tailwind dependency limits adoption.

**Solution**: New format option.

```bash
tokenctl build --format=css
```

**Output** (no Tailwind import, standard CSS):
```css
:root {
  --color-primary: #3b82f6;
  --spacing-md: 1rem;
}

@layer components {
  .btn { ... }
}
```

**Files to modify**:
- `cmd/tokenctl/build.go` - Add "css" format case
- `pkg/generators/css.go` - New generator (or flag in tailwind.go)

---

### 4. Token Search CLI

**Problem**: Finding relevant tokens requires reading files.

**Solution**: Search command.

```bash
tokenctl search "primary"
tokenctl search --type=color
tokenctl search --category=spacing
```

**Output**:
```
color.primary: #3b82f6
  Primary brand color. Use for buttons, links, focus rings.

color.primary-content: #ffffff
  Text color on primary backgrounds.
```

**Files to modify**:
- `cmd/tokenctl/search.go` - New command
- Uses existing loader/resolver infrastructure

---

### 5. Component Composition Metadata

**Problem**: LLMs don't know component relationships.

**Schema additions**:
```json
{
  "components": {
    "card": {
      "$type": "component",
      "$class": "card",
      "$contains": ["card-body", "card-title", "card-actions"],
      "$description": "Container for card content"
    },
    "card-body": {
      "$type": "component",
      "$class": "card-body",
      "$requires": "card",
      "$description": "Main content area, must be inside card"
    }
  }
}
```

**Catalog output** includes `contains` and `requires` fields.

**Files to modify**:
- `pkg/tokens/components.go` - Extract new fields
- `pkg/generators/catalog.go` - Include in output

---

### 6. Semantic Layer Validation

**Problem**: No enforcement of design system architecture.

**Schema additions**:
```json
{
  "color": {
    "$layer": "brand",
    "blue-500": { "$value": "#3b82f6" }
  },
  "semantic": {
    "$layer": "semantic",
    "primary": { "$value": "{color.blue-500}" }
  }
}
```

**Validation rules** (with `--strict-layers`):
- `brand` layer: Raw values only
- `semantic` layer: Can reference brand tokens
- `component` layer: Can only reference semantic tokens

**Files to modify**:
- `pkg/tokens/validator.go` - Add layer validation
- `cmd/tokenctl/validate.go` - Add `--strict-layers` flag

---

### 7. Responsive Tokens

**Problem**: Need structured responsive design without CSS chaos.

**Two approaches supported**:

#### 7a. Fluid Tokens (clamp)

Already works—just use clamp in values:
```json
{
  "spacing": {
    "section": {
      "$value": "clamp(2rem, 5vw, 6rem)",
      "$description": "Fluid section spacing"
    }
  }
}
```

No code changes needed.

#### 7b. Responsive Overrides

**Schema additions**:

Global breakpoints config:
```json
{
  "$breakpoints": {
    "sm": "640px",
    "md": "768px",
    "lg": "1024px",
    "xl": "1280px"
  }
}
```

Per-token responsive overrides:
```json
{
  "spacing": {
    "md": {
      "$value": "1rem",
      "$responsive": {
        "md": "1.25rem",
        "lg": "1.5rem"
      }
    }
  }
}
```

**Generated CSS**:
```css
:root {
  --spacing-md: 1rem;
}

@media (min-width: 768px) {
  :root {
    --spacing-md: 1.25rem;
  }
}

@media (min-width: 1024px) {
  :root {
    --spacing-md: 1.5rem;
  }
}
```

**Files to modify**:
- `pkg/tokens/loader.go` - Extract `$breakpoints` config
- `pkg/tokens/responsive.go` - New file for responsive token handling
- `pkg/generators/tailwind.go` - Generate media query blocks
- `pkg/generators/css.go` - Same for pure CSS output

---

## Testing Strategy

Each feature should include:
1. Unit tests for new functions
2. Integration test in `cmd/tokenctl/integration_test.go`
3. Example in `examples/` directory

---

## Documentation Deliverables

1. **HOWTO.md** - Comprehensive guide to the opinionated system
   - Philosophy: tokens over direct CSS
   - Architecture: brand → semantic → component layers
   - Responsive strategy: fluid-first, overrides when needed
   - LLM integration patterns
   - Migration from existing systems

2. **Update README.md** - Add new features to quick reference

3. **Update TOKENS.md** - Document new schema fields

---

## Implementation Order Rationale

1-3 (Manifests, Descriptions, Pure CSS): Quick wins that expand utility
4 (Search): Enables LLM tool use
5-6 (Components, Layers): Design system completeness
7 (Responsive): Modern CSS integration

This order ensures each phase delivers usable value.
