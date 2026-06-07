// Package fsrepo provides a filesystem-backed repository for PLM objects.
// Uses the transaction layer for all writes. Reads objects/*.md directly.
package fsrepo

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/Homiakus/go-plm/internal/core/object"
	"github.com/Homiakus/go-plm/internal/store/transaction"
)

// Repository provides read/write access to PLM objects on disk.
type Repository struct {
	Root string
	Tx   *transaction.Manager
}

// New creates a new Repository.
func New(root string, tx *transaction.Manager) *Repository {
	return &Repository{Root: root, Tx: tx}
}

// ObjectsDir returns the path to the objects/ directory.
func (r *Repository) ObjectsDir() string {
	return filepath.Join(r.Root, "objects")
}

// ObjectDir returns the path to a specific object's directory.
func (r *Repository) ObjectDir(id object.ID) string {
	return filepath.Join(r.ObjectsDir(), string(id))
}

// ObjectPath returns the path to an object's Markdown file.
func (r *Repository) ObjectPath(id object.ID) string {
	return filepath.Join(r.ObjectDir(id), string(id)+".md")
}

// HistoryPath returns the path to an object's history log.
func (r *Repository) HistoryPath(id object.ID) string {
	return filepath.Join(r.ObjectDir(id), "history.jsonl")
}

// GetObject reads an object from disk by ID.
func (r *Repository) GetObject(ctx context.Context, id object.ID) (object.Object, error) {
	if err := ctx.Err(); err != nil {
		return object.Object{}, err
	}

	path := r.ObjectPath(id)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return object.Object{}, fmt.Errorf("fsrepo: object %s not found", id)
		}
		return object.Object{}, fmt.Errorf("fsrepo: read %s: %w", path, err)
	}

	fm, _, err := splitFrontmatter(data)
	if err != nil {
		return object.Object{}, fmt.Errorf("fsrepo: parse frontmatter %s: %w", path, err)
	}

	var obj object.Object
	if err := yaml.Unmarshal(fm, &obj); err != nil {
		return object.Object{}, fmt.Errorf("fsrepo: unmarshal %s: %w", path, err)
	}

	return obj, nil
}

// SaveObject writes an object to disk atomically via the transaction layer.
func (r *Repository) SaveObject(ctx context.Context, obj object.Object) error {
	if err := obj.ValidateIdentity(); err != nil {
		return err
	}

	fm, err := yaml.Marshal(obj)
	if err != nil {
		return fmt.Errorf("fsrepo: marshal object: %w", err)
	}

	content := buildMarkdown(fm, "")
	mdPath := r.ObjectPath(obj.ID)

	plan := transaction.Plan{
		ID: transaction.NewID(),
		Writes: []transaction.WriteOp{
			{Path: mdPath, Content: content},
		},
	}

	return r.Tx.Execute(ctx, plan)
}

// ListObjects scans the objects/ directory and returns all objects.
// Uses bounded worker pool (8 goroutines) for parallel file reads.
func (r *Repository) ListObjects(ctx context.Context) ([]object.Object, []error) {
	objDir := r.ObjectsDir()
	entries, err := os.ReadDir(objDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, []error{fmt.Errorf("fsrepo: read objects dir: %w", err)}
	}

	// Collect valid object IDs first
	var ids []object.ID
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		id := object.ID(entry.Name())
		mdPath := r.ObjectPath(id)
		if _, err := os.Stat(mdPath); os.IsNotExist(err) {
			continue
		}
		ids = append(ids, id)
	}

	if len(ids) == 0 {
		return nil, nil
	}

	// Parallel reads with bounded concurrency
	const workers = 8
	sem := make(chan struct{}, workers)
	type result struct {
		obj object.Object
		err error
	}
	results := make([]result, len(ids))

	for i, id := range ids {
		sem <- struct{}{}
		go func(idx int, oid object.ID) {
			defer func() { <-sem }()
			obj, err := r.GetObject(ctx, oid)
			results[idx] = result{obj, err}
		}(i, id)
	}
	// Drain semaphore
	for i := 0; i < workers; i++ {
		sem <- struct{}{}
	}

	var objects []object.Object
	var errs []error
	for _, r := range results {
		if r.err != nil {
			errs = append(errs, r.err)
		} else {
			objects = append(objects, r.obj)
		}
	}

	return objects, errs
}

// Exists checks if an object exists on disk.
func (r *Repository) Exists(id object.ID) bool {
	_, err := os.Stat(r.ObjectPath(id))
	return err == nil
}

// DeleteObject removes an object directory via transaction.
func (r *Repository) DeleteObject(ctx context.Context, id object.ID) error {
	dir := r.ObjectDir(id)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("fsrepo: object %s not found", id)
	}

	plan := transaction.Plan{
		ID: transaction.NewID(),
		Deletes: []transaction.DeleteOp{
			{Path: dir},
		},
	}
	return r.Tx.Execute(ctx, plan)
}

// AppendHistory appends a JSON line to the history log.
func (r *Repository) AppendHistory(ctx context.Context, id object.ID, line string) error {
	path := r.HistoryPath(id)
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("fsrepo: open history %s: %w", path, err)
	}
	defer f.Close()

	if _, err := f.WriteString(line + "\n"); err != nil {
		return fmt.Errorf("fsrepo: append history: %w", err)
	}
	return nil
}

// splitFrontmatter extracts YAML frontmatter between --- delimiters.
func splitFrontmatter(data []byte) (frontmatter []byte, body []byte, err error) {
	text := string(data)

	if !strings.HasPrefix(text, "---") {
		return nil, nil, errors.New("fsrepo: missing frontmatter delimiter")
	}

	rest := text[3:]
	// Skip optional newline right after opening ---
	if strings.HasPrefix(rest, "\n") {
		rest = rest[1:]
	}
	// Handle empty frontmatter: "---\n---\n..."
	if strings.HasPrefix(rest, "---") {
		bd := rest[3:]
		if strings.HasPrefix(bd, "\n") {
			bd = bd[1:]
		}
		return []byte(""), []byte(bd), nil
	}

	idx := strings.Index(rest, "\n---")
	if idx < 0 {
		return nil, nil, errors.New("fsrepo: unclosed frontmatter delimiter")
	}

	fm := rest[:idx]
	bd := rest[idx+4:]
	return []byte(fm), []byte(bd), nil
}

// buildMarkdown assembles a full Markdown document from frontmatter YAML and body text.
func buildMarkdown(frontmatter []byte, body string) []byte {
	var b strings.Builder
	b.WriteString("---\n")
	b.Write(frontmatter)
	b.WriteString("\n---")
	if body != "" {
		b.WriteString("\n\n")
		b.WriteString(body)
	}
	b.WriteString("\n")
	return []byte(b.String())
}
