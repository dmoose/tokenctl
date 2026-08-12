// tokenctl/pkg/derive/typography.go
//
// Typography systems and density scaling, ported from the engine's
// TYPOGRAPHY_SYSTEMS / BASE_DIMENSIONS / generateTypographyTokens.
package derive

// TypographySystem is a font pairing plus the leading and tracking that
// suit it. Zero-valued optional fields fall back to the engine defaults.
type TypographySystem struct {
	Key         string
	Label       string
	Description string
	Sans        string
	Mono        string
	Serif       string // optional

	LeadingBody     float64 // default 1.5
	LeadingHeading  float64 // default 1.25
	TrackingHeading string  // default "-0.025em"
	TrackingBody    string  // default "0em"
	TrackingWide    string  // default "0.05em"
}

// TypographySystems are the engine's ten built-ins, in its order. The
// first entry is the fallback for an unrecognised key.
var TypographySystems = []TypographySystem{
	{
		Key: "system", Label: "Modern SaaS",
		Description: "Inter + JetBrains Mono — clean, professional UI",
		Sans:        "Inter, system-ui, -apple-system, sans-serif",
		Mono:        `"JetBrains Mono", ui-monospace, monospace`,
	},
	{
		Key: "editorial", Label: "Editorial Duo",
		Description:    "Source Serif 4 / Source Sans 3 — cohesive editorial system",
		Sans:           `"Source Sans 3", system-ui, sans-serif`,
		Mono:           `"Source Code Pro", ui-monospace, monospace`,
		Serif:          `"Source Serif 4", Georgia, serif`,
		LeadingBody:    1.6,
		LeadingHeading: 1.3,

		TrackingHeading: "-0.02em",
	},
	{
		Key: "serif", Label: "Premium Serif",
		Description:    "Merriweather + iA Writer Mono — elegant, trustworthy",
		Sans:           "Merriweather, Georgia, serif",
		Mono:           "ui-monospace, monospace",
		LeadingBody:    1.65,
		LeadingHeading: 1.3,

		TrackingHeading: "-0.015em",
		TrackingBody:    "0.01em",
	},
	{
		Key: "devdocs", Label: "Developer Docs",
		Description:  "IBM Plex Sans + Fira Code — technical, clear",
		Sans:         `"IBM Plex Sans", system-ui, sans-serif`,
		Mono:         `"Fira Code", ui-monospace, monospace`,
		TrackingBody: "0.005em",
	},
	{
		Key: "zeroload", Label: "Zero-Load System",
		Description: "System fonts only — instant rendering, no downloads",
		Sans:        `system-ui, -apple-system, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif`,
		Mono:        `ui-monospace, "SF Mono", "Cascadia Code", "Segoe UI Mono", Menlo, Monaco, monospace`,
	},
	{
		Key: "geometric", Label: "Geometric Minimal",
		Description: "DM Sans + DM Mono — modern, approachable",
		Sans:        `"DM Sans", system-ui, sans-serif`,
		Mono:        `"DM Mono", ui-monospace, monospace`,
	},
	{
		Key: "warm", Label: "Warm Editorial",
		Description:    "Lora + Open Sans + Inconsolata — friendly, inviting",
		Sans:           `"Open Sans", system-ui, sans-serif`,
		Mono:           "Inconsolata, ui-monospace, monospace",
		Serif:          "Lora, Georgia, serif",
		LeadingBody:    1.6,
		LeadingHeading: 1.3,

		TrackingHeading: "-0.015em",
	},
	{
		Key: "swiss", Label: "Swiss Precision",
		Description:  "Helvetica stack + Roboto Mono — corporate, precise",
		Sans:         `"Helvetica Neue", Helvetica, Arial, sans-serif`,
		Mono:         `"Roboto Mono", ui-monospace, monospace`,
		TrackingBody: "0.005em",
		TrackingWide: "0.06em",
	},
	{
		Key: "retro", Label: "Retro-Tech",
		Description:  "Space Grotesk + Space Mono — nostalgic, quirky",
		Sans:         `"Space Grotesk", system-ui, sans-serif`,
		Mono:         `"Space Mono", ui-monospace, monospace`,
		TrackingBody: "0.01em",
		TrackingWide: "0.06em",
	},
	{
		Key: "contemporary", Label: "Contemporary",
		Description:     "Satoshi fallback to system — bold, expressive",
		Sans:            "Satoshi, system-ui, -apple-system, sans-serif",
		Mono:            "Iosevka, ui-monospace, monospace",
		TrackingHeading: "-0.03em",
	},
}

