// Package command implements application-level commands that change system state.
// Commands orchestrate domain modules, store, validation, and gitops.
package command

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Homiakus/go-plm/internal/api/dto"
	"github.com/Homiakus/go-plm/internal/core/diagnostic"
	"github.com/Homiakus/go-plm/internal/core/event"
	"github.com/Homiakus/go-plm/internal/core/object"
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
	ObjectDir(id object.ID) string // path to object directory
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
func (s *ObjectService) CreateObject(ctx context.Context, req dto.CreateObjectRequest) (*dto.CreateObjectResponse, error) {
	// 1. Generate ID
	idStr := s.Naming.NextID(req.Class)
	parsed, err := naming.Parse(idStr)
	if err != nil {
		return nil, fmt.Errorf("command: generate id: %w", err)
	}

	// 2. Build domain object
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

	// 3. Persist to filesystem
	if err := s.Repo.SaveObject(ctx, obj); err != nil {
		return nil, fmt.Errorf("command: save object: %w", err)
	}

	// 4. Update index
	if err := s.Index.UpsertObject(ctx, obj); err != nil {
		return nil, fmt.Errorf("command: index object: %w", err)
	}

	// 5. Record event
	evt := event.New(
		fmt.Sprintf("evt-%s-%d", time.Now().Format("20060102"), time.Now().UnixNano()%1000000),
		"local-user",
		event.ObjectCreated,
		idStr,
		map[string]any{"class": req.Class, "title": req.Title},
	)
	evtJSON, _ := json.Marshal(evt)
	if err := s.Repo.AppendHistory(ctx, obj.ID, string(evtJSON)); err != nil {
		return nil, fmt.Errorf("command: append history: %w", err)
	}

	return &dto.CreateObjectResponse{
		ObjectID: idStr,
	}, nil
}

// RunTransition executes a lifecycle transition on an object.
func (s *ObjectService) RunTransition(ctx context.Context, req dto.TransitionRequest) (*dto.TransitionResponse, error) {
	obj, err := s.Repo.GetObject(ctx, object.ID(req.ObjectID))
	if err != nil {
		return nil, fmt.Errorf("command: get object: %w", err)
	}

	if !s.FSM.Can(string(obj.State), req.Transition) {
		return nil, fmt.Errorf("command: transition %q not available from state %q", req.Transition, obj.State)
	}

	newState, err := s.FSM.Next(string(obj.State), req.Transition)
	if err != nil {
		return nil, err
	}

	// Update state
	obj.State = object.State(newState)
	if err := s.Repo.SaveObject(ctx, obj); err != nil {
		return nil, err
	}
	if err := s.Index.UpsertObject(ctx, obj); err != nil {
		return nil, err
	}

	// Record event
	evt := event.New(
		fmt.Sprintf("evt-%s-%d", time.Now().Format("20060102"), time.Now().UnixNano()%1000000),
		"local-user",
		event.LifecycleTransitioned,
		string(obj.ID),
		map[string]any{"from": obj.State, "to": newState, "transition": req.Transition},
	)
	evtJSON, _ := json.Marshal(evt)
	s.Repo.AppendHistory(ctx, obj.ID, string(evtJSON))

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

// AttachArtifact copies a file into the object's files/ directory and updates the frontmatter.
func (s *ObjectService) AttachArtifact(ctx context.Context, objectID string, localPath string, kind, role string) error {
	obj, err := s.Repo.GetObject(ctx, object.ID(objectID))
	if err != nil {
		return fmt.Errorf("attach: get object: %w", err)
	}

	objDir := s.Repo.ObjectDir(object.ID(objectID))

	// Determine target subdirectory based on kind
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

	// Copy file
	src, err := os.ReadFile(localPath)
	if err != nil {
		return fmt.Errorf("attach: read source: %w", err)
	}
	if err := os.WriteFile(targetPath, src, 0644); err != nil {
		return fmt.Errorf("attach: write target: %w", err)
	}

	// Update object's artifact list
	relPath, _ := filepath.Rel(objDir, targetPath)
	obj.Artifacts = append(obj.Artifacts, object.ArtifactRef{
		ID:           fmt.Sprintf("art-%s-%d", objectID, len(obj.Artifacts)+1),
		Kind:         kind,
		Role:         role,
		Path:         relPath,
		OriginalName: base,
		Status:       "present",
	})

	if err := s.Repo.SaveObject(ctx, obj); err != nil {
		return fmt.Errorf("attach: save object: %w", err)
	}
	if err := s.Index.UpsertObject(ctx, obj); err != nil {
		return fmt.Errorf("attach: index: %w", err)
	}

	return nil
}
