package lifecycle

import "testing"

func makeObjLCM() *Machine {
	return NewMachine(StandardObjectLifecycle())
}

func TestMachineCan(t *testing.T) {
	m := makeObjLCM()

	if !m.Can("draft", "submit_review") {
		t.Error("draft should allow submit_review")
	}
	if m.Can("draft", "approve") {
		t.Error("draft should NOT allow approve")
	}
	if !m.Can("in_review", "approve") {
		t.Error("in_review should allow approve")
	}
	if !m.Can("draft", "block") {
		t.Error("draft should allow block (wildcard)")
	}
	if !m.Can("released", "block") {
		t.Error("released should allow block (wildcard)")
	}
}

func TestMachineNext(t *testing.T) {
	m := makeObjLCM()

	next, err := m.Next("draft", "submit_review")
	if err != nil {
		t.Fatal(err)
	}
	if next != "in_review" {
		t.Errorf("submit_review target = %q, want in_review", next)
	}

	_, err = m.Next("draft", "release")
	if err == nil {
		t.Error("expected error for invalid transition")
	}
}

func TestMachineAvailable(t *testing.T) {
	m := makeObjLCM()
	avail := m.Available("draft")
	if len(avail) < 2 {
		t.Errorf("draft should have at least 2 transitions, got %d", len(avail))
	}
	names := make(map[string]bool)
	for _, tr := range avail {
		names[tr.Name] = true
	}
	if !names["submit_review"] {
		t.Error("submit_review missing from draft transitions")
	}
	if !names["block"] {
		t.Error("block (wildcard) missing from draft transitions")
	}
}

func TestMachineGetTransition(t *testing.T) {
	m := makeObjLCM()
	tr := m.GetTransition("release")
	if tr == nil {
		t.Fatal("release transition not found")
	}
	if tr.To != "released" {
		t.Errorf("release target = %q", tr.To)
	}
	if len(tr.Guards) < 3 {
		t.Errorf("release should have >=3 guards, got %d", len(tr.Guards))
	}

	if m.GetTransition("nonexistent") != nil {
		t.Error("nonexistent transition should be nil")
	}
}

func TestStandardObjectLifecycleConsistency(t *testing.T) {
	def := StandardObjectLifecycle()
	if def.Initial != "draft" {
		t.Errorf("initial = %q", def.Initial)
	}
	states := make(map[string]bool)
	for _, s := range def.States {
		states[s] = true
	}
	for _, tr := range def.Transitions {
		if !states[tr.To] {
			t.Errorf("transition %q target %q is not in states list", tr.Name, tr.To)
		}
	}
}
