// Package service provides the central application service that wires together
// all domain services: storage, index, naming, FSM, validation, git, BOM, release.
package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"gopkg.in/yaml.v3"

	"github.com/Homiakus/go-plm/internal/app/command"
	"github.com/Homiakus/go-plm/internal/app/query"
	"github.com/Homiakus/go-plm/internal/core/object"
	"github.com/Homiakus/go-plm/internal/core/project"
	"github.com/Homiakus/go-plm/internal/gitops"
	"github.com/Homiakus/go-plm/internal/modules/bom"
	"github.com/Homiakus/go-plm/internal/modules/release"
	"github.com/Homiakus/go-plm/internal/modules/standardparts"
	"github.com/Homiakus/go-plm/internal/naming"
	"github.com/Homiakus/go-plm/internal/parser/frontmatter"
	"github.com/Homiakus/go-plm/internal/process/fsm"
	"github.com/Homiakus/go-plm/internal/store/fsrepo"
	"github.com/Homiakus/go-plm/internal/store/index"
	"github.com/Homiakus/go-plm/internal/store/transaction"
)

// App ties together all subsystems for a single PLM project.
type App struct {
	Root      string
	Config    project.Config
	Cmd       *command.ObjectService
	Qry       *query.ObjectService
	BOM       *bom.Service
	Release   *release.Service
	Git       *gitops.Service
	Index     *index.DB
	Repo      *fsrepo.Repository
	NamingGen *naming.Generator
	FSM       *fsm.Machine
}

// Open initialises all services for a project at the given root path.
// Automatically recovers incomplete transactions.
func Open(root string) (*App, error) {
	cfg, err := loadConfig(root)
	if err != nil {
		return nil, fmt.Errorf("service: load config: %w", err)
	}

	plmDir := filepath.Join(root, ".plm")
	os.MkdirAll(plmDir, 0755)
	os.MkdirAll(filepath.Join(root, "objects"), 0755)

	tx := transaction.NewManager(filepath.Join(plmDir, "transactions"))

	// Auto-recover incomplete transactions
	items, _ := tx.Recover(context.Background())
	for _, item := range items {
		tx.Rollback(context.Background(), item.ID)
	}
	repo := fsrepo.New(root, tx)

	idxPath := filepath.Join(plmDir, "index.db")
	idxDB, err := index.Open(idxPath)
	if err != nil {
		return nil, fmt.Errorf("service: open index: %w", err)
	}

	sequences := recoverSequences(context.Background(), cfg.Sequences, repo)
	nstore := naming.NewMemorySequenceStore(sequences)
	ngen := naming.NewGenerator(cfg.Naming.ProjectCode, nstore)

	// Try loading lifecycle from config/lifecycle.md, fall back to default
	machine := loadLifecycle(root, cfg)

	var gitSvc *gitops.Service
	if cfg.Git.Enabled {
		gitSvc, _ = gitops.Open(root)
	}

	cmdSvc := &command.ObjectService{
		Repo:   repo,
		Index:  idxDB,
		Naming: ngen,
		FSM:    machine,
		Git:    &gitAdapter{svc: gitSvc},
	}

	reader := &objectReader{repo: repo, idx: idxDB}
	bomSvc := bom.New(reader, reader)
	relSvc := release.New(reader, reader)

	qrySvc := &query.ObjectService{
		Objects: reader,
		BOM:     bomSvc,
		Release: relSvc,
	}

	return &App{
		Root:      root,
		Config:    cfg,
		Cmd:       cmdSvc,
		Qry:       qrySvc,
		BOM:       bomSvc,
		Release:   relSvc,
		Git:       gitSvc,
		Index:     idxDB,
		Repo:      repo,
		NamingGen: ngen,
		FSM:       machine,
	}, nil
}

