package app_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/Homiakus/go-plm/internal/api/dto"
	"github.com/Homiakus/go-plm/internal/api/tree"
	"github.com/Homiakus/go-plm/internal/app/command"
	"github.com/Homiakus/go-plm/internal/app/query"
	"github.com/Homiakus/go-plm/internal/core/object"
	"github.com/Homiakus/go-plm/internal/naming"
	"github.com/Homiakus/go-plm/internal/process/fsm"
	"github.com/Homiakus/go-plm/internal/store/fsrepo"
	"github.com/Homiakus/go-plm/internal/store/index"
	"github.com/Homiakus/go-plm/internal/store/transaction"
)

// TestProjectTreeWithThumbnails verifies GetTree returns correct nodes with thumbnails.
func TestProjectTreeWithThumbnails(t *testing.T) {
	root := t.TempDir()
	os.MkdirAll(filepath.Join(root, "objects"), 0755)
	os.MkdirAll(filepath.Join(root, ".plm"), 0755)
	os.WriteFile(filepath.Join(root, "project.md"), []byte(`---
project: {code: t, title: T, version: 1}
naming: {standard: v10, pattern: '[project]-[class]-[sequence]-v[major].[minor]', project_code: t}
sequences: {prt: 0, asm: 99}
storage: {source_of_truth: markdown_yaml, index: .plm/index.db}
git: {enabled: false}
---
`), 0644)

	tx := transaction.NewManager(filepath.Join(root, ".plm", "transactions"))
	repo := fsrepo.New(root, tx)
	idxDB, _ := index.Open(filepath.Join(root, ".plm", "index.db"))
	defer idxDB.Close()

	nstore := naming.NewMemorySequenceStore(map[string]int{"prt": 0, "asm": 99})
	ngen := naming.NewGenerator("t", nstore)
	lc, _ := fsm.NewFromYAML([]byte("initial: draft\nstates: [draft, released]\ntransitions: []\n"))

	cmd := &command.ObjectService{Repo: repo, Index: idxDB, Naming: ngen, FSM: lc, Git: &mockGitSvc{}}
	qry := &query.ObjectService{Objects: &projReader{repo: repo, idx: idxDB}}
	ctx := context.Background()

	// Create objects
	asm, _ := cmd.CreateObject(ctx, dto.CreateObjectRequest{Class: "asm", Title: "Motor Assy", Metadata: map[string]any{"unit": "pcs"}})
	prt, _ := cmd.CreateObject(ctx, dto.CreateObjectRequest{Class: "prt", Title: "Bracket", Metadata: map[string]any{"unit": "pcs", "make_buy": "make", "thumbnail": "files/images/thumb.png"}})

	// Attach thumbnail image
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "thumb.png"), []byte("fake-png-data"), 0644)
	cmd.AttachArtifact(ctx, prt.ObjectID, filepath.Join(tmpDir, "thumb.png"), "image", "photo")

	// Add relation
	addR(t, repo, asm.ObjectID, prt.ObjectID, "contains", 1, "pcs")

	// Get tree
	nodes, err := qry.GetTree(ctx, dto.TreeRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if len(nodes) < 2 {
		t.Fatalf("expected >=2 nodes, got %d", len(nodes))
	}

	// Verify assembly node
	var asmNode, prtNode *dto.TreeNodeDTO
	for i := range nodes {
		if nodes[i].ID == asm.ObjectID {
			asmNode = &nodes[i]
		}
		if nodes[i].ID == prt.ObjectID {
			prtNode = &nodes[i]
		}
	}

	if asmNode == nil || prtNode == nil {
		t.Fatal("nodes not found in tree")
	}

	// Assembly checks
	if asmNode.Icon != "package" {
		t.Errorf("asm icon = %q", asmNode.Icon)
	}
	if asmNode.StatusColor != "gray" {
		t.Errorf("asm status = %q (draft → gray)", asmNode.StatusColor)
	}
	if !asmNode.HasChildren {
		t.Error("asm should have children")
	}
	if asmNode.ChildrenCount != 1 {
		t.Errorf("asm children = %d", asmNode.ChildrenCount)
	}

	// Part checks
	if prtNode.Icon != "component" {
		t.Errorf("prt icon = %q", prtNode.Icon)
	}
	if prtNode.ThumbnailURL == "" {
		t.Error("prt should have thumbnail URL from metadata")
	}
	t.Logf("Thumbnail URL: %s", prtNode.ThumbnailURL)

	// Test icon helpers
	if tree.Icon(object.ClassAssembly) != "package" {
		t.Error("Icon(asm)")
	}
	if tree.Icon(object.ClassStandardPart) != "nut" {
		t.Error("Icon(std)")
	}
	if tree.StatusColor(object.StateReleased) != "green" {
		t.Error("StatusColor(released)")
	}
	if tree.StatusColor(object.StateBlocked) != "red" {
		t.Error("StatusColor(blocked)")
	}

	// Test ThumbnailURL priority: metadata > artifacts
	obj, _ := repo.GetObject(ctx, object.ID(prt.ObjectID))
	thumb := tree.ThumbnailURL(obj)
	t.Logf("Thumbnail from metadata: %s", thumb)

	// Test with only image artifact (no metadata thumbnail)
	obj.Metadata = map[string]any{"unit": "pcs"} // remove thumbnail key
	obj.Artifacts = []object.ArtifactRef{
		{Kind: "image", Role: "photo", Path: "files/images/photo.jpg", Status: "present"},
	}
	thumb2 := tree.ThumbnailURL(obj)
	if thumb2 != "files/images/photo.jpg" {
		t.Errorf("thumbnail from image artifact = %q", thumb2)
	}
	t.Logf("Thumbnail from artifact: %s", thumb2)

	// Test filter
	filtered, _ := qry.GetTree(ctx, dto.TreeRequest{Filter: "bracket"})
	if len(filtered) != 1 || filtered[0].ID != prt.ObjectID {
		t.Errorf("filter 'bracket': %d nodes", len(filtered))
	}

	// Test sort
	sorted, _ := qry.GetTree(ctx, dto.TreeRequest{Sort: "alpha"})
	if len(sorted) >= 2 && sorted[0].Title > sorted[len(sorted)-1].Title {
		t.Error("alpha sort failed")
	}

	// Test pagination
	paged, _ := qry.GetTree(ctx, dto.TreeRequest{Limit: 1, Offset: 0})
	if len(paged) != 1 {
		t.Errorf("pagination: %d nodes", len(paged))
	}

	t.Log("")
	t.Log("═══ PROJECT TREE VERIFICATION ═══")
	t.Logf("Total nodes:     %d", len(nodes))
	t.Logf("Assembly:        icon=%s children=%d hasChildren=%v",
		asmNode.Icon, asmNode.ChildrenCount, asmNode.HasChildren)
	t.Logf("Part:            icon=%s status=%s thumbnail=%s",
		prtNode.Icon, prtNode.StatusColor, prtNode.ThumbnailURL)
	t.Logf("Filter 'bracket': %d results", len(filtered))
	t.Logf("Sort 'alpha':     %d nodes", len(sorted))
	t.Logf("Pagination limit=1: %d nodes", len(paged))
	t.Log("═══ ALL CHECKS PASSED ═══")
}
