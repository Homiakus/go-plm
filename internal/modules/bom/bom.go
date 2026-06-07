// Package bom provides Bill of Materials construction and validation.
// BOM is built from 'contains' relations — no separate BOM table.
package bom

import (
	"context"
	"fmt"

	"github.com/Homiakus/go-plm/internal/core/diagnostic"
	"github.com/Homiakus/go-plm/internal/core/object"
)

// ObjectReader abstracts reading objects needed for BOM construction.
type ObjectReader interface {
	GetObject(ctx context.Context, id object.ID) (object.Object, error)
}

// RelationReader abstracts reading relations.
type RelationReader interface {
	ListRelations(ctx context.Context, fromID object.ID) ([]object.RelationRef, error)
}

// BOMRow represents one row in a BOM.
type BOMRow struct {
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

// Service builds and validates BOMs.
type Service struct {
	Objects   ObjectReader
	Relations RelationReader
}

// New creates a BOM service.
func New(objects ObjectReader, relations RelationReader) *Service {
	return &Service{Objects: objects, Relations: relations}
}

// GetStructured builds a hierarchical BOM from a root assembly.
// Returns rows with level/depth information for tree rendering.
func (s *Service) GetStructured(ctx context.Context, root object.ID) ([]BOMRow, error) {
	var rows []BOMRow
	visited := make(map[string]bool)
	if err := s.walkBOM(ctx, string(root), visited, 1.0, 0, &rows, ""); err != nil {
		return nil, err
	}
	return rows, nil
}

// GetFlat builds a flat BOM with rolled-up quantities (grouped by child_id).
func (s *Service) GetFlat(ctx context.Context, root object.ID) ([]BOMRow, error) {
	structured, err := s.GetStructured(ctx, root)
	if err != nil {
		return nil, err
	}

	// Group by child_id, sum quantities
	groups := make(map[string]*BOMRow)
	var order []string
	for _, row := range structured {
		if existing, ok := groups[row.ChildID]; ok {
			existing.Quantity += row.Quantity
		} else {
			cp := row
			cp.Level = 0 // flat
			groups[row.ChildID] = &cp
			order = append(order, row.ChildID)
		}
	}

	var flat []BOMRow
	for _, id := range order {
		flat = append(flat, *groups[id])
	}
	return flat, nil
}

// walkBOM recursively traverses `contains` relations.
func (s *Service) walkBOM(ctx context.Context, fromID string, visited map[string]bool,
	parentQty float64, level int, rows *[]BOMRow, position string) error {

	if visited[fromID] {
		return fmt.Errorf("bom: cycle detected at %s", fromID)
	}
	visited[fromID] = true

	rels, err := s.Relations.ListRelations(ctx, object.ID(fromID))
	if err != nil {
		return fmt.Errorf("bom: list relations for %s: %w", fromID, err)
	}

	for i, rel := range rels {
		if rel.Type != "contains" {
			continue
		}

		qty := 1.0
		if rel.Quantity != nil {
			qty = *rel.Quantity
		}
		effectiveQty := qty * parentQty

		childObj, err := s.Objects.GetObject(ctx, object.ID(rel.ToID))
		makeBuy := ""
		childClass := ""
		if err == nil {
			childClass = string(childObj.Class)
			if mb, ok := childObj.Metadata["make_buy"]; ok {
				makeBuy = fmt.Sprint(mb)
			}
		}

		pos := position
		if pos == "" {
			pos = fmt.Sprintf("%d", (i+1)*10)
		} else {
			pos = fmt.Sprintf("%s.%d", pos, i+1)
		}

		*rows = append(*rows, BOMRow{
			RowID:      fmt.Sprintf("bomrow-%s-%s", fromID, rel.ToID),
			ParentID:   fromID,
			ChildID:    rel.ToID,
			ChildClass: childClass,
			Quantity:   effectiveQty,
			Unit:       rel.Unit,
			Position:   pos,
			MakeBuy:    makeBuy,
			Level:      level,
		})

		// Recurse into child if it's an assembly
		if childClass == string(object.ClassAssembly) {
			newVisited := make(map[string]bool)
			for k, v := range visited {
				newVisited[k] = v
			}
			if err := s.walkBOM(ctx, rel.ToID, newVisited, effectiveQty, level+1, rows, pos); err != nil {
				return err
			}
		}
	}
	return nil
}

// DetectCycles checks for circular dependencies in the BOM.
func (s *Service) DetectCycles(ctx context.Context, root object.ID) ([]diagnostic.Diagnostic, error) {
	var diags []diagnostic.Diagnostic
	visited := make(map[string]bool)
	path := []string{}
	s.detectCycles(ctx, string(root), visited, path, &diags)
	return diags, nil
}

func (s *Service) detectCycles(ctx context.Context, id string, visited map[string]bool, path []string, diags *[]diagnostic.Diagnostic) {
	if visited[id] {
		*diags = append(*diags, diagnostic.Diagnostic{
			ID: "cycle-" + id, Severity: diagnostic.Blocker, Code: "BOM_CYCLE",
			ObjectID: id, Message: fmt.Sprintf("Cycle detected: %v", append(path, id)),
		})
		return
	}
	visited[id] = true
	path = append(path, id)

	rels, err := s.Relations.ListRelations(ctx, object.ID(id))
	if err != nil {
		return
	}
	for _, rel := range rels {
		if rel.Type == "contains" {
			s.detectCycles(ctx, rel.ToID, visited, path, diags)
		}
	}
	delete(visited, id)
}

// Validate checks BOM rows for common issues.
func (s *Service) Validate(rows []BOMRow) []diagnostic.Diagnostic {
	var diags []diagnostic.Diagnostic
	for i, row := range rows {
		if row.Quantity <= 0 {
			diags = append(diags, diagnostic.Diagnostic{
				ID: fmt.Sprintf("bom-qty-%d", i), Severity: diagnostic.Blocker,
				Code: "BOM_INVALID_QUANTITY", ObjectID: row.ChildID,
				Message: fmt.Sprintf("Invalid quantity %.2f for %s", row.Quantity, row.ChildID),
			})
		}
		if row.Unit == "" {
			diags = append(diags, diagnostic.Diagnostic{
				ID: fmt.Sprintf("bom-unit-%d", i), Severity: diagnostic.Warning,
				Code: "BOM_MISSING_UNIT", ObjectID: row.ChildID,
				Message: fmt.Sprintf("Missing unit for %s", row.ChildID),
			})
		}
	}
	return diags
}
