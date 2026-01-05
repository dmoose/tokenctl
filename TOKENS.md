# Tokctl Semantic Design System: Developer Guide

## Table of Contents
1. [Why Semantic Design Systems Matter](#why-semantic-design-systems-matter)
2. [Understanding the Semantic Approach](#understanding-the-semantic-approach)
3. [Getting Started with Tokctl](#getting-started-with-tokctl)
4. [Working with Semantic Tokens](#working-with-semantic-tokens)
5. [Building Components](#building-components)
6. [Theme Management](#theme-management)
7. [Advanced Patterns](#advanced-patterns)
8. [Best Practices](#best-practices)
9. [Troubleshooting](#troubleshooting)

## Why Semantic Design Systems Matter

### The Problem with Traditional CSS

If you've worked with CSS, you've probably written code like this:

```css
.header {
  background-color: #3b82f6;
  color: #ffffff;
}

.button-primary {
  background-color: #3b82f6;
  color: #ffffff;
}

.link-active {
  color: #3b82f6;
}
```

This approach has several problems:

1. **Magic Numbers**: `#3b82f6` appears everywhere, but what does it represent?
2. **Maintenance Nightmare**: Changing your brand color requires finding and replacing dozens of hex codes
3. **Inconsistency**: Different developers might use slightly different blues (`#3b82f6` vs `#2563eb`)
4. **No Context**: The color value tells you nothing about its purpose or when to use it

### The Utility-First Improvement

Tailwind CSS improved this with utility classes:

```html
<header class="bg-blue-500 text-white">
<button class="bg-blue-500 text-white">
<a class="text-blue-500">
```

This is better because:
- **Consistent Values**: `blue-500` always means the same color
- **Predictable Scale**: `blue-400`, `blue-500`, `blue-600` form a logical progression
- **Rapid Development**: No need to write custom CSS

But problems remain:
- **Arbitrary Names**: Why "blue-500"? What makes it "500"?
- **No Semantic Meaning**: `bg-blue-500` doesn't tell you this is your primary brand color
- **Theme Switching**: Supporting dark mode requires completely different class names

### The Semantic Solution

Semantic design systems solve these problems by using **meaningful names** that describe **purpose**, not appearance:

```html
<header class="bg-primary text-primary-content">
<button class="bg-primary text-primary-content">
<a class="text-primary">
```

Now the code tells a story:
- `bg-primary`: This uses our primary brand color
- `text-primary-content`: This text is designed to be readable on the primary color
- `text-primary`: This text uses the primary color for emphasis

## Understanding the Semantic Approach

### Semantic vs. Arbitrary Naming

| Arbitrary (Traditional) | Semantic (Better) | Why Semantic Wins |
|------------------------|-------------------|-------------------|
| `bg-blue-500` | `bg-primary` | Describes purpose, not appearance |
| `text-gray-900` | `text-base-content` | Adapts to themes automatically |
| `bg-green-100` | `bg-success` | Conveys meaning to developers |
| `border-red-500` | `border-error` | Self-documenting code |

### The Semantic Color System

Tokctl organizes colors into semantic categories:

#### Brand Colors
These represent your brand identity:
- `primary`: Your main brand color (logo, CTAs, key actions)
- `secondary`: Supporting brand color (accents, highlights)
- `accent`: Additional brand color (special elements, decorations)

#### Surface Colors
These create the foundation of your interface:
- `base-100`: Main background color (page background)
- `base-200`: Slightly darker (card backgrounds, subtle elevation)
- `base-300`: Even darker (borders, dividers)
- `base-content`: Text color that works on base colors

#### Semantic State Colors
These communicate status and feedback:
- `success`: Positive actions, confirmations (green family)
- `warning`: Caution, attention needed (yellow/orange family)
- `error`: Problems, destructive actions (red family)
- `info`: Neutral information, tips (blue family)

#### Content Colors
Every background color has a matching content color:
- `primary-content`: Text/icons that work on `primary` background
- `success-content`: Text/icons that work on `success` background
- `base-content`: Text/icons that work on `base` backgrounds

### Why This System Works

1. **Self-Documenting**: `bg-success` immediately tells you this indicates success
2. **Theme-Agnostic**: `primary` can be blue in light mode, purple in dark mode
3. **Consistent Relationships**: `primary-content` always works with `primary`
4. **Scalable**: Add new semantic colors without breaking existing ones

## Getting Started with Tokctl

### Installation and Setup

```bash
# Install the CLI tool
go install github.com/yourusername/tokctl/cmd/tokctl@latest

# Create a new project
mkdir my-app-design-system
cd my-app-design-system

# Initialize with semantic template
tokctl init --template=semantic .
```

This creates a structured token system:

```
tokens/
├── brand.tokens.json      # Your brand identity colors
├── surface.tokens.json    # Background and surface colors
├── semantic.tokens.json   # State colors (success, error, etc.)
├── spacing.tokens.json    # Spacing scale
├── typography.tokens.json # Font scales
└── themes/
    ├── light.tokens.json  # Light theme
    └── dark.tokens.json   # Dark theme
```

### Your First Token File

Let's examine `tokens/brand.tokens.json`:

```json
{
  "brand": {
    "$type": "color",
    "$description": "Core brand identity colors",
    
    "primary": {
      "$description": "Primary brand color - use for main CTAs, logos, key actions",
      "$value": {
        "colorSpace": "srgb",
        "components": [0.2, 0.4, 0.9],
        "hex": "#3366e6"
      }
    },
    "primary-content": {
      "$description": "Text/icon color that works on primary backgrounds",
      "$value": {
        "colorSpace": "srgb",
        "components": [1, 1, 1],
        "hex": "#ffffff"
      }
    },
    
    "secondary": {
      "$description": "Secondary brand color - use for supporting elements",
      "$value": {
        "colorSpace": "srgb",
        "components": [0.8, 0.2, 0.6],
        "hex": "#cc3399"
      }
    },
    "secondary-content": {
      "$description": "Text/icon color that works on secondary backgrounds",
      "$value": "{brand.primary-content}"
    }
  }
}
```

**Key Points:**
- `$type`: Tells Tokctl this is a color token
- `$description`: Documents when and how to use this token
- `$value`: The actual color value in W3C standard format
- `"{brand.primary-content}"`: A reference to another token (DRY principle)

### Generate Your First Output

```bash
# Generate Tailwind CSS theme
tokctl build --format=tailwind --output=./dist/theme.css
```

This creates:

**dist/theme.css**:
```css
@import "tailwindcss";

@theme {
  --color-primary: oklch(0.59 0.21 258.34);
  --color-primary-content: oklch(1 0 0);
  --color-secondary: oklch(0.61 0.24 328.36);
  --color-secondary-content: oklch(1 0 0);
}
```

## Working with Semantic Tokens

### Understanding Token Structure

Every token follows the W3C Design Tokens 2025.10 format:

```json
{
  "token-name": {
    "$type": "color|dimension|typography|...",
    "$value": "the actual value",
    "$description": "when and how to use this token",
    "$deprecated": false
  }
}
```

### Color Tokens

```json
{
  "semantic": {
    "$type": "color",
    
    "success": {
      "$description": "Use for positive feedback, confirmations, completed states",
      "$value": {
        "colorSpace": "srgb",
        "components": [0.1, 0.7, 0.3],
        "hex": "#1ab34d"
      }
    },
    "success-content": {
      "$description": "Text color for success backgrounds",
      "$value": {
        "colorSpace": "srgb",
        "components": [1, 1, 1],
        "hex": "#ffffff"
      }
    }
  }
}
```

**Usage in HTML:**
```html
<div class="bg-success text-success-content p-4 rounded">
  ✓ Your changes have been saved successfully!
</div>
```

### Spacing Tokens

```json
{
  "spacing": {
    "$type": "dimension",
    "$description": "Consistent spacing scale",
    
    "xs": {
      "$description": "Extra small spacing - use for tight layouts",
      "$value": {"value": 4, "unit": "px"}
    },
    "sm": {
      "$description": "Small spacing - use for compact components",
      "$value": {"value": 8, "unit": "px"}
    },
    "md": {
      "$description": "Medium spacing - default for most components",
      "$value": {"value": 16, "unit": "px"}
    },
    "lg": {
      "$description": "Large spacing - use for generous layouts",
      "$value": {"value": 24, "unit": "px"}
    }
  }
}
```

**Usage in HTML:**
```html
<div class="p-md">           <!-- padding: 16px -->
<div class="mb-lg">          <!-- margin-bottom: 24px -->
<div class="gap-sm">         <!-- gap: 8px -->
```

### Typography Tokens

```json
{
  "typography": {
    "$description": "Typography scale and styles",
    
    "heading": {
      "$type": "typography",
      
      "h1": {
        "$description": "Main page headings",
        "$value": {
          "fontFamily": ["Inter", "system-ui", "sans-serif"],
          "fontSize": {"value": 48, "unit": "px"},
          "fontWeight": 700,
          "lineHeight": 1.2,
          "letterSpacing": {"value": -0.5, "unit": "px"}
        }
      },
      "h2": {
        "$description": "Section headings",
        "$value": {
          "fontFamily": ["Inter", "system-ui", "sans-serif"],
          "fontSize": {"value": 36, "unit": "px"},
          "fontWeight": 600,
          "lineHeight": 1.3
        }
      }
    },
    
    "body": {
      "$type": "typography",
      
      "large": {
        "$description": "Large body text, introductions",
        "$value": {
          "fontFamily": ["Inter", "system-ui", "sans-serif"],
          "fontSize": {"value": 18, "unit": "px"},
          "fontWeight": 400,
          "lineHeight": 1.6
        }
      }
    }
  }
}
```

### Token References

Tokens can reference other tokens to maintain consistency:

```json
{
  "components": {
    "button": {
      "$type": "color",
      
      "primary-bg": {
        "$value": "{brand.primary}"
      },
      "primary-text": {
        "$value": "{brand.primary-content}"
      },
      "primary-hover": {
        "$description": "Slightly darker version of primary",
        "$value": {
          "colorSpace": "srgb",
          "components": [0.15, 0.35, 0.8]
        }
      }
    }
  }
}
```

**Benefits of References:**
- **DRY Principle**: Define colors once, use everywhere
- **Automatic Updates**: Change `brand.primary`, all references update
- **Consistency**: Impossible to have mismatched colors

## Building Components

### Component Token Patterns

Organize tokens around your components:

```json
{
  "components": {
    "button": {
      "$description": "Button component tokens",
      
      "variants": {
        "$description": "Different button styles",
        
        "primary": {
          "$type": "color",
          "background": {"$value": "{brand.primary}"},
          "text": {"$value": "{brand.primary-content}"},
          "border": {"$value": "{brand.primary}"},
          "hover": {
            "background": {"$value": "{brand.primary-focus}"}
          }
        },
        
        "secondary": {
          "$type": "color",
          "background": {"$value": "transparent"},
          "text": {"$value": "{brand.primary}"},
          "border": {"$value": "{brand.primary}"},
          "hover": {
            "background": {"$value": "{brand.primary}"},
            "text": {"$value": "{brand.primary-content}"}
          }
        },
        
        "success": {
          "$type": "color",
          "background": {"$value": "{semantic.success}"},
          "text": {"$value": "{semantic.success-content}"},
          "border": {"$value": "{semantic.success}"}
        }
      },
      
      "sizes": {
        "$description": "Button size variations",
        
        "small": {
          "$type": "dimension",
          "padding-x": {"$value": "{spacing.sm}"},
          "padding-y": {"$value": "{spacing.xs}"},
          "font-size": {"$value": {"value": 14, "unit": "px"}},
          "border-radius": {"$value": {"value": 4, "unit": "px"}}
        },
        
        "medium": {
          "$type": "dimension",
          "padding-x": {"$value": "{spacing.md}"},
          "padding-y": {"$value": "{spacing.sm}"},
          "font-size": {"$value": {"value": 16, "unit": "px"}},
          "border-radius": {"$value": {"value": 6, "unit": "px"}}
        },
        
        "large": {
          "$type": "dimension",
          "padding-x": {"$value": "{spacing.lg}"},
          "padding-y": {"$value": "{spacing.md}"},
          "font-size": {"$value": {"value": 18, "unit": "px"}},
          "border-radius": {"$value": {"value": 8, "unit": "px"}}
        }
      }
    }
  }
}
```

### Using Component Tokens in Go

Use the generated CSS variables or utility classes directly:

```html
<button class="bg-primary text-primary-content px-md py-sm rounded-md">
  Click Me
</button>
```

### Creating Reusable Components

```go
// components/button.templ
package components

import "github.com/a-h/templ"

type ButtonVariant string

const (
    ButtonPrimary   ButtonVariant = "primary"
    ButtonSecondary ButtonVariant = "secondary"
    ButtonSuccess   ButtonVariant = "success"
    ButtonError     ButtonVariant = "error"
)

type ButtonSize string

const (
    ButtonSmall  ButtonSize = "small"
    ButtonMedium ButtonSize = "medium"
    ButtonLarge  ButtonSize = "large"
)

templ Button(text string, variant ButtonVariant, size ButtonSize) {
    <button 
        class={
            "font-medium transition-colors focus:outline-none focus:ring-2",
            // Semantic classes
            "btn-" + string(variant),
            "btn-" + string(size),
        }
    >
        {text}
    </button>
}
```

**Usage:**
```go
// In your templ templates
@Button("Save Changes", ButtonSuccess, ButtonMedium)
@Button("Cancel", ButtonSecondary, ButtonMedium)
@Button("Delete", ButtonError, ButtonSmall)
```

## Theme Management

### Understanding Themes

Themes are variations of your design system that change the appearance while maintaining the same semantic structure.

### Creating Theme Variations

**themes/light.tokens.json** (default):
```json
{
  "light": {
    "$description": "Light theme - default appearance",
    
    "surface": {
      "$type": "color",
      
      "base-100": {
        "$description": "Main background - pure white",
        "$value": {
          "colorSpace": "srgb",
          "components": [1, 1, 1],
          "hex": "#ffffff"
        }
      },
      "base-200": {
        "$description": "Card backgrounds - light gray",
        "$value": {
          "colorSpace": "srgb",
          "components": [0.98, 0.98, 0.98],
          "hex": "#fafafa"
        }
      },
      "base-content": {
        "$description": "Text on base backgrounds - dark gray",
        "$value": {
          "colorSpace": "srgb",
          "components": [0.1, 0.1, 0.1],
          "hex": "#1a1a1a"
        }
      }
    }
  }
}
```

**themes/dark.tokens.json**:
```json
{
  "dark": {
    "$extends": "{light}",
    "$description": "Dark theme - inverted for low-light environments",
    
    "surface": {
      "$type": "color",
      
      "base-100": {
        "$description": "Main background - dark gray",
        "$value": {
          "colorSpace": "srgb",
          "components": [0.1, 0.1, 0.1],
          "hex": "#1a1a1a"
        }
      },
      "base-200": {
        "$description": "Card backgrounds - slightly lighter",
        "$value": {
          "colorSpace": "srgb",
          "components": [0.15, 0.15, 0.15],
          "hex": "#262626"
        }
      },
      "base-content": {
        "$description": "Text on base backgrounds - light gray",
        "$value": {
          "colorSpace": "srgb",
          "components": [0.9, 0.9, 0.9],
          "hex": "#e6e6e6"
        }
      }
    },
    
    "brand": {
      "$type": "color",
      
      "primary": {
        "$description": "Slightly brighter primary for dark backgrounds",
        "$value": {
          "colorSpace": "srgb",
          "components": [0.3, 0.5, 0.95],
          "hex": "#4d7fff"
        }
      }
    }
  }
}
```

### Theme Inheritance with $extends

The `$extends` property creates theme inheritance:

```json
{
  "dark": {
    "$extends": "{light}",
    // Only override what changes in dark mode
    // Everything else inherits from light theme
  }
}
```

This means:
- Dark theme starts with all light theme values
- Only specified tokens are overridden
- Unspecified tokens (like `success`, `warning`) remain the same

### Building Multi-Theme Output

```bash
# Build all themes
tokctl build --format=tailwind --output=./dist/

# This generates:
# ./dist/light.css
# ./dist/dark.css
# ./dist/theme-variables.css (shared variables)
```

**Generated CSS structure:**
```css
/* theme-variables.css - shared base */
@import "tailwindcss";

@theme inline {
  --color-primary: var(--theme-primary);
  --color-base-100: var(--theme-base-100);
  --color-base-content: var(--theme-base-content);
}

/* light.css */
:root {
  --theme-primary: oklch(0.59 0.21 258.34);
  --theme-base-100: oklch(1 0 0);
  --theme-base-content: oklch(0.1 0 0);
}

/* dark.css */
.dark {
  --theme-primary: oklch(0.69 0.21 258.34);
  --theme-base-100: oklch(0.1 0 0);
  --theme-base-content: oklch(0.9 0 0);
}
```

### Theme Switching in Your App

```html
<!DOCTYPE html>
<html class="light" data-theme="light">
<head>
    <link href="/dist/theme-variables.css" rel="stylesheet">
    <link href="/dist/light.css" rel="stylesheet">
    <link href="/dist/dark.css" rel="stylesheet">
</head>
<body class="bg-base-100 text-base-content">
    <button onclick="toggleTheme()">Toggle Theme</button>
    
    <script>
        function toggleTheme() {
            const html = document.documentElement;
            const currentTheme = html.getAttribute('data-theme');
            const newTheme = currentTheme === 'light' ? 'dark' : 'light';
            
            html.className = newTheme;
            html.setAttribute('data-theme', newTheme);
            
            localStorage.setItem('theme', newTheme);
        }
        
        // Load saved theme
        const savedTheme = localStorage.getItem('theme') || 'light';
        document.documentElement.className = savedTheme;
        document.documentElement.setAttribute('data-theme', savedTheme);
    </script>
</body>
</html>
```

## Advanced Patterns

### Composite Tokens

Some design elements require multiple values:

```json
{
  "effects": {
    "$description": "Visual effects and shadows",
    
    "shadow": {
      "$type": "shadow",
      
      "card": {
        "$description": "Subtle shadow for cards",
        "$value": {
          "color": {
            "colorSpace": "srgb",
            "components": [0, 0, 0],
            "alpha": 0.1
          },
          "offsetX": {"value": 0, "unit": "px"},
          "offsetY": {"value": 4, "unit": "px"},
          "blur": {"value": 8, "unit": "px"},
          "spread": {"value": 0, "unit": "px"}
        }
      },
      
      "focus": {
        "$description": "Focus ring for interactive elements",
        "$value": {
          "color": {
            "colorSpace": "srgb", 
            "components": [0.2, 0.4, 0.9],
            "alpha": 0.5
          },
          "offsetX": {"value": 0, "unit": "px"},
          "offsetY": {"value": 0, "unit": "px"},
          "blur": {"value": 0, "unit": "px"},
          "spread": {"value": 2, "unit": "px"}
        }
      }
    }
  }
}
```

### Border Tokens

```json
{
  "borders": {
    "$description": "Border styles and widths",
    
    "default": {
      "$type": "border",
      "$value": {
        "color": "{surface.base-300}",
        "width": {"value": 1, "unit": "px"},
        "style": "solid"
      }
    },
    
    "focus": {
      "$type": "border", 
      "$value": {
        "color": "{brand.primary}",
        "width": {"value": 2, "unit": "px"},
        "style": "solid"
      }
    }
  }
}
```

### Animation Tokens

```json
{
  "motion": {
    "$description": "Animation and transition tokens",
    
    "duration": {
      "$type": "duration",
      
      "fast": {
        "$description": "Quick transitions",
        "$value": {"value": 150, "unit": "ms"}
      },
      "normal": {
        "$description": "Standard transitions",
        "$value": {"value": 300, "unit": "ms"}
      },
      "slow": {
        "$description": "Deliberate transitions",
        "$value": {"value": 500, "unit": "ms"}
      }
    },
    
    "easing": {
      "$type": "cubicBezier",
      
      "ease-out": {
        "$description": "Natural deceleration",
        "$value": [0, 0, 0.2, 1]
      },
      "ease-in-out": {
        "$description": "Smooth acceleration and deceleration", 
        "$value": [0.4, 0, 0.2, 1]
      }
    }
  }
}
```

### Responsive Tokens

```json
{
  "breakpoints": {
    "$description": "Responsive breakpoints",
    
    "mobile": {
      "$type": "dimension",
      "$value": {"value": 640, "unit": "px"}
    },
    "tablet": {
      "$type": "dimension", 
      "$value": {"value": 768, "unit": "px"}
    },
    "desktop": {
      "$type": "dimension",
      "$value": {"value": 1024, "unit": "px"}
    }
  }
}
```

## Best Practices

### 1. Naming Conventions

**DO:**
- Use semantic names: `primary`, `success`, `base-content`
- Be consistent: always pair colors with `-content` variants
- Use hierarchical organization: `spacing.sm`, `typography.heading.h1`

**DON'T:**
- Use appearance-based names: `blue-500`, `dark-gray`
- Mix naming patterns: `primaryColor` vs `primary-color`
- Create orphaned tokens: `primary` without `primary-content`

### 2. Token Organization

**Group by purpose:**
```
tokens/
├── brand.tokens.json       # Brand identity
├── surface.tokens.json     # Backgrounds and surfaces
├── semantic.tokens.json    # State colors
├── typography.tokens.json  # Text styles
├── spacing.tokens.json     # Layout spacing
├── motion.tokens.json      # Animations
└── components/
    ├── button.tokens.json  # Button-specific tokens
    └── card.tokens.json    # Card-specific tokens
```

### 3. Reference Relationships

**Create logical relationships:**
```json
{
  "brand": {
    "primary": {"$value": "#3366e6"},
    "primary-content": {"$value": "#ffffff"}
  },
  "components": {
    "button": {
      "primary-bg": {"$value": "{brand.primary}"},
      "primary-text": {"$value": "{brand.primary-content}"}
    }
  }
}
```

### 4. Documentation

**Always include descriptions:**
```json
{
  "success": {
    "$description": "Use for positive feedback, confirmations, completed states. Examples: form success messages, completed progress indicators, positive status badges.",
    "$value": {"colorSpace": "srgb", "components": [0.1, 0.7, 0.3]}
  }
}
```

### 5. Validation

**Use Tokctl's built-in validation:**
```bash
# Check accessibility compliance
tokctl validate --wcag=AA

# Validate token structure
tokctl validate --check-references --check-cycles

# Check naming conventions
tokctl validate --naming-rules=semantic
```

## Troubleshooting

### Common Issues

#### 1. "Token reference not found"

**Error:** `Reference {brand.primary} not found`

**Solution:** Check that the referenced token exists and the path is correct:
```json
// ❌ Wrong
{"$value": "{brand.primary}"}

// ✅ Correct (if token is in same file)
{"$value": "{brand.primary}"}

// ✅ Correct (if token is in different file)
// Make sure brand.tokens.json defines brand.primary
```

#### 2. "Circular reference detected"

**Error:** `Circular reference: brand.primary -> brand.secondary -> brand.primary`

**Solution:** Remove the circular dependency:
```json
// ❌ Circular
{
  "primary": {"$value": "{brand.secondary}"},
  "secondary": {"$value": "{brand.primary}"}
}

// ✅ Fixed
{
  "primary": {"$value": "#3366e6"},
  "secondary": {"$value": "#cc3399"}
}
```

#### 3. "Invalid color format"

**Error:** `Invalid color value for token brand.primary`

**Solution:** Use proper W3C color format:
```json
// ❌ Wrong
{"$value": "#3366e6"}

// ✅ Correct
{
  "$value": {
    "colorSpace": "srgb",
    "components": [0.2, 0.4, 0.9],
    "hex": "#3366e6"
  }
}
```

#### 4. "Missing content color"

**Warning:** `Brand color 'primary' missing content color 'primary-content'`

**Solution:** Always create content colors for background colors:
```json
{
  "primary": {
    "$value": {"colorSpace": "srgb", "components": [0.2, 0.4, 0.9]}
  },
  "primary-content": {
    "$value": {"colorSpace": "srgb", "components": [1, 1, 1]}
  }
}
```

### Debugging Tips

#### 1. Validate frequently
```bash
tokctl validate
```
Run validation after making changes to catch issues early. Errors now include source file names:
```
[Error] color.primary [tokens/brand/colors.json]: reference not found: color.nonexistent
```

#### 2. Check generated output
```bash
tokctl build --format=tailwind --output=./debug
cat ./debug/tokens.css
```
Examine the generated CSS to understand how tokens are being processed.

#### 3. Use the examples
```bash
# Build and examine working examples
tokctl build examples/themes --output=./dist
tokctl build examples/components --output=./dist
```
The `examples/` directory contains working token systems demonstrating all features.

#### 4. Run the demo workflow
```bash
make demo
```
Runs a complete init → validate → build workflow to verify everything works.

### Getting Help

1. **Check the documentation**: Most issues are covered in this guide
2. **Validate your tokens**: Run `tokctl validate` to catch common problems
3. **Check examples**: Look at the `examples/` directory for working patterns
4. **Run the demo**: Use `make demo` to see a complete workflow in action
5. **Examine test fixtures**: The `testdata/` directory contains valid and invalid examples

---

This semantic design system approach will transform how you think about styling. Instead of managing hundreds of arbitrary color values, you'll work with a small set of meaningful, purpose-driven tokens that automatically adapt to themes and maintain consistency across your entire application.

The key is to think semantically: **what is this element's purpose?** rather than **what color should this be?** This mindset shift will make your code more maintainable, your designs more consistent, and your development process more efficient.
