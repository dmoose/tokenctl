package derive

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"testing"
)

type colorGolden struct {
	Cases []struct {
		Hex       string  `json:"hex"`
		L         float64 `json:"l"`
		C         float64 `json:"c"`
		H         float64 `json:"h"`
		RoundTrip string  `json:"roundTrip"`
	} `json:"cases"`
}

// The conversion is compared two ways. The emitted strings — what
// actually reaches a stylesheet — must match exactly. The raw
// components are additionally held to 1e-9, which is far tighter than
// any rounding boundary and would catch a matrix transcription error
// that happened to round the same way on these sixteen inputs.
func TestColorConvert_MatchesTypeScript(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile(filepath.Join(goldensDir, "_colorconvert.json"))
	if err != nil {
		t.Fatalf("read colour goldens: %v (run tools/derive-goldens/regenerate.sh)", err)
	}
	var g colorGolden
	if err := json.Unmarshal(data, &g); err != nil {
		t.Fatalf("parse colour goldens: %v", err)
	}
	if len(g.Cases) == 0 {
		t.Fatal("colour goldens are empty")
	}

	const tol = 1e-9

	for _, c := range g.Cases {
		t.Run(c.Hex, func(t *testing.T) {
			t.Parallel()

			l, ch, h, err := HexToOklchParts(c.Hex)
			if err != nil {
				t.Fatalf("HexToOklchParts(%s): %v", c.Hex, err)
			}

			// Formatted output: exact.
			if got, want := toFixed(l, 1), toFixed(c.L, 1); got != want {
				t.Errorf("lightness formats to %s, want %s", got, want)
			}
			if got, want := toFixed(ch, 3), toFixed(c.C, 3); got != want {
				t.Errorf("chroma formats to %s, want %s", got, want)
			}
			if got, want := toFixed(h, 0), toFixed(c.H, 0); got != want {
				t.Errorf("hue formats to %s, want %s", got, want)
			}

			// Components: tight tolerance for float noise only.
			if math.Abs(l-c.L) > tol {
				t.Errorf("lightness = %.12f, want %.12f (Δ%.2e)", l, c.L, l-c.L)
			}
			if math.Abs(ch-c.C) > tol {
				t.Errorf("chroma = %.12f, want %.12f (Δ%.2e)", ch, c.C, ch-c.C)
			}
			if math.Abs(h-c.H) > tol {
				t.Errorf("hue = %.12f, want %.12f (Δ%.2e)", h, c.H, h-c.H)
			}

			if got := OklchToHex(l, ch, h); got != c.RoundTrip {
				t.Errorf("round trip = %s, want %s", got, c.RoundTrip)
			}
		})
	}
}

// A greyscale colour has no hue. colorjs.io reports NaN and the engine
// coerced that to 0; if the port instead let rounding noise pick an
// angle, every neutral in the theme would drift toward it.
func TestColorConvert_GreyHasZeroHue(t *testing.T) {
	t.Parallel()

	for _, hex := range []string{"#000000", "#808080", "#ffffff", "#333333"} {
		_, c, h, err := HexToOklchParts(hex)
		if err != nil {
			t.Fatalf("%s: %v", hex, err)
		}
		if h != 0 {
			t.Errorf("%s: hue = %v, want 0 (chroma %v)", hex, h, c)
		}
	}
}

func TestParamsFromHex_FloorsChroma(t *testing.T) {
	t.Parallel()

	// Grey has essentially no chroma; the primary still needs enough to
	// be a colour, so the engine floors it at 0.08.
	p, err := ParamsFromHex("#808080")
	if err != nil {
		t.Fatalf("ParamsFromHex: %v", err)
	}
	if p.Chroma != 0.08 {
		t.Errorf("chroma = %v, want the 0.08 floor", p.Chroma)
	}

	// A saturated colour keeps its own chroma.
	p, err = ParamsFromHex("#3b6de0")
	if err != nil {
		t.Fatalf("ParamsFromHex: %v", err)
	}
	if p.Chroma <= 0.08 {
		t.Errorf("chroma = %v, want the measured chroma, not the floor", p.Chroma)
	}
}

func TestParseHex_Forms(t *testing.T) {
	t.Parallel()

	long, err := HexToOklchPartsAll(t, "#3b6de0")
	if err != nil {
		t.Fatal(err)
	}
	for _, form := range []string{"3b6de0", "#3B6DE0", "  #3b6de0  "} {
		got, err := HexToOklchPartsAll(t, form)
		if err != nil {
			t.Fatalf("%q: %v", form, err)
		}
		if got != long {
			t.Errorf("%q parsed to %v, want %v", form, got, long)
		}
	}

	if _, _, _, err := HexToOklchParts("#xyzxyz"); err == nil {
		t.Error("invalid hex should error, not fall back silently")
	}
	if _, _, _, err := HexToOklchParts("#1234"); err == nil {
		t.Error("four-digit hex should error")
	}

	// The three-digit shorthand expands.
	short, err := HexToOklchPartsAll(t, "#f00")
	if err != nil {
		t.Fatal(err)
	}
	full, err := HexToOklchPartsAll(t, "#ff0000")
	if err != nil {
		t.Fatal(err)
	}
	if short != full {
		t.Errorf("#f00 = %v, want it to equal #ff0000 = %v", short, full)
	}
}

func HexToOklchPartsAll(t *testing.T, hex string) ([3]float64, error) {
	t.Helper()
	l, c, h, err := HexToOklchParts(hex)
	return [3]float64{l, c, h}, err
}
