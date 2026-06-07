package release

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

func TestBuildScope(t *testing.T) {
	objects := mockObjects{
		"a-asm-0100-v1.0": {ID: "a-asm-0100-v1.0", Class: object.ClassAssembly, State: object.StateApproved},
		"a-prt-0001-v1.0": {ID: "a-prt-0001-v1.0", Class: object.ClassPart, State: object.StateApproved},
		"a-prt-0002-v1.0": {ID: "a-prt-0002-v1.0", Class: object.ClassPart, State: object.StateApproved},
	}
	relations := mockRelations{
		"a-asm-0100-v1.0": {
			{ToID: "a-prt-0001-v1.0", Type: "contains"},
			{ToID: "a-prt-0002-v1.0", Type: "contains"},
		},
	}

	svc := New(objects, relations)
	scope, err := svc.BuildScope(context.Background(), "a-asm-0100-v1.0")
	if err != nil {
		t.Fatal(err)
	}
	if len(scope.Objects) != 3 {
		t.Errorf("expected 3 objects in scope, got %d", len(scope.Objects))
	}
}

func TestCheckReadiness(t *testing.T) {
	objects := mockObjects{
		"a-prt-0001-v1.0": {ID: "a-prt-0001-v1.0", Class: object.ClassPart, State: object.StateReleased},
		"a-prt-0002-v1.0": {ID: "a-prt-0002-v1.0", Class: object.ClassPart, State: object.StateDraft}, // not ready
	}
	relations := mockRelations{}

	svc := New(objects, relations)
	scope := &Scope{RootID: "root", Objects: []string{"a-prt-0001-v1.0", "a-prt-0002-v1.0"}}

	readiness, err := svc.CheckReadiness(context.Background(), scope)
	if err != nil {
		t.Fatal(err)
	}
	if readiness.Ready {
		t.Error("should not be ready — a-prt-0002 is draft")
	}
	if len(readiness.Blockers) != 1 {
		t.Errorf("expected 1 blocker, got %d", len(readiness.Blockers))
	}
}

func TestGenerateManifest(t *testing.T) {
	objects := mockObjects{
		"a-prt-0001-v1.0": {ID: "a-prt-0001-v1.0", Class: object.ClassPart, State: object.StateReleased},
	}
	relations := mockRelations{}

	svc := New(objects, relations)
	scope := &Scope{RootID: "root", Objects: []string{"a-prt-0001-v1.0"}}

	manifest, err := svc.GenerateManifest(context.Background(), "a-rel-0001-v1.0", scope)
	if err != nil {
		t.Fatal(err)
	}
	if manifest.ReleaseID != "a-rel-0001-v1.0" {
		t.Errorf("release id = %q", manifest.ReleaseID)
	}
	if len(manifest.Objects) != 1 {
		t.Errorf("expected 1 object, got %d", len(manifest.Objects))
	}
}
