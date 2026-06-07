// Package dto provides Data Transfer Objects for the API layer.
// These types are used for frontend-backend communication.
package dto

import "github.com/Homiakus/go-plm/internal/core/diagnostic"

// ObjectDTO is the frontend-friendly object representation.
type ObjectDTO struct {
	ID        string         `json:"id"`
	Project   string         `json:"project"`
	Class     string         `json:"class"`
	Sequence  string         `json:"sequence,omitempty"`
	Version   string         `json:"version"`
	Revision  string         `json:"revision"`
	State     string         `json:"state"`
	Title     string         `json:"title"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	Relations []RelationDTO  `json:"relations,omitempty"`
	Artifacts []ArtifactDTO  `json:"artifacts,omitempty"`
}

// RelationDTO is a frontend representation of an object relation.
type RelationDTO struct {
	ToID     string   `json:"to"`
	Type     string   `json:"type"`
	Quantity *float64 `json:"quantity,omitempty"`
	Unit     string   `json:"unit,omitempty"`
}

// ArtifactDTO is a frontend representation of an attached file/artifact.
type ArtifactDTO struct {
	ID           string `json:"id"`
	Kind         string `json:"kind"`
	Role         string `json:"role"`
	Path         string `json:"path"`
	OriginalName string `json:"original_name,omitempty"`
	Checksum     string `json:"checksum,omitempty"`
	SizeBytes    int64  `json:"size_bytes,omitempty"`
	Generated    bool   `json:"generated"`
	Required     bool   `json:"required"`
	Status       string `json:"status"`
}

// ObjectDocumentDTO contains editable Markdown source split into frontmatter/body.
type ObjectDocumentDTO struct {
	ObjectID    string `json:"object_id"`
	Frontmatter string `json:"frontmatter"`
	Body        string `json:"body"`
}

// UpdateObjectDocumentRequest updates the Markdown source of an object.
type UpdateObjectDocumentRequest struct {
	ObjectID    string `json:"object_id"`
	Frontmatter string `json:"frontmatter"`
	Body        string `json:"body"`
}

// ProjectCreateRequest creates a new project on disk.
type ProjectCreateRequest struct {
	Path  string `json:"path"`
	Code  string `json:"code"`
	Title string `json:"title"`
}

// CreateObjectRequest is the request to create a new object.
type CreateObjectRequest struct {
	ProjectPath    string         `json:"project_path"`
	Class          string         `json:"class"`
	Title          string         `json:"title"`
	ParentObjectID string         `json:"parent_object_id,omitempty"`
	Metadata       map[string]any `json:"metadata,omitempty"`
}

// CreateObjectResponse is the response after creating an object.
type CreateObjectResponse struct {
	ObjectID    string                  `json:"object_id"`
	Diagnostics []diagnostic.Diagnostic `json:"diagnostics,omitempty"`
}

// Response wraps any API result with metadata.
type Response[T any] struct {
	OK          bool                    `json:"ok"`
	Result      T                       `json:"result,omitempty"`
	Diagnostics []diagnostic.Diagnostic `json:"diagnostics,omitempty"`
	Error       string                  `json:"error,omitempty"`
}

// SearchRequest is a search query.
type SearchRequest struct {
	Query string `json:"query"`
	Limit int    `json:"limit"`
}

// SearchResult is a single search result.
type SearchResult struct {
	ObjectID string `json:"object_id"`
	Title    string `json:"title"`
	Class    string `json:"class"`
	State    string `json:"state"`
	Excerpt  string `json:"excerpt,omitempty"`
}

// BOMRequest requests a BOM for an object.
type BOMRequest struct {
	RootID string `json:"root_id"`
	Flat   bool   `json:"flat"`
}

// BOMResponse contains BOM rows.
type BOMResponse struct {
	RootID string      `json:"root_id"`
	Rows   []BOMRowDTO `json:"rows"`
}

// BOMRowDTO is a BOM row for the frontend.
type BOMRowDTO struct {
	RowID      string  `json:"row_id"`
	ParentID   string  `json:"parent_id"`
	ChildID    string  `json:"child_id"`
	ChildClass string  `json:"child_class"`
	Quantity   float64 `json:"quantity"`
	Unit       string  `json:"unit"`
	Position   string  `json:"position,omitempty"`
	MakeBuy    string  `json:"make_buy,omitempty"`
	Level      int     `json:"level"`
}

// TransitionRequest executes a lifecycle transition.
type TransitionRequest struct {
	ObjectID   string `json:"object_id"`
	Transition string `json:"transition"`
}

// TransitionResponse is the result of a transition.
type TransitionResponse struct {
	NewState    string                  `json:"new_state"`
	Diagnostics []diagnostic.Diagnostic `json:"diagnostics,omitempty"`
}

// ReleaseRequest creates a release package.
type ReleaseRequest struct {
	RootObjectID string `json:"root_object_id"`
	ReleaseID    string `json:"release_id"`
}

// TreeNodeDTO is a tree node for the Project Tree UI.
type TreeNodeDTO struct {
	ID            string         `json:"id"`
	Class         string         `json:"class"`
	Title         string         `json:"title"`
	State         string         `json:"state"`
	Icon          string         `json:"icon"`
	StatusColor   string         `json:"status_color"`
	HasChildren   bool           `json:"has_children"`
	ChildrenCount int            `json:"children_count"`
	Warnings      int            `json:"warnings"`
	Blockers      int            `json:"blockers"`
	ThumbnailURL  string         `json:"thumbnail_url,omitempty"`
	Metadata      map[string]any `json:"metadata,omitempty"`
}

// TreeRequest is a request for tree data.
type TreeRequest struct {
	Filter string `json:"filter,omitempty"`
	Sort   string `json:"sort,omitempty"`
	Offset int    `json:"offset"`
	Limit  int    `json:"limit"`
}

// ReleaseResponse is the release result.
type ReleaseResponse struct {
	Ready    bool                    `json:"ready"`
	Blockers []diagnostic.Diagnostic `json:"blockers,omitempty"`
}
