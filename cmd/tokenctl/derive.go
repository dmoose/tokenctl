// tokenctl/cmd/tokenctl/derive.go
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dmoose/tokenctl/pkg/derive"
	"github.com/spf13/cobra"
)

var deriveCmd = &cobra.Command{
	Use:   "derive",
	Short: "Derive a semantic theme from a preset or explicit controls",
	Long: `Derive a full semantic token set from a handful of controls.

A theme is described by a hue and chroma plus four dials: how much of the
hue bleeds into the neutrals (tint), a global chroma multiplier
(saturation), light or dark mode, and a typography system with a density
scale. From those, derive produces every semantic colour token along with
the typography and density artifacts.

Start from a built-in preset, from a brand colour, or from explicit
values:

  tokenctl derive --preset=blue
  tokenctl derive --preset=teal --dark --density=115
  tokenctl derive --from-hex=#3b6de0 --tint=45
  tokenctl derive --hue=290 --chroma=0.18 --saturation=120

Output formats:
  json   W3C Design Tokens document, ready for tokenctl build (default)
  css    A custom-property block

Density scales the dimension tokens by density/100. Where a site
persists that scale is out of scope for this command — derive computes
the artifacts and writes them; it does not decide where they live.

Run with --list to see the presets and typography systems.`,
	Args: cobra.NoArgs,
	RunE: runDerive,
}

var (
	derivePreset      string
	deriveFromHex     string
	deriveHue         float64
	deriveChroma      float64
	deriveDark        bool
	deriveTint        float64
	deriveSaturation  float64
	deriveFontPairing string
	deriveDensity     float64
	deriveFormat      string
	deriveOutput      string
	deriveSelector    string
	deriveLayer       string
	deriveList        bool
)

func init() {
	f := deriveCmd.Flags()
	f.StringVar(&derivePreset, "preset", "", "Start from a built-in preset (see --list)")
	f.StringVar(&deriveFromHex, "from-hex", "", "Start from a brand colour, e.g. #3b6de0")
	f.Float64Var(&deriveHue, "hue", derive.DefaultParams.Hue, "OKLCH hue, 0-360")
	f.Float64Var(&deriveChroma, "chroma", derive.DefaultParams.Chroma, "OKLCH chroma of the primary")
	f.BoolVar(&deriveDark, "dark", false, "Derive the dark-mode variant")
	f.Float64Var(&deriveTint, "tint", derive.DefaultParams.Tint, "How much hue reaches the neutrals, 0-100")
	f.Float64Var(&deriveSaturation, "saturation", derive.DefaultParams.Saturation, "Global chroma multiplier, 0-150 (100 = normal)")
	f.StringVar(&deriveFontPairing, "type", derive.DefaultParams.FontPairing, "Typography system key (see --list)")
	f.Float64Var(&deriveDensity, "density", derive.DefaultParams.Density, "Dimension scale, 75-130 (100 = default)")
	f.StringVarP(&deriveFormat, "format", "f", "json", "Output format (json, css)")
	f.StringVarP(&deriveOutput, "output", "o", "", "Write to this file instead of stdout")
	f.StringVar(&deriveSelector, "selector", ":root", "CSS selector for --format=css")
	f.StringVar(&deriveLayer, "layer", "semantic", "$layer written into the JSON document (empty to omit)")
	f.BoolVar(&deriveList, "list", false, "List the built-in presets and typography systems")

	deriveCmd.MarkFlagsMutuallyExclusive("preset", "from-hex")

	rootCmd.AddCommand(deriveCmd)
}

func runDerive(cmd *cobra.Command, _ []string) error {
	if deriveList {
		printDeriveCatalog(cmd.OutOrStdout())
		return nil
	}

	params, err := resolveDeriveParams(cmd)
	if err != nil {
		return err
	}
	if err := params.Validate(); err != nil {
		return err
	}

	theme := derive.Generate(params)

	var content []byte
	switch deriveFormat {
	case "json":
		content, err = theme.ToTokenJSON(deriveLayer)
		if err != nil {
			return fmt.Errorf("rendering token JSON: %w", err)
		}
	case "css":
		selector := deriveSelector
		// A dark theme written to :root would override the light one it
		// is meant to sit beside, so give it the attribute selector
		// tokenctl's own theme output uses.
		if deriveDark && !cmd.Flags().Changed("selector") {
			selector = `[data-theme="dark"]`
		}
		content = []byte(theme.ToCSS(selector))
	default:
		return fmt.Errorf("unknown format: %s (valid: json, css)", deriveFormat)
	}

	if deriveOutput == "" {
		_, err := cmd.OutOrStdout().Write(content)
		return err
	}

	if dir := filepath.Dir(deriveOutput); dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("creating output dir: %w", err)
		}
	}
	if err := os.WriteFile(deriveOutput, content, 0o644); err != nil {
		return fmt.Errorf("writing output: %w", err)
	}
	_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Derived %d tokens into %s\n", theme.Len(), deriveOutput)
	return nil
}

// resolveDeriveParams builds the parameter set from the flags, layering
// an explicit --hue/--chroma over whatever a preset or brand colour set.
func resolveDeriveParams(cmd *cobra.Command) (derive.Params, error) {
	params := derive.DefaultParams

	switch {
	case derivePreset != "":
		preset, ok := derive.PresetByName(derivePreset)
		if !ok {
			return params, fmt.Errorf("unknown preset %q (valid: %s)",
				derivePreset, strings.Join(derive.PresetNames(), ", "))
		}
		params = derive.FromPreset(preset)
	case deriveFromHex != "":
		p, err := derive.ParamsFromHex(deriveFromHex)
		if err != nil {
			return params, err
		}
		params = p
	}

	// Explicit flags win over the preset or brand colour they follow.
	flags := cmd.Flags()
	if flags.Changed("hue") || (derivePreset == "" && deriveFromHex == "") {
		params.Hue = deriveHue
	}
	if flags.Changed("chroma") || (derivePreset == "" && deriveFromHex == "") {
		params.Chroma = deriveChroma
	}

	params.IsDark = deriveDark
	params.Tint = deriveTint
	params.Saturation = deriveSaturation
	params.FontPairing = deriveFontPairing
	params.Density = deriveDensity

	return params, nil
}

func printDeriveCatalog(w interface{ Write([]byte) (int, error) }) {
	_, _ = fmt.Fprintln(w, "Presets:")
	for _, p := range derive.Presets {
		_, _ = fmt.Fprintf(w, "  %-8s hue %-5g chroma %-5g %s\n",
			strings.ToLower(p.Name), p.Hue, p.Chroma, p.Swatch)
	}
	_, _ = fmt.Fprintln(w, "\nTypography systems:")
	for _, s := range derive.TypographySystems {
		_, _ = fmt.Fprintf(w, "  %-14s %-18s %s\n", s.Key, s.Label, s.Description)
	}
	_, _ = fmt.Fprintln(w, "\nRanges:")
	_, _ = fmt.Fprintf(w, "  tint        %g-%g\n", derive.TintMin, derive.TintMax)
	_, _ = fmt.Fprintf(w, "  saturation  %g-%g\n", derive.SaturationMin, derive.SaturationMax)
	_, _ = fmt.Fprintf(w, "  density     %g-%g\n", derive.DensityMin, derive.DensityMax)
}
