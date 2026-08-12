# tokenctl

[![Go](https://github.com/dmoose/tokenctl/actions/workflows/go.yml/badge.svg)](https://github.com/dmoose/tokenctl/actions/workflows/go.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/dmoose/tokenctl.svg)](https://pkg.go.dev/github.com/dmoose/tokenctl)
[![Go Report Card](https://goreportcard.com/badge/github.com/dmoose/tokenctl)](https://goreportcard.com/report/github.com/dmoose/tokenctl)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

A W3C Design Tokens CLI that generates CSS from JSON token definitions. Define your design system once, output to Tailwind 4, pure CSS, or JSON manifests.

## Key Features

- **W3C Compliant**: Uses the preview standard [W3C Design Token Format](https://tr.designtokens.org/format/)
- **Tailwind 4 Ready**: Generates modern `@theme` configurations with `@layer` support
- **Pure CSS Output**: Generate CSS without Tailwind dependency (`--format=css`)
- **Reference Resolution**: Deep referencing (`{color.brand.primary}`) with cycle detection
- **Theme Inheritance**: `$extends` for theme variations that inherit from parent themes
- **Computed Values**: `contrast()`, `darken()`, `lighten()`, `shade()`, `calc()` expressions
- **Scale Expansion**: `$scale` generates size variants automatically (xs, sm, md, lg, xl)
- **CSS @property**: `$property` field generates typed CSS custom properties for animations
- **CSS @keyframes**: Define animations in tokens, output to CSS `@keyframes` blocks
- **Responsive Tokens**: `$breakpoints` and `$responsive` for media query generation
- **Layer Validation**: `--strict-layers` enforces brand → semantic → component architecture
- **Token Search**: CLI search by name, type, or category
- **LLM Manifests**: Category-scoped JSON manifests for context-efficient LLM usage
- **Rich Metadata**: `$description`, `$usage`, `$avoid` fields for documentation
- **Component Composition**: `$contains`, `$requires` for component relationships
- **Component States**: `states` for semantic conditions (error, disabled, loading)
- **Container Queries**: `$container` for responsive component behavior within containers
- **Constraint Validation**: `$min`/`$max` bounds checking on dimension and number tokens
- **Type Validation**: Validates colors, dimensions, numbers, fontFamily, effect, duration
- **Source Tracking**: Validation errors include source file paths

## Installation

```bash
go install github.com/dmoose/tokenctl/cmd/tokenctl@latest
```

## Quick Start

### 1. Initialize a System

```bash
tokenctl init my-design-system
```

Creates:
```
my-design-system/
├── tokens/
│   ├── brand/colors.json
│   ├── surface/
│   ├── semantic/status.json
│   ├── typography/
│   ├── spacing/scale.json
│   └── themes/
```

### 2. Define Tokens

**tokens/brand/colors.json:**
```json
{
  "color": {
    "$type": "color",
    "primary": { "$value": "oklch(49.12% 0.309 275.75)" },
    "primary-content": { "$value": "contrast({color.primary})" },
    "secondary": { "$value": "#8b5cf6" }
  }
}
```

### 3. Create Theme Variations

**tokens/themes/dark.json:**
```json
{
  "$extends": "light",
  "color": {
    "primary": { "$value": "oklch(65% 0.2 275)" }
  }
}
```

### 4. Build

```bash
tokenctl build my-design-system --output=./dist
```

**Output (dist/tokens.css):**
```css
@import "tailwindcss";

@theme {
  --color-primary: oklch(49.12% 0.309 275.75);
  --color-primary-content: oklch(100% 0 0);
  --color-secondary: #8b5cf6;
}

@layer base {
  [data-theme="dark"] {
    --color-primary: oklch(65% 0.2 275);
  }
}
```

### 5. Multi-Directory Merge

Combine a base component library with project-specific extensions:

```bash
tokenctl build ./base-components ./dashboard-ext --output=./dist
tokenctl validate ./base-components ./dashboard-ext
```

Directories merge left-to-right: later directories extend or override earlier ones. See [MERGE.md](MERGE.md) for details.

## Token Features

### References

```json
{
  "button-bg": { "$value": "{color.primary}" }
}
```

### Computed Colors

```json
{
  "primary-content": { "$value": "contrast({color.primary})" },
  "primary-hover": { "$value": "darken({color.primary}, 10%)" },
  "base-200": { "$value": "shade({color.base-100}, 1)" }
}
```

### Scale Expansion

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

Generates: `--size-field`, `--size-field-xs`, `--size-field-sm`, etc.

### CSS @property

```json
{
  "color": {
    "primary": {
      "$value": "oklch(49% 0.3 275)",
      "$property": true
    }
  }
}
```

Enables animated theme transitions.

### Constraints

```json
{
  "size": {
    "field": {
      "$value": "2.5rem",
      "$min": "1rem",
      "$max": "5rem"
    }
  }
}
```

## Commands

```bash
tokenctl init [dir]                    # Initialize token system

tokenctl build [dir...]                # Build artifacts (multi-dir merge)
  --format=tailwind                  # Tailwind 4 CSS (default)
  --format=css                       # Pure CSS (no Tailwind import)
  --format=catalog                   # Full JSON catalog
  --format=manifest:CATEGORY         # Category-scoped manifest
  --output=<dir>                     # Output directory (default: dist)
  --customizable-only                # Only tokens marked $customizable: true
  --strict-unknown-keys              # Fail on input tokenctl doesn't consume

tokenctl derive                        # Derive a theme from a preset or controls
  --preset=<name>                    # Built-in preset (see --list)
  --from-hex=<#rrggbb>               # Derive hue and chroma from a brand colour
  --hue=<0-360> --chroma=<n>         # Explicit OKLCH controls
  --dark                             # Dark-mode variant
  --tint=<0-100>                     # How much hue reaches the neutrals
  --saturation=<0-150>               # Global chroma multiplier (100 = normal)
  --type=<key>                       # Typography system (see --list)
  --density=<75-130>                 # Dimension scale (100 = default)
  --format=json|css                  # Token document (default) or CSS block
  --output=<file>                    # Write to a file instead of stdout
  --list                             # Show presets, typography systems, ranges

tokenctl validate [dir...]             # Validate tokens (multi-dir merge)
  --strict-layers                    # Enforce layer reference rules
  --strict-unknown-keys              # Treat unconsumed keys as errors

tokenctl search [query]                # Search tokens
  --type=<type>                      # Filter by type (color, dimension, etc.)
  --category=<cat>                   # Filter by category
  --dir=<dir>                        # Token directory (default: .)

tokenctl version                       # Print version information
```

## Theme Derivation

`tokenctl derive` expands a handful of controls into a full semantic token
set — every colour token plus the typography and density artifacts:

```bash
tokenctl derive --preset=teal --output=tokens/semantic.json
tokenctl derive --preset=teal --dark --format=css
tokenctl derive --from-hex=#3b6de0 --tint=45 --density=115
```

The default JSON output is a W3C Design Tokens document that `tokenctl
build` consumes directly, and the variables it produces round-trip: the
CSS from `derive --format=css` and the CSS from building its JSON are
identical.

The derivation math is a port of the retired tokenctl-extension's
`theme-engine.ts`. Golden fixtures in `testdata/derive/goldens` are
generated by executing that original TypeScript, and the Go
implementation is required to match them exactly — no tolerance. See
`tools/derive-goldens/regenerate.sh`; regenerating is a deliberate act
and is never run by `go test`.

Density scales the dimension tokens by `density/100`. Where a site
persists that scale is out of scope for this command.

## Unconsumed Input

tokenctl reads a fixed vocabulary. A key outside it — a component
sub-block named `ratios` where the schema says `variants`, a `$selector`
that nothing reads, a `states` entry with no `$class` — parses and
validates cleanly, then contributes nothing to the output. `build` and
`validate` now name every such key and the file it came from; pass
`--strict-unknown-keys` to make it a failure instead of a warning.
Keys beginning with `//` are treated as in-file comments and never
reported.

## Catalog Format (v3.0)

The `--format=catalog` option generates a JSON catalog for external tool
integration — the shape studio's token browser and any other tool that
needs to know what the stylesheet contains reads.

**v3.0 is a shape change.** Component definitions previously serialized
to `{"$class": …}` and nothing else, and component-level state classes
never appeared at all; both are fixed below. `meta.generated_at` is now
opt-in, so the same tokens produce byte-identical output.

```json
{
  "meta": {
    "version": "3.0",
    "tokenctl_version": "v1.3.0"
  },
  "tokens": {
    "color.primary": "#3b82f6",
    "spacing.sm": "0.5rem"
  },
  "components": {
    "input": {
      "classes": ["input", "input-ghost", "input-sm", "input-error"],
      "definitions": {
        "input": {
          "$class": "input",
          "properties": { "display": "block" },
          "states": { "&:focus": { "outline": "2px solid var(--color-ring)" } }
        },
        "input-error": {
          "$class": "input-error",
          "properties": { "border-color": "{color.error}" }
        }
      }
    }
  },
  "themes": {
    "light": {
      "extends": null,
      "tokens": { "color.primary": "#60a5fa" },
      "diff": { "color.primary": "#60a5fa" }
    },
    "dark": {
      "extends": "light",
      "description": "Dark theme extends light",
      "tokens": { "color.primary": "#1e40af" },
      "diff": { "color.primary": "#1e40af" }
    }
  }
}
```

| Field | Description |
|-------|-------------|
| `meta.version` | Catalog schema version (semver) |
| `meta.generated_at` | Present only with `--generated-at`; see below |
| `meta.tokenctl_version` | Version of the binary that generated this catalog |
| `tokens` | Flattened, resolved base tokens |
| `components.<name>.classes` | Every class the stylesheet defines for this component: base, variants, sizes, and states, sorted |
| `components.<name>.definitions` | One entry per class, keyed by class name |
| `definitions.<class>.properties` | The class's CSS properties. Omitted when empty. |
| `definitions.<class>.states` | Pseudo-selector blocks (`&:hover`, `&:focus`), each a property map. Omitted when empty. |
| `themes.<name>.extends` | Parent theme name (null if extends base) |
| `themes.<name>.description` | Theme description from `$description` field |
| `themes.<name>.tokens` | Fully resolved token values for this theme |
| `themes.<name>.diff` | Only tokens that differ from parent/base |

### What changed in v3.0

**Definitions carry their styling.** `properties` and `states` are new.
Before, every definition serialized to `{"$class": "btn-primary"}` — a
consumer asking what a class does got its name back and nothing else.

The definition shape is deliberately *not* the authored shape. Tokens
are authored with properties inline beside `$class` and states keyed by
a leading `&` or `:`; the catalog splits them into named blocks so a
consumer does not have to repeat that sniffing (and so a property
literally named `$class` cannot collide with the class name). The
catalog is an export, never re-read by tokenctl.

**Component-level states reach `classes` and `definitions`.** The CSS
generator emits a class for every entry in a component's `states` map
(`.input-error`, `.tabs-trigger-active`). The catalog listed none of
them, so a tool asking about `.input-error` was told no such class
exists while the stylesheet in front of it defined one.

**Output is reproducible.** `meta.generated_at` was `time.Now()`, and
the class list was built by ranging Go maps, whose iteration order is
randomized. The same tokens now produce byte-identical output, so the
catalog can be diffed, hashed and cached. A caller that genuinely wants
a timestamp asks for one:

```bash
tokenctl build ./tokens --format=catalog                            # reproducible, no stamp
tokenctl build ./tokens --format=catalog --generated-at=now          # current UTC time
tokenctl build ./tokens --format=catalog --generated-at=2026-08-10   # a literal string
```

**`meta.tokenctl_version` tracks the binary.** It was a hand-maintained
const that read `1.2.0` whatever the binary was. It now comes from
`pkg/version`: an ldflags-injected version if the build set one,
otherwise the binary's own module version or VCS revision.

## Examples

```bash
tokenctl build examples/computed --output=dist/computed
tokenctl build examples/themes --output=dist/themes
tokenctl build examples/validation --output=dist/validation
tokenctl build examples/daisyui --output=dist/daisyui
```

See [examples/README.md](examples/README.md) for details.

## Token Types

| Type | Description | Example |
|------|-------------|---------|
| `color` | CSS colors | `#3b82f6`, `oklch(49% 0.3 275)` |
| `dimension` | Length values | `1rem`, `16px` |
| `number` | Numeric values | `400`, `0.5` |
| `fontFamily` | Font stacks | `["Inter", "sans-serif"]` |
| `duration` | Time values | `150ms`, `0.3s` |
| `effect` | Binary toggle | `0` or `1` |
| `component` | Component definition | See TOKENS.md |

## Documentation

- **Getting started** — You're here. Quick start above, examples below.
- **[HOWTO.md](HOWTO.md)** — Philosophy, architecture, migration, LLM integration
- **[TOKENS.md](TOKENS.md)** — Token format reference: types, expressions, components, themes
- **[MERGE.md](MERGE.md)** — Multi-directory merge for shared/multi-team systems
- **[CSS_PATTERNS.md](CSS_PATTERNS.md)** — CSS patterns for consuming tokenctl output (accessibility, typography, interaction states)
- **[examples/](examples/)** — 7 runnable examples
- **[testdata/](testdata/)** — Test fixtures and golden files

## Development

```bash
make build          # Build binary
make test           # Run tests
make coverage       # Coverage report
make demo           # Full workflow demo
make examples       # Build all examples
make help           # All targets
```

## License

Apache 2.0 - See [LICENSE](LICENSE) for details.
