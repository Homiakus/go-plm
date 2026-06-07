// Package command implements application-level commands that change system state.
// Commands orchestrate domain modules, store, validation, and gitops.
package command

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/Homiakus/go-plm/internal/api/dto"
	"github.com/Homiakus/go-plm/internal/core/diagnostic"
	"github.com/Homiakus/go-plm/internal/core/event"
	"github.com/Homiakus/go-plm/internal/core/object"
	"github.com/Homiakus/go-plm/internal/modules/standardparts"
	"github.com/Homiakus/go-plm/internal/naming"
	"github.com/Homiakus/go-plm/internal/process/fsm"
)

// ObjectService provides a unified interface for object-related commands.
type ObjectService struct {
	Repo      ObjectRepo
	Index     Indexer
	Naming    *naming.Generator
	FSM       *fsm.Machine
	Git       GitOps
	Validator Validator
}

// ObjectRepo abstracts object storage.
type ObjectRepo interface {
	SaveObject(ctx context.Context, obj object.Object) error
	GetObject(ctx context.Context, id object.ID) (object.Object, error)
	ListObjects(ctx context.Context) ([]object.Object, []error)
	DeleteObject(ctx context.Context, id object.ID) error
	AppendHistory(ctx context.Context, id object.ID, line string) error
	Exists(id object.ID) bool
	ObjectDir(id object.ID) string
}

// Indexer abstracts index operations.
type Indexer interface {
	UpsertObject(ctx context.Context, obj object.Object) error
	DeleteObject(ctx context.Context, id object.ID) error
	GetObject(ctx context.Context, id object.ID) (object.Object, error)
	RebuildIndex(ctx context.Context, objects []object.Object) error
}

// GitOps abstracts git operations.
type GitOps interface {
	Status() (modified, added, deleted []string, clean bool, err error)
	Checkpoint(ctx context.Context, message, author string) (string, error)
}

// Validator validates objects.
type Validator interface {
	ValidateObject(ctx context.Context, objID string) ([]diagnostic.Diagnostic, error)
}

// CreateObject creates a new PLM object — SaveObject, UpsertObject, and AppendHistory
// execute in parallel via errgroup (three independent I/O backends).
func (s *ObjectService) CreateObject(ctx context.Context, req dto.CreateObjectRequest) (*dto.CreateObjectResponse, error) {
	idStr := s.Naming.NextID(req.Class)
	parsed, err := naming.Parse(idStr)
	if err != nil {
		return nil, fmt.Errorf("command: generate id: %w", err)
	}

	obj := object.Object{
		ID:       object.ID(idStr),
		Project:  parsed.Project,
		Class:    object.Class(req.Class),
		Sequence: parsed.Sequence,
		Version:  fmt.Sprintf("%d.%d", parsed.Major, parsed.Minor),
		Revision: fmt.Sprintf("%d.%d", parsed.Major, parsed.Minor),
		State:    object.StateDraft,
		Title:    req.Title,
		Metadata: req.Metadata,
	}

	if err := obj.ValidateIdentity(); err != nil {
		return nil, fmt.Errorf("command: validate identity: %w", err)
	}

	// Duplicate detection for standard parts
	if obj.Class == object.ClassStandardPart {
		diags := standardparts.Validate(obj)
		for _, d := range diags {
			if d.IsBlocker() {
				return nil, fmt.Errorf("command: std validation: %s", d.Message)
			}
		}
	}

	evt := event.New(
		fmt.Sprintf("evt-%s-%d", time.Now().Format("20060102"), time.Now().UnixNano()%1000000),
		"local-user", event.ObjectCreated, idStr,
		map[string]any{"class": req.Class, "title": req.Title},
	)
	evtJSON, _ := json.Marshal(evt)

	// SaveObject must complete first — it creates the object directory
	if err := s.Repo.SaveObject(ctx, obj); err != nil {
		return nil, fmt.Errorf("command: save object: %w", err)
	}

	// Fan out: index + history can run in parallel after directory exists
	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error { return s.Index.UpsertObject(ctx, obj) })
	g.Go(func() error { return s.Repo.AppendHistory(ctx, obj.ID, string(evtJSON)) })

	if err := g.Wait(); err != nil {
		return nil, err
	}

	// Auto-create parent relation if requested
	if req.ParentObjectID != "" {
		qty := 1.0
		if err := s.AddRelation(ctx, req.ParentObjectID, idStr, "contains", &qty, "pcs"); err != nil {
			// Non-fatal: object is created, parent relation failed
			// In production this should be logged
		}
	}

	return &dto.CreateObjectResponse{ObjectID: idStr}, nil
}

