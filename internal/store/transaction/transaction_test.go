package transaction

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestExecuteWrite(t *testing.T) {
	dir := t.TempDir()
	txDir := filepath.Join(dir, ".plm", "transactions")
	m := NewManager(txDir)

	target := filepath.Join(dir, "test.md")
	content := []byte("# Hello")

	plan := Plan{
		ID: NewID(),
		Writes: []WriteOp{
			{Path: target, Content: content},
		},
	}

	if err := m.Execute(context.Background(), plan); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(content) {
		t.Errorf("content = %q, want %q", got, content)
	}

	// Work dir should be cleaned up
	_, err = os.Stat(filepath.Join(txDir, plan.ID))
	if !os.IsNotExist(err) {
		t.Error("work dir should be cleaned up after commit")
	}
}

func TestExecuteCopy(t *testing.T) {
	dir := t.TempDir()
	m := NewManager(filepath.Join(dir, ".plm", "transactions"))

	src := filepath.Join(dir, "source.txt")
	os.WriteFile(src, []byte("source"), 0644)
	dst := filepath.Join(dir, "dest.txt")

	plan := Plan{
		ID: NewID(),
		Copies: []CopyOp{
			{From: src, To: dst},
		},
	}

	if err := m.Execute(context.Background(), plan); err != nil {
		t.Fatal(err)
	}

	got, _ := os.ReadFile(dst)
	if string(got) != "source" {
		t.Errorf("dest content = %q", got)
	}
}

func TestExecuteDelete(t *testing.T) {
	dir := t.TempDir()
	m := NewManager(filepath.Join(dir, ".plm", "transactions"))

	target := filepath.Join(dir, "to_delete.txt")
	os.WriteFile(target, []byte("x"), 0644)

	plan := Plan{
		ID: NewID(),
		Deletes: []DeleteOp{
			{Path: target},
		},
	}

	if err := m.Execute(context.Background(), plan); err != nil {
		t.Fatal(err)
	}

	_, err := os.Stat(target)
	if !os.IsNotExist(err) {
		t.Error("file should be deleted")
	}
}

func TestExecuteRename(t *testing.T) {
	dir := t.TempDir()
	m := NewManager(filepath.Join(dir, ".plm", "transactions"))

	from := filepath.Join(dir, "old.txt")
	to := filepath.Join(dir, "new.txt")
	os.WriteFile(from, []byte("renamed"), 0644)

	plan := Plan{
		ID: NewID(),
		Renames: []RenameOp{
			{From: from, To: to},
		},
	}

	if err := m.Execute(context.Background(), plan); err != nil {
		t.Fatal(err)
	}

	_, err := os.Stat(from)
	if !os.IsNotExist(err) {
		t.Error("old file should not exist")
	}
	got, _ := os.ReadFile(to)
	if string(got) != "renamed" {
		t.Errorf("content = %q", got)
	}
}

func TestExecuteEmptyPlan(t *testing.T) {
	dir := t.TempDir()
	m := NewManager(filepath.Join(dir, ".plm", "transactions"))

	err := m.Execute(context.Background(), Plan{ID: ""})
	if err == nil {
		t.Error("expected error for empty ID")
	}
}

func TestRecoverAndRollback(t *testing.T) {
	dir := t.TempDir()
	txDir := filepath.Join(dir, ".plm", "transactions")
	m := NewManager(txDir)

	// Execute one plan to completion
	plan := Plan{
		ID: NewID(),
		Writes: []WriteOp{
			{Path: filepath.Join(dir, "ok.txt"), Content: []byte("ok")},
		},
	}
	if err := m.Execute(context.Background(), plan); err != nil {
		t.Fatal(err)
	}

	// Create an incomplete work dir manually
	incomplete := filepath.Join(txDir, "incomplete-tx")
	os.MkdirAll(incomplete, 0755)
	os.WriteFile(filepath.Join(incomplete, "junk.tmp"), []byte("partial"), 0644)

	items, err := m.Recover(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 incomplete, got %d", len(items))
	}
	if items[0].ID != "incomplete-tx" {
		t.Errorf("ID = %q", items[0].ID)
	}

	// Rollback
	if err := m.Rollback(context.Background(), items[0].ID); err != nil {
		t.Fatal(err)
	}
	_, err = os.Stat(incomplete)
	if !os.IsNotExist(err) {
		t.Error("incomplete work dir should be rolled back")
	}
}

func TestRollbackRejectsUnsafeID(t *testing.T) {
	dir := t.TempDir()
	m := NewManager(filepath.Join(dir, ".plm", "transactions"))

	if err := m.Rollback(context.Background(), ".."); err == nil {
		t.Fatal("expected unsafe rollback id error")
	}
	if err := m.Execute(context.Background(), Plan{ID: "..", Writes: []WriteOp{{Path: filepath.Join(dir, "x"), Content: []byte("x")}}}); err == nil {
		t.Fatal("expected unsafe execute id error")
	}
}

func TestIsCommitted(t *testing.T) {
	dir := t.TempDir()
	if IsCommitted(dir) {
		t.Error("empty dir should not be committed")
	}
	os.WriteFile(filepath.Join(dir, "committed"), []byte("ok"), 0644)
	if !IsCommitted(dir) {
		t.Error("dir with committed marker should be committed")
	}
}

func TestNewID(t *testing.T) {
	id1 := NewID()
	id2 := NewID()
	if id1 == id2 {
		t.Error("two IDs should be different")
	}
	if len(id1) < 10 {
		t.Errorf("ID too short: %q", id1)
	}
}
