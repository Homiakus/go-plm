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
// Uses bounded worker pool (8 goroutines) for parallel object reads.
func (s *Service) CheckReadiness(ctx context.Context, scope *Scope) (*Readiness, error) {
	const workers = 8
	sem := make(chan struct{}, workers)

	type result struct {
		id      string
		blocker *diagnostic.Diagnostic
	}
	results := make(chan result, len(scope.Objects))

	for _, id := range scope.Objects {
		sem <- struct{}{}
		go func(oid string) {
			defer func() { <-sem }()
			obj, err := s.Objects.GetObject(ctx, object.ID(oid))
			if err != nil {
				results <- result{oid, &diagnostic.Diagnostic{
					ID: "rel-obj-not-found-" + oid, Severity: diagnostic.Blocker,
					Code: "OBJECT_NOT_FOUND", ObjectID: oid,
					Message: fmt.Sprintf("Object %s not found", oid),
				}}
				return
			}
			if obj.State != object.StateApproved && obj.State != object.StateReleased {
				results <- result{oid, &diagnostic.Diagnostic{
					ID: "rel-state-" + oid, Severity: diagnostic.Blocker,
					Code: "OBJECT_NOT_APPROVED", ObjectID: oid,
					Message: fmt.Sprintf("Object %s is in state %s (must be approved or released)", oid, obj.State),
				}}
				return
			}
			results <- result{oid, nil}
		}(id)
	}
	// Drain semaphore
	for i := 0; i < workers; i++ {
		sem <- struct{}{}
	}
	close(results)

	var blockers []diagnostic.Diagnostic
	for r := range results {
		if r.blocker != nil {
			blockers = append(blockers, *r.blocker)
		}
	}

	return &Readiness{Ready: len(blockers) == 0, Blockers: blockers}, nil
}

// GenerateManifest creates a release manifest from a scope.
// Uses bounded worker pool (8 goroutines) for parallel object reads.
func (s *Service) GenerateManifest(ctx context.Context, releaseID string, scope *Scope) (*Manifest, error) {
	const workers = 8
	sem := make(chan struct{}, workers)

	type entry struct {
		idx int
		mo  ManifestObject
	}
	entries := make(chan entry, len(scope.Objects))

	for i, id := range scope.Objects {
		sem <- struct{}{}
		go func(idx int, oid string) {
			defer func() { <-sem }()
			mo := ManifestObject{ID: oid}
			obj, err := s.Objects.GetObject(ctx, object.ID(oid))
			if err == nil {
				mo.Class = string(obj.Class)
				mo.State = string(obj.State)
			}
			entries <- entry{idx, mo}
		}(i, id)
	}
	for i := 0; i < workers; i++ {
		sem <- struct{}{}
	}
	close(entries)

	// Sort back to original order
	manifestObjects := make([]ManifestObject, len(scope.Objects))
	for e := range entries {
		manifestObjects[e.idx] = e.mo
	}

	return &Manifest{
		ReleaseID:  releaseID,
		RootObject: scope.RootID,
		CreatedAt:  time.Now().UTC().Format(time.RFC3339),
		Objects:    manifestObjects,
	}, nil
}
