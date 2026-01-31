// tokenctl/pkg/colors/colors_test.go

package colors

import (
	"math"
	"testing"
)

// ============================================================================
// Color Parsing Tests
// ============================================================================

func TestParse_Hex(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantR   uint8
		wantG   uint8
		wantB   uint8
		wantFmt string
		wantErr bool
	}{
		{
			name:    "6-digit hex",
			input:   "#3b82f6",
			wantR:   59,
			wantG:   130,
			wantB:   246,
			wantFmt: FormatHex,
		},
		{
			name:    "6-digit hex uppercase",
			input:   "#3B82F6",
			wantR:   59,
			wantG:   130,
			wantB:   246,
			wantFmt: FormatHex,
		},
		{
			name:    "3-digit hex shorthand",
			input:   "#fff",
			wantR:   255,
			wantG:   255,
			wantB:   255,
			wantFmt: FormatHex,
		},
		{
			name:    "3-digit hex shorthand colors",
			input:   "#f00",
			wantR:   255,
			wantG:   0,
			wantB:   0,
			wantFmt: FormatHex,
		},
		{
			name:    "8-digit hex with alpha",
			input:   "#3b82f6ff",
			wantR:   59,
			wantG:   130,
			wantB:   246,
			wantFmt: FormatHex,
		},
		{
			name:    "black",
			input:   "#000000",
			wantR:   0,
			wantG:   0,
			wantB:   0,
			wantFmt: FormatHex,
		},
		{
			name:    "white",
			input:   "#ffffff",
			wantR:   255,
			wantG:   255,
			wantB:   255,
			wantFmt: FormatHex,
		},
		{
			name:    "invalid hex",
			input:   "#gggggg",
			wantErr: true,
		},
		{
			name:    "invalid hex length",
			input:   "#ff",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c, err := Parse(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("Parse(%q) expected error, got nil", tt.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("Parse(%q) unexpected error: %v", tt.input, err)
			}

			r, g, b := c.RGB255()
			if r != tt.wantR || g != tt.wantG || b != tt.wantB {
				t.Errorf("Parse(%q) = RGB(%d,%d,%d), want RGB(%d,%d,%d)",
					tt.input, r, g, b, tt.wantR, tt.wantG, tt.wantB)
			}

			if c.OriginalFormat() != tt.wantFmt {
				t.Errorf("Parse(%q) format = %q, want %q", tt.input, c.OriginalFormat(), tt.wantFmt)
			}
		})
	}
}

func TestParse_RGB(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantR   uint8
		wantG   uint8
		wantB   uint8
		wantErr bool
	}{
		{
			name:  "rgb with commas",
			input: "rgb(255, 128, 64)",
			wantR: 255,
			wantG: 128,
			wantB: 64,
		},
		{
			name:  "rgb with spaces",
			input: "rgb(255 128 64)",
			wantR: 255,
			wantG: 128,
			wantB: 64,
		},
		{
			name:  "rgba with alpha",
			input: "rgba(255, 128, 64, 0.5)",
			wantR: 255,
			wantG: 128,
			wantB: 64,
		},
		{
			name:  "rgb with percentages",
			input: "rgb(100%, 50%, 25%)",
			wantR: 255,
			wantG: 127,
			wantB: 63,
		},
		{
			name:  "rgb black",
			input: "rgb(0, 0, 0)",
			wantR: 0,
			wantG: 0,
			wantB: 0,
		},
		{
			name:  "rgb white",
			input: "rgb(255, 255, 255)",
			wantR: 255,
			wantG: 255,
			wantB: 255,
		},
		{
			name:    "invalid rgb",
			input:   "rgb(abc, def, ghi)",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c, err := Parse(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("Parse(%q) expected error, got nil", tt.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("Parse(%q) unexpected error: %v", tt.input, err)
			}

			r, g, b := c.RGB255()
			// Allow 1 unit tolerance for rounding
			if abs(int(r)-int(tt.wantR)) > 1 || abs(int(g)-int(tt.wantG)) > 1 || abs(int(b)-int(tt.wantB)) > 1 {
				t.Errorf("Parse(%q) = RGB(%d,%d,%d), want RGB(%d,%d,%d)",
					tt.input, r, g, b, tt.wantR, tt.wantG, tt.wantB)
			}

			if c.OriginalFormat() != FormatRGB {
				t.Errorf("Parse(%q) format = %q, want %q", tt.input, c.OriginalFormat(), FormatRGB)
			}
		})
	}
}

