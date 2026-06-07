package fsm

import "testing"

var objectLifecycle = []byte(`
name: object_lifecycle
initial: draft
states: [draft, in_review, approved, released, blocked, obsolete, archived]
transitions:
  - name: submit_review
    from: [draft]
    to: in_review
    guards: [valid_name, required_metadata]
  - name: reject
    from: [in_review]
    to: draft
  - name: approve
    from: [in_review]
    to: approved
    guards: [no_blocking_issues]
  - name: release
    from: [approved]
    to: released
    guards: [no_release_blockers, children_released]
    effects: [create_git_tag]
  - name: revise
    from: [released]
    to: draft
    effects: [create_new_revision]
  - name: block
    from: ["*"]
    to: blocked
  - name: unblock
    from: [blocked]
    to: draft
  - name: obsolete
    from: [released]
    to: obsolete
  - name: archive
    from: [obsolete]
    to: archived
`)

func TestNewFromYAML(t *testing.T) {
	m, err := NewFromYAML(objectLifecycle)
	if err != nil {
		t.Fatal(err)
	}
	if m.Def.Initial != "draft" {
		t.Errorf("initial = %q", m.Def.Initial)
	}
	if len(m.Def.Transitions) != 9 {
		t.Errorf("expected 9 transitions, got %d", len(m.Def.Transitions))
	}
}

func TestMachineCan(t *testing.T) {
	m, _ := NewFromYAML(objectLifecycle)
	if !m.Can("draft", "submit_review") {
		t.Error("draft should allow submit_review")
	}
	if m.Can("draft", "approve") {
		t.Error("draft should NOT allow approve")
	}
	if !m.Can("draft", "block") {
		t.Error("draft should allow block (wildcard)")
	}
}

func TestMachineNext(t *testing.T) {
	m, _ := NewFromYAML(objectLifecycle)
	next, err := m.Next("draft", "submit_review")
	if err != nil {
		t.Fatal(err)
	}
	if next != "in_review" {
		t.Errorf("next = %q", next)
	}
	_, err = m.Next("draft", "release")
	if err == nil {
		t.Error("expected error for invalid transition")
	}
}

func TestMachineAvailable(t *testing.T) {
	m, _ := NewFromYAML(objectLifecycle)
	avail := m.Available("draft")
	if len(avail) < 2 {
		t.Errorf("draft should have >=2 transitions, got %d", len(avail))
	}
	names := map[string]bool{}
	for _, tr := range avail {
		names[tr.Name] = true
	}
	if !names["submit_review"] {
		t.Error("submit_review missing")
	}
	if !names["block"] {
		t.Error("block missing")
	}
}

func TestGuardsAndEffects(t *testing.T) {
	m, _ := NewFromYAML(objectLifecycle)
	guards := m.GuardsFor("release")
	if len(guards) != 2 {
		t.Errorf("expected 2 guards for release, got %d", len(guards))
	}
	effects := m.EffectsFor("release")
	if len(effects) != 1 || effects[0] != "create_git_tag" {
		t.Errorf("effects = %v", effects)
	}
}
