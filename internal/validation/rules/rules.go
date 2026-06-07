// Package rules provides concrete validation rule implementations.
package rules

import (
	"context"
	"fmt"

	"github.com/Homiakus/go-plm/internal/core/diagnostic"
	"github.com/Homiakus/go-plm/internal/core/object"
	"github.com/Homiakus/go-plm/internal/naming"
	"github.com/Homiakus/go-plm/internal/validation/engine"
)

// ── ValidName ──

// ValidName checks that object ID conforms to Naming Standard v10.
type ValidName struct{}

func (r *ValidName) ID() string { return "valid_name" }

func (r *ValidName) Check(ctx context.Context, target engine.Target) ([]diagnostic.Diagnostic, error) {
	if err := naming.Validate(string(target.Object.ID)); err != nil {
		return []diagnostic.Diagnostic{{
			ID: "diag-name-" + string(target.Object.ID),
			Severity: diagnostic.Blocker, Code: "INVALID_NAME",
			ObjectID: string(target.Object.ID), Path: "id",
			Message: err.Error(), SuggestedActions: []string{"fix_id"},
		}}, nil
	}
	return nil, nil
}

// ── RequiredMetadata ──

// RequiredMetadata checks that all required metadata fields are present.
type RequiredMetadata struct {
	RequiredFields map[string][]string
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
				ID: fmt.Sprintf("diag-%s-%s", field, target.Object.ID),
				Severity: diagnostic.Blocker, Code: "REQUIRED_METADATA_MISSING",
				ObjectID: string(target.Object.ID),
				Path:     fmt.Sprintf("metadata.%s", field),
				Message:  fmt.Sprintf("Required metadata field '%s' is missing", field),
				SuggestedActions: []string{fmt.Sprintf("fill_%s", field)},
			})
		}
	}
	return diags, nil
}

// ── ValidRelations ──

// ValidRelations checks that all relations point to existing objects.
type ValidRelations struct{}

func (r *ValidRelations) ID() string { return "valid_relations" }

func (r *ValidRelations) Check(ctx context.Context, target engine.Target) ([]diagnostic.Diagnostic, error) {
	var diags []diagnostic.Diagnostic
	allIDs := map[string]bool{}
	for _, obj := range target.AllObjects {
		allIDs[string(obj.ID)] = true
	}

	for _, rel := range target.Object.Relations {
		if !allIDs[rel.ToID] {
			diags = append(diags, diagnostic.Diagnostic{
				ID: fmt.Sprintf("diag-dangling-%s-%s", target.Object.ID, rel.ToID),
				Severity: diagnostic.Blocker, Code: "DANGLING_RELATION",
				ObjectID: string(target.Object.ID),
				Path:     fmt.Sprintf("relations[to=%s]", rel.ToID),
				Message:  fmt.Sprintf("Relation target %q does not exist in project", rel.ToID),
				SuggestedActions: []string{"fix_target_id", "remove_relation"},
			})
		}
		if string(target.Object.ID) == rel.ToID {
			diags = append(diags, diagnostic.Diagnostic{
				ID: fmt.Sprintf("diag-self-loop-%s", target.Object.ID),
				Severity: diagnostic.Blocker, Code: "SELF_LOOP_RELATION",
				ObjectID: string(target.Object.ID),
				Path:     fmt.Sprintf("relations[to=%s]", rel.ToID),
				Message:  "Object has a self-referencing relation",
				SuggestedActions: []string{"remove_self_relation"},
			})
		}
	}
	return diags, nil
}

// ── ValidArtifacts ──

// ValidArtifacts checks that required artifacts are present and not outdated.
type ValidArtifacts struct{}

func (r *ValidArtifacts) ID() string { return "valid_artifacts" }

