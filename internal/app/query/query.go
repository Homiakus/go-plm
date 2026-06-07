// Package query implements application-level read operations.
package query

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/Homiakus/go-plm/internal/api/dto"
	"github.com/Homiakus/go-plm/internal/api/mapper"
	"github.com/Homiakus/go-plm/internal/api/tree"
	"github.com/Homiakus/go-plm/internal/core/object"
	"github.com/Homiakus/go-plm/internal/modules/bom"
	"github.com/Homiakus/go-plm/internal/modules/release"
	"gopkg.in/yaml.v3"
)

// ObjectReader reads objects for queries.
type ObjectReader interface {
	GetObject(ctx context.Context, id object.ID) (object.Object, error)
	ListObjects(ctx context.Context) ([]object.Object, []error)
	SearchObjects(ctx context.Context, query string, limit int) ([]object.ID, error)
}

// ObjectService provides query operations on objects.
type ObjectService struct {
	Objects    ObjectReader
	BOM        *bom.Service
	Release    *release.Service
}

// GetObject returns a single object as DTO.
func (s *ObjectService) GetObject(ctx context.Context, id string) (*dto.ObjectDTO, error) {
	obj, err := s.Objects.GetObject(ctx, object.ID(id))
	if err != nil {
		return nil, err
	}
	d := mapper.ObjectToDTO(obj)
	return &d, nil
}

// ListObjects returns all objects as DTOs.
func (s *ObjectService) ListObjects(ctx context.Context) ([]dto.ObjectDTO, error) {
	objects, errs := s.Objects.ListObjects(ctx)
	if len(errs) > 0 {
		return nil, errs[0]
	}
	return mapper.ObjectsToDTO(objects), nil
}

// SearchObjects performs full-text search.
func (s *ObjectService) SearchObjects(ctx context.Context, req dto.SearchRequest) ([]dto.SearchResult, error) {
	ids, err := s.Objects.SearchObjects(ctx, req.Query, req.Limit)
	if err != nil {
		return nil, err
	}
	var results []dto.SearchResult
	for _, id := range ids {
		obj, err := s.Objects.GetObject(ctx, id)
		r := dto.SearchResult{ObjectID: string(id)}
		if err == nil {
			r.Title = obj.Title
			r.Class = string(obj.Class)
			r.State = string(obj.State)
		}
		results = append(results, r)
	}
	return results, nil
}

// GetBOM builds a BOM for the given root object.
func (s *ObjectService) GetBOM(ctx context.Context, req dto.BOMRequest) (*dto.BOMResponse, error) {
	var rows []bom.BOMRow
	var err error

	if req.Flat {
		rows, err = s.BOM.GetFlat(ctx, object.ID(req.RootID))
	} else {
		rows, err = s.BOM.GetStructured(ctx, object.ID(req.RootID))
	}
	if err != nil {
		return nil, err
	}

	return &dto.BOMResponse{
		RootID: req.RootID,
		Rows:   mapper.BOMRowsToDTO(rows),
	}, nil
}

// CheckReleaseReadiness checks if a release can proceed.
func (s *ObjectService) CheckReleaseReadiness(ctx context.Context, rootID string) (*dto.ReleaseResponse, error) {
	scope, err := s.Release.BuildScope(ctx, object.ID(rootID))
	if err != nil {
		return nil, err
	}
	readiness, err := s.Release.CheckReadiness(ctx, scope)
	if err != nil {
		return nil, err
	}
	return &dto.ReleaseResponse{
		Ready:    readiness.Ready,
		Blockers: readiness.Blockers,
	}, nil
}

