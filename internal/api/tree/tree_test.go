// Package tree_test provides tests for the tree UI helpers.
package tree_test

import (
	"testing"

	"github.com/Homiakus/go-plm/internal/api/tree"
	"github.com/Homiakus/go-plm/internal/core/object"
)

func TestIcon(t *testing.T) {
	tests := []struct {
		class object.Class
		want  string
	}{
		{object.ClassAssembly, "package"},
		{object.ClassPart, "component"},
		{object.ClassDrawing, "ruler"},
		{object.ClassDocument, "file-text"},
		{object.ClassStandardPart, "nut"},
		{object.ClassCutting, "scissors"},
		{object.ClassBending, "corner-down-right"},
		{object.ClassNC, "cpu"},
		{object.ClassInspection, "search-check"},
		{object.ClassTechProcess, "clipboard-list"},
		{object.ClassWorkInstruction, "list-checks"},
		{object.ClassBOM, "table"},
		{object.ClassRelease, "tag"},
		{object.ClassChangeRequest, "git-pull-request"},
		{object.ClassMaterial, "brick-wall"},
		{object.Class("unknown"), "file"},
	}
	for _, tt := range tests {
		t.Run(string(tt.class), func(t *testing.T) {
			got := tree.Icon(tt.class)
			if got != tt.want {
				t.Errorf("Icon(%q) = %q, want %q", tt.class, got, tt.want)
			}
		})
	}
}

func TestStatusColor(t *testing.T) {
	tests := []struct {
		state object.State
		want  string
	}{
		{object.StateReleased, "green"},
		{object.StateApproved, "blue"},
		{object.StateInReview, "amber"},
		{object.StateDraft, "gray"},
		{object.StateBlocked, "red"},
		{object.StateObsolete, "slate"},
		{object.StateArchived, "slate"},
		{object.State("unknown"), "gray"},
	}
	for _, tt := range tests {
		t.Run(string(tt.state), func(t *testing.T) {
			got := tree.StatusColor(tt.state)
			if got != tt.want {
				t.Errorf("StatusColor(%q) = %q, want %q", tt.state, got, tt.want)
			}
		})
	}
}

func TestThumbnailURL(t *testing.T) {
	t.Run("metadata thumbnail", func(t *testing.T) {
		obj := object.Object{
			Metadata: map[string]any{"thumbnail": "https://example.com/thumb.png"},
		}
		if got := tree.ThumbnailURL(obj); got != "https://example.com/thumb.png" {
			t.Errorf("ThumbnailURL = %q", got)
		}
	})

	t.Run("image artifact fallback", func(t *testing.T) {
		obj := object.Object{
			Artifacts: []object.ArtifactRef{
				{Kind: "drawing", Path: "drawing.pdf", Status: "present"},
				{Kind: "image", Path: "photo.png", Status: "present"},
			},
		}
		if got := tree.ThumbnailURL(obj); got != "photo.png" {
			t.Errorf("ThumbnailURL = %q, want photo.png", got)
		}
	})

	t.Run("missing image fallback", func(t *testing.T) {
		obj := object.Object{
			Artifacts: []object.ArtifactRef{
				{Kind: "image", Path: "photo.png", Status: "missing"},
			},
		}
		if got := tree.ThumbnailURL(obj); got != "" {
			t.Errorf("ThumbnailURL = %q, want empty for missing image", got)
		}
	})

	t.Run("empty", func(t *testing.T) {
		if got := tree.ThumbnailURL(object.Object{}); got != "" {
			t.Errorf("ThumbnailURL = %q, want empty", got)
		}
	})

	t.Run("metadata takes priority", func(t *testing.T) {
		obj := object.Object{
			Metadata:  map[string]any{"thumbnail": "https://example.com/thumb.png"},
			Artifacts: []object.ArtifactRef{{Kind: "image", Path: "photo.png", Status: "present"}},
		}
		if got := tree.ThumbnailURL(obj); got != "https://example.com/thumb.png" {
			t.Errorf("ThumbnailURL = %q, metadata should take priority", got)
		}
	})
}
