// Package tree provides icon and status mappings for the Project Tree UI.
package tree

import "github.com/Homiakus/go-plm/internal/core/object"

// Icon returns the Lucide icon name for a given object class.
func Icon(class object.Class) string {
	switch class {
	case object.ClassAssembly:
		return "package"
	case object.ClassPart:
		return "component"
	case object.ClassDrawing:
		return "ruler"
	case object.ClassDocument:
		return "file-text"
	case object.ClassStandardPart:
		return "nut"
	case object.ClassCutting:
		return "scissors"
	case object.ClassBending:
		return "corner-down-right"
	case object.ClassNC:
		return "cpu"
	case object.ClassInspection:
		return "search-check"
	case object.ClassTechProcess:
		return "clipboard-list"
	case object.ClassWorkInstruction:
		return "list-checks"
	case object.ClassBOM:
		return "table"
	case object.ClassRelease:
		return "tag"
	case object.ClassChangeRequest:
		return "git-pull-request"
	case object.ClassMaterial:
		return "brick-wall"
	default:
		return "file"
	}
}

// StatusColor returns the CSS color for a given object state.
func StatusColor(state object.State) string {
	switch state {
	case object.StateReleased:
		return "green"
	case object.StateApproved:
		return "blue"
	case object.StateInReview:
		return "amber"
	case object.StateDraft:
		return "gray"
	case object.StateBlocked:
		return "red"
	case object.StateObsolete, object.StateArchived:
		return "slate"
	default:
		return "gray"
	}
}

// ThumbnailURL finds the best thumbnail URL for an object.
// Priority: 1) metadata.thumbnail, 2) first image artifact, 3) empty (use class icon).
func ThumbnailURL(obj object.Object) string {
	if url, ok := obj.Metadata["thumbnail"].(string); ok && url != "" {
		return url
	}
	for _, a := range obj.Artifacts {
		if a.Kind == "image" && a.Status == "present" {
			return a.Path
		}
	}
	return ""
}