func TestParse_HSL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantR   uint8
		wantG   uint8
		wantB   uint8
		wantErr bool
	}{
		{
			name:  "hsl red",
			input: "hsl(0, 100%, 50%)",
			wantR: 255,
			wantG: 0,
			wantB: 0,
		},
		{
			name:  "hsl green",
			input: "hsl(120, 100%, 50%)",
			wantR: 0,
			wantG: 255,
			wantB: 0,
		},
		{
			name:  "hsl blue",
			input: "hsl(240, 100%, 50%)",
			wantR: 0,
			wantG: 0,
			wantB: 255,
		},
		{
			name:  "hsl with deg",
			input: "hsl(180deg, 50%, 50%)",
			wantR: 64,
			wantG: 191,
			wantB: 191,
		},
		{
			name:  "hsla with alpha",
			input: "hsla(180, 50%, 50%, 0.5)",
			wantR: 64,
			wantG: 191,
			wantB: 191,
		},
		{
			name:    "invalid hsl",
			input:   "hsl(abc, 50%, 50%)",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c, err := Parse(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("Parse(%q) expected error, got nil", tt.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("Parse(%q) unexpected error: %v", tt.input, err)
			}

			r, g, b := c.RGB255()
			// Allow 2 units tolerance for HSL conversion rounding
			if abs(int(r)-int(tt.wantR)) > 2 || abs(int(g)-int(tt.wantG)) > 2 || abs(int(b)-int(tt.wantB)) > 2 {
				t.Errorf("Parse(%q) = RGB(%d,%d,%d), want RGB(%d,%d,%d)",
					tt.input, r, g, b, tt.wantR, tt.wantG, tt.wantB)
			}

			if c.OriginalFormat() != FormatHSL {
				t.Errorf("Parse(%q) format = %q, want %q", tt.input, c.OriginalFormat(), FormatHSL)
			}
		})
	}
}

func TestParse_OKLCH(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantL   float64
		wantC   float64
		wantH   float64
		wantErr bool
	}{
		{
			name:  "oklch with percentage lightness",
			input: "oklch(50% 0.2 180)",
			wantL: 0.50,
			wantC: 0.2,
			wantH: 180,
		},
		{
			name:  "oklch with decimal lightness",
			input: "oklch(0.5 0.2 180)",
			wantL: 0.50,
			wantC: 0.2,
			wantH: 180,
		},
		{
			name:  "oklch DaisyUI primary example",
			input: "oklch(49.12% 0.309 275.75)",
			wantL: 0.4912,
			wantC: 0.309,
			wantH: 275.75,
		},
		{
			name:  "oklch white",
			input: "oklch(100% 0 0)",
			wantL: 1.0,
			wantC: 0,
			wantH: -1, // Hue is undefined for achromatic colors, use -1 to skip check
		},
		{
			name:  "oklch black",
			input: "oklch(0% 0 0)",
			wantL: 0,
			wantC: 0,
			wantH: -1, // Hue is undefined for achromatic colors, use -1 to skip check
		},
		{
			name:    "invalid oklch",
			input:   "oklch(abc 0.2 180)",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c, err := Parse(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("Parse(%q) expected error, got nil", tt.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("Parse(%q) unexpected error: %v", tt.input, err)
			}

			l, ch, h := c.OkLch()
			// Allow small tolerance for floating point
			// Note: hue is undefined for achromatic colors (chroma=0), so skip hue check if wantH is -1
			hueOk := tt.wantH < 0 || math.Abs(h-tt.wantH) <= 0.5
			if math.Abs(l-tt.wantL) > 0.01 || math.Abs(ch-tt.wantC) > 0.01 || !hueOk {
				t.Errorf("Parse(%q) = OKLCH(%.3f, %.3f, %.2f), want OKLCH(%.3f, %.3f, %.2f)",
					tt.input, l, ch, h, tt.wantL, tt.wantC, tt.wantH)
			}

			if c.OriginalFormat() != FormatOKLCH {
				t.Errorf("Parse(%q) format = %q, want %q", tt.input, c.OriginalFormat(), FormatOKLCH)
			}
		})
	}
}