// RunTransition executes a lifecycle transition — SaveObject + UpsertObject in parallel.
// Also applies side effects (e.g., create_git_tag on release).
func (s *ObjectService) RunTransition(ctx context.Context, req dto.TransitionRequest) (*dto.TransitionResponse, error) {
	obj, err := s.Repo.GetObject(ctx, object.ID(req.ObjectID))
	if err != nil {
		return nil, fmt.Errorf("command: get object: %w", err)
	}

	if !s.FSM.Can(string(obj.State), req.Transition) {
		return nil, fmt.Errorf("command: transition %q not available from state %q", req.Transition, obj.State)
	}

	// Check effects before transition
	tr := s.FSM.GetTransition(req.Transition)
	hasCreateGitTag := false
	if tr != nil {
		for _, e := range tr.Effects {
			if e == "create_git_tag" {
				hasCreateGitTag = true
			}
		}
	}

	newState, err := s.FSM.Next(string(obj.State), req.Transition)
	if err != nil {
		return nil, err
	}

	oldState := string(obj.State)
	obj.State = object.State(newState)

	evt := event.New(
		fmt.Sprintf("evt-%s-%d", time.Now().Format("20060102"), time.Now().UnixNano()%1000000),
		"local-user", event.LifecycleTransitioned, string(obj.ID),
		map[string]any{"from": oldState, "to": newState, "transition": req.Transition},
	)
	evtJSON, _ := json.Marshal(evt)

	// SaveObject must complete first — it creates the object directory
	if err := s.Repo.SaveObject(ctx, obj); err != nil {
		return nil, err
	}

	// Fan out: index + history in parallel
	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error { return s.Index.UpsertObject(ctx, obj) })
	g.Go(func() error { return s.Repo.AppendHistory(ctx, obj.ID, string(evtJSON)) })

	if err := g.Wait(); err != nil {
		return nil, err
	}

	// Apply side effects
	if hasCreateGitTag && s.Git != nil {
		tagName := fmt.Sprintf("release/%s", req.ObjectID)
		s.Git.Checkpoint(ctx, fmt.Sprintf("Release %s — %s", req.ObjectID, req.Transition), "local-user")
		// Tag creation is done via gitops — set a best-effort tag
		if gitSvc, ok := s.Git.(interface{ CreateTag(name, msg string) error }); ok {
			gitSvc.CreateTag(tagName, fmt.Sprintf("Release transition: %s", req.Transition))
		}
	}

	return &dto.TransitionResponse{NewState: newState}, nil
}

// RebuildIndex triggers a full index rebuild from filesystem.
func (s *ObjectService) RebuildIndex(ctx context.Context) error {
	objects, errs := s.Repo.ListObjects(ctx)
	if len(errs) > 0 {
		return fmt.Errorf("command: list objects had %d errors", len(errs))
	}
	return s.Index.RebuildIndex(ctx, objects)
}

// CreateCheckpoint creates a git checkpoint.
func (s *ObjectService) CreateCheckpoint(ctx context.Context, message string) (string, error) {
	return s.Git.Checkpoint(ctx, message, "local-user")
}

// AttachArtifact copies a file into the object's files/ directory — SaveObject + UpsertObject in parallel.
func (s *ObjectService) AttachArtifact(ctx context.Context, objectID string, localPath string, kind, role string) error {
	obj, err := s.Repo.GetObject(ctx, object.ID(objectID))
	if err != nil {
		return fmt.Errorf("attach: get object: %w", err)
	}

	objDir := s.Repo.ObjectDir(object.ID(objectID))

	subDir := kind
	switch kind {
	case "cad":
		subDir = "files/cad"
	case "drawing":
		subDir = "files/drawings"
	case "manufacturing":
		subDir = "files/manufacturing"
	case "image":
		subDir = "files/images"
	case "certificate":
		subDir = "files/certificates"
	case "evidence":
		subDir = "files/evidence"
	default:
		subDir = "files"
	}

	targetDir := filepath.Join(objDir, subDir)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("attach: mkdir: %w", err)
	}

	base := filepath.Base(localPath)
	targetPath := filepath.Join(targetDir, base)

	src, err := os.ReadFile(localPath)
	if err != nil {
		return fmt.Errorf("attach: read source: %w", err)
	}
	if err := os.WriteFile(targetPath, src, 0644); err != nil {
		return fmt.Errorf("attach: write target: %w", err)
	}

	relPath, _ := filepath.Rel(objDir, targetPath)

	// Compute SHA-256 checksum
	hash := sha256.Sum256(src)
	checksum := hex.EncodeToString(hash[:])

	obj.Artifacts = append(obj.Artifacts, object.ArtifactRef{
		ID: fmt.Sprintf("art-%s-%d", objectID, len(obj.Artifacts)+1),
		Kind: kind, Role: role, Path: relPath,
		OriginalName: base, Checksum: checksum,
		SizeBytes: int64(len(src)), Generated: false,
		Required: false, Status: "present",
	})

	// SaveObject must complete first — it creates the object directory
	if err := s.Repo.SaveObject(ctx, obj); err != nil {
		return fmt.Errorf("attach: save object: %w", err)
	}

	// Fan out: index + history in parallel
	evt := event.New(
		fmt.Sprintf("evt-%s-%d", time.Now().Format("20060102"), time.Now().UnixNano()%1000000),
		"local-user", event.ArtifactAttached, objectID,
		map[string]any{"kind": kind, "role": role, "path": relPath},
	)
	evtJSON, _ := json.Marshal(evt)

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error { return s.Index.UpsertObject(ctx, obj) })
	g.Go(func() error { return s.Repo.AppendHistory(ctx, obj.ID, string(evtJSON)) })

	if err := g.Wait(); err != nil {
		return fmt.Errorf("attach: index/history: %w", err)
	}
	return nil
}

