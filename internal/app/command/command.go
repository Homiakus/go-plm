// Package command implements application-level commands that change system state.
// Commands orchestrate domain modules, store, validation, and gitops.
package command

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"
	"gopkg.in/yaml.v3"

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
	SaveObjectDocument(ctx context.Context, obj object.Object, body string) error
	GetObject(ctx context.Context, id object.ID) (object.Object, error)
	ListObjects(ctx context.Context) ([]object.Object, []error)
	DeleteObject(ctx context.Context, id object.ID) error
	AppendHistory(ctx context.Context, id object.ID, line string) error
	Exists(id object.ID) bool
	ObjectDir(id object.ID) string
	SaveObjectWithFile(ctx context.Context, obj object.Object, filePath string, fileContent []byte) error
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
	var parent *object.Object
	if req.ParentObjectID != "" {
		if p, err := s.Repo.GetObject(ctx, object.ID(req.ParentObjectID)); err == nil {
			parent = &p
		} else {
			return nil, fmt.Errorf("command: parent object %q not found: %w", req.ParentObjectID, err)
		}
	}

	idStr, err := s.nextAvailableID(req.Class)
	if err != nil {
		return nil, err
	}
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
		Metadata: autofillMetadata(req.Class, req.Metadata, parent),
	}
	obj.Relations = initialRelations(obj, parent)

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
		fmt.Sprintf("evt-%s-%s", time.Now().Format("20060102"), newEventSuffix()),
		"local-user", event.ObjectCreated, idStr,
		map[string]any{"class": req.Class, "title": req.Title},
	)
	evtJSON, _ := json.Marshal(evt)

	// SaveObjectDocument must complete first — it creates the object directory.
	if err := s.Repo.SaveObjectDocument(ctx, obj, defaultMarkdownBody(obj, parent)); err != nil {
		return nil, fmt.Errorf("command: save object: %w", err)
	}

	// Fan out: index + history can run in parallel after directory exists
	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error { return s.Index.UpsertObject(ctx, obj) })
	g.Go(func() error { return s.Repo.AppendHistory(ctx, obj.ID, string(evtJSON)) })

	if err := g.Wait(); err != nil {
		return nil, err
	}

	// Auto-create authoring relation in the parent when the parent owns the link.
	if req.ParentObjectID != "" {
		if err := s.addParentRelation(ctx, *parent, obj); err != nil {
			// Non-fatal: object is created, parent relation failed
			// In production this should be logged
		}
	}

	return &dto.CreateObjectResponse{ObjectID: idStr}, nil
}

func autofillMetadata(class string, user map[string]any, parent *object.Object) map[string]any {
	meta := map[string]any{
		"lifecycle_stage": "development",
		"owner":           "local-user",
	}
	if parent != nil {
		if stage, ok := parent.Metadata["lifecycle_stage"]; ok {
			meta["lifecycle_stage"] = stage
		}
	}

	switch class {
	case string(object.ClassPart):
		meta["unit"] = "pcs"
		meta["make_buy"] = "make"
		meta["material"] = ""
		meta["thickness_mm"] = nil
		meta["manufacturing_method"] = nil
	case string(object.ClassAssembly):
		meta["unit"] = "pcs"
		meta["assembly_type"] = "mechanical"
	case string(object.ClassDrawing):
		meta["format"] = "A3"
		meta["projection"] = "first_angle"
		if parent != nil {
			meta["drawing_for"] = string(parent.ID)
			meta["source_revision"] = parent.Revision
			if material, ok := parent.Metadata["material"]; ok {
				meta["material"] = material
			}
		}
	case string(object.ClassDocument):
		meta["doc_type"] = "engineering"
	case string(object.ClassStandardPart):
		meta["unit"] = "pcs"
		meta["make_buy"] = "buy"
		meta["std_class"] = ""
		meta["std_id"] = ""
		meta["standard"] = ""
	case string(object.ClassMaterial):
		meta["unit"] = "pcs"
		meta["make_buy"] = "buy"
		meta["material"] = ""
		meta["form"] = ""
	case string(object.ClassCutting):
		meta["process"] = "laser_cut"
		inheritSourceMetadata(meta, parent, "source_part")
	case string(object.ClassBending):
		meta["process"] = "bending"
		inheritSourceMetadata(meta, parent, "source_part")
	case string(object.ClassNC):
		meta["process"] = "machining"
		meta["machine"] = nil
		inheritSourceMetadata(meta, parent, "source_part")
	case string(object.ClassInspection):
		inheritSourceMetadata(meta, parent, "source_object")
	case string(object.ClassTechProcess), string(object.ClassWorkInstruction):
		inheritSourceMetadata(meta, parent, "source_object")
	}

	for k, v := range user {
		meta[k] = v
	}
	return meta
}

