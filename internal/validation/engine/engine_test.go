package engine

import (
	"context"
	"testing"

	"github.com/Homiakus/go-plm/internal/core/diagnostic"
	"github.com/Homiakus/go-plm/internal/core/object"
)

type testRule struct {
	id    string
	diags []diagnostic.Diagnostic
}

func (r *testRule) ID() string { return r.id }
func (r *testRule) Check(ctx context.Context, target Target) ([]diagnostic.Diagnostic, error) {
	return r.diags, nil
}

func TestEngineValidateObject(t *testing.T) {
	reg := NewRegistry()
	reg.Register(&testRule{
		id: "test_rule",
		diags: []diagnostic.Diagnostic{
			{ID: "d1", Severity: diagnostic.Blocker, Code: "ERR", Message: "test error"},
		},
	})

	eng := NewEngine(reg)
	obj := object.Object{ID: "demo-prt-0001-v1.0", Class: object.ClassPart, Title: "Test"}

	diags, err := eng.ValidateObject(context.Background(), obj, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(diags))
	}
	if diags[0].Code != "ERR" {
		t.Errorf("code = %q", diags[0].Code)
	}
}

func TestRegistryGet(t *testing.T) {
	reg := NewRegistry()
	rule := &testRule{id: "test"}
	reg.Register(rule)

	got, ok := reg.Get("test")
	if !ok {
		t.Fatal("rule not found")
	}
	if got.ID() != "test" {
		t.Errorf("ID = %q", got.ID())
	}

	_, ok = reg.Get("nonexistent")
	if ok {
		t.Error("should not find nonexistent rule")
	}
}
