package diagnostic

import "testing"

func TestDiagnosticSeverity(t *testing.T) {
	b := Diagnostic{Severity: Blocker}
	w := Diagnostic{Severity: Warning}
	e := Diagnostic{Severity: Error}
	i := Diagnostic{Severity: Info}

	if !b.IsBlocker() {
		t.Error("blocker.IsBlocker() = false")
	}
	if b.IsWarning() {
		t.Error("blocker.IsWarning() = true")
	}
	if !w.IsWarning() {
		t.Error("warning.IsWarning() = false")
	}
	if !e.IsError() {
		t.Error("error.IsError() = false")
	}
	if !i.IsInfo() {
		t.Error("info.IsInfo() = false")
	}
}

func TestDiagnosticFields(t *testing.T) {
	d := Diagnostic{
		ID:               "diag-001",
		Severity:         Blocker,
		Code:             "REQUIRED_METADATA_MISSING",
		ObjectID:         "a320-prt-0001-v1.0",
		Path:             "metadata.material",
		Message:          "Required metadata field 'material' is missing",
		SuggestedActions: []string{"fill_material"},
	}

	if d.ID != "diag-001" {
		t.Errorf("ID = %q", d.ID)
	}
	if !d.IsBlocker() {
		t.Error("should be blocker")
	}
	if len(d.SuggestedActions) != 1 {
		t.Error("expected 1 suggested action")
	}
}
