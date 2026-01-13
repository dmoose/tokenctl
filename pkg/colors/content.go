// tokenctl/pkg/colors/content.go

package colors

import (
	"github.com/lucasb-eyer/go-colorful"
)

// ContentColor generates an accessible foreground/content color for a given background
// This implements WCAG AA compliance (4.5:1 minimum contrast ratio for normal text)
//
// The algorithm:
// 1. Try white - most backgrounds work well with white text
// 2. Try black - for light backgrounds
// 3. If neither works (rare edge case), generate an optimal color by adjusting lightness
func ContentColor(background Color) Color {
	white := White()
	black := Black()

	// Check white first (common case for brand colors)
	if ContrastRatio(background, white) >= WCAGAANormal {
		return white
	}

	// Check black
	if ContrastRatio(background, black) >= WCAGAANormal {
		return black
	}

	// Edge case: neither pure white nor black provides sufficient contrast
	// This can happen with mid-tone colors around 50% luminance
	// Generate a color with adjusted lightness that maintains some relationship to the background

	l, c, h := background.OkLch()

	// Determine direction: go opposite of background lightness
	var targetL float64
	if l > 0.5 {
		// Dark content for light backgrounds
		// Reduce chroma to ensure readability
		targetL = 0.15
	} else {
		// Light content for dark backgrounds
		targetL = 0.95
	}

	// Reduce chroma significantly for content colors to ensure readability
	// Content colors should be more neutral than their backgrounds
	contentChroma := c * 0.15

	contentColor := FromOkLch(targetL, contentChroma, h)

	// Verify contrast and adjust if needed
	if ContrastRatio(background, contentColor) < WCAGAANormal {
		// Fall back to pure black or white, whichever is closer to our target
		if targetL > 0.5 {
			return white
		}
		return black
	}

	return contentColor.Clamped()
}

// ContentColorWithRatio generates a content color that meets a specific contrast ratio
// ratio: target contrast ratio (e.g., 4.5 for WCAG AA, 7.0 for WCAG AAA)
func ContentColorWithRatio(background Color, ratio float64) Color {
	white := White()
	black := Black()

	// Check white first
	if ContrastRatio(background, white) >= ratio {
		return white
	}

	// Check black
	if ContrastRatio(background, black) >= ratio {
		return black
	}

	// Neither pure white nor black achieves the ratio
	// This is rare - typically means the background is mid-tone
	// Try to find an intermediate color, but if we can't achieve the ratio,
	// return whichever of black/white gives better contrast
	l, c, h := background.OkLch()

	// For very high contrast requirements, reduce chroma to near zero
	contentChroma := c * 0.05

	// Search full range, going away from background lightness
	var minL, maxL float64
	var bestColor Color
	var bestRatio float64

	if l > 0.5 {
		// Light background: search dark range
		minL, maxL = 0.0, 0.3
	} else {
		// Dark background: search light range
		minL, maxL = 0.7, 1.0
	}

	// Binary search for optimal lightness
	for range 25 {
		midL := (minL + maxL) / 2
		testColor := FromOkLch(midL, contentChroma, h).Clamped()
		testRatio := ContrastRatio(background, testColor)

		if testRatio > bestRatio {
			bestRatio = testRatio
			bestColor = testColor
		}

		if testRatio >= ratio && testRatio <= ratio+1.0 {
			return testColor
		}

		// Adjust search based on whether we need more contrast
		if l > 0.5 {
			// Light background, searching dark range: lower L = more contrast
			if testRatio < ratio {
				maxL = midL
			} else {
				minL = midL
			}
		} else {
			// Dark background, searching light range: higher L = more contrast
			if testRatio < ratio {
				minL = midL
			} else {
				maxL = midL
			}
		}
	}

	// If we found something better than white/black, use it
	if bestRatio >= ratio {
		return bestColor
	}

	// Fall back to whichever of black/white gives better contrast
	whiteRatio := ContrastRatio(background, white)
	blackRatio := ContrastRatio(background, black)
	if whiteRatio >= blackRatio {
		return white
	}
	return black
}

// GenerateContentPair generates both a base color and its content color from a single input
// Useful for creating color pairs like primary/primary-content
func GenerateContentPair(base Color) (Color, Color) {
	return base, ContentColor(base)
}

// ContentColorPreserveHue generates a content color that preserves the hue of the background
// but adjusts lightness and reduces chroma for readability
// This creates a more harmonious pairing than pure black/white
func ContentColorPreserveHue(background Color) Color {
	l, c, h := background.OkLch()

	// Determine target lightness (opposite of background)
	var targetL float64
	if l > 0.6 {
		targetL = 0.2
	} else if l < 0.4 {
		targetL = 0.9
	} else {
		// Mid-tone: use whichever direction gives more contrast
		darkL := 0.15
		lightL := 0.9
		darkColor := FromOkLch(darkL, c*0.1, h)
		lightColor := FromOkLch(lightL, c*0.1, h)

		if ContrastRatio(background, darkColor) > ContrastRatio(background, lightColor) {
			targetL = darkL
		} else {
			targetL = lightL
		}
	}

	// Reduce chroma significantly for readability
	contentChroma := c * 0.12

	content := FromOkLch(targetL, contentChroma, h)

	// Verify WCAG compliance
	if ContrastRatio(background, content) < WCAGAANormal {
		// Fall back to optimal black/white
		return OptimalTextColor(background)
	}

	return content.Clamped()
}

// DaisyContentColor generates content colors in the style of DaisyUI
// DaisyUI typically uses very light or very dark colors with minimal chroma
func DaisyContentColor(background Color) Color {
	l, c, h := background.OkLch()

	var targetL, targetC float64

	if l > 0.5 {
		// Light background: dark content
		targetL = l * 0.2  // Very dark
		targetC = c * 0.15 // Minimal chroma
	} else {
		// Dark background: light content
		targetL = l + (1-l)*0.8 // Very light
		targetC = c * 0.15      // Minimal chroma
	}

	content := FromOkLch(targetL, targetC, h)

	// Ensure WCAG compliance
	if ContrastRatio(background, content) < WCAGAANormal {
		return OptimalTextColor(background)
	}

	return content.Clamped()
}

// CreateFromColorful wraps a go-colorful.Color for use with content generation
func CreateFromColorful(c colorful.Color) Color {
	return FromColorful(c, FormatOKLCH)
}
