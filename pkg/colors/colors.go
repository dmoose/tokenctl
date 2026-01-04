// tokenctl/pkg/colors/colors.go

package colors

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/lucasb-eyer/go-colorful"
)

// Color wraps go-colorful.Color with additional utilities for CSS color handling
type Color struct {
	colorful.Color
	originalFormat string // preserve input format for round-trip output
}

// Format constants for output
const (
	FormatHex   = "hex"
	FormatRGB   = "rgb"
	FormatHSL   = "hsl"
	FormatOKLCH = "oklch"
)

// Parse accepts any CSS color format and returns a normalized Color
// Supported formats:
//   - Hex: #fff, #ffffff, #ffffffff (with alpha)
//   - RGB: rgb(255, 128, 0), rgb(255 128 0), rgba(255, 128, 0, 0.5)
//   - HSL: hsl(180, 50%, 50%), hsla(180, 50%, 50%, 0.5)
//   - OKLCH: oklch(0.5 0.2 180), oklch(50% 0.2 180)
//   - Named colors: red, blue, etc. (limited set)
func Parse(input string) (Color, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return Color{}, fmt.Errorf("empty color string")
	}

	// Detect format and parse accordingly
	lower := strings.ToLower(input)

	// Hex format
	if strings.HasPrefix(input, "#") {
		c, err := parseHex(input)
		if err != nil {
			return Color{}, err
		}
		return Color{Color: c, originalFormat: FormatHex}, nil
	}

	// RGB/RGBA format
	if strings.HasPrefix(lower, "rgb") {
		c, err := parseRGB(input)
		if err != nil {
			return Color{}, err
		}
		return Color{Color: c, originalFormat: FormatRGB}, nil
	}

	// HSL/HSLA format
	if strings.HasPrefix(lower, "hsl") {
		c, err := parseHSL(input)
		if err != nil {
			return Color{}, err
		}
		return Color{Color: c, originalFormat: FormatHSL}, nil
	}

	// OKLCH format
	if strings.HasPrefix(lower, "oklch") {
		c, err := parseOKLCH(input)
		if err != nil {
			return Color{}, err
		}
		return Color{Color: c, originalFormat: FormatOKLCH}, nil
	}

	// Try named colors
	if c, ok := namedColors[lower]; ok {
		return Color{Color: c, originalFormat: FormatHex}, nil
	}

	return Color{}, fmt.Errorf("unrecognized color format: %s", input)
}

// MustParse is like Parse but panics on error (useful for tests and known-good values)
func MustParse(input string) Color {
	c, err := Parse(input)
	if err != nil {
		panic(err)
	}
	return c
}

// parseHex parses hex color formats: #rgb, #rrggbb, #rrggbbaa
func parseHex(input string) (colorful.Color, error) {
	// go-colorful's Hex() handles #rrggbb
	// We need to also handle #rgb shorthand and #rrggbbaa
	hex := strings.TrimPrefix(input, "#")

	switch len(hex) {
	case 3:
		// Expand #rgb to #rrggbb
		hex = string([]byte{hex[0], hex[0], hex[1], hex[1], hex[2], hex[2]})
	case 4:
		// Expand #rgba to #rrggbbaa (we ignore alpha for now)
		hex = string([]byte{hex[0], hex[0], hex[1], hex[1], hex[2], hex[2]})
	case 6:
		// Standard format, use as-is
	case 8:
		// Has alpha, strip it (go-colorful doesn't handle alpha)
		hex = hex[:6]
	default:
		return colorful.Color{}, fmt.Errorf("invalid hex color: %s", input)
	}

	return colorful.Hex("#" + hex)
}

// parseRGB parses rgb() and rgba() formats
// Supports both comma and space separators
var rgbRegex = regexp.MustCompile(`rgba?\s*\(\s*([0-9.]+%?)\s*[,\s]\s*([0-9.]+%?)\s*[,\s]\s*([0-9.]+%?)(?:\s*[,/]\s*([0-9.]+%?))?\s*\)`)