func TestParse_NamedColors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		wantR uint8
		wantG uint8
		wantB uint8
	}{
		{"black", "black", 0, 0, 0},
		{"white", "white", 255, 255, 255},
		{"red", "red", 255, 0, 0},
		{"blue", "blue", 0, 0, 255},
		{"Black uppercase", "BLACK", 0, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c, err := Parse(tt.input)
			if err != nil {
				t.Fatalf("Parse(%q) unexpected error: %v", tt.input, err)
			}

			r, g, b := c.RGB255()
			if r != tt.wantR || g != tt.wantG || b != tt.wantB {
				t.Errorf("Parse(%q) = RGB(%d,%d,%d), want RGB(%d,%d,%d)",
					tt.input, r, g, b, tt.wantR, tt.wantG, tt.wantB)
			}
		})
	}
}

func TestParse_Errors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
	}{
		{"empty string", ""},
		{"random string", "not a color"},
		{"invalid format", "xyz(1,2,3)"},
		{"malformed hex", "#zzz"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := Parse(tt.input)
			if err == nil {
				t.Errorf("Parse(%q) expected error, got nil", tt.input)
			}
		})
	}
}

// ============================================================================
// Color Output Tests
// ============================================================================

func TestColor_ToCSS(t *testing.T) {
	t.Parallel()

	c := MustParse("#3b82f6")

	tests := []struct {
		format string
		want   string
	}{
		{FormatHex, "#3b82f6"},
		{FormatRGB, "rgb(59, 130, 246)"},
	}

	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			t.Parallel()

			got := c.ToCSS(tt.format)
			if got != tt.want {
				t.Errorf("ToCSS(%q) = %q, want %q", tt.format, got, tt.want)
			}
		})
	}
}

func TestColor_ToOKLCH(t *testing.T) {
	t.Parallel()

	// Test that ToOKLCH produces valid output format
	c := MustParse("#3b82f6")
	oklch := c.ToOKLCH()

	// Should match pattern oklch(XX.XX% X.XXX XXX.XX)
	if len(oklch) < 10 {
		t.Errorf("ToOKLCH() = %q, expected longer string", oklch)
	}

	// Should be parseable back
	reparsed, err := Parse(oklch)
	if err != nil {
		t.Errorf("ToOKLCH() output %q not parseable: %v", oklch, err)
	}

	// Original format should be preserved on re-parse
	if reparsed.OriginalFormat() != FormatOKLCH {
		t.Errorf("Re-parsed format = %q, want %q", reparsed.OriginalFormat(), FormatOKLCH)
	}
}

func TestColor_RoundTrip(t *testing.T) {
	t.Parallel()

	// Test that colors can be round-tripped through various formats
	tests := []string{
		"#3b82f6",
		"#ffffff",
		"#000000",
		"#ff0000",
		"rgb(128, 64, 32)",
		"hsl(180, 50%, 50%)",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			t.Parallel()

			c1, err := Parse(input)
			if err != nil {
				t.Fatalf("Parse(%q) error: %v", input, err)
			}

			// Convert to hex and back
			hex := c1.Hex()
			c2, err := Parse(hex)
			if err != nil {
				t.Fatalf("Parse(%q) error: %v", hex, err)
			}

			r1, g1, b1 := c1.RGB255()
			r2, g2, b2 := c2.RGB255()

			if r1 != r2 || g1 != g2 || b1 != b2 {
				t.Errorf("Round-trip failed: RGB(%d,%d,%d) -> %q -> RGB(%d,%d,%d)",
					r1, g1, b1, hex, r2, g2, b2)
			}
		})
	}
}

