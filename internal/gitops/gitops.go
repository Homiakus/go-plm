// Package gitops provides Git integration: status, checkpoint, tag, log, diff.
// Wraps go-git for PLM-specific operations.
package gitops

import (
	"context"
	"fmt"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// Status holds the working tree status.
type Status struct {
	IsClean  bool     `json:"is_clean"`
	Modified []string `json:"modified"`
	Added    []string `json:"added"`
	Deleted  []string `json:"deleted"`
}

// CheckpointRequest describes a commit to create.
type CheckpointRequest struct {
	Message string   `json:"message"`
	Author  string   `json:"author"`
	Paths   []string `json:"paths,omitempty"` // empty = all
}

// CheckpointResult holds the result of a commit.
type CheckpointResult struct {
	CommitHash string `json:"commit_hash"`
}

// Commit represents a git commit for the log.
type Commit struct {
	Hash    string    `json:"hash"`
	Message string    `json:"message"`
	Author  string    `json:"author"`
	Time    time.Time `json:"time"`
}

// Service wraps a git repository.
type Service struct {
	repo *git.Repository
}

// Open opens a git repository at the given path.
func Open(path string) (*Service, error) {
	repo, err := git.PlainOpen(path)
	if err != nil {
		return nil, fmt.Errorf("gitops: open %s: %w", path, err)
	}
	return &Service{repo: repo}, nil
}

// Init initializes a new git repository.
func Init(path string) (*Service, error) {
	repo, err := git.PlainInit(path, false)
	if err != nil {
		return nil, fmt.Errorf("gitops: init %s: %w", path, err)
	}
	return &Service{repo: repo}, nil
}

// Status returns the working tree status.
func (s *Service) Status() (*Status, error) {
	wt, err := s.repo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("gitops: worktree: %w", err)
	}

	ws, err := wt.Status()
	if err != nil {
		return nil, fmt.Errorf("gitops: status: %w", err)
	}

	st := &Status{IsClean: ws.IsClean()}
	for path, fs := range ws {
		switch {
		case fs.Worktree == git.Modified:
			st.Modified = append(st.Modified, path)
		case fs.Worktree == git.Added:
			st.Added = append(st.Added, path)
		case fs.Worktree == git.Deleted:
			st.Deleted = append(st.Deleted, path)
		}
	}
	return st, nil
}

// Checkpoint creates a commit with the given message.
func (s *Service) Checkpoint(ctx context.Context, req CheckpointRequest) (*CheckpointResult, error) {
	wt, err := s.repo.Worktree()
	if err != nil {
		return nil, err
	}

	// Stage files
	for _, path := range req.Paths {
		if _, err := wt.Add(path); err != nil {
			return nil, fmt.Errorf("gitops: add %s: %w", path, err)
		}
	}
	if len(req.Paths) == 0 {
		if _, err := wt.Add("."); err != nil {
			return nil, fmt.Errorf("gitops: add all: %w", err)
		}
	}

	hash, err := wt.Commit(req.Message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  req.Author,
			Email: req.Author + "@plm.local",
			When:  time.Now(),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("gitops: commit: %w", err)
	}

	return &CheckpointResult{CommitHash: hash.String()}, nil
}

// CreateTag creates a lightweight tag.
func (s *Service) CreateTag(name, message string) error {
	h, err := s.repo.Head()
	if err != nil {
		return fmt.Errorf("gitops: head: %w", err)
	}

	tagOpts := &git.CreateTagOptions{
		Message: message,
		Tagger: &object.Signature{
			Name:  "go-plm",
			Email: "plm@local",
			When:  time.Now(),
		},
	}

	_, err = s.repo.CreateTag(name, h.Hash(), tagOpts)
	return err
}

// Log returns recent commits.
func (s *Service) Log(limit int) ([]Commit, error) {
	ref, err := s.repo.Head()
	if err != nil {
		return nil, err
	}

	iter, err := s.repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	var commits []Commit
	count := 0
	_ = iter.ForEach(func(c *object.Commit) error {
		if limit > 0 && count >= limit {
			return fmt.Errorf("stop")
		}
		commits = append(commits, Commit{
			Hash:    c.Hash.String(),
			Message: c.Message,
			Author:  c.Author.Name,
			Time:    c.Author.When,
		})
		count++
		return nil
	})
	// Ignore "stop" error — it's our own limit signal
	return commits, nil
}