// AddRelation adds a relation from one object to another.
func (s *ObjectService) AddRelation(ctx context.Context, fromID, toID, relType string, quantity *float64, unit string) error {
	obj, err := s.Repo.GetObject(ctx, object.ID(fromID))
	if err != nil {
		return fmt.Errorf("add relation: get from object: %w", err)
	}

	obj.Relations = append(obj.Relations, object.RelationRef{
		ToID: toID, Type: relType, Quantity: quantity, Unit: unit,
	})

	if err := s.Repo.SaveObject(ctx, obj); err != nil {
		return fmt.Errorf("add relation: save: %w", err)
	}

	evt := event.New(
		fmt.Sprintf("evt-%s-%d", time.Now().Format("20060102"), time.Now().UnixNano()%1000000),
		"local-user", event.RelationAdded, fromID,
		map[string]any{"to": toID, "type": relType},
	)
	evtJSON, _ := json.Marshal(evt)

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error { return s.Index.UpsertObject(ctx, obj) })
	g.Go(func() error { return s.Repo.AppendHistory(ctx, obj.ID, string(evtJSON)) })
	return g.Wait()
}

// DeleteObjectSafe checks where-used and deletes an object only if safe.
// Returns a list of blocking references if the object cannot be safely deleted.
func (s *ObjectService) DeleteObjectSafe(ctx context.Context, id object.ID, force bool) ([]string, error) {
	if !force {
		// Check where-used
		allObjects, errs := s.Repo.ListObjects(ctx)
		if len(errs) > 0 {
			return nil, errs[0]
		}

		var usedBy []string
		for _, obj := range allObjects {
			if string(obj.ID) == string(id) {
				continue
			}
			for _, rel := range obj.Relations {
				if rel.ToID == string(id) {
					usedBy = append(usedBy, string(obj.ID))
					break
				}
			}
		}
		if len(usedBy) > 0 {
			return usedBy, nil
		}
	}

	// Delete from index first, then filesystem
	if err := s.Index.DeleteObject(ctx, id); err != nil {
		return nil, fmt.Errorf("safe delete: index: %w", err)
	}
	if err := s.Repo.DeleteObject(ctx, id); err != nil {
		return nil, fmt.Errorf("safe delete: repo: %w", err)
	}
	return nil, nil
}

// UpdateObject saves an object after editing (title, metadata, body).
func (s *ObjectService) UpdateObject(ctx context.Context, id object.ID, title string, metadata map[string]any) error {
	obj, err := s.Repo.GetObject(ctx, id)
	if err != nil {
		return fmt.Errorf("update object: %w", err)
	}

	if title != "" {
		obj.Title = title
	}
	if metadata != nil {
		obj.Metadata = metadata
	}

	if err := s.Repo.SaveObject(ctx, obj); err != nil {
		return fmt.Errorf("update object: save: %w", err)
	}

	evt := event.New(
		fmt.Sprintf("evt-%s-%d", time.Now().Format("20060102"), time.Now().UnixNano()%1000000),
		"local-user", event.MetadataUpdated, string(obj.ID),
		map[string]any{"title": obj.Title},
	)
	evtJSON, _ := json.Marshal(evt)

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error { return s.Index.UpsertObject(ctx, obj) })
	g.Go(func() error { return s.Repo.AppendHistory(ctx, obj.ID, string(evtJSON)) })
	return g.Wait()
}