func inheritSourceMetadata(meta map[string]any, parent *object.Object, sourceKey string) {
	if parent == nil {
		return
	}
	meta[sourceKey] = string(parent.ID)
	meta["source_revision"] = parent.Revision
	for _, key := range []string{"material", "thickness_mm", "manufacturing_method"} {
		if value, ok := parent.Metadata[key]; ok {
			meta[key] = value
		}
	}
}

func initialRelations(obj object.Object, parent *object.Object) []object.RelationRef {
	if parent == nil {
		return nil
	}
	switch obj.Class {
	case object.ClassDrawing:
		return []object.RelationRef{{ToID: string(parent.ID), Type: "drawing_of"}}
	case object.ClassCutting, object.ClassBending, object.ClassNC:
		return []object.RelationRef{
			{ToID: string(parent.ID), Type: "manufacturing_file_for"},
			{ToID: string(parent.ID), Type: "derived_from"},
		}
	case object.ClassInspection:
		return []object.RelationRef{{ToID: string(parent.ID), Type: "inspection_for"}}
	case object.ClassTechProcess:
		return []object.RelationRef{{ToID: string(parent.ID), Type: "process_plan_for"}}
	case object.ClassWorkInstruction:
		return []object.RelationRef{{ToID: string(parent.ID), Type: "work_instruction_for"}}
	}
	return nil
}

func (s *ObjectService) addParentRelation(ctx context.Context, parent object.Object, child object.Object) error {
	switch child.Class {
	case object.ClassPart, object.ClassAssembly, object.ClassStandardPart:
		if parent.Class == object.ClassAssembly {
			qty := 1.0
			return s.AddRelation(ctx, string(parent.ID), string(child.ID), "contains", &qty, "pcs")
		}
	case object.ClassDrawing:
		if parent.Class == object.ClassPart || parent.Class == object.ClassAssembly || parent.Class == object.ClassStandardPart {
			return s.AddRelation(ctx, string(parent.ID), string(child.ID), "has_drawing", nil, "")
		}
	}
	return nil
}

func defaultMarkdownBody(obj object.Object, parent *object.Object) string {
	switch obj.Class {
	case object.ClassPart:
		return "# " + obj.Title + "\n\n## Purpose\n\n## Material\n\n## Main Dimensions\n\n## Manufacturing Notes\n\n## Quality Control\n\n## Related Documents\n"
	case object.ClassAssembly:
		return "# " + obj.Title + "\n\n## Purpose\n\n## Assembly Overview\n\n## BOM Notes\n\n## Interfaces\n\n## Release Notes\n"
	case object.ClassDrawing:
		return "# " + obj.Title + "\n\n## Drawing Scope\n\n## Source Object\n\n" + parentLine(parent) + "\n## Notes\n"
	case object.ClassCutting, object.ClassBending, object.ClassNC:
		return "# " + obj.Title + "\n\n## Source\n\n" + parentLine(parent) + "\n## Process Parameters\n\n## Machine Setup\n\n## Quality Checks\n"
	case object.ClassInspection:
		return "# " + obj.Title + "\n\n## Source Object\n\n" + parentLine(parent) + "\n## Inspection Plan\n\n## Critical Dimensions\n\n## Evidence\n"
	case object.ClassRelease:
		return "# " + obj.Title + "\n\n## Scope\n\n## Included Objects\n\n## Validation Summary\n\n## Approval\n"
	default:
		return "# " + obj.Title + "\n\n## Description\n\n## Notes\n"
	}
}

