// Package relation defines the Relation domain type — a named link between two objects.
// Pure domain model — no filesystem, target-existence checking, or index dependencies.
package relation

import "errors"

// Type represents the kind of relation between two objects.
type Type string

// Well-known relation types.
const (
	Contains              Type = "contains"
	HasDrawing            Type = "has_drawing"
	HasCAD                Type = "has_cad"
	ManufacturingFileFor  Type = "manufacturing_file_for"
	ProcessPlanFor        Type = "process_plan_for"
	WorkInstructionFor    Type = "work_instruction_for"
	InspectionFor         Type = "inspection_for"
	Substitutes           Type = "substitutes"
	EquivalentTo          Type = "equivalent_to"
	Replaces              Type = "replaces"
	Affects               Type = "affects"
	Includes              Type = "includes"
	DerivedFrom           Type = "derived_from"
)

// Inverse returns the canonical inverse relation type, if one exists.
func (t Type) Inverse() Type {
	switch t {
	case Contains:
		return "used_in"
	case HasDrawing:
		return "drawing_of"
	case ManufacturingFileFor:
		return "has_manufacturing_file"
	case HasCAD:
		return "cad_of"
	case ProcessPlanFor:
		return "has_process_plan"
	case WorkInstructionFor:
		return "has_work_instruction"
	case InspectionFor:
		return "has_inspection"
	case Substitutes:
		return "substituted_by"
	case Replaces:
		return "replaced_by"
	case Affects:
		return "affected_by"
	case Includes:
		return "released_in"
	default:
		return ""
	}
}

// IsBOMRelation returns true if this relation type contributes to BOM structure.
func (t Type) IsBOMRelation() bool {
	return t == Contains || t == Includes
}

// IsStored returns true if this relation type should be stored in the source object
// (not just projected from its inverse).
func (t Type) IsStored() bool {
	switch t {
	case Contains, HasDrawing, HasCAD, ManufacturingFileFor, ProcessPlanFor,
		WorkInstructionFor, InspectionFor, Substitutes, EquivalentTo, Replaces,
		Affects, Includes, DerivedFrom:
		return true
	default:
		return false
	}
}

// Relation is a named, optionally quantified link between two objects.
type Relation struct {
	FromID     string         `yaml:"from_id" json:"from_id"`
	ToID       string         `yaml:"to_id" json:"to_id"`
	Type       Type           `yaml:"type" json:"type"`
	Quantity   *float64       `yaml:"quantity,omitempty" json:"quantity,omitempty"`
	Unit       string         `yaml:"unit,omitempty" json:"unit,omitempty"`
	Metadata   map[string]any `yaml:"metadata,omitempty" json:"metadata,omitempty"`
}

// Common validation errors.
var (
	ErrEmptyFromID = errors.New("relation: from_id must not be empty")
	ErrEmptyToID   = errors.New("relation: to_id must not be empty")
	ErrEmptyType   = errors.New("relation: type must not be empty")
	ErrSelfLoop    = errors.New("relation: from_id and to_id must differ")
)

// ValidateShape checks structural validity of the relation.
func (r Relation) ValidateShape() error {
	if r.FromID == "" {
		return ErrEmptyFromID
	}
	if r.ToID == "" {
		return ErrEmptyToID
	}
	if r.Type == "" {
		return ErrEmptyType
	}
	if r.FromID == r.ToID {
		return ErrSelfLoop
	}
	return nil
}

// HasQuantity returns true if the relation has a positive quantity.
func (r Relation) HasQuantity() bool {
	return r.Quantity != nil && *r.Quantity > 0
}
