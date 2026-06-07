// Package gitops_test provides tests for Git integration operations.
package gitops_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/Homiakus/go-plm/internal/gitops"
)

func TestInit(t *testing.T) {
	dir := t.TempDir()
	svc, err := gitops.Init(dir)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	if svc == nil {
		t.Fatal("Init() returned nil service")
	}

	// Should be able to check status on fresh repo
	st, err := svc.Status()
	if err != nil {
		t.Fatalf("Status() error = %v", err)
	}
	if !st.IsClean {
		t.Log("fresh repo has untracked files — expected")
	}
}

func TestOpen(t *testing.T) {
	dir := t.TempDir()
	_, err := gitops.Init(dir)
	if err != nil {
		t.Fatal(err)
	}

	svc, err := gitops.Open(dir)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	if svc == nil {
		t.Fatal("Open() returned nil")
	}
}

func TestOpen_NonExistent(t *testing.T) {
	_, err := gitops.Open(t.TempDir())
	if err == nil {
		t.Error("expected error for non-existent repo")
	}
}

func TestCheckpoint(t *testing.T) {
	dir := t.TempDir()
	svc, err := gitops.Init(dir)
	if err != nil {
		t.Fatal(err)
	}

	// Create a file to commit
	p := filepath.Join(dir, "test.md")
	if err := os.WriteFile(p, []byte("# Test"), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := svc.Checkpoint(context.Background(), gitops.CheckpointRequest{
		Message: "initial commit",
		Author:  "dev",
	})
	if err != nil {
		t.Fatalf("Checkpoint() error = %v", err)
	}
	if result.CommitHash == "" {
		t.Error("CommitHash is empty")
	}
}

func TestCheckpoint_AllFiles(t *testing.T) {
	dir := t.TempDir()
	svc, err := gitops.Init(dir)
	if err != nil {
		t.Fatal(err)
	}

	os.WriteFile(filepath.Join(dir, "a.md"), []byte("a"), 0644)
	os.WriteFile(filepath.Join(dir, "b.md"), []byte("b"), 0644)

	// Empty Paths means stage all
	result, err := svc.Checkpoint(context.Background(), gitops.CheckpointRequest{
		Message: "add all",
		Author:  "dev",
	})
	if err != nil {
		t.Fatalf("Checkpoint(all) error = %v", err)
	}
	if result.CommitHash == "" {
		t.Error("CommitHash is empty")
	}
}

func TestCreateTag(t *testing.T) {
	dir := t.TempDir()
	svc, err := gitops.Init(dir)
	if err != nil {
		t.Fatal(err)
	}

	os.WriteFile(filepath.Join(dir, "test.md"), []byte("# Test"), 0644)
	svc.Checkpoint(context.Background(), gitops.CheckpointRequest{Message: "first", Author: "dev"})

	if err := svc.CreateTag("v1.0", "release v1.0"); err != nil {
		t.Fatalf("CreateTag() error = %v", err)
	}
}

func TestLog(t *testing.T) {
	dir := t.TempDir()
	svc, err := gitops.Init(dir)
	if err != nil {
		t.Fatal(err)
	}

	os.WriteFile(filepath.Join(dir, "test.md"), []byte("# Test"), 0644)
	svc.Checkpoint(context.Background(), gitops.CheckpointRequest{Message: "first", Author: "dev"})
	os.WriteFile(filepath.Join(dir, "test2.md"), []byte("# Test2"), 0644)
	svc.Checkpoint(context.Background(), gitops.CheckpointRequest{Message: "second", Author: "dev"})

	commits, err := svc.Log(5)
	if err != nil {
		t.Fatalf("Log() error = %v", err)
	}
	if len(commits) < 2 {
		t.Errorf("Log() returned %d commits, want >= 2", len(commits))
	}
}

func TestLog_Limit(t *testing.T) {
	dir := t.TempDir()
	svc, err := gitops.Init(dir)
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 5; i++ {
		p := filepath.Join(dir, "f"+string(rune('0'+i))+".md")
		os.WriteFile(p, []byte("data"), 0644)
		svc.Checkpoint(context.Background(), gitops.CheckpointRequest{Message: "c" + string(rune('0'+i)), Author: "dev"})
	}

	commits, err := svc.Log(2)
	if err != nil {
		t.Fatalf("Log(2) error = %v", err)
	}
	if len(commits) > 2 {
		t.Errorf("Log(2) returned %d commits, want <= 2", len(commits))
	}
}