func parentLine(parent *object.Object) string {
	if parent == nil {
		return ""
	}
	return "- " + string(parent.ID) + " — " + parent.Title + "\n"
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
		diags, err := s.checkTransitionGuards(ctx, obj, tr.Guards)
		if err != nil {
			return &dto.TransitionResponse{NewState: string(obj.State), Diagnostics: diags}, err
		}
		if hasBlockers(diags) {
			return &dto.TransitionResponse{NewState: string(obj.State), Diagnostics: diags},
				fmt.Errorf("command: transition %q blocked by %d diagnostics", req.Transition, len(diags))
		}
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
		fmt.Sprintf("evt-%s-%s", time.Now().Format("20060102"), newEventSuffix()),
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

	autoRoute := isAutoArtifactValue(kind) || isAutoArtifactValue(role)
	kind, role = inferArtifactKindRole(obj, localPath, kind, role)

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

	base := filepath.Base(localPath)
	ext := strings.ToLower(filepath.Ext(base))
	targetName := base
	if autoRoute && ext != "" {
		targetName = string(obj.ID) + ext
	}
	if targetName == "" {
		targetName = base
	}
	targetPath := filepath.Join(targetDir, targetName)

	src, err := os.ReadFile(localPath)
	if err != nil {
		return fmt.Errorf("attach: read source: %w", err)
	}

	relPath, _ := filepath.Rel(objDir, targetPath)

	// Compute SHA-256 checksum
	hash := sha256.Sum256(src)
	checksum := hex.EncodeToString(hash[:])

	obj.Artifacts = append(obj.Artifacts, object.ArtifactRef{
		ID:   fmt.Sprintf("art-%s-%d", objectID, len(obj.Artifacts)+1),
		Kind: kind, Role: role, Path: relPath,
		OriginalName: base, Checksum: checksum,
		SizeBytes: int64(len(src)), Generated: false,
		Required: false, Status: "present",
	})

	// Save file first and then object metadata in a single repository transaction.
	if err := s.Repo.SaveObjectWithFile(ctx, obj, targetPath, src); err != nil {
		return fmt.Errorf("attach: save object: %w", err)
	}

	// Fan out: index + history in parallel
	evt := event.New(
		fmt.Sprintf("evt-%s-%s", time.Now().Format("20060102"), newEventSuffix()),
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

func inferArtifactKindRole(obj object.Object, localPath string, kind, role string) (string, string) {
	kind = strings.TrimSpace(strings.ToLower(kind))
	role = strings.TrimSpace(strings.ToLower(role))
	if kind == "" || kind == "auto" {
		kind = inferArtifactKind(obj, localPath)
	}
	if role == "" || role == "auto" {
		role = inferArtifactRole(obj, localPath, kind)
	}
	return kind, role
}

func isAutoArtifactValue(value string) bool {
	value = strings.TrimSpace(strings.ToLower(value))
	return value == "" || value == "auto"
}

func inferArtifactKind(obj object.Object, localPath string) string {
	ext := strings.ToLower(filepath.Ext(localPath))
	switch ext {
	case ".step", ".stp", ".iges", ".igs", ".sldprt", ".sldasm":
		return "cad"
	case ".png", ".jpg", ".jpeg", ".webp":
		return "image"
	case ".nc", ".tap", ".gcode":
		return "manufacturing"
	case ".dxf", ".dwg":
		if obj.Class == object.ClassCutting || obj.Class == object.ClassBending || obj.Class == object.ClassNC {
			return "manufacturing"
		}
		return "drawing"
	case ".pdf":
		if obj.Class == object.ClassDrawing {
			return "drawing"
		}
		return "document"
	case ".md", ".txt":
		return "document"
	default:
		return "file"
	}
}

func inferArtifactRole(obj object.Object, localPath string, kind string) string {
	ext := strings.ToLower(filepath.Ext(localPath))
	switch kind {
	case "cad":
		return "primary_cad"
	case "image":
		return "reference_image"
	case "document":
		return "supporting_document"
	case "drawing":
		if ext == ".pdf" {
			return "drawing_pdf"
		}
		if ext == ".dxf" || ext == ".dwg" {
			return "drawing_dxf"
		}
		return "drawing_file"
	case "manufacturing":
		switch obj.Class {
		case object.ClassCutting:
			return "cutting_dxf"
		case object.ClassBending:
			return "bending_file"
		case object.ClassNC:
			return "nc_program"
		default:
			return "manufacturing_file"
		}
	default:
		return "reference"
	}
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
		fmt.Sprintf("evt-%s-%s", time.Now().Format("20060102"), newEventSuffix()),
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
		fmt.Sprintf("evt-%s-%s", time.Now().Format("20060102"), newEventSuffix()),
		"local-user", event.MetadataUpdated, string(obj.ID),
		map[string]any{"title": obj.Title},
	)
	evtJSON, _ := json.Marshal(evt)

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error { return s.Index.UpsertObject(ctx, obj) })
	g.Go(func() error { return s.Repo.AppendHistory(ctx, obj.ID, string(evtJSON)) })
	return g.Wait()
}