// ============================================================================
// Contrast Tests
// ============================================================================

func TestContrastRatio(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		c1      string
		c2      string
		wantMin float64
		wantMax float64
	}{
		{
			name:    "black and white",
			c1:      "#000000",
			c2:      "#ffffff",
			wantMin: 20.9,
			wantMax: 21.1,
		},
		{
			name:    "same color",
			c1:      "#3b82f6",
			c2:      "#3b82f6",
			wantMin: 1.0,
			wantMax: 1.0,
		},
		{
			name:    "white and light gray",
			c1:      "#ffffff",
			c2:      "#cccccc",
			wantMin: 1.5,
			wantMax: 1.7,
		},
		{
			name:    "dark blue and white",
			c1:      "#1e3a5f",
			c2:      "#ffffff",
			wantMin: 10.0,
			wantMax: 12.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c1 := MustParse(tt.c1)
			c2 := MustParse(tt.c2)

			ratio := ContrastRatio(c1, c2)

			if ratio < tt.wantMin || ratio > tt.wantMax {
				t.Errorf("ContrastRatio(%q, %q) = %.2f, want between %.2f and %.2f",
					tt.c1, tt.c2, ratio, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestContrastRatio_Symmetric(t *testing.T) {
	t.Parallel()

	// Contrast ratio should be the same regardless of argument order
	c1 := MustParse("#3b82f6")
	c2 := MustParse("#ffffff")

	ratio1 := ContrastRatio(c1, c2)
	ratio2 := ContrastRatio(c2, c1)

	if math.Abs(ratio1-ratio2) > 0.001 {
		t.Errorf("ContrastRatio not symmetric: %.4f vs %.4f", ratio1, ratio2)
	}
}

func TestMeetsWCAG(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		c1        string
		c2        string
		level     string
		largeText bool
		want      bool
	}{
		{
			name:  "black on white - AA normal",
			c1:    "#000000",
			c2:    "#ffffff",
			level: "AA",
			want:  true,
		},
		{
			name:  "black on white - AAA normal",
			c1:    "#000000",
			c2:    "#ffffff",
			level: "AAA",
			want:  true,
		},
		{
			name:  "light gray on white - AA fail",
			c1:    "#cccccc",
			c2:    "#ffffff",
			level: "AA",
			want:  false,
		},
		{
			name:      "medium gray on white - AA large",
			c1:        "#767676",
			c2:        "#ffffff",
			level:     "AA",
			largeText: true,
			want:      true,
		},
		{
			name:  "dark blue on white - AA pass",
			c1:    "#1e3a5f",
			c2:    "#ffffff",
			level: "AA",
			want:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c1 := MustParse(tt.c1)
			c2 := MustParse(tt.c2)

			got := MeetsWCAG(c1, c2, tt.level, tt.largeText)
			if got != tt.want {
				t.Errorf("MeetsWCAG(%q, %q, %q, %v) = %v, want %v",
					tt.c1, tt.c2, tt.level, tt.largeText, got, tt.want)
			}
		})
	}
}

func TestContrastLevel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		c1   string
		c2   string
		want string
	}{
		{"black/white", "#000000", "#ffffff", "AAA"},
		{"dark blue/white", "#1e3a5f", "#ffffff", "AAA"},
		{"medium gray/white", "#767676", "#ffffff", "AA"}, // #767676 on white is exactly 4.54:1, passes AA
		{"light gray/white", "#cccccc", "#ffffff", "Fail"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c1 := MustParse(tt.c1)
			c2 := MustParse(tt.c2)

			got := ContrastLevel(c1, c2)
			if got != tt.want {
				t.Errorf("ContrastLevel(%q, %q) = %q, want %q", tt.c1, tt.c2, got, tt.want)
			}
		})
	}
}

