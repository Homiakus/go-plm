// Package diagnostic defines the Diagnostic domain type — a validation result.
// Pure domain model — no validation rules, no file access, no auto-fix.
package diagnostic

// Severity indicates the criticality of a diagnostic finding.
type Severity string

const (
	Info    Severity = "info"
	Warning Severity = "warning"
	Blocker Severity = "blocker"
	Error   Severity = "error"
)

// Diagnostic is a single validation finding, linked to an object and a specific data path.
type Diagnostic struct {
	ID               string   `json:"id"`
	Severity         Severity `json:"severity"`
	Code             string   `json:"code"`
	ObjectID         string   `json:"object_id"`
	Path             string   `json:"path,omitempty"`
	Message          string   `json:"message"`
	SuggestedActions []string `json:"suggested_actions,omitempty"`
}

// IsBlocker returns true if this diagnostic blocks a lifecycle transition.
func (d Diagnostic) IsBlocker() bool {
	return d.Severity == Blocker
}

// IsWarning returns true if this is a non-blocking warning.
func (d Diagnostic) IsWarning() bool {
	return d.Severity == Warning
}

// IsError returns true if this is a technical error.
func (d Diagnostic) IsError() bool {
	return d.Severity == Error
}

// IsInfo returns true if this is informational.
func (d Diagnostic) IsInfo() bool {
	return d.Severity == Info
}
