package object

import "testing"

func TestObjectValidateIdentity(t *testing.T) {
	tests := []struct {
		name    string
		obj     Object
		wantErr error
	}{
		{
			name:    "valid object",
			obj:     Object{ID: "a320-prt-0001-v1.0", Class: ClassPart, Title: "Bracket"},
			wantErr: nil,
		},
		{
			name:    "empty id",
			obj:     Object{ID: "", Class: ClassPart, Title: "Bracket"},
			wantErr: ErrEmptyID,
		},
		{
			name:    "empty class",
			obj:     Object{ID: "a320-prt-0001-v1.0", Class: "", Title: "Bracket"},
			wantErr: ErrEmptyClass,
		},
		{
			name:    "empty title",
			obj:     Object{ID: "a320-prt-0001-v1.0", Class: ClassPart, Title: ""},
			wantErr: ErrEmptyTitle,
		},
		{
			name:    "whitespace title",
			obj:     Object{ID: "a320-prt-0001-v1.0", Class: ClassPart, Title: "   "},
			wantErr: ErrEmptyTitle,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.obj.ValidateIdentity()
			if err != tt.wantErr {
				t.Errorf("ValidateIdentity() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestObjectStateHelpers(t *testing.T) {
	draft := Object{ID: "a", Class: ClassPart, State: StateDraft}
	released := Object{ID: "b", Class: ClassPart, State: StateReleased}
	blocked := Object{ID: "c", Class: ClassPart, State: StateBlocked}
	approved := Object{ID: "d", Class: ClassPart, State: StateApproved}

	if !draft.IsDraft() {
		t.Error("draft.IsDraft() = false")
	}
	if draft.IsReleased() {
		t.Error("draft.IsReleased() = true")
	}
	if !released.IsReleased() {
		t.Error("released.IsReleased() = false")
	}
	if !released.IsApproved() {
		t.Error("released.IsApproved() = false")
	}
	if !approved.IsApproved() {
		t.Error("approved.IsApproved() = false")
	}
	if !blocked.IsBlocked() {
		t.Error("blocked.IsBlocked() = false")
	}
}

func TestObjectIsZero(t *testing.T) {
	zero := Object{}
	if !zero.IsZero() {
		t.Error("zero Object.IsZero() = false")
	}
	nonZero := Object{ID: "x"}
	if nonZero.IsZero() {
		t.Error("non-zero Object.IsZero() = true")
	}
}

func TestObjectWithState(t *testing.T) {
	obj := Object{ID: "x", Class: ClassPart, State: StateDraft}
	rel := obj.WithState(StateReleased)
	if rel.State != StateReleased {
		t.Errorf("WithState() state = %v, want %v", rel.State, StateReleased)
	}
	if obj.State != StateDraft {
		t.Error("WithState() mutated original")
	}
}

func TestObjectString(t *testing.T) {
	obj := Object{ID: "a320-prt-0001-v1.0", Class: ClassPart, Title: "Bracket", State: StateDraft}
	got := obj.String()
	want := "a320-prt-0001-v1.0 [prt] Bracket (draft)"
	if got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}
