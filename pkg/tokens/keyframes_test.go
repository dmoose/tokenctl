// tokenctl/pkg/tokens/keyframes_test.go

package tokens

import (
	"strings"
	"testing"
)

func TestExtractKeyframes_Basic(t *testing.T) {
	t.Parallel()
	dict := &Dictionary{
		Root: map[string]any{
			"keyframes": map[string]any{
				"pulse": map[string]any{
					"0%, 100%": map[string]any{"opacity": "1"},
					"50%":      map[string]any{"opacity": "0.5"},
				},
			},
		},
	}

	keyframes := ExtractKeyframes(dict)

	if len(keyframes) != 1 {
		t.Fatalf("Expected 1 keyframe, got %d", len(keyframes))
	}

	kf := keyframes[0]
	if kf.Name != "pulse" {
		t.Errorf("Expected name 'pulse', got '%s'", kf.Name)
	}

	if len(kf.Frames) != 2 {
		t.Errorf("Expected 2 frames, got %d", len(kf.Frames))
	}

	if kf.Frames["0%, 100%"]["opacity"] != "1" {
		t.Error("Missing or wrong 0%, 100% frame")
	}
	if kf.Frames["50%"]["opacity"] != "0.5" {
		t.Error("Missing or wrong 50% frame")
	}
}

func TestExtractKeyframes_Multiple(t *testing.T) {
	t.Parallel()
	dict := &Dictionary{
		Root: map[string]any{
			"keyframes": map[string]any{
				"spin": map[string]any{
					"from": map[string]any{"transform": "rotate(0deg)"},
					"to":   map[string]any{"transform": "rotate(360deg)"},
				},
				"fade": map[string]any{
					"from": map[string]any{"opacity": "0"},
					"to":   map[string]any{"opacity": "1"},
				},
			},
		},
	}

	keyframes := ExtractKeyframes(dict)

	if len(keyframes) != 2 {
		t.Fatalf("Expected 2 keyframes, got %d", len(keyframes))
	}

	// Should be sorted alphabetically
	if keyframes[0].Name != "fade" {
		t.Errorf("First keyframe should be 'fade', got '%s'", keyframes[0].Name)
	}
	if keyframes[1].Name != "spin" {
		t.Errorf("Second keyframe should be 'spin', got '%s'", keyframes[1].Name)
	}
}

func TestExtractKeyframes_Empty(t *testing.T) {
	t.Parallel()
	dict := &Dictionary{
		Root: map[string]any{
			"color": map[string]any{
				"primary": "#000",
			},
		},
	}

	keyframes := ExtractKeyframes(dict)

	if len(keyframes) != 0 {
		t.Errorf("Expected 0 keyframes, got %d", len(keyframes))
	}
}

func TestGenerateKeyframesCSS(t *testing.T) {
	t.Parallel()
	keyframes := []KeyframeDefinition{
		{
			Name: "pulse",
			Frames: map[string]map[string]string{
				"0%, 100%": {"opacity": "1"},
				"50%":      {"opacity": "0.5"},
			},
		},
	}

	css := GenerateKeyframesCSS(keyframes)

	if !strings.Contains(css, "@keyframes pulse {") {
		t.Error("Missing @keyframes declaration")
	}
	if !strings.Contains(css, "0%, 100% {") {
		t.Error("Missing 0%, 100% selector")
	}
	if !strings.Contains(css, "opacity: 1;") {
		t.Error("Missing opacity: 1")
	}
}

func TestGenerateKeyframesCSS_FrameOrder(t *testing.T) {
	t.Parallel()
	keyframes := []KeyframeDefinition{
		{
			Name: "slide",
			Frames: map[string]map[string]string{
				"100%": {"left": "100%"},
				"0%":   {"left": "0"},
				"50%":  {"left": "50%"},
			},
		},
	}

	css := GenerateKeyframesCSS(keyframes)

	// Frames should be ordered by percentage
	idx0 := strings.Index(css, "0% {")
	idx50 := strings.Index(css, "50% {")
	idx100 := strings.Index(css, "100% {")

	if idx0 > idx50 || idx50 > idx100 {
		t.Error("Frames should be ordered by percentage (0% < 50% < 100%)")
	}
}

func TestGenerateKeyframesCSS_FromTo(t *testing.T) {
	t.Parallel()
	keyframes := []KeyframeDefinition{
		{
			Name: "fade",
			Frames: map[string]map[string]string{
				"to":   {"opacity": "1"},
				"from": {"opacity": "0"},
			},
		},
	}

	css := GenerateKeyframesCSS(keyframes)

	// from should come before to
	idxFrom := strings.Index(css, "from {")
	idxTo := strings.Index(css, "to {")

	if idxFrom > idxTo {
		t.Error("'from' should come before 'to'")
	}
}

func TestGenerateKeyframesCSS_Empty(t *testing.T) {
	t.Parallel()
	css := GenerateKeyframesCSS(nil)
	if css != "" {
		t.Error("Empty keyframes should produce empty string")
	}

	css = GenerateKeyframesCSS([]KeyframeDefinition{})
	if css != "" {
		t.Error("Empty keyframes slice should produce empty string")
	}
}
