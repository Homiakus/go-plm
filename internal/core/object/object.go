// Package object defines the core Object domain type for the PLM system.
// An Object is the fundamental engineering entity: Part, Assembly, Drawing, etc.
// Pure domain model — no filesystem, SQLite, or UI dependencies.
package object

import (
	"errors"
	"fmt"
	"strings"
)

// ID is a unique object identifier in the format [project]-[class]-[sequence]-v[major].[minor].
type ID string

// Class represents the type of engineering object.
type Class string

// State represents the current lifecycle state of an object.
type State string

// Well-known classes.
const (
	ClassPart           Class = "prt"
	ClassAssembly       Class = "asm"
	ClassDrawing        Class = "drw"
	ClassDocument       Class = "doc"
	ClassStandardPart   Class = "std"
	ClassMaterial       Class = "mat"
	ClassCutting        Class = "cut"
	ClassBending        Class = "bnd"
	ClassNC             Class = "nc"
	ClassInspection     Class = "ins"
	ClassTechProcess    Class = "tpc"
	ClassWorkInstruction Class = "wi"
	ClassBOM            Class = "bom"
	ClassRelease        Class = "rel"
	ClassChangeRequest  Class = "cr"
	ClassBuild          Class = "build"
)

// Well-known states.
const (
	StateDraft     State = "draft"
	StateInReview  State = "in_review"
	StateApproved  State = "approved"
	StateReleased  State = "released"
	StateBlocked   State = "blocked"
	StateObsolete  State = "obsolete"
	StateArchived  State = "archived"
)

// RelationRef is a lightweight reference to a relation stored in an object's frontmatter.
type RelationRef struct {
	ToID     string  `yaml:"to" json:"to"`
	Type     string  `yaml:"type" json:"type"`
	Quantity *float64 `yaml:"quantity,omitempty" json:"quantity,omitempty"`
	Unit     string  `yaml:"unit,omitempty" json:"unit,omitempty"`
}

// ArtifactRef is a lightweight reference to an artifact stored in an object's frontmatter.
type ArtifactRef struct {
	ID           string `yaml:"id" json:"id"`
	Kind         string `yaml:"kind" json:"kind"`
	Role         string `yaml:"role" json:"role"`
	Path         string `yaml:"path" json:"path"`
	OriginalName string `yaml:"original_name,omitempty" json:"original_name,omitempty"`
	Checksum     string `yaml:"checksum,omitempty" json:"checksum,omitempty"`
	SizeBytes    int64  `yaml:"size_bytes,omitempty" json:"size_bytes,omitempty"`
	Generated    bool   `yaml:"generated" json:"generated"`
	Required     bool   `yaml:"required" json:"required"`
	Status       string `yaml:"status" json:"status"`
}

// Object is the core engineering entity in the PLM system.
type Object struct {
	ID        ID              `yaml:"id" json:"id"`
	Project   string          `yaml:"project" json:"project"`
	Class     Class           `yaml:"class" json:"class"`
	Sequence  string          `yaml:"sequence" json:"sequence"`
	Version   string          `yaml:"version" json:"version"`
	Revision  string          `yaml:"revision" json:"revision"`
	State     State           `yaml:"state" json:"state"`
	Title     string          `yaml:"title" json:"title"`
	Metadata  map[string]any  `yaml:"metadata,omitempty" json:"metadata,omitempty"`
	Relations []RelationRef   `yaml:"relations,omitempty" json:"relations,omitempty"`
	Artifacts []ArtifactRef   `yaml:"artifacts,omitempty" json:"artifacts,omitempty"`
}

// Common validation errors.
var (
	ErrEmptyID    = errors.New("object: id must not be empty")
	ErrEmptyClass = errors.New("object: class must not be empty")
	ErrEmptyTitle = errors.New("object: title must not be empty")
)

// ValidateIdentity checks that the object has a valid ID, class, and title.
func (o Object) ValidateIdentity() error {
	if o.ID == "" {
		return ErrEmptyID
	}
	if o.Class == "" {
		return ErrEmptyClass
	}
	if strings.TrimSpace(o.Title) == "" {
		return ErrEmptyTitle
	}
	return nil
}

// IsZero returns true if the object is the zero value.
func (o Object) IsZero() bool {
	return o.ID == "" && o.Class == "" && o.State == ""
}

// IsDraft returns true if the object is in draft state.
func (o Object) IsDraft() bool {
	return o.State == StateDraft || o.State == ""
}

// IsReleased returns true if the object is released.
func (o Object) IsReleased() bool {
	return o.State == StateReleased
}

// IsApproved returns true if the object is approved or released.
func (o Object) IsApproved() bool {
	return o.State == StateApproved || o.State == StateReleased
}

// IsBlocked returns true if the object is blocked.
func (o Object) IsBlocked() bool {
	return o.State == StateBlocked
}

// WithState returns a copy of the object with the given state.
func (o Object) WithState(state State) Object {
	o.State = state
	return o
}

// String returns a human-readable representation.
func (o Object) String() string {
	return fmt.Sprintf("%s [%s] %s (%s)", o.ID, o.Class, o.Title, o.State)
}
