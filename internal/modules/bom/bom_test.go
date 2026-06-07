package bom

import (
	"context"
	"testing"

	"github.com/Homiakus/go-plm/internal/core/object"
)

type mockObjects map[string]object.Object
type mockRelations map[string][]object.RelationRef

func (m mockObjects) GetObject(ctx context.Context, id object.ID) (object.Object, error) {
	return m[string(id)], nil
}

func (m mockRelations) ListRelations(ctx context.Context, fromID object.ID) ([]object.RelationRef, error) {
	return m[string(fromID)], nil
}

func TestGetStructured(t *testing.T) {
	objects := mockObjects{
		"a-asm-0100-v1.0": {ID: "a-asm-0100-v1.0", Class: object.ClassAssembly, Title: "Motor"},
		"a-prt-0001-v1.0": {ID: "a-prt-0001-v1.0", Class: object.ClassPart, Title: "Bracket", Metadata: map[string]any{"make_buy": "make"}},
		"a-prt-0002-v1.0": {ID: "a-prt-0002-v1.0", Class: object.ClassPart, Title: "Shaft"},
	}
	qty2 := 2.0
	relations := mockRelations{
		"a-asm-0100-v1.0": {
			{ToID: "a-prt-0001-v1.0", Type: "contains", Quantity: &qty2, Unit: "pcs"},
			{ToID: "a-prt-0002-v1.0", Type: "contains", Quantity: nil, Unit: "pcs"},
		},
	}

	svc := New(objects, relations)
	rows, err := svc.GetStructured(context.Background(), "a-asm-0100-v1.0")
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}
	if rows[0].ChildID != "a-prt-0001-v1.0" || rows[0].Quantity != 2.0 {
		t.Errorf("row 0: child=%s qty=%.1f", rows[0].ChildID, rows[0].Quantity)
	}
}

func TestGetStructuredNested(t *testing.T) {
	objects := mockObjects{
		"a-asm-0100-v1.0": {ID: "a-asm-0100-v1.0", Class: object.ClassAssembly, Title: "Top"},
		"a-asm-0101-v1.0": {ID: "a-asm-0101-v1.0", Class: object.ClassAssembly, Title: "Sub"},
		"a-prt-0001-v1.0": {ID: "a-prt-0001-v1.0", Class: object.ClassPart, Title: "Screw"},
	}
	relations := mockRelations{
		"a-asm-0100-v1.0": {
			{ToID: "a-asm-0101-v1.0", Type: "contains"},
		},
		"a-asm-0101-v1.0": {
			{ToID: "a-prt-0001-v1.0", Type: "contains"},
		},
	}

	svc := New(objects, relations)
	rows, err := svc.GetStructured(context.Background(), "a-asm-0100-v1.0")
	if err != nil {
		t.Fatal(err)
	}
	// Sub-assembly + screw = 2 rows
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}
	if rows[1].ChildID != "a-prt-0001-v1.0" {
		t.Errorf("leaf = %s", rows[1].ChildID)
	}
	if rows[1].Level != 1 {
		t.Errorf("level = %d", rows[1].Level)
	}
}

func TestDetectCycle(t *testing.T) {
	objects := mockObjects{
		"a-asm-0100-v1.0": {ID: "a-asm-0100-v1.0", Class: object.ClassAssembly},
		"a-prt-0001-v1.0": {ID: "a-prt-0001-v1.0", Class: object.ClassAssembly},
	}
	relations := mockRelations{
		"a-asm-0100-v1.0": {{ToID: "a-prt-0001-v1.0", Type: "contains"}},
		"a-prt-0001-v1.0": {{ToID: "a-asm-0100-v1.0", Type: "contains"}},
	}

	svc := New(objects, relations)
	diags, err := svc.DetectCycles(context.Background(), "a-asm-0100-v1.0")
	if err != nil {
		t.Fatal(err)
	}
	if len(diags) == 0 {
		t.Error("expected cycle detection")
	}
}

func TestGetStructuredAllowsSharedSubassembly(t *testing.T) {
	objects := mockObjects{
		"a-asm-0100-v1.0": {ID: "a-asm-0100-v1.0", Class: object.ClassAssembly},
		"a-asm-0101-v1.0": {ID: "a-asm-0101-v1.0", Class: object.ClassAssembly},
		"a-asm-0102-v1.0": {ID: "a-asm-0102-v1.0", Class: object.ClassAssembly},
		"a-prt-0001-v1.0": {ID: "a-prt-0001-v1.0", Class: object.ClassPart},
	}
	relations := mockRelations{
		"a-asm-0100-v1.0": {
			{ToID: "a-asm-0101-v1.0", Type: "contains"},
			{ToID: "a-asm-0102-v1.0", Type: "contains"},
		},
		"a-asm-0101-v1.0": {{ToID: "a-prt-0001-v1.0", Type: "contains"}},
		"a-asm-0102-v1.0": {{ToID: "a-prt-0001-v1.0", Type: "contains"}},
	}

	rows, err := New(objects, relations).GetStructured(context.Background(), "a-asm-0100-v1.0")
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 4 {
		t.Fatalf("expected shared subassembly DAG to produce 4 rows, got %d", len(rows))
	}
}

func TestGetFlat(t *testing.T) {
	qty2 := 2.0
	qty3 := 3.0
	objects := mockObjects{
		"a-asm-0100-v1.0": {ID: "a-asm-0100-v1.0", Class: object.ClassAssembly},
		"a-prt-0001-v1.0": {ID: "a-prt-0001-v1.0", Class: object.ClassPart},
	}
	relations := mockRelations{
		"a-asm-0100-v1.0": {
			{ToID: "a-prt-0001-v1.0", Type: "contains", Quantity: &qty2, Unit: "pcs"},
			{ToID: "a-prt-0001-v1.0", Type: "contains", Quantity: &qty3, Unit: "pcs"},
		},
	}

	svc := New(objects, relations)
	flat, err := svc.GetFlat(context.Background(), "a-asm-0100-v1.0")
	if err != nil {
		t.Fatal(err)
	}
	if len(flat) != 1 {
		t.Fatalf("expected 1 flat row, got %d", len(flat))
	}
	if flat[0].Quantity != 5.0 {
		t.Errorf("rolled-up qty = %.1f, want 5.0", flat[0].Quantity)
	}
}

func TestValidate(t *testing.T) {
	svc := &Service{}
	rows := []BOMRow{
		{ChildID: "a", Quantity: 0, Unit: ""},
		{ChildID: "b", Quantity: 5, Unit: "pcs"},
	}
	diags := svc.Validate(rows)
	if len(diags) != 2 {
		t.Errorf("expected 2 diagnostics, got %d", len(diags))
	}
}
