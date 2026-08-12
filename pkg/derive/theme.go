// tokenctl/pkg/derive/theme.go
//
// The colour derivation itself: OKLCH lightness/chroma/hue triples for
// every semantic colour, with the neutrals tinted by the tint control
// and the actions scaled by the saturation control.
package derive

import "math"

// Theme is a derived token set. Order preserves the engine's emission
// order so CSS output reads the same way the original did; Values is the
// lookup used for comparison and emission.
type Theme struct {
	Params Params
	Order  []string
	Values map[string]string
}

func newTheme(p Params) *Theme {
	return &Theme{Params: p, Values: make(map[string]string, 64)}
}

func (t *Theme) set(name, value string) {
	if _, seen := t.Values[name]; !seen {
		t.Order = append(t.Order, name)
	}
	t.Values[name] = value
}

// Get returns a token value by CSS variable name.
func (t *Theme) Get(name string) (string, bool) {
	v, ok := t.Values[name]
	return v, ok
}

// Len reports how many tokens the theme carries.
func (t *Theme) Len() int { return len(t.Order) }

// Generate derives a complete theme — colours plus typography and
// density artifacts — from a full parameter set. Params must already be
// valid; call Params.Validate first.
func Generate(p Params) *Theme {
	t := newTheme(p)
	generateColors(p, t)
	generateTypographyTokens(p.FontPairing, p.Density, t)
	return t
}

// GenerateColors derives only the semantic colour tokens.
func GenerateColors(p Params) *Theme {
	t := newTheme(p)
	generateColors(p, t)
	return t
}

// FromPreset builds params from a named preset, taking the remaining
// controls from defaults.
func FromPreset(pr Preset) Params {
	p := DefaultParams
	p.Hue = pr.Hue
	p.Chroma = pr.Chroma
	return p
}

// oklch renders an OKLCH triple the way the engine does: lightness to
// one decimal as a percentage, chroma to three, hue to none.
func oklch(l, c, h float64) string {
	return "oklch(" + toFixed(l, 1) + "% " + toFixed(c, 3) + " " + toFixed(h, 0) + ")"
}

func clampL(l float64) float64 { return math.Max(0, math.Min(100, l)) }
func clampC(c float64) float64 { return math.Max(0, math.Min(0.4, c)) }

func generateColors(p Params, t *Theme) {
	// Saturation 0–150 maps to a 0.0–1.5x multiplier on action chroma.
	sat := p.Saturation / 100
	// Tint 0–100 maps to 0.0–1.0: how much of the hue reaches neutrals.
	// At 0 the neutrals are pure grey; at 100 they are strongly tinted.
	tintF := p.Tint / 100

	neutralC := func(v float64) float64 { return clampC(v * tintF) }
	actionC := func(v float64) float64 { return clampC(v * sat) }

	if p.IsDark {
		generateDark(p.Hue, p.Chroma, neutralC, actionC, t)
		return
	}
	generateLight(p.Hue, p.Chroma, neutralC, actionC, t)
}