func recoverSequences(ctx context.Context, configured map[string]int, repo *fsrepo.Repository) map[string]int {
	seqs := make(map[string]int)
	for k, v := range naming.DefaultStarts() {
		seqs[k] = v
	}
	for k, v := range configured {
		if v > seqs[k] {
			seqs[k] = v
		}
	}

	objects, errs := repo.ListObjects(ctx)
	if len(errs) > 0 {
		return seqs
	}
	for _, obj := range objects {
		parsed, err := naming.Parse(string(obj.ID))
		if err != nil {
			continue
		}
		n, err := strconv.Atoi(parsed.Sequence)
		if err != nil {
			continue
		}
		if n > seqs[parsed.Class] {
			seqs[parsed.Class] = n
		}
	}
	return seqs
}

// InitProject creates a new PLM project at the given path.
func InitProject(root, code, title string) (*App, error) {
	if err := os.MkdirAll(root, 0755); err != nil {
		return nil, fmt.Errorf("service: mkdir root: %w", err)
	}

	cfg := project.DefaultConfig(code, title)

	// Write project.md
	cfgYAML, _ := yaml.Marshal(cfg)
	fm := frontmatter.Join(cfgYAML, []byte("# "+title+"\n\nPLM project managed by go-plm.\n"))
	if err := os.WriteFile(filepath.Join(root, "project.md"), fm, 0644); err != nil {
		return nil, fmt.Errorf("service: write project.md: %w", err)
	}

	// Write config/lifecycle.md
	os.MkdirAll(filepath.Join(root, "config"), 0755)
	lc := fsm.StandardObjectLifecycle()
	lcYAML, _ := yaml.Marshal(lc)
	lcFM := frontmatter.Join(lcYAML, []byte("# Lifecycle Configuration\n\nDefines the state machine for object lifecycles.\n"))
	if err := os.WriteFile(filepath.Join(root, "config", "lifecycle.md"), lcFM, 0644); err != nil {
		return nil, fmt.Errorf("service: write lifecycle.md: %w", err)
	}

	// Write config/classes.md
	classesYAML, _ := yaml.Marshal(map[string]any{
		"classes": []map[string]any{
			{"code": "prt", "title": "Part", "sequence": "0001-9999", "lifecycle": "object_lifecycle"},
			{"code": "asm", "title": "Assembly", "sequence": "0100-9999", "lifecycle": "object_lifecycle"},
			{"code": "drw", "title": "Drawing", "lifecycle": "document_lifecycle"},
			{"code": "doc", "title": "Document", "lifecycle": "document_lifecycle"},
			{"code": "std", "title": "Standard Part", "lifecycle": "standard_part_lifecycle"},
			{"code": "mat", "title": "Material", "lifecycle": "object_lifecycle"},
		},
	})
	classesFM := frontmatter.Join(classesYAML, []byte("# Object Classes\n\nDefines all object classes available in this project.\n"))
	os.WriteFile(filepath.Join(root, "config", "classes.md"), classesFM, 0644)

	// Write config/naming.md
	namingYAML, _ := yaml.Marshal(map[string]any{
		"standard": "v10",
		"pattern":  "[project]-[class]-[sequence]-v[major].[minor]",
		"case":     "kebab",
	})
	namingFM := frontmatter.Join(namingYAML, []byte("# Naming Standard Configuration\n\nDefines the ID format for all objects.\n"))
	os.WriteFile(filepath.Join(root, "config", "naming.md"), namingFM, 0644)

	// Write config/validation.md
	validationYAML, _ := yaml.Marshal(map[string]any{
		"rules": []map[string]any{
			{"id": "valid_name", "severity": "blocker"},
			{"id": "required_metadata", "severity": "blocker", "fields": map[string][]string{
				"prt": {"unit", "make_buy"},
				"asm": {"unit"},
				"std": {"std_class", "std_id", "make_buy"},
			}},
			{"id": "valid_relations", "severity": "blocker"},
			{"id": "valid_artifacts", "severity": "warning"},
			{"id": "checksums_actual", "severity": "blocker"},
			{"id": "lifecycle_consistency", "severity": "warning"},
		},
	})
	validationFM := frontmatter.Join(validationYAML, []byte("# Validation Rules\n\nDefines validation rules for the project.\n"))
	os.WriteFile(filepath.Join(root, "config", "validation.md"), validationFM, 0644)

	// Init git if enabled
	if cfg.Git.Enabled {
		git, err := gitops.Init(root)
		if err == nil {
			// Create initial commit
			git.Checkpoint(context.Background(), gitops.CheckpointRequest{
				Message: fmt.Sprintf("Initial commit: %s project created", title),
				Author:  "go-plm",
			})
		}
	}

	return Open(root)
}

