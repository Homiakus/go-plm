// Package guards provides FSM guard implementations.
// Guards check preconditions before a lifecycle transition is allowed.
package guards

import (
	"context"

	"github.com/Homiakus/go-plm/internal/core/diagnostic"
	"github.com/Homiakus/go-plm/internal/naming"
	"github.com/Homiakus/go-plm/internal/validation/engine"
)

// GuardInput carries the data needed to evaluate a guard.
type GuardInput struct {
	ObjectID    string
	Class       string
	State       string
	Metadata    map[string]any
	AllObjects  []engine.Target
}

// Guard is a precondition check for a transition.
type Guard interface {
	ID() string
	Check(ctx context.Context, input GuardInput) ([]diagnostic.Diagnostic, error)
}

// ValidName checks that the object ID conforms to naming standard.
type ValidName struct{}

func (g *ValidName) ID() string { return "valid_name" }

func (g *ValidName) Check(ctx context.Context, input GuardInput) ([]diagnostic.Diagnostic, error) {
	if err := naming.Validate(input.ObjectID); err != nil {
		return []diagnostic.Diagnostic{{
			ID: "guard-valid-name", Severity: diagnostic.Blocker, Code: "INVALID_NAME",
			ObjectID: input.ObjectID, Message: err.Error(),
		}}, nil
	}
	return nil, nil
}

// RequiredMetadata checks that all required fields for the object class are filled.
type RequiredMetadata struct {
	Required map[string][]string
}

func (g *RequiredMetadata) ID() string { return "required_metadata" }

func (g *RequiredMetadata) Check(ctx context.Context, input GuardInput) ([]diagnostic.Diagnostic, error) {
	fields := g.Required[input.Class]
	var diags []diagnostic.Diagnostic
	for _, f := range fields {
		if _, ok := input.Metadata[f]; !ok {
			diags = append(diags, diagnostic.Diagnostic{
				ID: "guard-metadata-" + f, Severity: diagnostic.Blocker,
				Code: "REQUIRED_METADATA_MISSING", ObjectID: input.ObjectID,
				Message: "Required field " + f + " is missing",
			})
		}
	}
	return diags, nil
}

// NoBlockingIssues ensures there are no blocker diagnostics.
type NoBlockingIssues struct {
	Validator interface {
		ValidateObject(ctx context.Context, objID string) ([]diagnostic.Diagnostic, error)
	}
}

func (g *NoBlockingIssues) ID() string { return "no_blocking_issues" }

func (g *NoBlockingIssues) Check(ctx context.Context, input GuardInput) ([]diagnostic.Diagnostic, error) {
	if g.Validator == nil {
		return nil, nil
	}
	diags, err := g.Validator.ValidateObject(ctx, input.ObjectID)
	if err != nil {
		return nil, err
	}
	var blockers []diagnostic.Diagnostic
	for _, d := range diags {
		if d.IsBlocker() {
			blockers = append(blockers, d)
		}
	}
	return blockers, nil
}

// DefaultGuards returns a map of guard name -> implementation for MVP.
func DefaultGuards() map[string]Guard {
	return map[string]Guard{
		"valid_name":        &ValidName{},
		"required_metadata": &RequiredMetadata{Required: map[string][]string{
			"prt": {"unit", "make_buy"},
			"asm": {"unit"},
			"std": {"std_class", "std_id", "make_buy"},
		}},
	}
}
