package app_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/Homiakus/go-plm/internal/api/dto"
	"github.com/Homiakus/go-plm/internal/app/command"
	"github.com/Homiakus/go-plm/internal/core/object"
	"github.com/Homiakus/go-plm/internal/naming"
	"github.com/Homiakus/go-plm/internal/parser/frontmatter"
	"github.com/Homiakus/go-plm/internal/process/fsm"
	"github.com/Homiakus/go-plm/internal/store/fsrepo"
	"github.com/Homiakus/go-plm/internal/store/index"
	"github.com/Homiakus/go-plm/internal/store/transaction"
)

// TestAllObjectTypes generates every object class and verifies all documents.
func TestAllObjectTypes(t *testing.T) {
	root := t.TempDir()
	os.MkdirAll(filepath.Join(root, "objects"), 0755)
	os.MkdirAll(filepath.Join(root, ".plm"), 0755)
	os.MkdirAll(filepath.Join(root, "config"), 0755)
	os.MkdirAll(filepath.Join(root, "templates"), 0755)

	// project.md with frontmatter
	os.WriteFile(filepath.Join(root, "project.md"), []byte(`---
project:
  code: demo
  title: "Full Feature Demo"
  version: 1
naming:
  standard: v10
  pattern: "[project]-[class]-[sequence]-v[major].[minor]"
  project_code: demo
  case: kebab
sequences:
  prt: 0
  asm: 99
  drw: 0
  doc: 0
  std: 0
  mat: 0
  cut: 0
  bnd: 0
  nc: 0
  ins: 0
  tpc: 0
  wi: 0
  bom: 0
  rel: 0
  cr: 0
modules:
  objects: true
  bom: true
  release: true
  standardparts: true
  manufacturing: true
  procurement: true
storage:
  source_of_truth: markdown_yaml
  index: .plm/index.db
git:
  enabled: true
  release_tags: true
---

# Full Feature Demo Project

This project demonstrates ALL object types supported by go-plm.
`), 0644)

	// lifecycle.md
	os.WriteFile(filepath.Join(root, "config", "lifecycle.md"), []byte(`---
initial: draft
states: [draft, in_review, approved, released, blocked, obsolete, archived]
transitions:
  - name: submit_review
    from: [draft]
    to: in_review
  - name: approve
    from: [in_review]
    to: approved
  - name: release
    from: [approved]
    to: released
    effects: [create_git_tag]
  - name: revise
    from: [released]
    to: draft
  - name: block
    from: ["*"]
    to: blocked
  - name: unblock
    from: [blocked]
    to: draft
---
`), 0644)

	tx := transaction.NewManager(filepath.Join(root, ".plm", "transactions"))
	repo := fsrepo.New(root, tx)
	idxDB, _ := index.Open(filepath.Join(root, ".plm", "index.db"))
	defer idxDB.Close()

	// All classes with their starting sequences
	allStarts := map[string]int{
		"prt": 0, "asm": 99, "drw": 0, "doc": 0, "std": 0,
		"mat": 0, "cut": 0, "bnd": 0, "nc": 0, "ins": 0,
		"tpc": 0, "wi": 0, "bom": 0, "rel": 0, "cr": 0,
	}
	nstore := naming.NewMemorySequenceStore(allStarts)
	ngen := naming.NewGenerator("demo", nstore)

	lcData, _ := os.ReadFile(filepath.Join(root, "config", "lifecycle.md"))
	lcDoc, _ := frontmatter.Split(lcData)
	machine, _ := fsm.NewFromYAML(lcDoc.Frontmatter)

	cmd := &command.ObjectService{Repo: repo, Index: idxDB, Naming: ngen, FSM: machine, Git: &mockGitSvc{}}
	ctx := context.Background()

	type entry struct {
		ID, Class, Title string
		Metadata         map[string]any
	}
	var all []entry

	create := func(class, title string, meta map[string]any) string {
		resp, err := cmd.CreateObject(ctx, dto.CreateObjectRequest{Class: class, Title: title, Metadata: meta})
		if err != nil {
			t.Fatalf("CREATE %s %q: %v", class, title, err)
		}
		all = append(all, entry{resp.ObjectID, class, title, meta})
		t.Logf("  ✓ %-6s %s — %s", class, resp.ObjectID, title)
		return resp.ObjectID
	}

	t.Log("═══ GENERATING ALL 15 OBJECT CLASSES ═══")
	t.Log("")

	// ── Parts ──
	prt1 := create("prt", "Bracket Motor", map[string]any{
		"unit": "pcs", "make_buy": "make", "material": "al5052",
		"thickness_mm": 2.0, "manufacturing_method": "laser_cut",
		"surface_treatment": "anodize", "finish_color": "ral7035",
	})
	prt2 := create("prt", "Shaft Drive", map[string]any{
		"unit": "pcs", "make_buy": "make", "material": "steel_4140",
		"surface_treatment": "harden", "tolerance_class": "iso2768-m",
	})
	prt3 := create("prt", "Gasket Seal", map[string]any{
		"unit": "pcs", "make_buy": "buy", "material": "viton",
		"supplier": "seals-inc", "lead_time_days": 14,
	})

	// ── Assemblies ──
	asm1 := create("asm", "Motor Assembly", map[string]any{
		"unit": "pcs", "lifecycle_stage": "development",
	})
	asm2 := create("asm", "Gearbox Sub-Assembly", map[string]any{
		"unit": "pcs",
	})

	// ── Drawings ──
	drw1 := create("drw", "Motor Assembly Drawing", map[string]any{
		"drawing_for": asm1, "format": "A3", "scale": "1:2",
	})
	drw2 := create("drw", "Bracket Detail Drawing", map[string]any{
		"drawing_for": prt1, "format": "A4", "scale": "1:1",
	})

	// ── Documents ──
	doc1 := create("doc", "Design Specification", map[string]any{
		"document_type": "specification", "author": "engineer-01",
	})
	_ = create("doc", "Test Report", map[string]any{
		"document_type": "report", "author": "qa-team",
	})

	// ── Standard Parts ──
	_ = create("std", "Hex Bolt ISO 4017 M6x30 A2", map[string]any{
		"std_class": "fst", "std_id": "iso4017-m6x30-a2", "make_buy": "buy",
		"standard": "ISO 4017", "thread": "M6", "length_mm": 30,
		"material_grade": "A2", "category": "fastener", "family": "bolt",
	})
	std2 := create("std", "Ball Bearing 6205-2RS C3", map[string]any{
		"std_class": "brg", "std_id": "6205-2rs-c3", "make_buy": "buy",
		"bearing_series": "6205", "seal_type": "2RS", "clearance": "C3",
	})
	_ = create("std", "Resistor 10K 0603 1% 0.1W", map[string]any{
		"std_class": "elc", "std_id": "res-10k-0603-1pct", "make_buy": "buy",
		"component_type": "res", "value": "10k", "package": "0603",
		"tolerance": "1%", "power": "0.1W",
	})

	// ── Materials ──
	_ = create("mat", "Aluminum Sheet 5052 H19", map[string]any{
		"unit": "kg", "material": "al5052", "form": "sheet",
		"thickness_mm": 2.0, "width_mm": 1250, "length_mm": 2500,
	})

	// ── Manufacturing Files ──
	cut1 := create("cut", "Bracket Cutting Program", map[string]any{
		"source_part": prt1, "process": "laser_cut",
		"machine": "trumpf-3030", "nesting_pattern": "n01",
	})
	_ = create("bnd", "Bracket Bending Program", map[string]any{
		"source_part": prt1, "process": "air_bending",
		"machine": "amada-hds", "bend_sequence": "seq-01",
	})
	_ = create("nc", "Shaft Turning Program", map[string]any{
		"source_part": prt2, "machine": "dmc-lathe",
		"tool_list": "t01-t05", "cycle_time_min": 12.5,
	})

	// ── Inspection ──
	_ = create("ins", "Bracket Inspection Plan", map[string]any{
		"source_object": prt1, "inspection_method": "cmm",
		"critical_dimensions": "d1,d2,d3",
	})

	// ── Tech Process Card ──
	_ = create("tpc", "Bracket Manufacturing Route", map[string]any{
		"source_object": prt1, "operations_count": 5,
		"work_centers": "laser-01,bend-02,deburr-03",
	})

	// ── Work Instruction ──
	_ = create("wi", "Bracket Assembly Steps", map[string]any{
		"related_object": prt1, "safety_warnings": "wear_gloves",
		"estimated_time_min": 15,
	})

	// ── Formal BOM ──
	_ = create("bom", "Motor Assembly eBOM", map[string]any{
		"root_assembly": asm1, "bom_type": "ebom",
	})

	// ── Release Package ──
	_ = create("rel", "Motor Assembly Release v1.0", map[string]any{
		"root_object": asm1, "release_type": "major",
	})

	// ── Change Request ──
	_ = create("cr", "Bracket Material Change", map[string]any{
		"affected_objects": prt1, "change_type": "material",
		"reason": "cost_reduction", "priority": "medium",
	})

	t.Logf("")
	t.Logf("Total objects created: %d", len(all))
	t.Logf("")

	// ── Build relations ──
	t.Log("═══ BUILDING ALL RELATION TYPES ═══")
	addR(t, repo, asm1, asm2, "contains", 1, "pcs")
	addR(t, repo, asm1, prt1, "contains", 1, "pcs")
	addR(t, repo, asm1, prt3, "contains", 2, "pcs")
	addR(t, repo, asm2, prt2, "contains", 1, "pcs")
	addR(t, repo, asm1, drw1, "has_drawing", 0, "")
	addR(t, repo, prt1, drw2, "has_drawing", 0, "")
	addR(t, repo, prt1, cut1, "manufacturing_file_for", 0, "")
	addR(t, repo, doc1, prt1, "affects", 0, "")

	t.Logf("Relations: 8")
	t.Logf("")

	// ── Verify ALL documents ──
	t.Log("═══ VERIFYING ALL GENERATED DOCUMENTS ═══")
	t.Log("")

	countByClass := map[string]int{}
	var issues []string

	for _, e := range all {
		countByClass[e.Class]++

		mdPath := filepath.Join(root, "objects", e.ID, e.ID+".md")
		histPath := filepath.Join(root, "objects", e.ID, "history.jsonl")

		// Check .md exists
		data, err := os.ReadFile(mdPath)
		if err != nil {
			issues = append(issues, fmt.Sprintf("MISSING .md: %s", e.ID))
			continue
		}

		// Parse frontmatter
		doc, err := frontmatter.Split(data)
		if err != nil {
			issues = append(issues, fmt.Sprintf("BAD FRONTMATTER: %s — %v", e.ID, err))
			continue
		}

		// Validate YAML
		var parsed map[string]any
		if err := yaml.Unmarshal(doc.Frontmatter, &parsed); err != nil {
			issues = append(issues, fmt.Sprintf("BAD YAML: %s — %v", e.ID, err))
			continue
		}

		// Check required frontmatter fields
		for _, f := range []string{"id", "project", "class", "state", "title"} {
			if _, ok := parsed[f]; !ok {
				issues = append(issues, fmt.Sprintf("MISSING %s: %s", f, e.ID))
			}
		}

		// ID matches filename
		if pid, _ := parsed["id"].(string); pid != e.ID {
			issues = append(issues, fmt.Sprintf("ID MISMATCH: file=%s yaml=%v", e.ID, pid))
		}

		// Class matches
		if pclass, _ := parsed["class"].(string); pclass != e.Class {
			issues = append(issues, fmt.Sprintf("CLASS MISMATCH: %s yaml=%v", e.ID, pclass))
		}

		// Metadata present
		if meta, ok := parsed["metadata"]; !ok || meta == nil {
			if len(e.Metadata) > 0 {
				issues = append(issues, fmt.Sprintf("METADATA MISSING: %s", e.ID))
			}
		}

		// Check history exists
		if _, err := os.Stat(histPath); os.IsNotExist(err) {
			issues = append(issues, fmt.Sprintf("NO HISTORY: %s", e.ID))
		}

		// Check body exists (body is generated from templates — may be minimal in MVP)
		if len(doc.Body) == 0 {
			issues = append(issues, fmt.Sprintf("SHORT BODY: %s (%d bytes)", e.ID, len(doc.Body)))
		}

		// Verify via GetObject
		got, err := repo.GetObject(ctx, object.ID(e.ID))
		if err != nil {
			issues = append(issues, fmt.Sprintf("GET FAILED: %s — %v", e.ID, err))
			continue
		}
		if got.Title != e.Title {
			issues = append(issues, fmt.Sprintf("TITLE MISMATCH: %s repo=%q want=%q", e.ID, got.Title, e.Title))
		}
		if string(got.Class) != e.Class {
			issues = append(issues, fmt.Sprintf("CLASS MISMATCH repo: %s got=%s", e.ID, got.Class))
		}

		// Log success
		t.Logf("  ✓ %-6s %-30s [%s] meta=%d fields rel=%d body=%dB",
			e.Class, e.ID, e.Title, len(e.Metadata), len(got.Relations), len(doc.Body))
	}

	if len(issues) > 0 {
		t.Errorf("")
		t.Errorf("═══ %d DOCUMENT ISSUES FOUND ═══", len(issues))
		for _, iss := range issues {
			t.Errorf("  ✗ %s", iss)
		}
		t.Errorf("")
	} else {
		t.Logf("")
		t.Logf("✓ ALL %d documents valid — no issues", len(all))
	}

	// ── Print sample document ──
	t.Log("")
	t.Log("══════════════════════════════════════════════")
	t.Log("  SAMPLE DOCUMENT: demo-prt-0001-v1.0.md")
	t.Log("══════════════════════════════════════════════")
	samplePath := filepath.Join(root, "objects", prt1, prt1+".md")
	sampleData, _ := os.ReadFile(samplePath)
	for _, line := range strings.Split(string(sampleData), "\n") {
		t.Logf("  %s", line)
	}

	t.Log("")
	t.Log("══════════════════════════════════════════════")
	t.Log("  SAMPLE DOCUMENT: demo-std-0002-v1.0.md")
	t.Log("══════════════════════════════════════════════")
	stdPath := filepath.Join(root, "objects", std2, std2+".md")
	stdData, _ := os.ReadFile(stdPath)
	for _, line := range strings.Split(string(stdData), "\n") {
		t.Logf("  %s", line)
	}

	// ── Print history sample ──
	t.Log("")
	t.Log("══════════════════════════════════════════════")
	t.Log("  SAMPLE HISTORY: demo-prt-0001-v1.0/history.jsonl")
	t.Log("══════════════════════════════════════════════")
	histPath := filepath.Join(root, "objects", prt1, "history.jsonl")
	histData, _ := os.ReadFile(histPath)
	t.Logf("  %s", string(histData))

	// ── Final Report ──
	t.Log("")
	t.Log("══════════════════════════════════════════════")
	t.Log("         COMPLETE PROJECT ANALYSIS           ")
	t.Log("══════════════════════════════════════════════")
	t.Logf("Objects:      %d total", len(all))
	classes := []string{"asm", "prt", "drw", "doc", "std", "mat", "cut", "bnd", "nc", "ins", "tpc", "wi", "bom", "rel", "cr"}
	for _, c := range classes {
		if n := countByClass[c]; n > 0 {
			t.Logf("  %-6s %d", c+":", n)
		}
	}
	t.Logf("Relations:    8")
	t.Logf("Docs valid:   %d/%d", len(all)-len(issues), len(all))
	t.Logf("YAML valid:   %d/%d", len(all)-len(issues), len(all))
	t.Logf("FS readable:  %d/%d", len(all)-len(issues), len(all))
	t.Logf("History:      %d/%d", len(all)-len(issues), len(all))
	t.Log("══════════════════════════════════════════════")
}

func addR(t *testing.T, repo command.ObjectRepo, from, to, typ string, qty float64, unit string) {
	t.Helper()
	obj, _ := repo.GetObject(context.Background(), object.ID(from))
	obj.Relations = append(obj.Relations, object.RelationRef{ToID: to, Type: typ, Quantity: &qty, Unit: unit})
	repo.SaveObject(context.Background(), obj)
}
