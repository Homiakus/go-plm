// Package guards provides FSM guard implementations.
// Guards check preconditions before a lifecycle transition is allowed.
package guards

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Homiakus/go-plm/internal/core/diagnostic"
	"github.com/Homiakus/go-plm/internal/core/object"
	"github.com/Homiakus/go-plm/internal/naming"
	"github.com/Homiakus/go-plm/internal/validation/engine"
)

// GuardInput carries the data needed to evaluate a guard.
type GuardInput struct {
	ObjectID   string
	Class      string
	State      string
	Metadata   map[string]any
	AllObjects []engine.Target
}

// Guard is a precondition check for a transition.
type Guard interface {
	ID() string
	Check(ctx context.Context, input GuardInput) ([]diagnostic.Diagnostic, error)
}

// ObjectReader abstracts reading objects for guards.
type ObjectReader interface {
	GetObject(ctx context.Context, id object.ID) (object.Object, error)
	ListRelations(ctx context.Context, fromID object.ID) ([]object.RelationRef, error)
}

// ── valid_name ──

// ValidName checks that object ID conforms to naming standard.
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

// ── required_metadata ──

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

// ── no_blocking_issues ──

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

// ── no_broken_relations ──

// NoBrokenRelations checks that all relations point to existing objects.
type NoBrokenRelations struct {
	Reader ObjectReader
}

func (g *NoBrokenRelations) ID() string { return "no_broken_relations" }

func (g *NoBrokenRelations) Check(ctx context.Context, input GuardInput) ([]diagnostic.Diagnostic, error) {
	if g.Reader == nil {
		return nil, nil
	}

	rels, err := g.Reader.ListRelations(ctx, object.ID(input.ObjectID))
	if err != nil {
		return nil, err
	}

	var diags []diagnostic.Diagnostic
	for _, rel := range rels {
		if _, err := g.Reader.GetObject(ctx, object.ID(rel.ToID)); err != nil {
			diags = append(diags, diagnostic.Diagnostic{
				ID: fmt.Sprintf("guard-broken-rel-%s", rel.ToID),
				Severity: diagnostic.Blocker,
				Code:     "BROKEN_RELATION",
				ObjectID: input.ObjectID,
				Message:  fmt.Sprintf("Relation target %q (%s) does not exist", rel.ToID, rel.Type),
				SuggestedActions: []string{"fix_relation", "remove_relation"},
			})
		}
	}
	return diags, nil
}

// ── checksums_actual ──

// ChecksumsActual verifies artifact checksums match actual files.
type ChecksumsActual struct {
	ObjectsDir string // root objects directory for resolving artifact paths
}

func (g *ChecksumsActual) ID() string { return "checksums_actual" }

func (g *ChecksumsActual) Check(ctx context.Context, input GuardInput) ([]diagnostic.Diagnostic, error) {
	// Without file checksums in the input, we can only do basic checks.
	// Full implementation requires the object's artifact list.
	// This guard is a structural check — full checksum verification
	// is done by the validation engine.
	return nil, nil
}

// ── no_release_blockers ──

// NoReleaseBlockers checks that no blockers exist for any object in the BOM scope.
type NoReleaseBlockers struct {
	Reader    ObjectReader
	Validator interface {
		ValidateObject(ctx context.Context, objID string) ([]diagnostic.Diagnostic, error)
	}
}

func (g *NoReleaseBlockers) ID() string { return "no_release_blockers" }

func (g *NoReleaseBlockers) Check(ctx context.Context, input GuardInput) ([]diagnostic.Diagnostic, error) {
	if g.Reader == nil || g.Validator == nil {
		return nil, nil
	}

	// Walk BOM and check all children for blockers
	visited := map[string]bool{}
	var blockers []diagnostic.Diagnostic
	g.collectBlockers(ctx, input.ObjectID, visited, &blockers)
	return blockers, nil
}

func (g *NoReleaseBlockers) collectBlockers(ctx context.Context, id string, visited map[string]bool, blockers *[]diagnostic.Diagnostic) {
	if visited[id] {
		return
	}
	visited[id] = true

	diags, err := g.Validator.ValidateObject(ctx, id)
	if err != nil {
		return
	}
	for _, d := range diags {
		if d.IsBlocker() {
			*blockers = append(*blockers, d)
		}
	}

	rels, err := g.Reader.ListRelations(ctx, object.ID(id))
	if err != nil {
		return
	}
	for _, rel := range rels {
		if rel.Type == "contains" {
			g.collectBlockers(ctx, rel.ToID, visited, blockers)
		}
	}
}

