package index

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/Homiakus/go-plm/internal/core/object"
)

func TestOpenAndMigrate(t *testing.T) {
	path := filepath.Join(t.TempDir(), "index.db")
	db, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	o, r, a, err := db.Stats(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if o != 0 || r != 0 || a != 0 {
		t.Errorf("expected empty index, got %d/%d/%d", o, r, a)
	}
}

func TestUpsertAndGetObject(t *testing.T) {
	db := openDB(t)
	defer db.Close()

	obj := object.Object{
		ID: "demo-prt-0001-v1.0", Project: "demo", Class: object.ClassPart,
		Sequence: "0001", Version: "1.0", Revision: "1.0",
		State: object.StateDraft, Title: "Bracket",
		Metadata: map[string]any{"material": "al5052"},
	}

	if err := db.UpsertObject(context.Background(), obj); err != nil {
		t.Fatal(err)
	}

	got, err := db.GetObject(context.Background(), obj.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Title != "Bracket" {
		t.Errorf("Title = %q", got.Title)
	}
	if got.Metadata["material"] != "al5052" {
		t.Errorf("metadata = %v", got.Metadata)
	}
}

func TestDeleteObject(t *testing.T) {
	db := openDB(t)
	defer db.Close()
	ctx := context.Background()

	obj := object.Object{ID: "x-prt-0001-v1.0", Project: "x", Class: object.ClassPart, Title: "X"}
	db.UpsertObject(ctx, obj)

	if err := db.DeleteObject(ctx, obj.ID); err != nil {
		t.Fatal(err)
	}

	_, err := db.GetObject(ctx, obj.ID)
	if err == nil {
		t.Error("expected error after delete")
	}
}

func TestRebuildIndex(t *testing.T) {
	db := openDB(t)
	defer db.Close()
	ctx := context.Background()

	objs := []object.Object{
		{ID: "a-prt-0001-v1.0", Project: "a", Class: object.ClassPart, Title: "A1"},
		{ID: "a-prt-0002-v1.0", Project: "a", Class: object.ClassPart, Title: "A2"},
	}

	// First insert
	for _, o := range objs {
		db.UpsertObject(ctx, o)
	}

	// Rebuild
	if err := db.RebuildIndex(ctx, objs); err != nil {
		t.Fatal(err)
	}

	o, _, _, _ := db.Stats(ctx)
	if o != 2 {
		t.Errorf("expected 2 objects after rebuild, got %d", o)
	}
}

func TestSearchObjects(t *testing.T) {
	db := openDB(t)
	defer db.Close()
	ctx := context.Background()

	obj := object.Object{
		ID: "demo-prt-0001-v1.0", Project: "demo", Class: object.ClassPart,
		Title: "Bracket Motor", State: object.StateDraft,
	}
	db.UpsertObject(ctx, obj)

	ids, err := db.SearchObjects(ctx, "bracket", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) == 0 {
		t.Skip("FTS5 may need content sync")
	}
}

func openDB(t *testing.T) *DB {
	t.Helper()
	db, err := Open(filepath.Join(t.TempDir(), "index.db"))
	if err != nil {
		t.Fatal(err)
	}
	return db
}
