// Package release provides release package construction, readiness checks, and manifest generation.
package release

import (
	"context"
	"fmt"
	"time"

	"github.com/Homiakus/go-plm/internal/core/diagnostic"
	"github.com/Homiakus/go-plm/internal/core/object"
)

// ObjectReader abstracts reading objects.
type ObjectReader interface {
	GetObject(ctx context.Context, id object.ID) (object.Object, error)
}

// RelationLister abstracts reading relations.
type RelationLister interface {
	ListRelations(ctx context.Context, fromID object.ID) ([]object.RelationRef, error)
}

// Scope contains the list of objects included in a release.
type Scope struct {
	RootID  string   `json:"root_id"`
	Objects []string `json:"objects"`
}

// Readiness describes the release readiness of a scope.
type Readiness struct {
	Ready    bool                  `json:"ready"`
	Blockers []diagnostic.Diagnostic `json:"blockers,omitempty"`
}

// Manifest is the release package manifest.
type Manifest struct {
	ReleaseID  string           `json:"release_id"`
	RootObject string           `json:"root_object"`
	CreatedAt  string           `json:"created_at"`
	Objects    []ManifestObject `json:"objects"`
}

// ManifestObject is an object entry in the manifest.
type ManifestObject struct {
	ID       string `json:"id"`
	Class    string `json:"class"`
	State    string `json:"state"`
	Checksum string `json:"checksum,omitempty"`
}

// Service builds release packages.
type Service struct {
	Objects   ObjectReader
	Relations RelationLister
}

// New creates a release service.
func New(objects ObjectReader, relations RelationLister) *Service {
	return &Service{Objects: objects, Relations: relations}
}

// BuildScope collects all objects reachable from the root via 'contains' relations.
func (s *Service) BuildScope(ctx context.Context, rootID object.ID) (*Scope, error) {
	visited := make(map[string]bool)
	ids := []string{}
	if err := s.collectScope(ctx, string(rootID), visited, &ids); err != nil {
		return nil, err
	}
	return &Scope{RootID: string(rootID), Objects: ids}, nil
}

func (s *Service) collectScope(ctx context.Context, id string, visited map[string]bool, ids *[]string) error {
	if visited[id] {
		return nil
	}
	visited[id] = true
	*ids = append(*ids, id)

	rels, err := s.Relations.ListRelations(ctx, object.ID(id))
	if err != nil {
		return err
	}
	for _, rel := range rels {
		if rel.Type == "contains" {
			if err := s.collectScope(ctx, rel.ToID, visited, ids); err != nil {
				return err
			}
		}
	}
	return nil
}

// CheckReadiness verifies that all objects in the scope are ready for release.
func (s *Service) CheckReadiness(ctx context.Context, scope *Scope) (*Readiness, error) {
	var blockers []diagnostic.Diagnostic

	for _, id := range scope.Objects {
		obj, err := s.Objects.GetObject(ctx, object.ID(id))
		if err != nil {
			blockers = append(blockers, diagnostic.Diagnostic{
				ID: "rel-obj-not-found-" + id, Severity: diagnostic.Blocker,
				Code: "OBJECT_NOT_FOUND", ObjectID: id,
				Message: fmt.Sprintf("Object %s not found", id),
			})
			continue
		}
		if obj.State != object.StateApproved && obj.State != object.StateReleased {
			blockers = append(blockers, diagnostic.Diagnostic{
				ID: "rel-state-" + id, Severity: diagnostic.Blocker,
				Code: "OBJECT_NOT_APPROVED", ObjectID: id,
				Message: fmt.Sprintf("Object %s is in state %s (must be approved or released)", id, obj.State),
			})
		}
	}

	return &Readiness{
		Ready:    len(blockers) == 0,
		Blockers: blockers,
	}, nil
}

// GenerateManifest creates a release manifest from a scope.
func (s *Service) GenerateManifest(ctx context.Context, releaseID string, scope *Scope) (*Manifest, error) {
	manifest := &Manifest{
		ReleaseID:  releaseID,
		RootObject: scope.RootID,
		CreatedAt:  time.Now().UTC().Format(time.RFC3339),
	}

	for _, id := range scope.Objects {
		obj, err := s.Objects.GetObject(ctx, object.ID(id))
		mo := ManifestObject{ID: id}
		if err == nil {
			mo.Class = string(obj.Class)
			mo.State = string(obj.State)
		}
		manifest.Objects = append(manifest.Objects, mo)
	}

	return manifest, nil
}
