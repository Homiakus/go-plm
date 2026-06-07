package rules

import (
	"context"
	"testing"

	"github.com/Homiakus/go-plm/internal/core/diagnostic"
	"github.com/Homiakus/go-plm/internal/core/object"
	"github.com/Homiakus/go-plm/internal/validation/engine"
)

func TestValidName(t *testing.T) {
	rule := &ValidName{}
	target := engine.Target{
		Object: object.Object{ID: "demo-prt-0001-v1.0", Class: object.ClassPart, Title: "T"},
	}
	diags, err := rule.Check(context.Background(), target)
	if err != nil {
		t.Fatal(err)
	}
	if len(diags) != 0 {
		t.Errorf("expected no diagnostics for valid ID, got %d", len(diags))
	}
}

func TestValidNameInvalid(t *testing.T) {
	rule := &ValidName{}
	target := engine.Target{
		Object: object.Object{ID: "BAD_ID", Class: object.ClassPart, Title: "T"},
	}
	diags, _ := rule.Check(context.Background(), target)
	if len(diags) != 1 || diags[0].Severity != diagnostic.Blocker {
		t.Errorf("expected 1 blocker diagnostic, got %d", len(diags))
	}
}

func TestRequiredMetadata(t *testing.T) {
	rule := &RequiredMetadata{
		RequiredFields: map[string][]string{"prt": {"material"}},
	}

	// Missing field
	target := engine.Target{
		Object: object.Object{
			ID: "demo-prt-0001-v1.0", Class: object.ClassPart, Title: "T",
			Metadata: map[string]any{},
		},
	}
	diags, _ := rule.Check(context.Background(), target)
	if len(diags) != 1 || diags[0].Code != "REQUIRED_METADATA_MISSING" {
		t.Errorf("expected REQUIRED_METADATA_MISSING, got %d diags", len(diags))
	}

	// Field present
	target.Object.Metadata = map[string]any{"material": "al5052"}
	diags, _ = rule.Check(context.Background(), target)
	if len(diags) != 0 {
		t.Errorf("expected no diagnostics, got %d", len(diags))
	}

	// Unknown class — no checks
	target.Object.Class = "zzz"
	diags, _ = rule.Check(context.Background(), target)
	if len(diags) != 0 {
		t.Errorf("expected no diagnostics for unknown class, got %d", len(diags))
	}
}

func TestDefaultRegistry(t *testing.T) {
	reg := DefaultRegistry()
	if _, ok := reg.Get("valid_name"); !ok {
		t.Error("valid_name rule not registered")
	}
	if _, ok := reg.Get("required_metadata"); !ok {
		t.Error("required_metadata rule not registered")
	}
}
