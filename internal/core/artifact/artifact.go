// Package artifact defines the Artifact domain type — a file attached to an Object.
// Pure domain model — no filesystem I/O, checksum computation, or file operations.
package artifact

// Kind classifies the type of artifact file.
type Kind string

// Well-known artifact kinds.
const (
	KindCAD          Kind = "cad"
	KindDrawing      Kind = "drawing"
	KindManufacturing Kind = "manufacturing"
	KindImage        Kind = "image"
	KindCertificate  Kind = "certificate"
	KindEvidence     Kind = "evidence"
	KindGenerated    Kind = "generated"
	KindDocument     Kind = "document"
)

// Role describes the functional role of an artifact (e.g., "primary_step", "drawing_pdf").
type Role string

// Status indicates whether the artifact file is present, missing, or has an outdated checksum.
type Status string

const (
	StatusPresent  Status = "present"
	StatusMissing  Status = "missing"
	StatusOutdated Status = "outdated"
)

// Artifact represents a file associated with an engineering object.
type Artifact struct {
	ID           string `yaml:"id" json:"id"`
	Kind         Kind   `yaml:"kind" json:"kind"`
	Role         Role   `yaml:"role" json:"role"`
	Path         string `yaml:"path" json:"path"`
	OriginalName string `yaml:"original_name,omitempty" json:"original_name,omitempty"`
	Checksum     string `yaml:"checksum,omitempty" json:"checksum,omitempty"`
	SizeBytes    int64  `yaml:"size_bytes,omitempty" json:"size_bytes,omitempty"`
	Generated    bool   `yaml:"generated" json:"generated"`
	Required     bool   `yaml:"required" json:"required"`
	Status       Status `yaml:"status" json:"status"`
}

// IsPresent returns true if the artifact file exists and checksum is current.
func (a Artifact) IsPresent() bool {
	return a.Status == StatusPresent
}

// IsMissing returns true if the artifact is required but not yet created.
func (a Artifact) IsMissing() bool {
	return a.Required && a.Status == StatusMissing
}

// IsPlaceholder returns true if this is a placeholder for a required-but-missing file.
func (a Artifact) IsPlaceholder() bool {
	return a.Required && a.Path == "" && a.Status == StatusMissing
}