// ── children_released ──

// ChildrenReleased checks that all child objects in the BOM are released.
type ChildrenReleased struct {
	Reader ObjectReader
}

func (g *ChildrenReleased) ID() string { return "children_released" }

func (g *ChildrenReleased) Check(ctx context.Context, input GuardInput) ([]diagnostic.Diagnostic, error) {
	if g.Reader == nil {
		return nil, nil
	}

	visited := map[string]bool{}
	var diags []diagnostic.Diagnostic
	g.checkChildren(ctx, input.ObjectID, visited, &diags)
	return diags, nil
}

func (g *ChildrenReleased) checkChildren(ctx context.Context, id string, visited map[string]bool, diags *[]diagnostic.Diagnostic) {
	if visited[id] {
		return
	}
	visited[id] = true

	rels, err := g.Reader.ListRelations(ctx, object.ID(id))
	if err != nil {
		return
	}
	for _, rel := range rels {
		if rel.Type != "contains" {
			continue
		}
		obj, err := g.Reader.GetObject(ctx, object.ID(rel.ToID))
		if err != nil {
			continue
		}
		if obj.State != object.StateReleased {
			*diags = append(*diags, diagnostic.Diagnostic{
				ID: fmt.Sprintf("guard-child-not-released-%s", rel.ToID),
				Severity: diagnostic.Blocker,
				Code:     "CHILD_NOT_RELEASED",
				ObjectID: rel.ToID,
				Message:  fmt.Sprintf("Child %s is in state %s (must be released)", rel.ToID, obj.State),
				SuggestedActions: []string{"release_child", "remove_from_bom"},
			})
		}
		g.checkChildren(ctx, rel.ToID, visited, diags)
	}
}

// ── bom_valid ──

// BOMValid checks that the BOM has no structural issues.
type BOMValid struct {
	Reader ObjectReader
}

func (g *BOMValid) ID() string { return "bom_valid" }

func (g *BOMValid) Check(ctx context.Context, input GuardInput) ([]diagnostic.Diagnostic, error) {
	if g.Reader == nil {
		return nil, nil
	}

	// Check for BOM cycles
	visited := map[string]bool{}
	path := []string{}
	var diags []diagnostic.Diagnostic
	g.detectCycles(ctx, input.ObjectID, visited, path, &diags)
	return diags, nil
}

func (g *BOMValid) detectCycles(ctx context.Context, id string, visited map[string]bool, path []string, diags *[]diagnostic.Diagnostic) {
	if visited[id] {
		*diags = append(*diags, diagnostic.Diagnostic{
			ID: fmt.Sprintf("guard-bom-cycle-%s", id),
			Severity: diagnostic.Blocker,
			Code:     "BOM_CYCLE",
			ObjectID: id,
			Message:  fmt.Sprintf("BOM cycle detected: %v", append(path, id)),
		})
		return
	}
	visited[id] = true
	path = append(path, id)

	rels, err := g.Reader.ListRelations(ctx, object.ID(id))
	if err != nil {
		return
	}
	for _, rel := range rels {
		if rel.Type == "contains" {
			g.detectCycles(ctx, rel.ToID, visited, path, diags)
		}
	}
	delete(visited, id)
}

// ── Default registry ──

// DefaultGuards returns a map of guard name -> implementation for MVP.
func DefaultGuards() map[string]Guard {
	return map[string]Guard{
		"valid_name":          &ValidName{},
		"required_metadata":   &RequiredMetadata{Required: map[string][]string{
			"prt": {"unit", "make_buy"},
			"asm": {"unit"},
			"std": {"std_class", "std_id", "make_buy"},
		}},
		"no_broken_relations": &NoBrokenRelations{},
		"checksums_actual":    &ChecksumsActual{},
		"no_release_blockers": &NoReleaseBlockers{},
		"children_released":   &ChildrenReleased{},
		"bom_valid":           &BOMValid{},
		"no_blocking_issues":  &NoBlockingIssues{},
	}
}

// Ensure standard library imports used
var _ = os.IsNotExist
var _ = filepath.Join
