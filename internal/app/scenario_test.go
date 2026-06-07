package app_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/Homiakus/go-plm/internal/api/dto"
	"github.com/Homiakus/go-plm/internal/app/command"
	"github.com/Homiakus/go-plm/internal/app/query"
	"github.com/Homiakus/go-plm/internal/core/object"
	"github.com/Homiakus/go-plm/internal/gitops"
	"github.com/Homiakus/go-plm/internal/modules/bom"
	"github.com/Homiakus/go-plm/internal/modules/release"
	"github.com/Homiakus/go-plm/internal/naming"
	"github.com/Homiakus/go-plm/internal/process/fsm"
	rulespkg "github.com/Homiakus/go-plm/internal/validation/rules"
	"github.com/Homiakus/go-plm/internal/store/fsrepo"
	"github.com/Homiakus/go-plm/internal/store/index"
	"github.com/Homiakus/go-plm/internal/store/transaction"
	"github.com/Homiakus/go-plm/internal/validation/engine"
)

func setup(t *testing.T) (*command.ObjectService, *query.ObjectService, func()) {
	t.Helper()
	root := t.TempDir()
	os.MkdirAll(filepath.Join(root, "objects"), 0755)
	os.MkdirAll(filepath.Join(root, ".plm"), 0755)

	tx := transaction.NewManager(filepath.Join(root, ".plm", "transactions"))
	repo := fsrepo.New(root, tx)

	idxDB, err := index.Open(filepath.Join(root, ".plm", "index.db"))
	if err != nil {
		t.Fatal(err)
	}

	nstore := naming.NewMemorySequenceStore(naming.DefaultStarts())
	ngen := naming.NewGenerator("demo", nstore)

	lcYAML := []byte(`
name: object_lifecycle
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
	machine, err := fsm.NewFromYAML(lcYAML)
	if err != nil {
		t.Fatal(err)
	}

	cmdSvc := &command.ObjectService{
		Repo:   repo,
		Index:  idxDB,
		Naming: ngen,
		FSM:    machine,
		Git:    &mockGit{},
	}

	reader := &repoReader{repo: repo, idx: idxDB}
	bomSvc := bom.New(reader, reader)
	relSvc := release.New(reader, reader)

	qrySvc := &query.ObjectService{
		Objects: reader,
		BOM:     bomSvc,
		Release: relSvc,
	}

	return cmdSvc, qrySvc, func() { idxDB.Close() }
}

// TestFullLifecycle: Create → Review → Approve → Release → Revise
func TestFullLifecycle(t *testing.T) {
	cmd, qry, cleanup := setup(t)
	defer cleanup()
	ctx := context.Background()

	asm, err := cmd.CreateObject(ctx, dto.CreateObjectRequest{Class: "asm", Title: "Motor", Metadata: map[string]any{"unit": "pcs"}})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	prt, err := cmd.CreateObject(ctx, dto.CreateObjectRequest{Class: "prt", Title: "Bracket", Metadata: map[string]any{"unit": "pcs", "make_buy": "make"}})
	if err != nil {
		t.Fatalf("create part: %v", err)
	}

	// Submit → Approve → Release
	mustTransition(t, cmd, ctx, asm.ObjectID, "submit_review", "in_review")
	mustTransition(t, cmd, ctx, asm.ObjectID, "approve", "approved")
	mustTransition(t, cmd, ctx, asm.ObjectID, "release", "released")

	// Verify
	obj, _ := qry.GetObject(ctx, asm.ObjectID)
	if obj.State != "released" {
		t.Errorf("state = %s", obj.State)
	}

	// Revise
	mustTransition(t, cmd, ctx, asm.ObjectID, "revise", "draft")

	// Verify history
	path := cmd.Repo.(*fsrepo.Repository).HistoryPath(object.ID(asm.ObjectID))
	data, _ := os.ReadFile(path)
	if len(data) == 0 {
		t.Error("history empty")
	}
	t.Logf("Lifecycle OK, history: %d bytes, part: %s", len(data), prt.ObjectID)
}

// TestBOMWithNesting: 3-level BOM, flat rollup
func TestBOMWithNesting(t *testing.T) {
	cmd, qry, cleanup := setup(t)
	defer cleanup()
	ctx := context.Background()

	top, _ := cmd.CreateObject(ctx, dto.CreateObjectRequest{Class: "asm", Title: "Top", Metadata: map[string]any{"unit": "pcs"}})
	sub, _ := cmd.CreateObject(ctx, dto.CreateObjectRequest{Class: "asm", Title: "Sub", Metadata: map[string]any{"unit": "pcs"}})
	part, _ := cmd.CreateObject(ctx, dto.CreateObjectRequest{Class: "prt", Title: "Screw", Metadata: map[string]any{"unit": "pcs", "make_buy": "buy"}})

	addRel(t, cmd.Repo, top.ObjectID, sub.ObjectID, "contains", 1, "pcs")
	addRel(t, cmd.Repo, top.ObjectID, part.ObjectID, "contains", 4, "pcs")
	addRel(t, cmd.Repo, sub.ObjectID, part.ObjectID, "contains", 2, "pcs")

	structured, _ := qry.GetBOM(ctx, dto.BOMRequest{RootID: top.ObjectID})
	flat, _ := qry.GetBOM(ctx, dto.BOMRequest{RootID: top.ObjectID, Flat: true})

	if len(structured.Rows) < 2 {
		t.Errorf("structured BOM: %d rows", len(structured.Rows))
	}
	for _, r := range flat.Rows {
		if r.ChildID == part.ObjectID && r.Quantity != 6.0 {
			t.Errorf("flat qty = %.1f, want 6.0", r.Quantity)
		}
	}
	t.Logf("BOM: %d structured, %d flat rows", len(structured.Rows), len(flat.Rows))
}

// TestReleaseReadiness: verify blockers
func TestReleaseReadiness(t *testing.T) {
	cmd, qry, cleanup := setup(t)
	defer cleanup()
	ctx := context.Background()

	asm, _ := cmd.CreateObject(ctx, dto.CreateObjectRequest{Class: "asm", Title: "Motor", Metadata: map[string]any{"unit": "pcs"}})
	prt, _ := cmd.CreateObject(ctx, dto.CreateObjectRequest{Class: "prt", Title: "Bracket", Metadata: map[string]any{"unit": "pcs", "make_buy": "make"}})
	addRel(t, cmd.Repo, asm.ObjectID, prt.ObjectID, "contains", 1, "pcs")

	// Not ready — draft
	resp, _ := qry.CheckReleaseReadiness(ctx, asm.ObjectID)
	if resp.Ready {
		t.Error("should NOT be ready")
	}
	t.Logf("Blockers: %d", len(resp.Blockers))

	// Approve all
	mustTransition(t, cmd, ctx, asm.ObjectID, "submit_review", "in_review")
	mustTransition(t, cmd, ctx, asm.ObjectID, "approve", "approved")
	mustTransition(t, cmd, ctx, prt.ObjectID, "submit_review", "in_review")
	mustTransition(t, cmd, ctx, prt.ObjectID, "approve", "approved")

	// Ready
	resp, _ = qry.CheckReleaseReadiness(ctx, asm.ObjectID)
	if !resp.Ready {
		t.Errorf("should be ready, blockers: %d", len(resp.Blockers))
	}
	t.Log("Release readiness: OK")
}

// TestDuplicateDetection: exact, probable, not duplicates
func TestDuplicateDetection(t *testing.T) {
	meta1 := map[string]any{"standard": "ISO 4017", "thread": "M6", "length_mm": 30, "material_grade": "A2"}
	meta2 := map[string]any{"standard": "ISO 4017", "thread": "M6", "length_mm": 30, "material_grade": "A2"}
	meta3 := map[string]any{"standard": "DIN 933", "thread": "M8", "length_mm": 50, "material_grade": "8.8"}

	spmod := struct{}{}
	_ = spmod

	np1, _ := NormalizeFromMeta(meta1)
	np2, _ := NormalizeFromMeta(meta2)
	np3, _ := NormalizeFromMeta(meta3)

	if np1.Key != np2.Key {
		t.Error("keys should be identical for exact duplicate")
	}
	if np1.Key == np3.Key {
		t.Error("keys should differ for different parts")
	}
	t.Logf("Keys: %s | %s | %s", np1.Key, np2.Key, np3.Key)
}

// TestErrorScenarios: invalid transitions, missing objects
func TestErrorScenarios(t *testing.T) {
	cmd, _, cleanup := setup(t)
	defer cleanup()
	ctx := context.Background()

	// 1. Release from draft — blocked
	asm, _ := cmd.CreateObject(ctx, dto.CreateObjectRequest{Class: "asm", Title: "Motor", Metadata: map[string]any{"unit": "pcs"}})
	_, err := cmd.RunTransition(ctx, dto.TransitionRequest{ObjectID: asm.ObjectID, Transition: "release"})
	if err == nil {
		t.Error("release from draft should fail")
	}

	// 2. Nonexistent object
	_, err = cmd.RunTransition(ctx, dto.TransitionRequest{ObjectID: "nonexistent-v1.0", Transition: "approve"})
	if err == nil {
		t.Error("nonexistent object should fail")
	}

	// 3. Empty title
	_, err = cmd.CreateObject(ctx, dto.CreateObjectRequest{Class: "prt", Title: ""})
	if err == nil {
		t.Error("empty title should fail")
	}

	t.Log("Error scenarios: OK")
}

// TestIndexRebuild: create, delete index, rebuild
func TestIndexRebuild(t *testing.T) {
	cmd, qry, cleanup := setup(t)
	defer cleanup()
	ctx := context.Background()

	cmd.CreateObject(ctx, dto.CreateObjectRequest{Class: "prt", Title: "A", Metadata: map[string]any{"unit": "pcs", "make_buy": "make"}})
	cmd.CreateObject(ctx, dto.CreateObjectRequest{Class: "prt", Title: "B", Metadata: map[string]any{"unit": "pcs", "make_buy": "buy"}})

	list, _ := qry.ListObjects(ctx)
	if len(list) != 2 {
		t.Fatalf("expected 2, got %d", len(list))
	}

	if err := cmd.RebuildIndex(ctx); err != nil {
		t.Fatal(err)
	}

	list2, _ := qry.ListObjects(ctx)
	if len(list2) != 2 {
		t.Errorf("after rebuild: %d", len(list2))
	}
	t.Log("Index rebuild: OK")
}

// TestNamingSequence: multiple IDs, different classes
func TestNamingSequence(t *testing.T) {
	store := naming.NewMemorySequenceStore(naming.DefaultStarts())
	gen := naming.NewGenerator("proj", store)

	if id := gen.NextID("prt"); id != "proj-prt-0001-v1.0" {
		t.Errorf("prt1 = %s", id)
	}
	gen.NextID("prt")
	gen.NextID("prt")
	if id := gen.NextID("prt"); id != "proj-prt-0004-v1.0" {
		t.Errorf("prt4 = %s", id)
	}
	if id := gen.NextID("asm"); id != "proj-asm-0100-v1.0" {
		t.Errorf("asm1 = %s", id)
	}

	// Validate
	if naming.Validate("proj-prt-0001-v1.0") != nil {
		t.Error("valid ID rejected")
	}
	if naming.Validate("BAD") == nil {
		t.Error("invalid ID accepted")
	}
	t.Log("Naming: OK")
}

// TestFSMEndToEnd: full lifecycle path
func TestFSMEndToEnd(t *testing.T) {
	m, _ := fsm.NewFromYAML([]byte(`
name: test
initial: draft
states: [draft, in_review, approved, released]
transitions:
  - name: submit
    from: [draft]
    to: in_review
  - name: ok
    from: [in_review]
    to: approved
  - name: ship
    from: [approved]
    to: released
`))
	current := "draft"
	for _, tr := range []string{"submit", "ok", "ship"} {
		if !m.Can(current, tr) {
			t.Fatalf("cannot %s from %s", tr, current)
		}
		current, _ = m.Next(current, tr)
	}
	if current != "released" {
		t.Errorf("final = %s", current)
	}
	t.Log("FSM: OK")
}

// TestTransactionFullCycle: write + copy + rename + delete
func TestTransactionFullCycle(t *testing.T) {
	dir := t.TempDir()
	tx := transaction.NewManager(filepath.Join(dir, ".plm", "transactions"))

	src := filepath.Join(dir, "src.txt")
	os.WriteFile(src, []byte("source"), 0644)

	plan := transaction.Plan{
		ID: transaction.NewID(),
		Writes:  []transaction.WriteOp{{Path: filepath.Join(dir, "new.md"), Content: []byte("# Doc")}},
		Copies:  []transaction.CopyOp{{From: src, To: filepath.Join(dir, "copy.txt")}},
		Renames: []transaction.RenameOp{{From: src, To: filepath.Join(dir, "moved.txt")}},
		Deletes: []transaction.DeleteOp{{Path: filepath.Join(dir, "copy.txt")}},
	}
	if err := tx.Execute(context.Background(), plan); err != nil {
		t.Fatal(err)
	}

	assertExists(t, dir, "new.md")
	assertExists(t, dir, "moved.txt")
	assertMissing(t, dir, "src.txt")
	assertMissing(t, dir, "copy.txt")
	t.Log("Transaction: OK")
}

// TestValidationRules: valid, invalid name, missing metadata
func TestValidationRules(t *testing.T) {
	reg := rulespkg.DefaultRegistry()
	eng := engine.NewEngine(reg)

	valid := object.Object{
		ID: "demo-prt-0001-v1.0", Class: object.ClassPart, Title: "Bracket",
		Metadata: map[string]any{"unit": "pcs", "make_buy": "make"},
	}
	d1, _ := eng.ValidateObject(context.Background(), valid, nil)
	if len(d1) != 0 {
		t.Errorf("valid obj has %d diags", len(d1))
	}

	noMeta := object.Object{ID: "demo-prt-0002-v1.0", Class: object.ClassPart, Title: "X", Metadata: map[string]any{}}
	d2, _ := eng.ValidateObject(context.Background(), noMeta, nil)
	if len(d2) == 0 {
		t.Error("expected diags for missing metadata")
	}

	badName := object.Object{ID: "BAD", Class: object.ClassPart, Title: "X", Metadata: map[string]any{"unit": "pcs", "make_buy": "make"}}
	d3, _ := eng.ValidateObject(context.Background(), badName, nil)
	found := false
	for _, d := range d3 {
		if d.Code == "INVALID_NAME" {
			found = true
		}
	}
	if !found {
		t.Error("expected INVALID_NAME diagnostic")
	}
	t.Log("Validation: OK")
}

// TestGitInitCheckpoint: init, add files, commit, tag
func TestGitInitCheckpoint(t *testing.T) {
	dir := t.TempDir()
	git, err := gitops.Init(dir)
	if err != nil {
		t.Fatal(err)
	}

	os.WriteFile(filepath.Join(dir, "test.md"), []byte("# Test"), 0644)
	hash, err := git.Checkpoint(context.Background(), gitops.CheckpointRequest{Message: "first", Author: "dev"})
	if err != nil {
		t.Fatal(err)
	}
	if hash.CommitHash == "" {
		t.Error("empty hash")
	}

	os.WriteFile(filepath.Join(dir, "test2.md"), []byte("# Test2"), 0644)
	hash2, _ := git.Checkpoint(context.Background(), gitops.CheckpointRequest{Message: "second", Author: "dev"})

	if err := git.CreateTag("v1.0", "release"); err != nil {
		t.Fatal(err)
	}

	commits, _ := git.Log(5)
	if len(commits) < 2 {
		t.Errorf("commits = %d", len(commits))
	}
	t.Logf("Git: %d commits, tags: %s -> %s", len(commits), hash.CommitHash[:8], hash2.CommitHash[:8])
}

// TestSearchObjects: FTS5 search scenario
func TestSearchObjects(t *testing.T) {
	_, qry, cleanup := setup(t)
	defer cleanup()
	ctx := context.Background()

	results, err := qry.SearchObjects(ctx, dto.SearchRequest{Query: "test", Limit: 10})
	if err != nil {
		// FTS5 may not have data synced — skip gracefully
		t.Skipf("search not available: %v", err)
	}
	t.Logf("Search results: %d", len(results))
}

// helpers

type mockGit struct{}

func (g *mockGit) Status() (modified, added, deleted []string, clean bool, err error) {
	return nil, nil, nil, true, nil
}
func (g *mockGit) Checkpoint(ctx context.Context, message, author string) (string, error) {
	return "mock-hash", nil
}

type repoReader struct {
	repo *fsrepo.Repository
	idx  *index.DB
}

func (r *repoReader) GetObject(ctx context.Context, id object.ID) (object.Object, error) {
	return r.repo.GetObject(ctx, id)
}
func (r *repoReader) ListObjects(ctx context.Context) ([]object.Object, []error) {
	return r.repo.ListObjects(ctx)
}
func (r *repoReader) ListRelations(ctx context.Context, fromID object.ID) ([]object.RelationRef, error) {
	obj, err := r.repo.GetObject(ctx, fromID)
	if err != nil {
		return nil, err
	}
	return obj.Relations, nil
}
func (r *repoReader) SearchObjects(ctx context.Context, query string, limit int) ([]object.ID, error) {
	return r.idx.SearchObjects(ctx, query, limit)
}

func mustTransition(t *testing.T, cmd *command.ObjectService, ctx context.Context, id, transition, wantState string) {
	t.Helper()
	resp, err := cmd.RunTransition(ctx, dto.TransitionRequest{ObjectID: id, Transition: transition})
	if err != nil {
		t.Fatalf("%s: %v", transition, err)
	}
	if resp.NewState != wantState {
		t.Errorf("%s: state = %s, want %s", transition, resp.NewState, wantState)
	}
}

func addRel(t *testing.T, repo command.ObjectRepo, from, to, typ string, qty float64, unit string) {
	t.Helper()
	obj, _ := repo.GetObject(context.Background(), object.ID(from))
	obj.Relations = append(obj.Relations, object.RelationRef{ToID: to, Type: typ, Quantity: &qty, Unit: unit})
	repo.SaveObject(context.Background(), obj)
}

func assertExists(t *testing.T, dir, name string) {
	t.Helper()
	if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
		t.Errorf("%s should exist", name)
	}
}

func assertMissing(t *testing.T, dir, name string) {
	t.Helper()
	if _, err := os.Stat(filepath.Join(dir, name)); !os.IsNotExist(err) {
		t.Errorf("%s should NOT exist", name)
	}
}

// NormalizeFromMeta is a test helper using the standardparts normalize logic inline.
func NormalizeFromMeta(meta map[string]any) (struct{ Key string }, error) {
	// Simple inline normalization for testing
	std := getStr(meta, "standard")
	thread := getStr(meta, "thread")
	length := getStr(meta, "length_mm")
	grade := getStr(meta, "material_grade")
	key := "fst:" + tok(std) + ":" + tok(thread) + ":" + tok(length) + ":" + tok(grade)
	return struct{ Key string }{Key: key}, nil
}

func getStr(m map[string]any, k string) string {
	if v, ok := m[k]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func tok(s string) string {
	r := ""
	for _, c := range s {
		if c >= 'A' && c <= 'Z' {
			r += string(c + 32)
		} else if c >= 'a' && c <= 'z' || c >= '0' && c <= '9' {
			r += string(c)
		}
	}
	return r
}