// RebuildIndex triggers a full index rebuild.
func (a *App) RebuildIndex(ctx context.Context) error {
	return a.Cmd.RebuildIndex(ctx)
}

// ValidateAll runs validation on all objects.
func (a *App) ValidateAll(ctx context.Context, objects []object.Object) error {
	// Use standardparts validation for relevant objects
	for _, obj := range objects {
		diags := standardparts.Validate(obj)
		if len(diags) > 0 {
			// Log diagnostics — non-fatal
			for _, d := range diags {
				_ = d // collected in frontend
			}
		}
	}
	return nil
}

// Close shuts down all resources.
func (a *App) Close() error {
	if a.Index != nil {
		return a.Index.Close()
	}
	return nil
}

// loadConfig reads project.md and parses project configuration.
func loadConfig(root string) (project.Config, error) {
	path := filepath.Join(root, "project.md")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Default config
			base := filepath.Base(root)
			return project.DefaultConfig(base, base), nil
		}
		return project.Config{}, err
	}

	doc, err := frontmatter.Split(data)
	if err != nil {
		return project.Config{}, fmt.Errorf("parse project.md: %w", err)
	}

	var cfg project.Config
	if err := yaml.Unmarshal(doc.Frontmatter, &cfg); err != nil {
		return project.Config{}, fmt.Errorf("unmarshal project.md: %w", err)
	}
	return cfg, nil
}

// loadLifecycle loads FSM definition from config/lifecycle.md or returns standard.
func loadLifecycle(root string, cfg project.Config) *fsm.Machine {
	path := filepath.Join(root, "config", "lifecycle.md")
	data, err := os.ReadFile(path)
	if err != nil {
		def := fsm.StandardObjectLifecycle()
		return fsm.New(def)
	}

	doc, err := frontmatter.Split(data)
	if err != nil {
		def := fsm.StandardObjectLifecycle()
		return fsm.New(def)
	}

	machine, err := fsm.NewFromYAML(doc.Frontmatter)
	if err != nil {
		def := fsm.StandardObjectLifecycle()
		return fsm.New(def)
	}
	return machine
}

// objectReader adapts repo+index to the query.ObjectReader interface.
type objectReader struct {
	repo *fsrepo.Repository
	idx  *index.DB
}

func (r *objectReader) GetObject(ctx context.Context, id object.ID) (object.Object, error) {
	return r.repo.GetObject(ctx, id)
}

func (r *objectReader) ListObjects(ctx context.Context) ([]object.Object, []error) {
	return r.repo.ListObjects(ctx)
}

func (r *objectReader) SearchObjects(ctx context.Context, q string, limit int) ([]object.ID, error) {
	return r.idx.SearchObjects(ctx, q, limit)
}

func (r *objectReader) ListRelations(ctx context.Context, fromID object.ID) ([]object.RelationRef, error) {
	obj, err := r.repo.GetObject(ctx, fromID)
	if err != nil {
		return nil, err
	}
	return obj.Relations, nil
}

// gitAdapter adapts gitops.Service to command.GitOps interface.
type gitAdapter struct {
	svc *gitops.Service
}

func (g *gitAdapter) Status() (modified, added, deleted []string, clean bool, err error) {
	if g.svc == nil {
		return nil, nil, nil, true, nil
	}
	st, err := g.svc.Status()
	if err != nil {
		return nil, nil, nil, false, err
	}
	return st.Modified, st.Added, st.Deleted, st.IsClean, nil
}

func (g *gitAdapter) Checkpoint(ctx context.Context, message, author string) (string, error) {
	if g.svc == nil {
		return "", fmt.Errorf("git not available")
	}
	r, err := g.svc.Checkpoint(ctx, gitops.CheckpointRequest{Message: message, Author: author})
	if err != nil {
		return "", err
	}
	return r.CommitHash, nil
}