func TestOptimalTextColor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		background string
		wantWhite  bool
	}{
		{"dark blue", "#1e3a5f", true},
		{"black", "#000000", true},
		{"white", "#ffffff", false},
		{"yellow", "#ffff00", false},
		{"dark red", "#8b0000", true},
		{"light pink", "#ffb6c1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			bg := MustParse(tt.background)
			result := OptimalTextColor(bg)

			r, g, b := result.RGB255()
			isWhite := r == 255 && g == 255 && b == 255

			if isWhite != tt.wantWhite {
				t.Errorf("OptimalTextColor(%q) returned white=%v, want white=%v",
					tt.background, isWhite, tt.wantWhite)
			}
		})
	}
}

// ============================================================================
// Content Color Tests
// ============================================================================

func TestContentColor_WCAGCompliance(t *testing.T) {
	t.Parallel()

	// Test that ContentColor always produces WCAG AA compliant results
	testColors := []string{
		"#3b82f6", // Blue
		"#ef4444", // Red
		"#10b981", // Green
		"#f59e0b", // Orange
		"#8b5cf6", // Purple
		"#000000", // Black
		"#ffffff", // White
		"#808080", // Mid gray
		"#ffff00", // Yellow (light, tricky)
		"#00ffff", // Cyan
	}

	for _, hex := range testColors {
		t.Run(hex, func(t *testing.T) {
			t.Parallel()

			bg := MustParse(hex)
			content := ContentColor(bg)

			ratio := ContrastRatio(bg, content)

			if ratio < WCAGAANormal {
				t.Errorf("ContentColor(%q) contrast ratio = %.2f, want >= %.1f",
					hex, ratio, WCAGAANormal)
			}
		})
	}
}

func TestContentColor_DaisyUIColors(t *testing.T) {
	t.Parallel()

	// Test with actual DaisyUI default theme colors
	daisyColors := []struct {
		name  string
		value string
	}{
		{"primary", "oklch(49.12% 0.309 275.75)"},
		{"secondary", "oklch(69.71% 0.329 342.55)"},
		{"accent", "oklch(76.76% 0.184 183.61)"},
		{"neutral", "oklch(20% 0.024 255.701)"},
		{"base-100", "oklch(100% 0 0)"},
		{"info", "oklch(72.06% 0.191 231.6)"},
		{"success", "oklch(64.8% 0.15 160)"},
		{"warning", "oklch(84.71% 0.199 83.87)"},
		{"error", "oklch(71.76% 0.221 22.18)"},
	}

	for _, dc := range daisyColors {
		t.Run(dc.name, func(t *testing.T) {
			t.Parallel()

			bg, err := Parse(dc.value)
			if err != nil {
				t.Fatalf("Parse(%q) error: %v", dc.value, err)
			}

			content := ContentColor(bg)
			ratio := ContrastRatio(bg, content)

			if ratio < WCAGAANormal {
				t.Errorf("ContentColor for %s (%s) contrast ratio = %.2f, want >= %.1f",
					dc.name, dc.value, ratio, WCAGAANormal)
			}
		})
	}
}

func TestContentColorWithRatio(t *testing.T) {
	t.Parallel()

	// Test with a dark color where AAA (7:1) is achievable with white
	bg := MustParse("#1e3a5f") // Dark blue
	content := ContentColorWithRatio(bg, WCAGAAANormal)
	ratio := ContrastRatio(bg, content)

	if ratio < WCAGAAANormal {
		t.Errorf("ContentColorWithRatio() ratio = %.2f, want >= %.1f", ratio, WCAGAAANormal)
	}

	// Test that when target ratio is unachievable, we get the best possible
	midTone := MustParse("#3b82f6") // Medium blue - max contrast is ~5.7
	contentMid := ContentColorWithRatio(midTone, WCAGAAANormal)
	ratioMid := ContrastRatio(midTone, contentMid)

	// Should return black (best possible at ~5.7), not fail
	// Just verify we get a reasonable contrast
	if ratioMid < WCAGAANormal {
		t.Errorf("ContentColorWithRatio() for mid-tone ratio = %.2f, want >= %.1f (best effort)", ratioMid, WCAGAANormal)
	}
}

