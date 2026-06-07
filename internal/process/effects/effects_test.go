// Package effects_test provides tests for FSM effect implementations.
package effects_test

import (
	"context"
	"testing"

	"github.com/Homiakus/go-plm/internal/process/effects"
)

func TestCreateGitTag(t *testing.T) {
	e := &effects.CreateGitTag{}
	if e.ID() != "create_git_tag" {
		t.Errorf("ID() = %q, want %q", e.ID(), "create_git_tag")
	}
	err := e.Apply(context.Background(), effects.EffectInput{
		ObjectID: "demo-prt-0001-v1.0", TransitionName: "release",
		FromState: "approved", ToState: "released",
	})
	if err != nil {
		t.Errorf("Apply() error = %v", err)
	}
}

func TestCreateNewRevision(t *testing.T) {
	e := &effects.CreateNewRevision{}
	if e.ID() != "create_new_revision" {
		t.Errorf("ID() = %q, want %q", e.ID(), "create_new_revision")
	}
	err := e.Apply(context.Background(), effects.EffectInput{
		ObjectID: "demo-prt-0001-v1.0", TransitionName: "revise",
		FromState: "released", ToState: "draft",
	})
	if err != nil {
		t.Errorf("Apply() error = %v", err)
	}
}

func TestCreateChangeEvent(t *testing.T) {
	e := &effects.CreateChangeEvent{}
	if e.ID() != "create_change_event" {
		t.Errorf("ID() = %q, want %q", e.ID(), "create_change_event")
	}
	err := e.Apply(context.Background(), effects.EffectInput{
		ObjectID: "demo-prt-0001-v1.0", TransitionName: "revise",
		FromState: "released", ToState: "draft",
	})
	if err != nil {
		t.Errorf("Apply() error = %v", err)
	}
}

func TestLogTransition(t *testing.T) {
	e := &effects.LogTransition{}
	if e.ID() != "log_transition" {
		t.Errorf("ID() = %q, want %q", e.ID(), "log_transition")
	}
	err := e.Apply(context.Background(), effects.EffectInput{
		ObjectID: "demo-prt-0001-v1.0", TransitionName: "submit_review",
		FromState: "draft", ToState: "in_review",
	})
	if err != nil {
		t.Errorf("Apply() error = %v", err)
	}
}

func TestDefaultEffects(t *testing.T) {
	defs := effects.DefaultEffects()
	expected := []string{"create_git_tag", "create_new_revision", "create_change_event", "log_transition"}
	for _, name := range expected {
		if _, ok := defs[name]; !ok {
			t.Errorf("DefaultEffects missing %q", name)
		}
	}
	if len(defs) != len(expected) {
		t.Errorf("DefaultEffects has %d entries, want %d", len(defs), len(expected))
	}
}
