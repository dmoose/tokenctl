// tokenctl/pkg/colors/contrast.go

package colors

import (
	"math"
)

// WCAG contrast ratio thresholds
const (
	WCAGAANormal    = 4.5 // AA compliance for normal text
	WCAGAALarge     = 3.0 // AA compliance for large text (18pt+ or 14pt bold)
	WCAGAAANormal   = 7.0 // AAA compliance for normal text
	WCAGAAALarge    = 4.5 // AAA compliance for large text
	WCAGUIComponent = 3.0 // AA compliance for UI components and graphical objects
)

// RelativeLuminance calculates the relative luminance of a color
// according to WCAG 2.1 definition
// Returns a value between 0 (darkest black) and 1 (lightest white)
// Formula: L = 0.2126 * R + 0.7152 * G + 0.0722 * B
// where R, G, B are linearized sRGB values
func RelativeLuminance(c Color) float64 {
	// Get linear RGB values (go-colorful provides this)
	r, g, b := c.LinearRgb()

	// WCAG luminance formula
	return 0.2126*r + 0.7152*g + 0.0722*b
}

// ContrastRatio calculates the WCAG contrast ratio between two colors
// Returns a value between 1.0 (identical) and 21.0 (black/white)
// Formula: (L1 + 0.05) / (L2 + 0.05) where L1 is the lighter color
func ContrastRatio(c1, c2 Color) float64 {
	l1 := RelativeLuminance(c1)
	l2 := RelativeLuminance(c2)

	// Ensure l1 is the lighter luminance
	if l1 < l2 {
		l1, l2 = l2, l1
	}

	return (l1 + 0.05) / (l2 + 0.05)
}

// MeetsWCAG checks if the contrast between two colors meets WCAG accessibility standards
// level: "AA" or "AAA"
// largeText: true for large text (18pt+ or 14pt bold)
func MeetsWCAG(c1, c2 Color, level string, largeText bool) bool {
	ratio := ContrastRatio(c1, c2)

	switch level {
	case "AAA":
		if largeText {
			return ratio >= WCAGAAALarge
		}
		return ratio >= WCAGAAANormal
	case "AA":
		fallthrough
	default:
		if largeText {
			return ratio >= WCAGAALarge
		}
		return ratio >= WCAGAANormal
	}
}

// MeetsWCAG_AA is a convenience function that checks for AA compliance with normal text
func MeetsWCAG_AA(c1, c2 Color) bool {
	return MeetsWCAG(c1, c2, "AA", false)
}

// MeetsWCAG_AAA is a convenience function that checks for AAA compliance with normal text
func MeetsWCAG_AAA(c1, c2 Color) bool {
	return MeetsWCAG(c1, c2, "AAA", false)
}

// SufficientContrast checks if two colors have at least the specified contrast ratio
func SufficientContrast(c1, c2 Color, minRatio float64) bool {
	return ContrastRatio(c1, c2) >= minRatio
}

// ContrastLevel returns a human-readable description of the contrast level
func ContrastLevel(c1, c2 Color) string {
	ratio := ContrastRatio(c1, c2)

	switch {
	case ratio >= WCAGAAANormal:
		return "AAA"
	case ratio >= WCAGAANormal:
		return "AA"
	case ratio >= WCAGAALarge:
		return "AA Large"
	default:
		return "Fail"
	}
}

// OptimalTextColor returns either black or white, whichever provides better contrast
// against the given background color
func OptimalTextColor(background Color) Color {
	white := White()
	black := Black()

	whiteContrast := ContrastRatio(background, white)
	blackContrast := ContrastRatio(background, black)

	if whiteContrast >= blackContrast {
		return white
	}
	return black
}

// AdjustLightnessForContrast adjusts a color's lightness to achieve the target contrast
// against a reference color. Returns the adjusted color.
// direction: 1 for lighter, -1 for darker, 0 for auto (away from reference)
func AdjustLightnessForContrast(c Color, reference Color, targetRatio float64, direction int) Color {
	// Get OKLCH components for perceptually uniform adjustments
	l, ch, h := c.OkLch()
	refL, _, _ := reference.OkLch()

	// Determine direction if auto
	if direction == 0 {
		if refL > 0.5 {
			direction = -1 // Make darker if reference is light
		} else {
			direction = 1 // Make lighter if reference is dark
		}
	}

	// Binary search for the right lightness
	minL, maxL := 0.0, 1.0
	if direction > 0 {
		minL = l
	} else {
		maxL = l
	}

	// Iterate to find optimal lightness
	for range 20 {
		testL := (minL + maxL) / 2
		testColor := FromOkLch(testL, ch, h)

		ratio := ContrastRatio(testColor, reference)

		if math.Abs(ratio-targetRatio) < 0.1 {
			return testColor.Clamped()
		}

		// Adjust search range based on whether we need more or less contrast
		needMoreContrast := ratio < targetRatio

		if direction > 0 {
			// Going lighter increases contrast against dark backgrounds
			if needMoreContrast {
				minL = testL
			} else {
				maxL = testL
			}
		} else {
			// Going darker increases contrast against light backgrounds
			if needMoreContrast {
				maxL = testL
			} else {
				minL = testL
			}
		}
	}

	// Return best approximation
	return FromOkLch((minL+maxL)/2, ch, h).Clamped()
}