func parseRGB(input string) (colorful.Color, error) {
	matches := rgbRegex.FindStringSubmatch(input)
	if matches == nil {
		return colorful.Color{}, fmt.Errorf("invalid rgb color: %s", input)
	}

	r, err := parseColorComponent(matches[1], 255)
	if err != nil {
		return colorful.Color{}, fmt.Errorf("invalid red component: %w", err)
	}

	g, err := parseColorComponent(matches[2], 255)
	if err != nil {
		return colorful.Color{}, fmt.Errorf("invalid green component: %w", err)
	}

	b, err := parseColorComponent(matches[3], 255)
	if err != nil {
		return colorful.Color{}, fmt.Errorf("invalid blue component: %w", err)
	}

	// Alpha is ignored for now (matches[4])

	return colorful.Color{R: r, G: g, B: b}, nil
}

// parseHSL parses hsl() and hsla() formats
var hslRegex = regexp.MustCompile(`hsla?\s*\(\s*([0-9.]+)(?:deg)?\s*[,\s]\s*([0-9.]+)%\s*[,\s]\s*([0-9.]+)%(?:\s*[,/]\s*([0-9.]+%?))?\s*\)`)

func parseHSL(input string) (colorful.Color, error) {
	matches := hslRegex.FindStringSubmatch(input)
	if matches == nil {
		return colorful.Color{}, fmt.Errorf("invalid hsl color: %s", input)
	}

	h, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return colorful.Color{}, fmt.Errorf("invalid hue: %w", err)
	}

	s, err := strconv.ParseFloat(matches[2], 64)
	if err != nil {
		return colorful.Color{}, fmt.Errorf("invalid saturation: %w", err)
	}

	l, err := strconv.ParseFloat(matches[3], 64)
	if err != nil {
		return colorful.Color{}, fmt.Errorf("invalid lightness: %w", err)
	}

	// go-colorful expects s and l in 0-1 range
	return colorful.Hsl(h, s/100, l/100), nil
}

// parseOKLCH parses oklch() format
// CSS syntax: oklch(L C H) where L is 0-1 or 0%-100%, C is typically 0-0.4, H is 0-360
var oklchRegex = regexp.MustCompile(`oklch\s*\(\s*([0-9.]+)(%?)\s+([0-9.]+)\s+([0-9.]+)\s*\)`)

func parseOKLCH(input string) (colorful.Color, error) {
	matches := oklchRegex.FindStringSubmatch(input)
	if matches == nil {
		return colorful.Color{}, fmt.Errorf("invalid oklch color: %s", input)
	}

	l, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return colorful.Color{}, fmt.Errorf("invalid lightness: %w", err)
	}

	// If percentage, convert to 0-1
	if matches[2] == "%" {
		l = l / 100
	}

	c, err := strconv.ParseFloat(matches[3], 64)
	if err != nil {
		return colorful.Color{}, fmt.Errorf("invalid chroma: %w", err)
	}

	h, err := strconv.ParseFloat(matches[4], 64)
	if err != nil {
		return colorful.Color{}, fmt.Errorf("invalid hue: %w", err)
	}

	return colorful.OkLch(l, c, h), nil
}

// parseColorComponent parses a color component value (0-255 or 0%-100%)
func parseColorComponent(s string, max float64) (float64, error) {
	s = strings.TrimSpace(s)
	if strings.HasSuffix(s, "%") {
		v, err := strconv.ParseFloat(strings.TrimSuffix(s, "%"), 64)
		if err != nil {
			return 0, err
		}
		return v / 100, nil
	}

	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, err
	}
	return v / max, nil
}

// ToCSS outputs the color in the specified format
func (c Color) ToCSS(format string) string {
	switch format {
	case FormatHex:
		return c.Hex()
	case FormatRGB:
		return c.ToRGB()
	case FormatHSL:
		return c.ToHSL()
	case FormatOKLCH:
		return c.ToOKLCH()
	default:
		return c.Hex()
	}
}

