// Package guards_test provides tests for FSM guard implementations.
package guards_test

import (
	"context"
	"testing"

	"github.com/Homiakus/go-plm/internal/process/guards"
)

func TestValidName_Valid(t *testing.T) {
	g := &guards.ValidName{}
	diags, err := g.Check(context.Background(), guards.GuardInput{
		ObjectID: "demo-prt-0001-v1.0",
		Class:    "prt",
		State:    "draft",
	})
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if len(diags) != 0 {
		t.Errorf("expected 0 diags, got %d: %v", len(diags), diags)
	}
}

func TestValidName_Invalid(t *testing.T) {
	g := &guards.ValidName{}
	diags, err := g.Check(context.Background(), guards.GuardInput{
		ObjectID: "BAD-ID",
		Class:    "prt",
		State:    "draft",
	})
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if len(diags) == 0 {
		t.Error("expected diagnostics for invalid name")
	}
	if diags[0].Code != "INVALID_NAME" {
		t.Errorf("code = %q, want INVALID_NAME", diags[0].Code)
	}
}

func TestValidName_ID(t *testing.T) {
	g := &guards.ValidName{}
	if g.ID() != "valid_name" {
		t.Errorf("ID() = %q, want %q", g.ID(), "valid_name")
	}
}

func TestRequiredMetadata_Present(t *testing.T) {
	g := &guards.RequiredMetadata{
		Required: map[string][]string{
			"prt": {"unit", "make_buy"},
		},
	}
	diags, err := g.Check(context.Background(), guards.GuardInput{
		ObjectID: "demo-prt-0001-v1.0",
		Class:    "prt",
		State:    "draft",
		Metadata: map[string]any{"unit": "pcs", "make_buy": "make"},
	})
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if len(diags) != 0 {
		t.Errorf("expected 0 diags, got %d", len(diags))
	}
}

func TestRequiredMetadata_Missing(t *testing.T) {
	g := &guards.RequiredMetadata{
		Required: map[string][]string{
			"prt": {"unit", "make_buy"},
		},
	}
	diags, err := g.Check(context.Background(), guards.GuardInput{
		ObjectID: "demo-prt-0001-v1.0",
		Class:    "prt",
		State:    "draft",
		Metadata: map[string]any{"unit": "pcs"},
	})
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if len(diags) != 1 {
		t.Errorf("expected 1 diagnostic, got %d", len(diags))
	}
	if diags[0].Code != "REQUIRED_METADATA_MISSING" {
		t.Errorf("code = %q", diags[0].Code)
	}
}

func TestRequiredMetadata_NoRequirements(t *testing.T) {
	g := &guards.RequiredMetadata{
		Required: map[string][]string{},
	}
	diags, err := g.Check(context.Background(), guards.GuardInput{
		ObjectID: "demo-asm-0100-v1.0",
		Class:    "asm",
		State:    "draft",
		Metadata: map[string]any{},
	})
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if len(diags) != 0 {
		t.Errorf("expected 0 diags, got %d", len(diags))
	}
}

func TestNoBlockingIssues_NilValidator(t *testing.T) {
	g := &guards.NoBlockingIssues{Validator: nil}
	diags, err := g.Check(context.Background(), guards.GuardInput{
		ObjectID: "demo-prt-0001-v1.0",
	})
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if len(diags) != 0 {
		t.Errorf("expected 0 diags with nil validator, got %d", len(diags))
	}
}

func TestDefaultGuards(t *testing.T) {
	defs := guards.DefaultGuards()
	expected := []string{"valid_name", "required_metadata"}
	for _, name := range expected {
		if _, ok := defs[name]; !ok {
			t.Errorf("DefaultGuards missing %q", name)
		}
	}
}
