package tokens

import (
	"strings"
	"testing"
)

func dictFromJSON(t *testing.T, s string) *Dictionary {
	t.Helper()
	d, err := ParseJSON(strings.NewReader(s))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	return d
}

func kinds(findings []Finding) map[FindingKind]int {
	out := map[FindingKind]int{}
	for _, f := range findings {
		out[f.Kind]++
	}
	return out
}

func paths(findings []Finding) []string {
	out := make([]string, 0, len(findings))
	for _, f := range findings {
		out = append(out, f.Path)
	}
	return out
}

// The production incident: a component wrote "ratios" where the schema
// says "variants". It parsed, it validated, it built — and five classes
// shipped with no declarations at all.
func TestAuditUnknownKeys_RatiosIncident(t *testing.T) {
	t.Parallel()

	d := dictFromJSON(t, `{
      "$layer": "component",
      "components": {
        "aspect-ratio": {
          "$type": "component",
          "$class": "aspect-ratio",
          "base": { "position": "relative" },
          "ratios": {
            "square": { "$class": "aspect-ratio-square", "aspect-ratio": "1 / 1" },
            "video":  { "$class": "aspect-ratio-video",  "aspect-ratio": "16 / 9" }
          }
        }
      }
    }`)

	findings := AuditUnknownKeys(d)
	if got := kinds(findings)[FindingUnknownComponentKey]; got != 1 {
		t.Fatalf("want 1 unknown-component-key finding, got %d (%v)", got, paths(findings))
	}
	var f Finding
	for _, cand := range findings {
		if cand.Kind == FindingUnknownComponentKey {
			f = cand
		}
	}
	if f.Path != "components.aspect-ratio.ratios" {
		t.Errorf("path = %q, want components.aspect-ratio.ratios", f.Path)
	}
	if !strings.Contains(f.Hint, "variants") {
		t.Errorf("hint should point at the real sub-blocks, got %q", f.Hint)
	}
}

func TestAuditUnknownKeys_CleanComponentIsSilent(t *testing.T) {
	t.Parallel()

	d := dictFromJSON(t, `{
      "$layer": "component",
      "components": {
        "aspect-ratio": {
          "$type": "component",
          "$class": "aspect-ratio",
          "$description": "fixed ratio box",
          "base": { "position": "relative", "&:hover": { "opacity": "0.9" } },
          "variants": {
            "square": { "$class": "aspect-ratio-square", "aspect-ratio": "1 / 1" }
          },
          "sizes":  { "sm": { "$class": "aspect-ratio-sm", "width": "4rem" } },
          "states": { "busy": { "$class": "aspect-ratio-busy", "opacity": "0.5" } }
        }
      }
    }`)

	if findings := AuditUnknownKeys(d); len(findings) != 0 {
		t.Errorf("clean component produced findings: %v", findings)
	}
}

// The token files carry schema notes as "// why" keys on purpose. The
// audit must never flag them, or every build turns noisy and the real
// findings get tuned out.
func TestAuditUnknownKeys_CommentKeysAreLegal(t *testing.T) {
	t.Parallel()

	d := dictFromJSON(t, `{
      "components": {
        "sr-only": {
          "$type": "component",
          "$class": "sr-only",
          "// why": "visually hidden but readable by AT",
          "base": { "position": "absolute" }
        }
      }
    }`)

	if findings := AuditUnknownKeys(d); len(findings) != 0 {
		t.Errorf("comment keys must not be findings, got %v", findings)
	}
}

func TestAuditUnknownKeys_UnknownMetadataKey(t *testing.T) {
	t.Parallel()

	d := dictFromJSON(t, `{
      "components": {
        "breadcrumb-link": {
          "$type": "component",
          "$class": "breadcrumb-link",
          "base": { "color": "red" },
          "states": {
            "hover": { "$class": "breadcrumb-link-hover", "$selector": ":hover", "color": "blue" }
          }
        }
      }
    }`)

	findings := AuditUnknownKeys(d)
	if got := kinds(findings)[FindingUnknownMetadataKey]; got != 1 {
		t.Fatalf("want 1 unknown-metadata-key finding, got %d (%v)", got, findings)
	}
	if p := paths(findings)[0]; !strings.HasSuffix(p, "$selector") {
		t.Errorf("finding path = %q, want it to end in $selector", p)
	}
}