// ToOriginalFormat outputs the color in its original parsed format
func (c Color) ToOriginalFormat() string {
	return c.ToCSS(c.originalFormat)
}

// OriginalFormat returns the format the color was originally parsed from
func (c Color) OriginalFormat() string {
	return c.originalFormat
}

// Hex returns the color as a hex string (#rrggbb)
func (c Color) Hex() string {
	return c.Color.Hex()
}

// ToRGB returns the color as an rgb() CSS string
func (c Color) ToRGB() string {
	r, g, b := c.Color.RGB255()
	return fmt.Sprintf("rgb(%d, %d, %d)", r, g, b)
}

// ToHSL returns the color as an hsl() CSS string
func (c Color) ToHSL() string {
	h, s, l := c.Color.Hsl()
	return fmt.Sprintf("hsl(%.1f, %.1f%%, %.1f%%)", h, s*100, l*100)
}

// ToOKLCH returns the color as an oklch() CSS string
// Format matches DaisyUI convention: oklch(L% C H)
func (c Color) ToOKLCH() string {
	l, ch, h := c.Color.OkLch()
	return fmt.Sprintf("oklch(%.2f%% %.3f %.2f)", l*100, ch, h)
}

// RGB255 returns the RGB components as 0-255 integers
func (c Color) RGB255() (r, g, b uint8) {
	return c.Color.RGB255()
}

// OkLch returns the OKLCH components
func (c Color) OkLch() (l, chroma, h float64) {
	return c.Color.OkLch()
}

// IsValid returns true if the color is a valid RGB color (all components in 0-1)
func (c Color) IsValid() bool {
	return c.Color.IsValid()
}

// Clamped returns a copy of the color clamped to valid RGB range
func (c Color) Clamped() Color {
	return Color{Color: c.Color.Clamped(), originalFormat: c.originalFormat}
}

// namedColors maps CSS named colors to their colorful.Color values
var namedColors = map[string]colorful.Color{
	"black":   {R: 0, G: 0, B: 0},
	"white":   {R: 1, G: 1, B: 1},
	"red":     {R: 1, G: 0, B: 0},
	"green":   {R: 0, G: 0.502, B: 0}, // CSS green is #008000
	"blue":    {R: 0, G: 0, B: 1},
	"yellow":  {R: 1, G: 1, B: 0},
	"cyan":    {R: 0, G: 1, B: 1},
	"magenta": {R: 1, G: 0, B: 1},
	"orange":  {R: 1, G: 0.647, B: 0},
	"purple":  {R: 0.502, G: 0, B: 0.502},
	"pink":    {R: 1, G: 0.753, B: 0.796},
	"gray":    {R: 0.502, G: 0.502, B: 0.502},
	"grey":    {R: 0.502, G: 0.502, B: 0.502},
	// Add more as needed
}

// White returns a white color
func White() Color {
	return Color{Color: colorful.Color{R: 1, G: 1, B: 1}, originalFormat: FormatHex}
}

// Black returns a black color
func Black() Color {
	return Color{Color: colorful.Color{R: 0, G: 0, B: 0}, originalFormat: FormatHex}
}

// FromColorful wraps a go-colorful.Color into our Color type
func FromColorful(c colorful.Color, format string) Color {
	return Color{Color: c, originalFormat: format}
}

// FromOkLch creates a Color from OKLCH values
func FromOkLch(l, c, h float64) Color {
	return Color{Color: colorful.OkLch(l, c, h), originalFormat: FormatOKLCH}
}

// FromRGB creates a Color from RGB values (0-1 range)
func FromRGB(r, g, b float64) Color {
	return Color{Color: colorful.Color{R: r, G: g, B: b}, originalFormat: FormatRGB}
}

// FromRGB255 creates a Color from RGB values (0-255 range)
func FromRGB255(r, g, b uint8) Color {
	return Color{
		Color:          colorful.Color{R: float64(r) / 255, G: float64(g) / 255, B: float64(b) / 255},
		originalFormat: FormatRGB,
	}
}
