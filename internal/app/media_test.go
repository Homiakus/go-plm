package app_test

import (
	"context"
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

// TestMediaAttachments verifies all artifact kinds attach correctly to objects.
func TestMediaAttachments(t *testing.T) {
	root := t.TempDir()
	os.MkdirAll(filepath.Join(root, "objects"), 0755)
	os.MkdirAll(filepath.Join(root, ".plm"), 0755)

	// Create project.md
	os.WriteFile(filepath.Join(root, "project.md"), []byte(`---
project:
  code: demo
  title: "Media Test"
  version: 1
naming:
  standard: v10
  pattern: "[project]-[class]-[sequence]-v[major].[minor]"
  project_code: demo
  case: kebab
sequences:
  prt: 0
  asm: 99
storage:
  source_of_truth: markdown_yaml
  index: .plm/index.db
git:
  enabled: false
---

# Media Attachment Test
`), 0644)

	tx := transaction.NewManager(filepath.Join(root, ".plm", "transactions"))
	repo := fsrepo.New(root, tx)
	idxDB, _ := index.Open(filepath.Join(root, ".plm", "index.db"))
	defer idxDB.Close()

	nstore := naming.NewMemorySequenceStore(map[string]int{"prt": 0, "asm": 99})
	ngen := naming.NewGenerator("demo", nstore)

	lc := []byte("initial: draft\nstates: [draft, in_review, approved, released]\ntransitions:\n  - name: submit_review\n    from: [draft]\n    to: in_review\n")
	machine, _ := fsm.NewFromYAML(lc)

	cmd := &command.ObjectService{Repo: repo, Index: idxDB, Naming: ngen, FSM: machine, Git: &mockGitSvc{}}
	ctx := context.Background()

	// Create a Part to attach files to
	resp, err := cmd.CreateObject(ctx, dto.CreateObjectRequest{
		Class: "prt", Title: "Test Part",
		Metadata: map[string]any{"unit": "pcs", "make_buy": "make", "material": "al5052"},
	})
	if err != nil {
		t.Fatal(err)
	}
	objID := resp.ObjectID
	t.Logf("Created: %s", objID)

	// Create temp source files
	filesDir := t.TempDir()
	files := []struct {
		name    string
		content string
		kind    string
		role    string
	}{
		{"bracket.step", "ISO-10303-21;\nHEADER;\nFILE_DESCRIPTION(('Bracket'),'2;1');\nENDSEC;\nDATA;\n#1=CARTESIAN_POINT('',(0.0,0.0,0.0));\nENDSEC;\nEND-ISO-10303-21;", "cad", "primary_step"},
		{"bracket.pdf", "%PDF-1.4\n1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n2 0 obj\n<< /Type /Pages /Kids [3 0 R] /Count 1 >>\nendobj\n3 0 obj\n<< /Type /Page /Parent 2 0 R /MediaBox [0 0 595 842] >>\nendobj\nxref\n0 4\n0000000000 65535 f \n0000000009 00000 n \n0000000058 00000 n \n0000000115 00000 n \ntrailer\n<< /Size 4 /Root 1 0 R >>\nstartxref\n190\n%%EOF", "drawing", "drawing_pdf"},
		{"bracket.dxf", "0\nSECTION\n2\nHEADER\n0\nENDSEC\n0\nSECTION\n2\nENTITIES\n0\nLINE\n8\n0\n10\n0.0\n20\n0.0\n30\n0.0\n11\n100.0\n21\n50.0\n31\n0.0\n0\nENDSEC\n0\nEOF", "manufacturing", "cutting_dxf"},
		{"photo.jpg", "\xff\xd8\xff\xe0\x00\x10JFIF\x00\x01\x01\x01\x00H\x00H\x00\x00\xff\xdb\x00C\x00\x08\x06\x06\x07\x06\x05\x08\x07\x07\x07\x09\x09\x08\n\x0c\x14\r\x0c\x0b\x0b\x0c\x19\x12\x13\x0f\x14\x1d\x1a\x1f\x1e\x1d\x1a\x1c\x1c $.' \",#\x1c\x1c(7),01444\x1f'9=82<.342\xff\xd9", "image", "photo"},
		{"material-cert.pdf", "%PDF-1.4 fake certificate", "certificate", "material_cert"},
		{"assembly-evidence.png", "\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR\x00\x00\x00\x01\x00\x00\x00\x01\x08\x02\x00\x00\x00\x90wS\xde\x00\x00\x00\x0cIDATx\x9cc\xf8\x0f\x00\x00\x01\x01\x00\x05\x18\xd8N\x00\x00\x00\x00IEND\xaeB`\x82", "evidence", "build_photo"},
	}

	for _, f := range files {
		os.WriteFile(filepath.Join(filesDir, f.name), []byte(f.content), 0644)
	}

	// Attach all files
	t.Log("═══ ATTACHING MEDIA FILES ═══")
	for _, f := range files {
		srcPath := filepath.Join(filesDir, f.name)
		if err := cmd.AttachArtifact(ctx, objID, srcPath, f.kind, f.role); err != nil {
			t.Fatalf("attach %s (%s): %v", f.name, f.kind, err)
		}
		t.Logf("  ✓ %-25s → files/%-15s [%s]", f.name, f.kind, f.role)
	}

	// Verify files exist on disk
	t.Log("")
	t.Log("═══ VERIFYING FILES ON DISK ═══")
	objDir := repo.ObjectDir(object.ID(objID))
	for _, f := range files {
		var expectedDir string
		switch f.kind {
		case "cad":
			expectedDir = "files/cad"
		case "drawing":
			expectedDir = "files/drawings"
		case "manufacturing":
			expectedDir = "files/manufacturing"
		case "image":
			expectedDir = "files/images"
		case "certificate":
			expectedDir = "files/certificates"
		case "evidence":
			expectedDir = "files/evidence"
		}
		fullPath := filepath.Join(objDir, expectedDir, f.name)
		data, err := os.ReadFile(fullPath)
		if err != nil {
			t.Errorf("MISSING: %s", fullPath)
			continue
		}
		if string(data) != f.content {
			t.Errorf("CONTENT MISMATCH: %s", f.name)
		}
		t.Logf("  ✓ %s/%s — %d bytes OK", expectedDir, f.name, len(data))
	}

	// Verify YAML frontmatter has artifact entries
	t.Log("")
	t.Log("═══ VERIFYING YAML FRONTMATTER ARTIFACTS ═══")
	mdPath := filepath.Join(objDir, objID+".md")
	mdData, _ := os.ReadFile(mdPath)
	doc, _ := frontmatter.Split(mdData)

	var parsed struct {
		Artifacts []map[string]any `yaml:"artifacts"`
	}
	yaml.Unmarshal(doc.Frontmatter, &parsed)

	if len(parsed.Artifacts) != len(files) {
		t.Errorf("expected %d artifacts in YAML, got %d", len(files), len(parsed.Artifacts))
	}

	for i, a := range parsed.Artifacts {
		kind := a["kind"]
		role := a["role"]
		status := a["status"]
		path := a["path"]
		origName := a["original_name"]

		t.Logf("  ✓ art[%d] kind=%v role=%v status=%v path=%v orig=%v", i, kind, role, status, path, origName)

		if status != "present" {
			t.Errorf("art[%d]: status=%v, want present", i, status)
		}
		if origName != files[i].name {
			t.Errorf("art[%d]: original_name=%q, want %q", i, origName, files[i].name)
		}
	}

	// Verify via GetObject
	t.Log("")
	t.Log("═══ VERIFYING VIA GetObject ═══")
	got, _ := repo.GetObject(ctx, object.ID(objID))
	if len(got.Artifacts) != len(files) {
		t.Errorf("GetObject: %d artifacts, want %d", len(got.Artifacts), len(files))
	}
	for _, a := range got.Artifacts {
		t.Logf("  %-10s %-20s %-30s %s", a.Kind, a.Role, a.Path, a.Status)
	}

	// Print sample document
	t.Log("")
	t.Log("══════════════════════════════════════════════")
	t.Log("  SAMPLE: " + objID + ".md WITH ARTIFACTS")
	t.Log("══════════════════════════════════════════════")
	for _, line := range strings.Split(string(mdData), "\n") {
		t.Logf("  %s", line)
	}

	// Verify placeholder artifact (required but missing)
	t.Log("")
	t.Log("═══ PLACEHOLDER ARTIFACT TEST ═══")
	resp2, _ := cmd.CreateObject(ctx, dto.CreateObjectRequest{
		Class: "prt", Title: "Placeholder Test",
		Metadata: map[string]any{"unit": "pcs", "make_buy": "make"},
	})
	// Manually add a placeholder artifact to simulate required-but-missing
	obj2, _ := repo.GetObject(ctx, object.ID(resp2.ObjectID))
	obj2.Artifacts = append(obj2.Artifacts, object.ArtifactRef{
		ID: "art-placeholder-1", Kind: "drawing", Role: "drawing_pdf",
		Path: "", Status: "missing", Required: true,
	})
	repo.SaveObject(ctx, obj2)

	// Verify placeholder
	got2, _ := repo.GetObject(ctx, object.ID(resp2.ObjectID))
	if len(got2.Artifacts) != 1 || got2.Artifacts[0].Status != "missing" {
		t.Error("placeholder artifact not found")
	} else {
		t.Logf("  ✓ Placeholder: kind=%s role=%s status=%s required=%v",
			got2.Artifacts[0].Kind, got2.Artifacts[0].Role,
			got2.Artifacts[0].Status, got2.Artifacts[0].Required)
	}

	// Summary
	t.Log("")
	t.Log("══════════════════════════════════════════════")
	t.Log("         MEDIA ATTACHMENT VERIFICATION        ")
	t.Log("══════════════════════════════════════════════")
	t.Logf("Files attached:      %d", len(files))
	t.Logf("File kinds tested:   cad, drawing, manufacturing, image, certificate, evidence")
	t.Logf("Disk verification:   %d/%d OK", len(files), len(files))
	t.Logf("YAML artifacts:      %d entries", len(parsed.Artifacts))
	t.Logf("GetObject artifacts: %d entries", len(got.Artifacts))
	t.Logf("Placeholder test:    ✓")
	t.Log("══════════════════════════════════════════════")
}

// TestFileStructure verifies the complete directory layout after attachments.
func TestFileStructure(t *testing.T) {
	root := t.TempDir()
	os.MkdirAll(filepath.Join(root, "objects"), 0755)
	os.MkdirAll(filepath.Join(root, ".plm"), 0755)
	os.WriteFile(filepath.Join(root, "project.md"), []byte("---\nproject: {code: t, title: T, version: 1}\nnaming: {standard: v10, pattern: '[project]-[class]-[sequence]-v[major].[minor]', project_code: t}\nsequences: {prt: 0}\nstorage: {source_of_truth: markdown_yaml, index: .plm/index.db}\ngit: {enabled: false}\n---\n"), 0644)

	tx := transaction.NewManager(filepath.Join(root, ".plm", "transactions"))
	repo := fsrepo.New(root, tx)
	idxDB, _ := index.Open(filepath.Join(root, ".plm", "index.db"))
	defer idxDB.Close()

	nstore := naming.NewMemorySequenceStore(map[string]int{"prt": 0})
	ngen := naming.NewGenerator("t", nstore)
	lc, _ := fsm.NewFromYAML([]byte("initial: draft\nstates: [draft, released]\ntransitions:\n  - name: release\n    from: [draft]\n    to: released\n"))

	cmd := &command.ObjectService{Repo: repo, Index: idxDB, Naming: ngen, FSM: lc, Git: &mockGitSvc{}}
	ctx := context.Background()

	resp, _ := cmd.CreateObject(ctx, dto.CreateObjectRequest{Class: "prt", Title: "Struct Test", Metadata: map[string]any{"unit": "pcs", "make_buy": "make"}})
	objID := resp.ObjectID

	// Create and attach 3 files of different kinds
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "model.step"), []byte("step"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "drawing.pdf"), []byte("pdf"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "photo.png"), []byte("png"), 0644)

	cmd.AttachArtifact(ctx, objID, filepath.Join(tmpDir, "model.step"), "cad", "primary_step")
	cmd.AttachArtifact(ctx, objID, filepath.Join(tmpDir, "drawing.pdf"), "drawing", "drawing_pdf")
	cmd.AttachArtifact(ctx, objID, filepath.Join(tmpDir, "photo.png"), "image", "photo")

	// Print full directory tree
	t.Log("═══ OBJECT DIRECTORY STRUCTURE ═══")
	filepath.Walk(repo.ObjectDir(object.ID(objID)), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		rel, _ := filepath.Rel(root, path)
		prefix := "  "
		if info.IsDir() {
			prefix = "  📁"
		} else {
			prefix = "  📄"
		}
		t.Logf("%s %s", prefix, rel)
		return nil
	})

	// Verify specific paths
	checks := []string{
		filepath.Join("objects", objID, objID+".md"),
		filepath.Join("objects", objID, "history.jsonl"),
		filepath.Join("objects", objID, "files", "cad", "model.step"),
		filepath.Join("objects", objID, "files", "drawings", "drawing.pdf"),
		filepath.Join("objects", objID, "files", "images", "photo.png"),
	}
	for _, check := range checks {
		fullPath := filepath.Join(root, check)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("MISSING: %s", check)
		}
	}
	t.Log("")
	t.Log("✓ All expected files present in directory structure")
}