func (r *ValidArtifacts) Check(ctx context.Context, target engine.Target) ([]diagnostic.Diagnostic, error) {
	var diags []diagnostic.Diagnostic
	for _, a := range target.Object.Artifacts {
		if a.Required && a.Path == "" && a.Status == "missing" {
			diags = append(diags, diagnostic.Diagnostic{
				ID: fmt.Sprintf("diag-art-missing-%s-%s", target.Object.ID, a.ID),
				Severity: diagnostic.Warning, Code: "ARTIFACT_MISSING",
				ObjectID: string(target.Object.ID),
				Path:     fmt.Sprintf("artifacts[%s]", a.ID),
				Message:  fmt.Sprintf("Required artifact %q (%s) is not yet created", a.ID, a.Kind),
				SuggestedActions: []string{"create_artifact", "upload_file"},
			})
		}
		if a.Status == "outdated" {
			diags = append(diags, diagnostic.Diagnostic{
				ID: fmt.Sprintf("diag-art-outdated-%s-%s", target.Object.ID, a.ID),
				Severity: diagnostic.Warning, Code: "ARTIFACT_OUTDATED",
				ObjectID: string(target.Object.ID),
				Path:     fmt.Sprintf("artifacts[%s]", a.ID),
				Message:  fmt.Sprintf("Artifact %q checksum is outdated", a.ID),
				SuggestedActions: []string{"update_checksum", "reupload_file"},
			})
		}
	}
	return diags, nil
}

// ── ChecksumsActual ──

// ChecksumsActual verifies artifact file checksums are current.
type ChecksumsActual struct {
	ObjectsDir string
}

func (r *ChecksumsActual) ID() string { return "checksums_actual" }

func (r *ChecksumsActual) Check(ctx context.Context, target engine.Target) ([]diagnostic.Diagnostic, error) {
	// Full checksum computation requires reading files — done by upper layers.
	// Here we check that artifacts with checksums haven't been flagged as outdated.
	var diags []diagnostic.Diagnostic
	for _, a := range target.Object.Artifacts {
		if a.Status == "outdated" {
			diags = append(diags, diagnostic.Diagnostic{
				ID: fmt.Sprintf("diag-checksum-%s-%s", target.Object.ID, a.ID),
				Severity: diagnostic.Blocker, Code: "CHECKSUM_OUTDATED",
				ObjectID: string(target.Object.ID),
				Path:     fmt.Sprintf("artifacts[%s].checksum", a.ID),
				Message:  fmt.Sprintf("Checksum for %q is outdated; file may have changed", a.Path),
				SuggestedActions: []string{"recompute_checksum"},
			})
		}
	}
	return diags, nil
}

// ── LifecycleConsistency ──

// LifecycleConsistency checks that the object's state is consistent.
type LifecycleConsistency struct{}

func (r *LifecycleConsistency) ID() string { return "lifecycle_consistency" }

func (r *LifecycleConsistency) Check(ctx context.Context, target engine.Target) ([]diagnostic.Diagnostic, error) {
	var diags []diagnostic.Diagnostic

	// Check: released objects need version > 1.0
	if target.Object.State == object.StateReleased {
		// Released objects are valid — no additional checks in MVP
	}

	// Check: approved/released objects should have all required metadata
	if target.Object.IsApproved() {
		if len(target.Object.Metadata) == 0 {
			diags = append(diags, diagnostic.Diagnostic{
				ID: fmt.Sprintf("diag-lifecycle-no-meta-%s", target.Object.ID),
				Severity: diagnostic.Warning, Code: "APPROVED_NO_METADATA",
				ObjectID: string(target.Object.ID),
				Message:  "Object is approved/released but has no metadata",
				SuggestedActions: []string{"fill_metadata"},
			})
		}
	}

	return diags, nil
}

// ── DefaultRegistry ──

// DefaultRegistry returns a registry with all MVP validation rules registered.
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
	reg.Register(&ValidRelations{})
	reg.Register(&ValidArtifacts{})
	reg.Register(&ChecksumsActual{})
	reg.Register(&LifecycleConsistency{})
	return reg
}
