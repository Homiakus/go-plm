package relation

import "testing"

func TestRelationValidateShape(t *testing.T) {
	qty := 2.0
	tests := []struct {
		name    string
		r       Relation
		wantErr error
	}{
		{
			name:    "valid",
			r:       Relation{FromID: "asm-1", ToID: "prt-1", Type: Contains, Quantity: &qty, Unit: "pcs"},
			wantErr: nil,
		},
		{"empty from", Relation{FromID: "", ToID: "prt-1", Type: Contains}, ErrEmptyFromID},
		{"empty to", Relation{FromID: "asm-1", ToID: "", Type: Contains}, ErrEmptyToID},
		{"empty type", Relation{FromID: "asm-1", ToID: "prt-1", Type: ""}, ErrEmptyType},
		{"self loop", Relation{FromID: "x", ToID: "x", Type: Contains}, ErrSelfLoop},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.r.ValidateShape()
			if err != tt.wantErr {
				t.Errorf("ValidateShape() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestRelationHasQuantity(t *testing.T) {
	qty := 2.0
	zero := 0.0
	r2 := Relation{Quantity: &qty}
	if !r2.HasQuantity() {
		t.Error("HasQuantity() with 2.0 = false")
	}
	r0 := Relation{Quantity: &zero}
	if r0.HasQuantity() {
		t.Error("HasQuantity() with 0.0 = true")
	}
	rNil := Relation{}
	if rNil.HasQuantity() {
		t.Error("HasQuantity() with nil = true")
	}
}

func TestTypeInverse(t *testing.T) {
	if Contains.Inverse() != "used_in" {
		t.Errorf("Contains.Inverse() = %q", Contains.Inverse())
	}
	if HasDrawing.Inverse() != "drawing_of" {
		t.Errorf("HasDrawing.Inverse() = %q", HasDrawing.Inverse())
	}
	if DerivedFrom.Inverse() != "" {
		t.Error("DerivedFrom should have no inverse")
	}
}

func TestIsBOMRelation(t *testing.T) {
	if !Contains.IsBOMRelation() {
		t.Error("Contains.IsBOMRelation() = false")
	}
	if !Includes.IsBOMRelation() {
		t.Error("Includes.IsBOMRelation() = false")
	}
	if EquivalentTo.IsBOMRelation() {
		t.Error("EquivalentTo.IsBOMRelation() = true")
	}
}

func TestIsStored(t *testing.T) {
	if !Contains.IsStored() {
		t.Error("Contains.IsStored() = false")
	}
	if Type("used_in").IsStored() {
		t.Error("used_in.IsStored() = true (projected only)")
	}
}
