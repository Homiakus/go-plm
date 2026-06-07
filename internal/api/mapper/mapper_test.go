// Package mapper_test provides tests for domain ↔ DTO conversion.
package mapper_test

import (
	"testing"

	"github.com/Homiakus/go-plm/internal/api/mapper"
	"github.com/Homiakus/go-plm/internal/core/object"
	"github.com/Homiakus/go-plm/internal/modules/bom"
	"github.com/Homiakus/go-plm/internal/modules/release"
)

func TestObjectToDTO(t *testing.T) {
	obj := object.Object{
		ID:       "demo-prt-0001-v1.0",
		Project:  "demo",
		Class:    object.ClassPart,
		Sequence: "0001",
		Version:  "1.0",
		Revision: "1.0",
		State:    object.StateDraft,
		Title:    "Bracket",
		Metadata: map[string]any{"unit": "pcs"},
	}
	dto := mapper.ObjectToDTO(obj)
	if dto.ID != "demo-prt-0001-v1.0" {
		t.Errorf("ID = %q", dto.ID)
	}
	if dto.Class != "prt" {
		t.Errorf("Class = %q", dto.Class)
	}
	if dto.State != "draft" {
		t.Errorf("State = %q", dto.State)
	}
	if dto.Title != "Bracket" {
		t.Errorf("Title = %q", dto.Title)
	}
}

func TestDTOToObject(t *testing.T) {
	dto := mapper.DTOToObject(mapper.ObjectToDTO(object.Object{
		ID: "demo-asm-0100-v1.0", Project: "demo", Class: object.ClassAssembly,
		Title: "Motor", State: object.StateDraft,
	}))
	if dto.ID != "demo-asm-0100-v1.0" {
		t.Errorf("ID = %q", dto.ID)
	}
	if dto.Class != object.ClassAssembly {
		t.Errorf("Class = %q", dto.Class)
	}
}

func TestObjectsToDTO(t *testing.T) {
	objects := []object.Object{
		{ID: "a", Title: "A"},
		{ID: "b", Title: "B"},
	}
	dtos := mapper.ObjectsToDTO(objects)
	if len(dtos) != 2 {
		t.Fatalf("len = %d, want 2", len(dtos))
	}
	if dtos[0].Title != "A" || dtos[1].Title != "B" {
		t.Errorf("titles = %q, %q", dtos[0].Title, dtos[1].Title)
	}
}

func TestObjectsToDTO_Empty(t *testing.T) {
	dtos := mapper.ObjectsToDTO(nil)
	if len(dtos) != 0 {
		t.Errorf("expected empty slice, got len=%d", len(dtos))
	}
}

func TestBOMRowsToDTO(t *testing.T) {
	qty := 3.0
	rows := []bom.BOMRow{
		{RowID: "r1", ParentID: "asm", ChildID: "prt1", ChildClass: "prt", Quantity: qty, Unit: "pcs", Position: "10", Level: 0},
		{RowID: "r2", ParentID: "asm", ChildID: "prt2", ChildClass: "prt", Quantity: 5.0, Unit: "pcs", Level: 0},
	}
	dtos := mapper.BOMRowsToDTO(rows)
	if len(dtos) != 2 {
		t.Fatalf("len = %d", len(dtos))
	}
	if dtos[0].Quantity != 3.0 {
		t.Errorf("qty = %.1f", dtos[0].Quantity)
	}
	if dtos[0].Position != "10" {
		t.Errorf("position = %q", dtos[0].Position)
	}
}

func TestManifestToResponse(t *testing.T) {
	m := &release.Manifest{
		ReleaseID:  "rel-1",
		RootObject: "asm-1",
	}
	resp := mapper.ManifestToResponse(m, true)
	if !resp.Ready {
		t.Error("ready should be true")
	}
	resp2 := mapper.ManifestToResponse(m, false)
	if resp2.Ready {
		t.Error("ready should be false")
	}
}
