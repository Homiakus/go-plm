package standardparts

import (
	"testing"

	"github.com/Homiakus/go-plm/internal/core/object"
)

func TestNormalizeFastener(t *testing.T) {
	meta := map[string]any{
		"standard":       "ISO 4017",
		"thread":         "M6",
		"length_mm":      30,
		"material_grade": "A2",
	}
	np, err := Normalize(Fst, meta)
	if err != nil {
		t.Fatal(err)
	}
	if np.Key != "fst:iso4017:m6:30:a2" {
		t.Errorf("key = %q, want fst:iso4017:m6:30:a2", np.Key)
	}
}

func TestNormalizeBearing(t *testing.T) {
	meta := map[string]any{
		"bearing_series": "6205",
		"seal_type":      "2RS",
		"clearance":      "C3",
	}
	np, err := Normalize(Brg, meta)
	if err != nil {
		t.Fatal(err)
	}
	if np.Key != "brg:6205:2rs:c3" {
		t.Errorf("key = %q", np.Key)
	}
}

func TestNormalizeElectronics(t *testing.T) {
	meta := map[string]any{
		"component_type": "res",
		"value":          "10k",
		"package":        "0603",
	}
	np, err := Normalize(Elc, meta)
	if err != nil {
		t.Fatal(err)
	}
	if np.Key != "elc:res:10k:0603" {
		t.Errorf("key = %q", np.Key)
	}
}

func TestExactDuplicate(t *testing.T) {
	meta := map[string]any{"standard": "ISO4017", "thread": "M6", "length_mm": 30, "material_grade": "A2"}
	np1, _ := Normalize(Fst, meta)
	np2, _ := Normalize(Fst, meta)

	candidates := FindDuplicates(np1, []NormalizedPart{np2})
	if len(candidates) != 1 || candidates[0].Level != ExactDuplicate {
		t.Errorf("expected exact duplicate, got %+v", candidates)
	}
}

func TestValidateStandardPart(t *testing.T) {
	// Valid
	obj := object.Object{
		ID: "a-std-fst-iso4017-m6x30-v1.0", Class: object.ClassStandardPart,
		Title: "Bolt", Metadata: map[string]any{"std_class": "fst"},
	}
	if diags := Validate(obj); len(diags) != 0 {
		t.Errorf("expected no diags, got %d", len(diags))
	}

	// No metadata
	obj.Metadata = nil
	if diags := Validate(obj); len(diags) == 0 {
		t.Error("expected diags for nil metadata")
	}

	// Not a std part
	obj.Class = object.ClassPart
	if diags := Validate(obj); len(diags) != 0 {
		t.Error("non-std part should skip validation")
	}
}
