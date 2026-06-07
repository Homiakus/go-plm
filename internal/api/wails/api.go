// Package wailsapi exposes PLM operations to the Wails frontend via bound methods.
package wailsapi

import (
	"context"

	"github.com/Homiakus/go-plm/internal/api/dto"
	"github.com/Homiakus/go-plm/internal/app/service"
	"github.com/Homiakus/go-plm/internal/core/diagnostic"
	"github.com/Homiakus/go-plm/internal/core/object"
	"github.com/Homiakus/go-plm/internal/modules/bom"
)

// API is the Wails-bound API surface.
// Every exported method becomes a callable frontend function.
type API struct {
	app *service.App
}

// NewAPI creates a Wails API from an App instance.
func NewAPI(app *service.App) *API {
	return &API{app: app}
}

// ── Objects ──

// CreateObject creates a new engineering object.
func (a *API) CreateObject(req dto.CreateObjectRequest) (*dto.CreateObjectResponse, error) {
	return a.app.Cmd.CreateObject(context.Background(), req)
}

// GetObject retrieves a single object by ID.
func (a *API) GetObject(id string) (*dto.ObjectDTO, error) {
	return a.app.Qry.GetObject(context.Background(), id)
}

// ListObjects returns all objects.
func (a *API) ListObjects() ([]dto.ObjectDTO, error) {
	return a.app.Qry.ListObjects(context.Background())
}

// SearchObjects performs FTS5 full-text search.
func (a *API) SearchObjects(req dto.SearchRequest) ([]dto.SearchResult, error) {
	return a.app.Qry.SearchObjects(context.Background(), req)
}

// DeleteObject removes an object by ID.
func (a *API) DeleteObject(id string) error {
	return a.app.Repo.DeleteObject(context.Background(), object.ID(id))
}

// ── Lifecycle Transitions ──

// RunTransition executes a lifecycle state change.
func (a *API) RunTransition(req dto.TransitionRequest) (*dto.TransitionResponse, error) {
	return a.app.Cmd.RunTransition(context.Background(), req)
}

// AvailableTransitions returns transitions available from the current state.
func (a *API) AvailableTransitions(objectID string) ([]map[string]any, error) {
	obj, err := a.app.Qry.GetObject(context.Background(), objectID)
	if err != nil {
		return nil, err
	}
	transitions := a.app.FSM.Available(obj.State)
	result := make([]map[string]any, len(transitions))
	for i, tr := range transitions {
		result[i] = map[string]any{
			"name":    tr.Name,
			"to":      tr.To,
			"guards":  tr.Guards,
			"effects": tr.Effects,
		}
	}
	return result, nil
}

// ── BOM ──

// GetBOM builds a structured or flat BOM.
func (a *API) GetBOM(req dto.BOMRequest) (*dto.BOMResponse, error) {
	return a.app.Qry.GetBOM(context.Background(), req)
}

// ValidateBOM checks BOM rows for issues.
func (a *API) ValidateBOM(rootID string) ([]diagnostic.Diagnostic, error) {
	resp, err := a.app.Qry.GetBOM(context.Background(), dto.BOMRequest{RootID: rootID})
	if err != nil {
		return nil, err
	}
	rows := make([]bom.BOMRow, len(resp.Rows))
	for i, r := range resp.Rows {
		rows[i] = bom.BOMRow{
			RowID: r.RowID, ParentID: r.ParentID, ChildID: r.ChildID,
			ChildClass: r.ChildClass, Quantity: r.Quantity, Unit: r.Unit,
			Position: r.Position, MakeBuy: r.MakeBuy, Level: r.Level,
		}
	}
	return a.app.BOM.Validate(rows), nil
}

// DetectBOMCycles checks for circular dependencies.
func (a *API) DetectBOMCycles(rootID string) ([]diagnostic.Diagnostic, error) {
	return a.app.BOM.DetectCycles(context.Background(), object.ID(rootID))
}

// ── Release ──

// CheckReleaseReadiness checks if all objects in scope are ready.
func (a *API) CheckReleaseReadiness(rootID string) (*dto.ReleaseResponse, error) {
	return a.app.Qry.CheckReleaseReadiness(context.Background(), rootID)
}

// BuildReleaseScope collects all objects in a release scope.
func (a *API) BuildReleaseScope(rootID string) ([]string, error) {
	scope, err := a.app.Release.BuildScope(context.Background(), object.ID(rootID))
	if err != nil {
		return nil, err
	}
	return scope.Objects, nil
}

// ── Project Tree ──

