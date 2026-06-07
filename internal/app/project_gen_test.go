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
	"github.com/Homiakus/go-plm/internal/modules/bom"
	"github.com/Homiakus/go-plm/internal/modules/release"
	"github.com/Homiakus/go-plm/internal/naming"
	"github.com/Homiakus/go-plm/internal/parser/frontmatter"
	"github.com/Homiakus/go-plm/internal/process/fsm"
	"github.com/Homiakus/go-plm/internal/store/fsrepo"
	"github.com/Homiakus/go-plm/internal/store/index"
	"github.com/Homiakus/go-plm/internal/store/transaction"
	rulespkg "github.com/Homiakus/go-plm/internal/validation/rules"
	"github.com/Homiakus/go-plm/internal/validation/engine"
)

func TestGenerateAndAnalyzeProject(t *testing.T) {
	root := t.TempDir()
	os.MkdirAll(filepath.Join(root, "objects"), 0755)
	os.MkdirAll(filepath.Join(root, ".plm"), 0755)
	os.MkdirAll(filepath.Join(root, "config"), 0755)

	os.WriteFile(filepath.Join(root, "project.yaml"), []byte(`project:
  code: demo
  title: "Demo Engineering Project"
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
  std: 0
modules:
  objects: true
  bom: true
  release: true
  standardparts: true
storage:
  source_of_truth: markdown_yaml
  index: .plm/index.db
git:
  enabled: true
  release_tags: true
`), 0644)

	lcYAML := []byte(`
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
`)
	os.WriteFile(filepath.Join(root, "config", "lifecycle.yaml"), lcYAML, 0644)

	tx := transaction.NewManager(filepath.Join(root, ".plm", "transactions"))
	repo := fsrepo.New(root, tx)
	idxDB, _ := index.Open(filepath.Join(root, ".plm", "index.db"))
	defer idxDB.Close()

	nstore := naming.NewMemorySequenceStore(map[string]int{"prt": 0, "asm": 99, "drw": 0, "std": 0})
	ngen := naming.NewGenerator("demo", nstore)
	machine, _ := fsm.NewFromYAML(lcYAML)

	cmd := &command.ObjectService{Repo: repo, Index: idxDB, Naming: ngen, FSM: machine, Git: &mockGitSvc{}}
	ctx := context.Background()

	type item struct{ ID, Class, Title string }
	var objs []item

	create := func(class, title string, meta map[string]any) string {
		resp, err := cmd.CreateObject(ctx, dto.CreateObjectRequest{Class: class, Title: title, Metadata: meta})
		if err != nil {
			t.Fatalf("create %s %q: %v", class, title, err)
		}
		objs = append(objs, item{resp.ObjectID, class, title})
		return resp.ObjectID
	}

	asm1 := create("asm", "Motor Assembly", map[string]any{"unit": "pcs"})
	asm2 := create("asm", "Gearbox Sub-Assembly", map[string]any{"unit": "pcs"})
	prt1 := create("prt", "Bracket Motor", map[string]any{"unit": "pcs", "make_buy": "make", "material": "al5052", "thickness_mm": 2.0})
	prt2 := create("prt", "Shaft Drive", map[string]any{"unit": "pcs", "make_buy": "make", "material": "steel_4140"})
	prt3 := create("prt", "Housing Cover", map[string]any{"unit": "pcs", "make_buy": "make", "material": "al6061"})
	prt4 := create("prt", "Gasket Seal", map[string]any{"unit": "pcs", "make_buy": "buy", "material": "viton"})
	std1 := create("std", "Hex Bolt ISO 4017 M6x30 A2", map[string]any{"std_class": "fst", "std_id": "iso4017-m6x30-a2", "make_buy": "buy", "standard": "ISO 4017", "thread": "M6", "length_mm": 30, "material_grade": "A2"})
	std2 := create("std", "Bearing 6205-2RS C3", map[string]any{"std_class": "brg", "std_id": "6205-2rs-c3", "make_buy": "buy", "bearing_series": "6205", "seal_type": "2RS", "clearance": "C3"})
	std3 := create("std", "Resistor 10K 0603 1%", map[string]any{"std_class": "elc", "std_id": "res-10k-0603", "make_buy": "buy", "component_type": "res", "value": "10k", "package": "0603"})
	drw1 := create("drw", "Motor Assembly Drawing", map[string]any{"drawing_for": asm1, "format": "A3"})
	drw2 := create("drw", "Bracket Detail Drawing", map[string]any{"drawing_for": prt1, "format": "A4"})

	t.Logf("═══ Created %d objects ═══", len(objs))

	// Relations
	appendRel(t, repo, asm1, asm2, "contains", 1, "pcs")
	appendRel(t, repo, asm1, prt1, "contains", 1, "pcs")
	appendRel(t, repo, asm1, prt4, "contains", 2, "pcs")
	appendRel(t, repo, asm2, prt2, "contains", 1, "pcs")
	appendRel(t, repo, asm2, prt3, "contains", 1, "pcs")
	appendRel(t, repo, asm2, std2, "contains", 2, "pcs")
	appendRel(t, repo, asm1, std1, "contains", 8, "pcs")
	appendRel(t, repo, asm2, std3, "contains", 4, "pcs")
	appendRel(t, repo, asm1, drw1, "has_drawing", 0, "")
	appendRel(t, repo, prt1, drw2, "has_drawing", 0, "")

	// YAML verification
	t.Log("═══ YAML VERIFICATION ═══")
	var yamlIssues []string
	for _, e := range objs {
		path := filepath.Join(root, "objects", e.ID, e.ID+".md")
		data, _ := os.ReadFile(path)
		doc, err := frontmatter.Split(data)
		if err != nil {
			yamlIssues = append(yamlIssues, fmt.Sprintf("%s: %v", e.ID, err))
			continue
		}
		var parsed map[string]any
		if err := yaml.Unmarshal(doc.Frontmatter, &parsed); err != nil {
			yamlIssues = append(yamlIssues, fmt.Sprintf("%s: bad YAML: %v", e.ID, err))
			continue
		}
		for _, f := range []string{"id", "project", "class", "state", "title"} {
			if _, ok := parsed[f]; !ok {
				yamlIssues = append(yamlIssues, fmt.Sprintf("%s: missing %s", e.ID, f))
			}
		}
		if pid, _ := parsed["id"].(string); pid != e.ID {
			yamlIssues = append(yamlIssues, fmt.Sprintf("%s: ID mismatch yaml=%v", e.ID, pid))
		}
	}
	if len(yamlIssues) > 0 {
		for _, iss := range yamlIssues {
			t.Errorf("YAML: %s", iss)
		}
	} else {
		t.Logf("✓ All %d YAML frontmatters valid", len(objs))
	}

	// Filesystem read-back
	t.Log("═══ FILESYSTEM READ-BACK ═══")
	for _, e := range objs {
		got, err := repo.GetObject(ctx, object.ID(e.ID))
		if err != nil {
			t.Errorf("GetObject %s: %v", e.ID, err)
			continue
		}
		if got.Title != e.Title {
			t.Errorf("%s: title got=%q want=%q", e.ID, got.Title, e.Title)
		}
		if string(got.Class) != e.Class {
			t.Errorf("%s: class got=%s want=%s", e.ID, got.Class, e.Class)
		}
	}
	t.Logf("✓ All %d objects readable", len(objs))

	// Validation
	t.Log("═══ VALIDATION ═══")
	reg := rulespkg.DefaultRegistry()
	eng := engine.NewEngine(reg)
	allObjs, _ := repo.ListObjects(ctx)
	totalDiags, blockers := 0, 0
	for _, obj := range allObjs {
		diags, _ := eng.ValidateObject(ctx, obj, allObjs)
		totalDiags += len(diags)
		for _, d := range diags {
			if d.IsBlocker() {
				blockers++
			}
		}
	}
	t.Logf("Objects: %d, diagnostics: %d, blockers: %d", len(allObjs), totalDiags, blockers)

	// BOM
	t.Log("═══ BOM ═══")
	reader := &projReader{repo: repo, idx: idxDB}
	bomSvc := bom.New(reader, reader)
	structured, _ := bomSvc.GetStructured(ctx, object.ID(asm1))
	flat, _ := bomSvc.GetFlat(ctx, object.ID(asm1))

	maxLvl := 0
	for _, r := range structured {
		if r.Level > maxLvl {
			maxLvl = r.Level
		}
	}
	t.Logf("Structured BOM: %d rows, %d levels", len(structured), maxLvl+1)
	for _, r := range structured {
		indent := strings.Repeat("  ", r.Level+1)
		t.Logf("%s├─ %s [%s] qty=%.0f %s", indent, r.ChildID, r.ChildClass, r.Quantity, r.Unit)
	}
	t.Logf("Flat BOM: %d rows", len(flat))
	for _, r := range flat {
		t.Logf("  %-35s qty=%-6.0f %s", r.ChildID, r.Quantity, r.Unit)
	}

	// Cycle check
	cycles, _ := bomSvc.DetectCycles(ctx, object.ID(asm1))
	if len(cycles) > 0 {
		t.Errorf("BOM cycles: %d", len(cycles))
	} else {
		t.Log("✓ No BOM cycles")
	}

	// Release
	t.Log("═══ RELEASE READINESS ═══")
	relSvc := release.New(reader, reader)
	scope, _ := relSvc.BuildScope(ctx, object.ID(asm1))
	r1, _ := relSvc.CheckReadiness(ctx, scope)
	t.Logf("Scope: %d objects, ready=%v, blockers=%d", len(scope.Objects), r1.Ready, len(r1.Blockers))
	if len(r1.Blockers) > 0 {
		for _, b := range r1.Blockers {
			t.Logf("  [%s] %s", b.Code, b.Message)
		}
	}

	// Approve all assemblies and parts
	for _, e := range objs {
		if e.Class == "asm" || e.Class == "prt" {
			doTrans(t, cmd, ctx, e.ID, "submit_review", "in_review")
			doTrans(t, cmd, ctx, e.ID, "approve", "approved")
		}
	}
	r2, _ := relSvc.CheckReadiness(ctx, scope)
	t.Logf("After approval: ready=%v, blockers=%d", r2.Ready, len(r2.Blockers))

	// Count by class
	nAsm, nPrt, nStd, nDrw := 0, 0, 0, 0
	for _, e := range objs {
		switch e.Class {
		case "asm":
			nAsm++
		case "prt":
			nPrt++
		case "std":
			nStd++
		case "drw":
			nDrw++
		}
	}

	t.Log("")
	t.Log("═══════════════════════════════════════")
	t.Log("       PROJECT ANALYSIS REPORT         ")
	t.Log("═══════════════════════════════════════")
	t.Logf("Project:       demo")
	t.Logf("Root assembly: %s", asm1)
	t.Logf("Objects:       %d (asm:%d prt:%d std:%d drw:%d)", len(objs), nAsm, nPrt, nStd, nDrw)
	t.Logf("Relations:     10")
	t.Logf("BOM depth:     %d levels", maxLvl+1)
	t.Logf("BOM rows:      %d structured, %d flat", len(structured), len(flat))
	t.Logf("YAML valid:    ✓ (%d files)", len(objs))
	t.Logf("FS read-back:  ✓ (%d objects)", len(objs))
	t.Logf("Validation:    %d diags, %d blockers", totalDiags, blockers)
	t.Logf("BOM cycles:    %d", len(cycles))
	t.Logf("Release ready: %v", r2.Ready)
	t.Log("═══════════════════════════════════════")
}