func generateLight(hue, chroma float64, neutralC, actionC func(float64) float64, t *Theme) {
	const primaryL = 55
	primaryC := actionC(chroma)

	t.set("--color-primary", oklch(primaryL, primaryC, hue))
	t.set("--color-secondary", oklch(55, actionC(chroma*0.15), hue))
	t.set("--color-destructive", oklch(55, actionC(0.22), 25))
	t.set("--color-success", oklch(55, actionC(0.18), 145))
	t.set("--color-warning", oklch(75, actionC(0.18), 75))
	t.set("--color-error", oklch(55, actionC(0.22), 25))
	t.set("--color-info", oklch(primaryL, primaryC, hue))

	t.set("--color-background", oklch(clampL(99), neutralC(0.010), hue))
	t.set("--color-surface", oklch(clampL(96), neutralC(0.015), hue))
	t.set("--color-accent", oklch(clampL(96), neutralC(0.015), hue))
	t.set("--color-muted", oklch(clampL(92), neutralC(0.015), hue))
	t.set("--color-popover", oklch(clampL(99), neutralC(0.010), hue))

	t.set("--color-foreground", oklch(clampL(18), neutralC(0.020), hue))
	t.set("--color-text", oklch(clampL(18), neutralC(0.020), hue))

	t.set("--color-border", oklch(clampL(80), neutralC(0.015), hue))
	t.set("--color-ring", oklch(primaryL, primaryC, hue))
	t.set("--color-backdrop", "rgba(0, 0, 0, 0.5)")

	t.set("--color-shadow", "rgba(0, 0, 0, 0.1)")
	t.set("--color-primary-foreground", oklch(100, 0, 0))
	t.set("--color-secondary-foreground", oklch(clampL(18), neutralC(0.02), hue))
	t.set("--color-destructive-foreground", oklch(100, 0, 0))
	t.set("--color-success-foreground", oklch(100, 0, 0))
	t.set("--color-warning-foreground", oklch(clampL(18), neutralC(0.02), hue))
	t.set("--color-info-foreground", oklch(100, 0, 0))
	t.set("--color-success-subtle", oklch(95, actionC(0.05), 145))
	t.set("--color-warning-subtle", oklch(95, actionC(0.05), 75))
	t.set("--color-error-subtle", oklch(95, actionC(0.03), 25))
	t.set("--color-info-subtle", oklch(95, actionC(0.03), hue))
}

func generateDark(hue, chroma float64, neutralC, actionC func(float64) float64, t *Theme) {
	const primaryL = 70
	primaryC := actionC(chroma * 0.9)

	t.set("--color-primary", oklch(primaryL, primaryC, hue))
	t.set("--color-secondary", oklch(75, actionC(chroma*0.15), hue))
	t.set("--color-destructive", oklch(65, actionC(0.20), 25))
	t.set("--color-success", oklch(65, actionC(0.16), 145))
	t.set("--color-warning", oklch(80, actionC(0.16), 75))
	t.set("--color-error", oklch(65, actionC(0.20), 25))
	t.set("--color-info", oklch(primaryL, primaryC, hue))

	t.set("--color-background", oklch(clampL(12), neutralC(0.015), hue))
	t.set("--color-surface", oklch(clampL(20), neutralC(0.015), hue))
	t.set("--color-accent", oklch(clampL(25), neutralC(0.015), hue))
	t.set("--color-muted", oklch(clampL(25), neutralC(0.015), hue))
	t.set("--color-popover", oklch(clampL(16), neutralC(0.015), hue))

	t.set("--color-foreground", oklch(clampL(95), neutralC(0.010), hue))
	t.set("--color-text", oklch(clampL(95), neutralC(0.010), hue))

	t.set("--color-border", oklch(clampL(40), neutralC(0.015), hue))
	t.set("--color-ring", oklch(primaryL, primaryC, hue))
	t.set("--color-backdrop", "rgba(0, 0, 0, 0.7)")

	t.set("--color-shadow", "rgba(0, 0, 0, 0.4)")
	t.set("--color-primary-foreground", oklch(clampL(12), neutralC(0.015), hue))
	t.set("--color-secondary-foreground", oklch(clampL(95), neutralC(0.01), hue))
	t.set("--color-destructive-foreground", oklch(12, 0, 0))
	t.set("--color-success-foreground", oklch(12, 0, 0))
	t.set("--color-warning-foreground", oklch(clampL(12), neutralC(0.015), hue))
	t.set("--color-info-foreground", oklch(12, 0, 0))
	t.set("--color-success-subtle", oklch(20, actionC(0.05), 145))
	t.set("--color-warning-subtle", oklch(22, actionC(0.05), 75))
	t.set("--color-error-subtle", oklch(20, actionC(0.05), 25))
	t.set("--color-info-subtle", oklch(20, actionC(0.04), hue))
}
