package fsrepo

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/Homiakus/go-plm/internal/core/object"
	"github.com/Homiakus/go-plm/internal/store/transaction"
)

func setupRepo(t *testing.T) *Repository {
	t.Helper()
	root := t.TempDir()
	os.MkdirAll(filepath.Join(root, "objects"), 0755)
	txDir := filepath.Join(root, ".plm", "transactions")
	return New(root, transaction.NewManager(txDir))
}

func TestSaveAndGetObject(t *testing.T) {
	repo := setupRepo(t)
	ctx := context.Background()

	obj := object.Object{
		ID:    "demo-prt-0001-v1.0",
		Class: object.ClassPart,
		Title: "Test Part",
		State: object.StateDraft,
	}

	if err := repo.SaveObject(ctx, obj); err != nil {
		t.Fatal(err)
	}

	got, err := repo.GetObject(ctx, obj.ID)
	if err != nil {
		t.Fatal(err)
	}

	if got.ID != obj.ID {
		t.Errorf("ID = %q", got.ID)
	}
	if got.Title != "Test Part" {
		t.Errorf("Title = %q", got.Title)
	}
	if got.Class != object.ClassPart {
		t.Errorf("Class = %q", got.Class)
	}
}

func TestListObjects(t *testing.T) {
	repo := setupRepo(t)
	ctx := context.Background()

	objs := []object.Object{
		{ID: "demo-prt-0001-v1.0", Class: object.ClassPart, Title: "P1"},
		{ID: "demo-prt-0002-v1.0", Class: object.ClassPart, Title: "P2"},
		{ID: "demo-asm-0100-v1.0", Class: object.ClassAssembly, Title: "A1"},
	}

	for _, obj := range objs {
		if err := repo.SaveObject(ctx, obj); err != nil {
			t.Fatal(err)
		}
	}

	got, errs := repo.ListObjects(ctx)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if len(got) != 3 {
		t.Fatalf("expected 3 objects, got %d", len(got))
	}
}

func TestExists(t *testing.T) {
	repo := setupRepo(t)
	ctx := context.Background()

	if repo.Exists("demo-prt-0001-v1.0") {
		t.Error("should not exist before save")
	}

	repo.SaveObject(ctx, object.Object{ID: "demo-prt-0001-v1.0", Class: object.ClassPart, Title: "T"})

	if !repo.Exists("demo-prt-0001-v1.0") {
		t.Error("should exist after save")
	}
}

func TestDeleteObject(t *testing.T) {
	repo := setupRepo(t)
	ctx := context.Background()

	obj := object.Object{ID: "demo-prt-0001-v1.0", Class: object.ClassPart, Title: "T"}
	repo.SaveObject(ctx, obj)

	if err := repo.DeleteObject(ctx, obj.ID); err != nil {
		t.Fatal(err)
	}

	if repo.Exists(obj.ID) {
		t.Error("should not exist after delete")
	}
}

func TestGetObjectNotFound(t *testing.T) {
	repo := setupRepo(t)
	_, err := repo.GetObject(context.Background(), "nonexistent-v1.0")
	if err == nil {
		t.Error("expected error for nonexistent object")
	}
}

func TestAppendHistory(t *testing.T) {
	repo := setupRepo(t)
	ctx := context.Background()

	obj := object.Object{ID: "demo-prt-0001-v1.0", Class: object.ClassPart, Title: "T"}
	repo.SaveObject(ctx, obj)

	if err := repo.AppendHistory(ctx, obj.ID, `{"event":"test"}`); err != nil {
		t.Fatal(err)
	}

	path := repo.HistoryPath(obj.ID)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Error("history file should not be empty")
	}
}

func TestSplitFrontmatter(t *testing.T) {
	basic := "---\nid: test\n---\n\n# Body\n"
	fm, body, err := splitFrontmatter([]byte(basic))
	if err != nil {
		t.Fatal(err)
	}
	if string(fm) != "id: test" {
		t.Errorf("frontmatter = %q, want %q", fm, "id: test")
	}
	if len(body) == 0 {
		t.Error("body should not be empty")
	}

	// No frontmatter
	_, _, err = splitFrontmatter([]byte("# Just body\n"))
	if err == nil {
		t.Error("expected error for missing frontmatter")
	}

	// Empty frontmatter
	empty := "---\n---\nBody\n"
	fm2, _, err := splitFrontmatter([]byte(empty))
	if err != nil {
		t.Fatal(err)
	}
	if len(fm2) != 0 {
		t.Errorf("expected empty frontmatter, got %q", fm2)
	}
}