func TestGenerateContentPair(t *testing.T) {
	t.Parallel()

	base := MustParse("#3b82f6")
	gotBase, gotContent := GenerateContentPair(base)

	// Base should be unchanged
	if gotBase.Hex() != base.Hex() {
		t.Errorf("GenerateContentPair base changed: %q -> %q", base.Hex(), gotBase.Hex())
	}

	// Content should have sufficient contrast
	ratio := ContrastRatio(gotBase, gotContent)
	if ratio < WCAGAANormal {
		t.Errorf("GenerateContentPair content ratio = %.2f, want >= %.1f", ratio, WCAGAANormal)
	}
}

func TestDaisyContentColor(t *testing.T) {
	t.Parallel()

	// Test that DaisyContentColor produces valid results
	testColors := []string{
		"#3b82f6",
		"#ef4444",
		"#ffffff",
		"#000000",
		"#808080",
	}

	for _, hex := range testColors {
		t.Run(hex, func(t *testing.T) {
			t.Parallel()

			bg := MustParse(hex)
			content := DaisyContentColor(bg)

			ratio := ContrastRatio(bg, content)
			if ratio < WCAGAANormal {
				t.Errorf("DaisyContentColor(%q) ratio = %.2f, want >= %.1f",
					hex, ratio, WCAGAANormal)
			}
		})
	}
}

// ============================================================================
// Helper Functions
// ============================================================================

func TestMustParse_Panic(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			t.Error("MustParse with invalid input should panic")
		}
	}()

	MustParse("not a color")
}

func TestWhiteAndBlack(t *testing.T) {
	t.Parallel()

	white := White()
	black := Black()

	wr, wg, wb := white.RGB255()
	if wr != 255 || wg != 255 || wb != 255 {
		t.Errorf("White() = RGB(%d,%d,%d), want RGB(255,255,255)", wr, wg, wb)
	}

	br, bg, bb := black.RGB255()
	if br != 0 || bg != 0 || bb != 0 {
		t.Errorf("Black() = RGB(%d,%d,%d), want RGB(0,0,0)", br, bg, bb)
	}
}

func TestFromOkLch(t *testing.T) {
	t.Parallel()

	c := FromOkLch(0.5, 0.2, 180)

	l, ch, h := c.OkLch()
	if math.Abs(l-0.5) > 0.01 || math.Abs(ch-0.2) > 0.01 || math.Abs(h-180) > 0.5 {
		t.Errorf("FromOkLch(0.5, 0.2, 180) = OKLCH(%.3f, %.3f, %.2f)", l, ch, h)
	}

	if c.OriginalFormat() != FormatOKLCH {
		t.Errorf("FromOkLch format = %q, want %q", c.OriginalFormat(), FormatOKLCH)
	}
}

func TestFromRGB255(t *testing.T) {
	t.Parallel()

	c := FromRGB255(128, 64, 32)

	r, g, b := c.RGB255()
	if r != 128 || g != 64 || b != 32 {
		t.Errorf("FromRGB255(128,64,32) = RGB(%d,%d,%d)", r, g, b)
	}
}

func TestIsValid(t *testing.T) {
	t.Parallel()

	valid := MustParse("#3b82f6")
	if !valid.IsValid() {
		t.Error("Valid color reported as invalid")
	}
}

func TestClamped(t *testing.T) {
	t.Parallel()

	// Create a potentially invalid color from OKLCH
	// Some OKLCH values produce out-of-gamut RGB
	c := FromOkLch(0.9, 0.4, 150)
	clamped := c.Clamped()

	// Clamped version should always be valid
	if !clamped.IsValid() {
		t.Error("Clamped color should be valid")
	}
}

func TestRelativeLuminance(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		color   string
		wantMin float64
		wantMax float64
	}{
		{"white", "#ffffff", 0.99, 1.01},
		{"black", "#000000", 0.0, 0.01},
		{"mid gray", "#808080", 0.2, 0.25},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c := MustParse(tt.color)
			lum := RelativeLuminance(c)

			if lum < tt.wantMin || lum > tt.wantMax {
				t.Errorf("RelativeLuminance(%q) = %.4f, want between %.4f and %.4f",
					tt.color, lum, tt.wantMin, tt.wantMax)
			}
		})
	}
}

// abs returns the absolute value of an int
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
