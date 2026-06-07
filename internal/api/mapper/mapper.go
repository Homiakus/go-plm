// Package mapper converts between domain types and DTOs.
package mapper

import (
	"github.com/Homiakus/go-plm/internal/api/dto"
	"github.com/Homiakus/go-plm/internal/core/object"
	"github.com/Homiakus/go-plm/internal/modules/bom"
	"github.com/Homiakus/go-plm/internal/modules/release"
)

// ObjectToDTO converts a domain Object to an ObjectDTO.
func ObjectToDTO(obj object.Object) dto.ObjectDTO {
	relations := make([]dto.RelationDTO, len(obj.Relations))
	for i, rel := range obj.Relations {
		relations[i] = dto.RelationDTO{
			ToID:     rel.ToID,
			Type:     rel.Type,
			Quantity: rel.Quantity,
			Unit:     rel.Unit,
		}
	}
	artifacts := make([]dto.ArtifactDTO, len(obj.Artifacts))
	for i, artifact := range obj.Artifacts {
		artifacts[i] = dto.ArtifactDTO{
			ID:           artifact.ID,
			Kind:         artifact.Kind,
			Role:         artifact.Role,
			Path:         artifact.Path,
			OriginalName: artifact.OriginalName,
			Checksum:     artifact.Checksum,
			SizeBytes:    artifact.SizeBytes,
			Generated:    artifact.Generated,
			Required:     artifact.Required,
			Status:       artifact.Status,
		}
	}
	return dto.ObjectDTO{
		ID:        string(obj.ID),
		Project:   obj.Project,
		Class:     string(obj.Class),
		Sequence:  obj.Sequence,
		Version:   obj.Version,
		Revision:  obj.Revision,
		State:     string(obj.State),
		Title:     obj.Title,
		Metadata:  obj.Metadata,
		Relations: relations,
		Artifacts: artifacts,
	}
}

// DTOToObject converts an ObjectDTO to a domain Object.
func DTOToObject(d dto.ObjectDTO) object.Object {
	relations := make([]object.RelationRef, len(d.Relations))
	for i, rel := range d.Relations {
		relations[i] = object.RelationRef{
			ToID:     rel.ToID,
			Type:     rel.Type,
			Quantity: rel.Quantity,
			Unit:     rel.Unit,
		}
	}
	artifacts := make([]object.ArtifactRef, len(d.Artifacts))
	for i, artifact := range d.Artifacts {
		artifacts[i] = object.ArtifactRef{
			ID:           artifact.ID,
			Kind:         artifact.Kind,
			Role:         artifact.Role,
			Path:         artifact.Path,
			OriginalName: artifact.OriginalName,
			Checksum:     artifact.Checksum,
			SizeBytes:    artifact.SizeBytes,
			Generated:    artifact.Generated,
			Required:     artifact.Required,
			Status:       artifact.Status,
		}
	}
	return object.Object{
		ID:        object.ID(d.ID),
		Project:   d.Project,
		Class:     object.Class(d.Class),
		Sequence:  d.Sequence,
		Version:   d.Version,
		Revision:  d.Revision,
		State:     object.State(d.State),
		Title:     d.Title,
		Metadata:  d.Metadata,
		Relations: relations,
		Artifacts: artifacts,
	}
}

// BOMRowsToDTO converts domain BOM rows to DTOs.
func BOMRowsToDTO(rows []bom.BOMRow) []dto.BOMRowDTO {
	result := make([]dto.BOMRowDTO, len(rows))
	for i, r := range rows {
		result[i] = dto.BOMRowDTO{
			RowID:      r.RowID,
			ParentID:   r.ParentID,
			ChildID:    r.ChildID,
			ChildClass: r.ChildClass,
			Quantity:   r.Quantity,
			Unit:       r.Unit,
			Position:   r.Position,
			MakeBuy:    r.MakeBuy,
			Level:      r.Level,
		}
	}
	return result
}

// ObjectsToDTO converts a slice of Objects to DTOs.
// Returns empty slice (not nil) for empty input.
func ObjectsToDTO(objects []object.Object) []dto.ObjectDTO {
	if objects == nil {
		return []dto.ObjectDTO{}
	}
	result := make([]dto.ObjectDTO, len(objects))
	for i, obj := range objects {
		result[i] = ObjectToDTO(obj)
	}
	return result
}

// ManifestToResponse converts a release manifest to a DTO.
func ManifestToResponse(m *release.Manifest, ready bool) dto.ReleaseResponse {
	return dto.ReleaseResponse{
		Ready: ready,
	}
}