// A states/variants/sizes entry without $class has no selector to hang
// its declarations on, so the generator drops the entry entirely.
func TestAuditUnknownKeys_EntryMissingClass(t *testing.T) {
	t.Parallel()

	d := dictFromJSON(t, `{
      "components": {
        "breadcrumb-link": {
          "$type": "component",
          "$class": "breadcrumb-link",
          "base": { "color": "red" },
          "states": { "hover": { "color": "blue" } },
          "variants": { "muted": { "$class": "breadcrumb-link-muted", "color": "gray" } }
        }
      }
    }`)

	findings := AuditUnknownKeys(d)
	if got := kinds(findings)[FindingMissingClass]; got != 1 {
		t.Fatalf("want 1 missing-class finding, got %d (%v)", got, findings)
	}
	for _, f := range findings {
		if f.Kind == FindingMissingClass && f.Path != "components.breadcrumb-link.states.hover" {
			t.Errorf("path = %q, want components.breadcrumb-link.states.hover", f.Path)
		}
	}
}

// base is a property bag, not a block of named entries — its children
// must never be checked for $class.
func TestAuditUnknownKeys_BaseIsNotAVariantBlock(t *testing.T) {
	t.Parallel()

	d := dictFromJSON(t, `{
      "components": {
        "card": {
          "$type": "component",
          "$class": "card",
          "base": { "padding": "1rem", "&:hover": { "opacity": "0.9" }, "& p": { "margin": "0" } }
        }
      }
    }`)

	if findings := AuditUnknownKeys(d); len(findings) != 0 {
		t.Errorf("base children must not be audited as entries, got %v", findings)
	}
}

// Plain token groups are not components; their keys are free-form and
// must not be flagged.
func TestAuditUnknownKeys_TokenGroupsAreFreeForm(t *testing.T) {
	t.Parallel()

	d := dictFromJSON(t, `{
      "$layer": "brand",
      "color":   { "primary": { "$value": "#3b82f6", "$type": "color" } },
      "spacing": { "md": { "$value": "1rem", "$type": "dimension" } },
      "ratios":  { "golden": { "$value": "1.618", "$type": "number" } }
    }`)

	if findings := AuditUnknownKeys(d); len(findings) != 0 {
		t.Errorf("token groups must not be flagged, got %v", findings)
	}
}

func TestAuditUnknownKeys_Deterministic(t *testing.T) {
	t.Parallel()

	src := `{
      "components": {
        "b": { "$type": "component", "$class": "b", "zzz": {}, "aaa": {} },
        "a": { "$type": "component", "$class": "a", "mmm": {}, "$bogus": 1 }
      }
    }`

	first := paths(AuditUnknownKeys(dictFromJSON(t, src)))
	for range 20 {
		got := paths(AuditUnknownKeys(dictFromJSON(t, src)))
		if len(got) != len(first) {
			t.Fatalf("finding count varies: %d vs %d", len(got), len(first))
		}
		for j := range got {
			if got[j] != first[j] {
				t.Fatalf("order varies at %d: %q vs %q", j, got[j], first[j])
			}
		}
	}
	if len(first) != 4 {
		t.Errorf("want 4 findings (3 unknown keys + 1 bogus metadata), got %d: %v", len(first), first)
	}
}

func TestAuditUnknownKeys_NilDictionary(t *testing.T) {
	t.Parallel()

	if findings := AuditUnknownKeys(nil); findings != nil {
		t.Errorf("nil dictionary should produce no findings, got %v", findings)
	}
}