// GetTree returns tree nodes for the Project Tree UI.
func (s *ObjectService) GetTree(ctx context.Context, req dto.TreeRequest) ([]dto.TreeNodeDTO, error) {
	objects, errs := s.Objects.ListObjects(ctx)
	if len(errs) > 0 {
		return nil, errs[0]
	}

	nodes := make([]dto.TreeNodeDTO, 0)
	for _, obj := range objects {
		// Basic filtering
		if req.Filter != "" {
			if !matchesFilter(obj, req.Filter) {
				continue
			}
		}

		thumb := tree.ThumbnailURL(obj)
		childrenCount := 0
		hasChildren := false
		if obj.Class == object.ClassAssembly {
			for _, r := range obj.Relations {
				if r.Type == "contains" {
					childrenCount++
					hasChildren = true
				}
			}
		}

		nodes = append(nodes, dto.TreeNodeDTO{
			ID:            string(obj.ID),
			Class:         string(obj.Class),
			Title:         obj.Title,
			State:         string(obj.State),
			Icon:          tree.Icon(obj.Class),
			StatusColor:   tree.StatusColor(obj.State),
			HasChildren:   hasChildren,
			ChildrenCount: childrenCount,
			ThumbnailURL:  thumb,
			Metadata:      obj.Metadata,
		})
	}

	// Sort
	switch req.Sort {
	case "alpha":
		sort.Slice(nodes, func(i, j int) bool { return nodes[i].Title < nodes[j].Title })
	case "class":
		sort.Slice(nodes, func(i, j int) bool { return nodes[i].Class < nodes[j].Class })
	case "status":
		sort.Slice(nodes, func(i, j int) bool { return nodes[i].State < nodes[j].State })
	}

	// Pagination
	if req.Limit > 0 {
		start := req.Offset
		end := start + req.Limit
		if start >= len(nodes) {
			return nil, nil
		}
		if end > len(nodes) {
			end = len(nodes)
		}
		nodes = nodes[start:end]
	}

	return nodes, nil
}

func matchesFilter(obj object.Object, filter string) bool {
	// Simple substring match on ID, title, class, state
	s := string(obj.ID) + " " + obj.Title + " " + string(obj.Class) + " " + string(obj.State)
	return containsFold(s, filter)
}

func containsFold(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		len(substr) == 0 ||
		findFold(s, substr) >= 0)
}

func findFold(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			sc := s[i+j]
			tc := substr[j]
			if sc >= 'A' && sc <= 'Z' {
				sc += 32
			}
			if tc >= 'A' && tc <= 'Z' {
				tc += 32
			}
			if sc != tc {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}

// WhereUsed returns all objects that reference the given object.
func (s *ObjectService) WhereUsed(ctx context.Context, id string) ([]dto.ObjectDTO, error) {
	objects, errs := s.Objects.ListObjects(ctx)
	if len(errs) > 0 {
		return nil, errs[0]
	}

	var usedIn []dto.ObjectDTO
	for _, obj := range objects {
		if string(obj.ID) == id {
			continue
		}
		for _, rel := range obj.Relations {
			if rel.ToID == id {
				usedIn = append(usedIn, mapper.ObjectToDTO(obj))
				break
			}
		}
	}
	return usedIn, nil
}

// ExportBOM exports a BOM in the specified format (csv, json, yaml).
func (s *ObjectService) ExportBOM(ctx context.Context, rootID string, format string) ([]byte, error) {
	rows, err := s.BOM.GetStructured(ctx, object.ID(rootID))
	if err != nil {
		return nil, fmt.Errorf("export BOM: %w", err)
	}

	switch strings.ToLower(format) {
	case "csv":
		return exportBOMCSV(rows)
	case "json":
		return json.MarshalIndent(rows, "", "  ")
	case "yaml", "yml":
		return yaml.Marshal(rows)
	default:
		return nil, fmt.Errorf("export BOM: unsupported format %q (use csv, json, yaml)", format)
	}
}

func exportBOMCSV(rows []bom.BOMRow) ([]byte, error) {
	var buf strings.Builder
	w := csv.NewWriter(&buf)
	w.Write([]string{"level", "position", "child_id", "child_class", "quantity", "unit", "make_buy"})
	for _, r := range rows {
		w.Write([]string{
			fmt.Sprintf("%d", r.Level),
			r.Position,
			r.ChildID,
			r.ChildClass,
			fmt.Sprintf("%.2f", r.Quantity),
			r.Unit,
			r.MakeBuy,
		})
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, err
	}
	return []byte(buf.String()), nil
}
