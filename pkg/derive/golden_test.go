package derive

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"testing"
)

// goldensDir holds fixtures produced by running the original TypeScript
// engine (tools/derive-goldens/regenerate.sh). This test only ever reads
// them: regenerating is a deliberate act, not a side effect of testing,
// so a Go-side regression cannot quietly rewrite its own reference.
const goldensDir = "../../testdata/derive/goldens"

type goldenCase struct {
	Name   string `json:"name"`
	Params struct {
		Hue         float64 `json:"hue"`
		Chroma      float64 `json:"chroma"`
		IsDark      bool    `json:"isDark"`
		Tint        float64 `json:"tint"`
		Saturation  float64 `json:"saturation"`
		FontPairing string  `json:"fontPairing"`
		Density     float64 `json:"density"`
	} `json:"params"`
	Tokens map[string]string `json:"tokens"`
}

func (g goldenCase) toParams() Params {
	return Params{
		Hue:         g.Params.Hue,
		Chroma:      g.Params.Chroma,
		IsDark:      g.Params.IsDark,
		Tint:        g.Params.Tint,
		Saturation:  g.Params.Saturation,
		FontPairing: g.Params.FontPairing,
		Density:     g.Params.Density,
	}
}

func loadIndex(t *testing.T) []string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(goldensDir, "_index.json"))
	if err != nil {
		t.Fatalf("read golden index: %v (run tools/derive-goldens/regenerate.sh)", err)
	}
	var idx struct {
		Count int      `json:"count"`
		Cases []string `json:"cases"`
	}
	if err := json.Unmarshal(data, &idx); err != nil {
		t.Fatalf("parse golden index: %v", err)
	}
	if len(idx.Cases) != idx.Count {
		t.Fatalf("golden index is inconsistent: count %d, cases %d", idx.Count, len(idx.Cases))
	}
	if idx.Count == 0 {
		t.Fatal("golden index is empty")
	}
	return idx.Cases
}

// The Go port must reproduce the TypeScript engine's output exactly —
// same token names, same value strings, no tolerance. Values are
// formatted decimal strings, so an exact comparison is the honest one;
// anything looser would hide a real divergence in the math.
func TestGoldens_ExactAgreement(t *testing.T) {
	t.Parallel()

	for _, name := range loadIndex(t) {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			data, err := os.ReadFile(filepath.Join(goldensDir, name+".json"))
			if err != nil {
				t.Fatalf("read golden: %v", err)
			}
			var g goldenCase
			if err := json.Unmarshal(data, &g); err != nil {
				t.Fatalf("parse golden: %v", err)
			}

			params := g.toParams()
			if err := params.Validate(); err != nil {
				t.Fatalf("golden params rejected by Validate: %v", err)
			}

			got := Generate(params)

			if got.Len() != len(g.Tokens) {
				t.Errorf("token count = %d, want %d", got.Len(), len(g.Tokens))
			}

			wantNames := make([]string, 0, len(g.Tokens))
			for k := range g.Tokens {
				wantNames = append(wantNames, k)
			}
			sort.Strings(wantNames)

			for _, key := range wantNames {
				want := g.Tokens[key]
				have, ok := got.Get(key)
				if !ok {
					t.Errorf("missing token %s (want %q)", key, want)
					continue
				}
				if have != want {
					t.Errorf("%s = %q, want %q", key, have, want)
				}
			}

			for _, key := range got.Order {
				if _, ok := g.Tokens[key]; !ok {
					v, _ := got.Get(key)
					t.Errorf("extra token %s = %q not present in golden", key, v)
				}
			}
		})
	}
}

// Every preset must appear in the golden set in both modes; a preset
// added to Go without a corresponding golden would otherwise be untested.
func TestGoldens_CoverEveryPreset(t *testing.T) {
	t.Parallel()

	have := map[string]bool{}
	for _, name := range loadIndex(t) {
		have[name] = true
	}
	for _, p := range Presets {
		for _, mode := range []string{"light", "dark"} {
			want := "preset-" + lower(p.Name) + "-" + mode
			if !have[want] {
				t.Errorf("no golden for preset %s in %s mode (expected %s.json)", p.Name, mode, want)
			}
		}
	}
}

// Same for typography systems: each key needs an axis golden.
func TestGoldens_CoverEveryTypographySystem(t *testing.T) {
	t.Parallel()

	have := map[string]bool{}
	for _, name := range loadIndex(t) {
		have[name] = true
	}
	for _, key := range TypographyKeys() {
		for _, mode := range []string{"light", "dark"} {
			want := "axis-fontPairing-" + key + "-" + mode
			if !have[want] {
				t.Errorf("no golden for typography system %q in %s mode (expected %s.json)", key, mode, want)
			}
		}
	}
}

func lower(s string) string {
	b := []byte(s)
	for i := range b {
		if 'A' <= b[i] && b[i] <= 'Z' {
			b[i] += 'a' - 'A'
		}
	}
	return string(b)
}