// TypographySystemByKey looks a system up by its exact key.
func TypographySystemByKey(key string) (TypographySystem, bool) {
	for _, s := range TypographySystems {
		if s.Key == key {
			return s, true
		}
	}
	return TypographySystem{}, false
}

// TypographyKeys lists the system keys in declaration order.
func TypographyKeys() []string {
	keys := make([]string, len(TypographySystems))
	for i, s := range TypographySystems {
		keys[i] = s.Key
	}
	return keys
}

// dimension is one density-scaled rem value at density 100.
type dimension struct {
	name string
	base float64
}

// baseDimensions is the engine's BASE_DIMENSIONS, kept as an ordered
// slice because the emitted order is part of the golden comparison.
var baseDimensions = []dimension{
	{"--spacing-xs", 0.25},
	{"--spacing-sm", 0.5},
	{"--spacing-md", 1.0},
	{"--spacing-lg", 1.5},
	{"--radius-sm", 0.125},
	{"--radius-md", 0.375},
	{"--radius-lg", 0.5},
	{"--font-size-xs", 0.75},
	{"--font-size-sm", 0.875},
	{"--font-size-base", 1.0},
	{"--font-size-lg", 1.125},
	{"--font-size-xl", 1.25},
	{"--font-size-2xl", 1.5},
	{"--font-size-3xl", 2.0},
	{"--font-size-4xl", 2.5},
	{"--font-size-h1", 2.0},
	{"--font-size-h2", 1.5},
	{"--font-size-h3", 1.25},
	{"--font-size-h4", 1.125},
}

// generateTypographyTokens produces the font, dimension, leading,
// tracking and weight tokens for a pairing at a density.
//
// Density scales every dimension token by density/100. Where the scaled
// values are *persisted* on a site is deliberately not decided here —
// this function computes the artifacts and nothing more.
func generateTypographyTokens(fontPairing string, density float64, out *Theme) {
	system, ok := TypographySystemByKey(fontPairing)
	if !ok {
		system = TypographySystems[0]
	}
	scale := density / 100

	out.set("--font-family-sans", system.Sans)
	out.set("--font-family-mono", system.Mono)

	for _, d := range baseDimensions {
		out.set(d.name, trimFixed(d.base*scale, 3)+"rem")
	}

	bodyLeading := 1.5
	if system.LeadingBody != 0 {
		bodyLeading = system.LeadingBody
	}
	headingLeading := 1.25
	if system.LeadingHeading != 0 {
		headingLeading = system.LeadingHeading
	}

	out.set("--leading-none", "1")
	out.set("--leading-tight", numberToString(headingLeading))
	out.set("--leading-snug", trimFixed((headingLeading+bodyLeading)/2, 3))
	out.set("--leading-normal", numberToString(bodyLeading))
	out.set("--leading-relaxed", trimFixed(bodyLeading+0.125, 3))

	out.set("--tracking-tight", orDefault(system.TrackingHeading, "-0.025em"))
	out.set("--tracking-snug", "-0.02em")
	out.set("--tracking-normal", orDefault(system.TrackingBody, "0em"))
	out.set("--tracking-wide", "0.025em")
	out.set("--tracking-wider", orDefault(system.TrackingWide, "0.05em"))

	out.set("--font-weight-normal", "400")
	out.set("--font-weight-medium", "500")
	out.set("--font-weight-semibold", "600")
	out.set("--font-weight-bold", "700")
}

func orDefault(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}
