package event

import (
	"testing"
	"time"
)

func TestEventNew(t *testing.T) {
	payload := map[string]any{"class": "prt", "title": "Bracket"}
	ev := New("evt-001", "test-user", ObjectCreated, "a320-prt-0001-v1.0", payload)

	if ev.ID != "evt-001" {
		t.Errorf("ID = %q", ev.ID)
	}
	if ev.Actor != "test-user" {
		t.Errorf("Actor = %q", ev.Actor)
	}
	if ev.Type != ObjectCreated {
		t.Errorf("Type = %q", ev.Type)
	}
	if ev.ObjectID != "a320-prt-0001-v1.0" {
		t.Errorf("ObjectID = %q", ev.ObjectID)
	}
	if time.Since(ev.Time) > time.Second {
		t.Error("Time is not recent")
	}
	if ev.Payload["class"] != "prt" {
		t.Error("Payload class mismatch")
	}
}

func TestEventTypesNotEmpty(t *testing.T) {
	types := []Type{
		ObjectCreated, ObjectDeleted, ObjectDuplicated,
		MetadataUpdated, RelationAdded, RelationRemoved,
		ArtifactAttached, ArtifactDetached,
		LifecycleTransitioned, CheckpointCreated, ReleasePublished,
	}
	for _, typ := range types {
		if typ == "" {
			t.Error("empty event type constant")
		}
	}
}