// BumpRevision creates a new revision of an object (minor bump).
func (s *ObjectService) BumpRevision(ctx context.Context, objectID string) (string, error) {
	obj, err := s.Repo.GetObject(ctx, object.ID(objectID))
	if err != nil {
		return "", fmt.Errorf("bump revision: %w", err)
	}

	parsed, err := naming.Parse(string(obj.ID))
	if err != nil {
		return "", fmt.Errorf("bump revision: parse id: %w", err)
	}

	newVersion := fmt.Sprintf("%d.%d", parsed.Major, parsed.Minor+1)
	newID := object.ID(fmt.Sprintf("%s-%s-%s-v%s", parsed.Project, parsed.Class, parsed.Sequence, newVersion))

	newObj := obj
	newObj.ID = newID
	newObj.Version = newVersion
	newObj.Revision = newVersion
	newObj.State = object.StateDraft
	newObj.Relations = append(newObj.Relations, object.RelationRef{
		ToID: objectID, Type: "derived_from",
	})

	if err := s.Repo.SaveObject(ctx, newObj); err != nil {
		return "", fmt.Errorf("bump revision: save new: %w", err)
	}

	evt := event.New(
		fmt.Sprintf("evt-%s-%d", time.Now().Format("20060102"), time.Now().UnixNano()%1000000),
		"local-user", event.ObjectCreated, string(newID),
		map[string]any{"class": string(newObj.Class), "title": newObj.Title, "derived_from": objectID},
	)
	evtJSON, _ := json.Marshal(evt)

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error { return s.Index.UpsertObject(ctx, newObj) })
	g.Go(func() error { return s.Repo.AppendHistory(ctx, newObj.ID, string(evtJSON)) })
	if err := g.Wait(); err != nil {
		return "", err
	}

	return string(newID), nil
}

// DuplicateObject creates a copy of an object with a new ID.
func (s *ObjectService) DuplicateObject(ctx context.Context, sourceID object.ID) (string, error) {
	obj, err := s.Repo.GetObject(ctx, sourceID)
	if err != nil {
		return "", fmt.Errorf("duplicate: %w", err)
	}

	newIDStr := s.Naming.NextID(string(obj.Class))
	newObj := obj
	newObj.ID = object.ID(newIDStr)
	newObj.State = object.StateDraft
	newObj.Title = obj.Title + " (Copy)"
	newObj.Relations = append(newObj.Relations, object.RelationRef{
		ToID: string(sourceID), Type: "derived_from",
	})

	if err := s.Repo.SaveObject(ctx, newObj); err != nil {
		return "", fmt.Errorf("duplicate: save: %w", err)
	}

	evt := event.New(
		fmt.Sprintf("evt-%s-%d", time.Now().Format("20060102"), time.Now().UnixNano()%1000000),
		"local-user", event.ObjectDuplicated, newIDStr,
		map[string]any{"source": string(sourceID), "class": string(obj.Class)},
	)
	evtJSON, _ := json.Marshal(evt)

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error { return s.Index.UpsertObject(ctx, newObj) })
	g.Go(func() error { return s.Repo.AppendHistory(ctx, newObj.ID, string(evtJSON)) })
	if err := g.Wait(); err != nil {
		return "", err
	}
	return newIDStr, nil
}

// CreateReleasePackage creates a release package object with scope.
func (s *ObjectService) CreateReleasePackage(ctx context.Context, rootID object.ID, releaseTitle string) (string, error) {
	objects, errs := s.Repo.ListObjects(ctx)
	if len(errs) > 0 {
		return "", fmt.Errorf("release: list objects: %w", errs[0])
	}

	// Collect scope via contains relations
	visited := map[string]bool{}
	scopeIDs := []string{}
	var collect func(id string)
	collect = func(id string) {
		if visited[id] {
			return
		}
		visited[id] = true
		scopeIDs = append(scopeIDs, id)
		for _, obj := range objects {
			if string(obj.ID) == id {
				for _, rel := range obj.Relations {
					if rel.Type == "contains" {
						collect(rel.ToID)
					}
				}
				break
			}
		}
	}
	collect(string(rootID))

	relID := s.Naming.NextID("rel")
	relObj := object.Object{
		ID: object.ID(relID), Project: s.Naming.Project,
		Class: object.ClassRelease, State: object.StateReleased,
		Title: releaseTitle, Version: "1.0", Revision: "1.0",
		Metadata: map[string]any{
			"root_object": string(rootID),
			"scope":       scopeIDs,
			"scope_count": len(scopeIDs),
		},
	}

	if err := s.Repo.SaveObject(ctx, relObj); err != nil {
		return "", fmt.Errorf("release: save: %w", err)
	}

	evt := event.New(
		fmt.Sprintf("evt-%s-%d", time.Now().Format("20060102"), time.Now().UnixNano()%1000000),
		"local-user", event.ReleasePublished, relID,
		map[string]any{"root": string(rootID), "scope_count": len(scopeIDs)},
	)
	evtJSON, _ := json.Marshal(evt)

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error { return s.Index.UpsertObject(ctx, relObj) })
	g.Go(func() error { return s.Repo.AppendHistory(ctx, relObj.ID, string(evtJSON)) })
	if err := g.Wait(); err != nil {
		return "", err
	}
	return relID, nil
}
