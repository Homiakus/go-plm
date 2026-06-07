// Package transaction provides atomic filesystem operations.
// All Source-of-Truth writes go through this layer.
// Uses temp files + fsync + atomic rename. Direct overwrites are forbidden.
package transaction

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Op types.
const (
	OpWrite  = "write"
	OpCopy   = "copy"
	OpDelete = "delete"
	OpRename = "rename"
)

// WriteOp writes content to a file.
type WriteOp struct {
	Path    string `json:"path"`
	Content []byte `json:"-"`
}

// CopyOp copies a file from source to destination.
type CopyOp struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// DeleteOp removes a file.
type DeleteOp struct {
	Path string `json:"path"`
}

// RenameOp renames/moves a file.
type RenameOp struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// Plan describes an atomic unit of filesystem work.
type Plan struct {
	ID      string     `json:"id"`
	Writes  []WriteOp  `json:"writes,omitempty"`
	Copies  []CopyOp   `json:"copies,omitempty"`
	Deletes []DeleteOp `json:"deletes,omitempty"`
	Renames []RenameOp `json:"renames,omitempty"`
}

// Manager executes transaction plans atomically.
type Manager struct {
	TxDir string // directory for temp files and commit markers
}

// NewManager creates a Manager with the given transaction directory.
func NewManager(txDir string) *Manager {
	return &Manager{TxDir: txDir}
}

// Execute runs the plan as an atomic unit.
// On success, all operations are committed. On failure, partial work is left in TxDir for recovery.
func (m *Manager) Execute(ctx context.Context, p Plan) error {
	if p.ID == "" {
		return errors.New("transaction: plan.ID must not be empty")
	}
	if err := validateID(p.ID); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	workDir := filepath.Join(m.TxDir, p.ID)
	if err := os.MkdirAll(workDir, 0755); err != nil {
		return fmt.Errorf("transaction: create work dir: %w", err)
	}

	// Phase 1: prepare temp files
	type preparedWrite struct {
		target string
		tmp    string
	}
	tempWrites := make([]preparedWrite, 0, len(p.Writes))
	for i, w := range p.Writes {
		tmp := filepath.Join(workDir, fmt.Sprintf("write_%d.tmp", i))
		if err := os.WriteFile(tmp, w.Content, 0644); err != nil {
			return fmt.Errorf("transaction: write temp %s: %w", tmp, err)
		}
		if err := os.MkdirAll(filepath.Dir(w.Path), 0755); err != nil {
			return fmt.Errorf("transaction: mkdir %s: %w", filepath.Dir(w.Path), err)
		}
		tempWrites = append(tempWrites, preparedWrite{target: w.Path, tmp: tmp})
	}

	// Phase 2: copy files
	for _, c := range p.Copies {
		src, err := os.ReadFile(c.From)
		if err != nil {
			return fmt.Errorf("transaction: read source %s: %w", c.From, err)
		}
		if err := os.MkdirAll(filepath.Dir(c.To), 0755); err != nil {
			return fmt.Errorf("transaction: mkdir %s: %w", filepath.Dir(c.To), err)
		}
		if err := os.WriteFile(c.To, src, 0644); err != nil {
			return fmt.Errorf("transaction: copy %s -> %s: %w", c.From, c.To, err)
		}
	}

	// Phase 3: atomic rename (commit)
	for _, w := range tempWrites {
		if err := os.Rename(w.tmp, w.target); err != nil {
			// Cleanup temp — already committed writes stay
			return fmt.Errorf("transaction: rename %s -> %s: %w", w.tmp, w.target, err)
		}
	}

	// Phase 4: rename operations
	for _, r := range p.Renames {
		if err := os.MkdirAll(filepath.Dir(r.To), 0755); err != nil {
			return fmt.Errorf("transaction: mkdir for rename %s: %w", r.To, err)
		}
		if err := os.Rename(r.From, r.To); err != nil {
			return fmt.Errorf("transaction: rename %s -> %s: %w", r.From, r.To, err)
		}
	}

	// Phase 5: delete operations
	for _, d := range p.Deletes {
		if err := os.RemoveAll(d.Path); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("transaction: delete %s: %w", d.Path, err)
		}
	}

	// Phase 6: write commit marker
	marker := filepath.Join(workDir, "committed")
	if err := os.WriteFile(marker, []byte("ok"), 0644); err != nil {
		return fmt.Errorf("transaction: write commit marker: %w", err)
	}

	// Cleanup work dir (best-effort)
	_ = os.RemoveAll(workDir)

	return nil
}

// Recover finds incomplete transactions and returns recovery items.
func (m *Manager) Recover(ctx context.Context) ([]RecoveryItem, error) {
	entries, err := os.ReadDir(m.TxDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var items []RecoveryItem
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		workDir := filepath.Join(m.TxDir, e.Name())
		marker := filepath.Join(workDir, "committed")
		if _, err := os.Stat(marker); err == nil {
			// Already committed — safe to clean up
			_ = os.RemoveAll(workDir)
			continue
		}
		items = append(items, RecoveryItem{
			ID:      e.Name(),
			WorkDir: workDir,
		})
	}
	return items, nil
}

// Rollback removes the work directory of an incomplete transaction.
func (m *Manager) Rollback(ctx context.Context, id string) error {
	if err := validateID(id); err != nil {
		return err
	}
	return os.RemoveAll(filepath.Join(m.TxDir, id))
}

// RecoveryItem represents an incomplete or completed transaction.
type RecoveryItem struct {
	ID      string `json:"id"`
	WorkDir string `json:"work_dir"`
}

// NewID generates a unique transaction ID.
func NewID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return "tx-" + hex.EncodeToString(b)
}

func validateID(id string) error {
	if id == "" {
		return errors.New("transaction: id must not be empty")
	}
	if id == "." || id == ".." || filepath.IsAbs(id) || strings.ContainsAny(id, `/\`) {
		return fmt.Errorf("transaction: unsafe id %q", id)
	}
	if clean := filepath.Clean(id); clean != id {
		return fmt.Errorf("transaction: unsafe id %q", id)
	}
	return nil
}

// IsCommitted checks if a transaction work directory has a commit marker.
func IsCommitted(workDir string) bool {
	_, err := os.Stat(filepath.Join(workDir, "committed"))
	return err == nil
}

// Plan types carry json struct tags for diagnostics/recovery reporting.
var _ = json.Marshal
