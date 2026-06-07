// Package rules provides concrete validation rule implementations.
package rules

import (
	"context"
	"fmt"

	"github.com/Homiakus/go-plm/internal/core/diagnostic"
	"github.com/Homiakus/go-plm/internal/naming"
	"github.com/Homiakus/go-plm/internal/validation/engine"
)

// RequiredMetadata checks that all required metadata fields are present per object class.
type RequiredMetadata struct {
	RequiredFields map[string][]string // class -> required fields
}

func (r *RequiredMetadata) ID() string { return "required_metadata" }

func (r *RequiredMetadata) Check(ctx context.Context, target engine.Target) ([]diagnostic.Diagnostic, error) {
	class := string(target.Object.Class)
	required, ok := r.RequiredFields[class]
	if !ok {
		return nil, nil
	}

	var diags []diagnostic.Diagnostic
	for _, field := range required {
		if _, exists := target.Object.Metadata[field]; !exists {
			diags = append(diags, diagnostic.Diagnostic{
				ID:       fmt.Sprintf("diag-%s-%s", field, target.Object.ID),
				Severity: diagnostic.Blocker,
				Code:     "REQUIRED_METADATA_MISSING",
				ObjectID: string(target.Object.ID),
				Path:     fmt.Sprintf("metadata.%s", field),
				Message:  fmt.Sprintf("Required metadata field '%s' is missing", field),
				SuggestedActions: []string{fmt.Sprintf("fill_%s", field)},
			})
		}
	}
	return diags, nil
}

// ValidName checks that object ID conforms to Naming Standard v10.
type ValidName struct{}

func (r *ValidName) ID() string { return "valid_name" }

func (r *ValidName) Check(ctx context.Context, target engine.Target) ([]diagnostic.Diagnostic, error) {
	if err := naming.Validate(string(target.Object.ID)); err != nil {
		return []diagnostic.Diagnostic{{
			ID:               "diag-name-" + string(target.Object.ID),
			Severity:         diagnostic.Blocker,
			Code:             "INVALID_NAME",
			ObjectID:         string(target.Object.ID),
			Path:             "id",
			Message:          err.Error(),
			SuggestedActions: []string{"fix_id"},
		}}, nil
	}
	return nil, nil
}

// DefaultRegistry returns a registry with MVP validation rules registered.
func DefaultRegistry() *engine.Registry {
	reg := engine.NewRegistry()
	reg.Register(&ValidName{})
	reg.Register(&RequiredMetadata{
		RequiredFields: map[string][]string{
			"prt": {"unit", "make_buy"},
			"asm": {"unit"},
			"std": {"std_class", "std_id", "make_buy"},
		},
	})
	return reg
}