// UpdateObjectDocument saves a full Markdown document through backend validation.
func (s *ObjectService) UpdateObjectDocument(ctx context.Context, req dto.UpdateObjectDocumentRequest) error {
	fm := strings.TrimSpace(req.Frontmatter)
	if strings.HasPrefix(fm, "---") {
		fm = strings.TrimPrefix(fm, "---")
		fm = strings.TrimSpace(fm)
		fm = strings.TrimSuffix(fm, "---")
		fm = strings.TrimSpace(fm)
	}

	var obj object.Object
	if err := yaml.Unmarshal([]byte(fm), &obj); err != nil {
		return fmt.Errorf("update document: parse frontmatter: %w", err)
	}
	if obj.ID == "" {
		obj.ID = object.ID(req.ObjectID)
	}
	if string(obj.ID) != req.ObjectID {
		return fmt.Errorf("update document: frontmatter id %q does not match %q", obj.ID, req.ObjectID)
	}
	if err := obj.ValidateIdentity(); err != nil {
		return fmt.Errorf("update document: validate: %w", err)
	}

	if err := s.Repo.SaveObjectDocument(ctx, obj, req.Body); err != nil {
		return fmt.Errorf("update document: save: %w", err)
	}

	evt := event.New(
		fmt.Sprintf("evt-%s-%s", time.Now().Format("20060102"), newEventSuffix()),
		"local-user", event.DocumentUpdated, string(obj.ID),
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
	if s.Repo.Exists(newID) {
		return "", fmt.Errorf("bump revision: target revision already exists: %s", newID)
	}

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
		fmt.Sprintf("evt-%s-%s", time.Now().Format("20060102"), newEventSuffix()),
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

	newIDStr, err := s.nextAvailableID(string(obj.Class))
	if err != nil {
		return "", err
	}
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
		fmt.Sprintf("evt-%s-%s", time.Now().Format("20060102"), newEventSuffix()),
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

	relID, err := s.nextAvailableID("rel")
	if err != nil {
		return "", err
	}
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
		fmt.Sprintf("evt-%s-%s", time.Now().Format("20060102"), newEventSuffix()),
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

func (s *ObjectService) nextAvailableID(class string) (string, error) {
	for i := 0; i < 10000; i++ {
		id := s.Naming.NextID(class)
		if !s.Repo.Exists(object.ID(id)) {
			return id, nil
		}
	}
	return "", fmt.Errorf("command: could not allocate unique id for class %q", class)
}

func (s *ObjectService) checkTransitionGuards(ctx context.Context, obj object.Object, guards []string) ([]diagnostic.Diagnostic, error) {
	var all []diagnostic.Diagnostic
	for _, guard := range guards {
		var diags []diagnostic.Diagnostic
		var err error
		switch guard {
		case "valid_name":
			diags = checkValidName(obj)
		case "required_metadata":
			diags = checkRequiredMetadata(obj)
		case "no_broken_relations":
			diags, err = s.checkNoBrokenRelations(ctx, obj)
		case "no_blocking_issues", "checksums_actual":
			diags, err = s.validateObjectBlockers(ctx, obj)
		case "no_release_blockers":
			diags, err = s.checkNoReleaseBlockers(ctx, obj)
		case "children_released":
			diags, err = s.checkChildrenReleased(ctx, obj)
		case "bom_valid":
			diags, err = s.checkBOMValid(ctx, obj)
		default:
			return all, fmt.Errorf("command: guard %q is not registered", guard)
		}
		if err != nil {
			return all, err
		}
		all = append(all, diags...)
	}
	return all, nil
}

func checkValidName(obj object.Object) []diagnostic.Diagnostic {
	if err := naming.Validate(string(obj.ID)); err != nil {
		return []diagnostic.Diagnostic{{
			ID: "guard-valid-name-" + string(obj.ID), Severity: diagnostic.Blocker,
			Code: "INVALID_NAME", ObjectID: string(obj.ID), Path: "id", Message: err.Error(),
		}}
	}
	return nil
}

func checkRequiredMetadata(obj object.Object) []diagnostic.Diagnostic {
	required := map[string][]string{
		"prt": {"unit", "make_buy"},
		"asm": {"unit"},
		"std": {"std_class", "std_id", "make_buy"},
	}
	fields := required[string(obj.Class)]
	var diags []diagnostic.Diagnostic
	for _, field := range fields {
		if _, ok := obj.Metadata[field]; !ok {
			diags = append(diags, diagnostic.Diagnostic{
				ID: "guard-required-" + field + "-" + string(obj.ID), Severity: diagnostic.Blocker,
				Code: "REQUIRED_METADATA_MISSING", ObjectID: string(obj.ID), Path: "metadata." + field,
				Message: "Required metadata field " + field + " is missing",
			})
		}
	}
	return diags
}

func (s *ObjectService) checkNoBrokenRelations(ctx context.Context, obj object.Object) ([]diagnostic.Diagnostic, error) {
	var diags []diagnostic.Diagnostic
	for _, rel := range obj.Relations {
		if _, err := s.Repo.GetObject(ctx, object.ID(rel.ToID)); err != nil {
			diags = append(diags, diagnostic.Diagnostic{
				ID: "guard-broken-relation-" + rel.ToID, Severity: diagnostic.Blocker,
				Code: "BROKEN_RELATION", ObjectID: string(obj.ID), Path: "relations",
				Message: fmt.Sprintf("Relation target %q (%s) does not exist", rel.ToID, rel.Type),
			})
		}
	}
	return diags, nil
}

func (s *ObjectService) validateObjectBlockers(ctx context.Context, obj object.Object) ([]diagnostic.Diagnostic, error) {
	var diags []diagnostic.Diagnostic
	diags = append(diags, checkValidName(obj)...)
	diags = append(diags, checkRequiredMetadata(obj)...)
	broken, err := s.checkNoBrokenRelations(ctx, obj)
	if err != nil {
		return nil, err
	}
	diags = append(diags, broken...)
	for _, a := range obj.Artifacts {
		if a.Status == "outdated" {
			diags = append(diags, diagnostic.Diagnostic{
				ID: "guard-checksum-" + a.ID, Severity: diagnostic.Blocker,
				Code: "CHECKSUM_OUTDATED", ObjectID: string(obj.ID), Path: "artifacts." + a.ID,
				Message: fmt.Sprintf("Artifact %q checksum is outdated", a.Path),
			})
		}
		if a.Required && (a.Path == "" || a.Status == "missing") {
			diags = append(diags, diagnostic.Diagnostic{
				ID: "guard-artifact-missing-" + a.ID, Severity: diagnostic.Blocker,
				Code: "ARTIFACT_MISSING", ObjectID: string(obj.ID), Path: "artifacts." + a.ID,
				Message: fmt.Sprintf("Required artifact %q is missing", a.ID),
			})
		}
	}
	return onlyBlockers(diags), nil
}

func (s *ObjectService) checkNoReleaseBlockers(ctx context.Context, obj object.Object) ([]diagnostic.Diagnostic, error) {
	var ids []object.ID
	if err := s.collectContainsScope(ctx, obj.ID, map[object.ID]bool{}, &ids); err != nil {
		return nil, err
	}
	var all []diagnostic.Diagnostic
	for _, id := range ids {
		o, err := s.Repo.GetObject(ctx, id)
		if err != nil {
			return nil, err
		}
		diags, err := s.validateObjectBlockers(ctx, o)
		if err != nil {
			return nil, err
		}
		all = append(all, diags...)
	}
	return all, nil
}

func (s *ObjectService) checkChildrenReleased(ctx context.Context, obj object.Object) ([]diagnostic.Diagnostic, error) {
	var diags []diagnostic.Diagnostic
	var walk func(object.ID) error
	visited := map[object.ID]bool{}
	walk = func(id object.ID) error {
		if visited[id] {
			return nil
		}
		visited[id] = true
		parent, err := s.Repo.GetObject(ctx, id)
		if err != nil {
			return err
		}
		for _, rel := range parent.Relations {
			if rel.Type != "contains" {
				continue
			}
			child, err := s.Repo.GetObject(ctx, object.ID(rel.ToID))
			if err != nil {
				diags = append(diags, diagnostic.Diagnostic{
					ID: "guard-child-missing-" + rel.ToID, Severity: diagnostic.Blocker,
					Code: "CHILD_NOT_FOUND", ObjectID: rel.ToID,
					Message: fmt.Sprintf("Child %s does not exist", rel.ToID),
				})
				continue
			}
			if child.State != object.StateReleased {
				diags = append(diags, diagnostic.Diagnostic{
					ID: "guard-child-not-released-" + rel.ToID, Severity: diagnostic.Blocker,
					Code: "CHILD_NOT_RELEASED", ObjectID: rel.ToID,
					Message: fmt.Sprintf("Child %s is in state %s (must be released)", rel.ToID, child.State),
				})
			}
			if err := walk(child.ID); err != nil {
				return err
			}
		}
		return nil
	}
	return diags, walk(obj.ID)
}

func (s *ObjectService) checkBOMValid(ctx context.Context, obj object.Object) ([]diagnostic.Diagnostic, error) {
	var diags []diagnostic.Diagnostic
	var walk func(object.ID, map[object.ID]bool, []object.ID) error
	walk = func(id object.ID, path map[object.ID]bool, stack []object.ID) error {
		if path[id] {
			diags = append(diags, diagnostic.Diagnostic{
				ID: "guard-bom-cycle-" + string(id), Severity: diagnostic.Blocker,
				Code: "BOM_CYCLE", ObjectID: string(id),
				Message: fmt.Sprintf("BOM cycle detected: %v", append(stack, id)),
			})
			return nil
		}
		path[id] = true
		stack = append(stack, id)
		o, err := s.Repo.GetObject(ctx, id)
		if err != nil {
			return err
		}
		for _, rel := range o.Relations {
			if rel.Type == "contains" {
				if err := walk(object.ID(rel.ToID), path, stack); err != nil {
					return err
				}
			}
		}
		delete(path, id)
		return nil
	}
	return diags, walk(obj.ID, map[object.ID]bool{}, nil)
}

func (s *ObjectService) collectContainsScope(ctx context.Context, id object.ID, visited map[object.ID]bool, ids *[]object.ID) error {
	if visited[id] {
		return nil
	}
	visited[id] = true
	*ids = append(*ids, id)
	obj, err := s.Repo.GetObject(ctx, id)
	if err != nil {
		return err
	}
	for _, rel := range obj.Relations {
		if rel.Type == "contains" {
			if err := s.collectContainsScope(ctx, object.ID(rel.ToID), visited, ids); err != nil {
				return err
			}
		}
	}
	return nil
}

func hasBlockers(diags []diagnostic.Diagnostic) bool {
	for _, d := range diags {
		if d.IsBlocker() {
			return true
		}
	}
	return false
}

func onlyBlockers(diags []diagnostic.Diagnostic) []diagnostic.Diagnostic {
	var blockers []diagnostic.Diagnostic
	for _, d := range diags {
		if d.IsBlocker() {
			blockers = append(blockers, d)
		}
	}
	return blockers
}

// newEventSuffix generates a cryptographically random suffix for event IDs.
// Avoids collision-prone UnixNano()%1000000 (~20 bits entropy).
func newEventSuffix() string {
	n, err := rand.Int(rand.Reader, big.NewInt(999999))
	if err != nil {
		return "000000"
	}
	return fmt.Sprintf("%06d", n.Int64())
}