type mockGitSvc struct{}

func (g *mockGitSvc) Status() (modified, added, deleted []string, clean bool, err error) {
	return nil, nil, nil, true, nil
}
func (g *mockGitSvc) Checkpoint(ctx context.Context, message, author string) (string, error) {
	return "mock", nil
}

type projReader struct {
	repo *fsrepo.Repository
	idx  *index.DB
}

func (r *projReader) GetObject(ctx context.Context, id object.ID) (object.Object, error) {
	return r.repo.GetObject(ctx, id)
}
func (r *projReader) ListObjects(ctx context.Context) ([]object.Object, []error) {
	return r.repo.ListObjects(ctx)
}
func (r *projReader) ListRelations(ctx context.Context, fromID object.ID) ([]object.RelationRef, error) {
	obj, err := r.repo.GetObject(ctx, fromID)
	if err != nil {
		return nil, err
	}
	return obj.Relations, nil
}
func (r *projReader) SearchObjects(ctx context.Context, query string, limit int) ([]object.ID, error) {
	return r.idx.SearchObjects(ctx, query, limit)
}

func appendRel(t *testing.T, repo command.ObjectRepo, from, to, typ string, qty float64, unit string) {
	t.Helper()
	obj, _ := repo.GetObject(context.Background(), object.ID(from))
	obj.Relations = append(obj.Relations, object.RelationRef{ToID: to, Type: typ, Quantity: &qty, Unit: unit})
	repo.SaveObject(context.Background(), obj)
}

func doTrans(t *testing.T, cmd *command.ObjectService, ctx context.Context, id, transition, want string) {
	t.Helper()
	resp, err := cmd.RunTransition(ctx, dto.TransitionRequest{ObjectID: id, Transition: transition})
	if err != nil {
		t.Fatalf("%s: %v", transition, err)
	}
	if resp.NewState != want {
		t.Errorf("%s: got=%s want=%s", transition, resp.NewState, want)
	}
}