// GetTree returns tree nodes for the project explorer.
func (a *API) GetTree(req dto.TreeRequest) ([]dto.TreeNodeDTO, error) {
	return a.app.Qry.GetTree(context.Background(), req)
}

// ── Git ──

// CreateCheckpoint creates a git commit.
func (a *API) CreateCheckpoint(message string) (string, error) {
	return a.app.Cmd.CreateCheckpoint(context.Background(), message)
}

// GitStatus returns working tree status.
func (a *API) GitStatus() (map[string]any, error) {
	mod, add, del, clean, err := a.app.Cmd.Git.Status()
	return map[string]any{
		"modified": mod,
		"added":    add,
		"deleted":  del,
		"clean":    clean,
	}, err
}

// ── Index ──

// RebuildIndex triggers a full index rebuild.
func (a *API) RebuildIndex() error {
	return a.app.RebuildIndex(context.Background())
}

// Stats returns index statistics.
func (a *API) Stats() (map[string]int, error) {
	objs, rels, arts, err := a.app.Index.Stats(context.Background())
	return map[string]int{"objects": objs, "relations": rels, "artifacts": arts}, err
}

// ── Project ──

// ProjectInfo returns project configuration.
func (a *API) ProjectInfo() map[string]any {
	return map[string]any{
		"code":        a.app.Config.Project.Code,
		"title":       a.app.Config.Project.Title,
		"description": a.app.Config.Project.Description,
		"version":     a.app.Config.Project.Version,
		"root":        a.app.Root,
	}
}

// GetNamingInfo returns naming context for ID generation.
func (a *API) GetNamingInfo() map[string]any {
	return map[string]any{
		"project_code": a.app.Config.Naming.ProjectCode,
		"standard":     a.app.Config.Naming.Standard,
		"pattern":      a.app.Config.Naming.Pattern,
	}
}

// ── Attachments ──

// AttachArtifact copies a file into the object's directory.
func (a *API) AttachArtifact(objectID, localPath, kind, role string) error {
	return a.app.Cmd.AttachArtifact(context.Background(), objectID, localPath, kind, role)
}

// UpdateObject saves changes to an object's title and metadata.
func (a *API) UpdateObject(id, title string, metadata map[string]any) error {
	return a.app.Cmd.UpdateObject(context.Background(), object.ID(id), title, metadata)
}

// DuplicateObject creates a copy of an object.
func (a *API) DuplicateObject(sourceID string) (string, error) {
	return a.app.Cmd.DuplicateObject(context.Background(), object.ID(sourceID))
}

// CreateReleasePackage creates a release package object.
func (a *API) CreateReleasePackage(rootID, title string) (string, error) {
	return a.app.Cmd.CreateReleasePackage(context.Background(), object.ID(rootID), title)
}

// ── Sequences ──

// GetNextSequence returns the next sequence value for a class.
func (a *API) GetNextSequence(class string) int {
	return a.app.NamingGen.Store.Current(class) + 1
}

// ── Where-Used ──

// WhereUsed returns all objects referencing the given object.
func (a *API) WhereUsed(objectID string) ([]dto.ObjectDTO, error) {
	return a.app.Qry.WhereUsed(context.Background(), objectID)
}

// ── BOM Export ──

// ExportBOM exports a BOM in the specified format.
func (a *API) ExportBOM(rootID, format string) (string, error) {
	data, err := a.app.Qry.ExportBOM(context.Background(), rootID, format)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ── Revision Bump ──

// BumpRevision creates a new revision of an object.
func (a *API) BumpRevision(objectID string) (string, error) {
	return a.app.Cmd.BumpRevision(context.Background(), objectID)
}

// ── Relations ──

// AddRelation adds a relation between two objects.
func (a *API) AddRelation(fromID, toID, relType string, quantity *float64, unit string) error {
	return a.app.Cmd.AddRelation(context.Background(), fromID, toID, relType, quantity, unit)
}

// ── Recovery ──

// RecoverTransactions finds and recovers incomplete transactions.
func (a *API) RecoverTransactions() ([]map[string]string, error) {
	items, err := a.app.Repo.Tx.Recover(context.Background())
	if err != nil {
		return nil, err
	}
	result := make([]map[string]string, len(items))
	for i, item := range items {
		result[i] = map[string]string{"id": item.ID, "work_dir": item.WorkDir}
	}
	return result, nil
}

// RollbackTransaction rolls back an incomplete transaction.
func (a *API) RollbackTransaction(id string) error {
	return a.app.Repo.Tx.Rollback(context.Background(), id)
}
