// Package ports defines the interfaces that the application layer depends on.
// All implementations (fsrepo, index, gitops) satisfy these interfaces.
package ports

import (
	"context"

	"github.com/Homiakus/go-plm/internal/core/object"
)

// ObjectRepository reads and writes objects.
type ObjectRepository interface {
	GetObject(ctx context.Context, id object.ID) (object.Object, error)
	SaveObject(ctx context.Context, obj object.Object) error
	ListObjects(ctx context.Context) ([]object.Object, []error)
	DeleteObject(ctx context.Context, id object.ID) error
	Exists(id object.ID) bool
	AppendHistory(ctx context.Context, id object.ID, line string) error
}

// Indexer provides fast read access and search.
type Indexer interface {
	UpsertObject(ctx context.Context, obj object.Object) error
	DeleteObject(ctx context.Context, id object.ID) error
	GetObject(ctx context.Context, id object.ID) (object.Object, error)
	SearchObjects(ctx context.Context, query string, limit int) ([]object.ID, error)
	RebuildIndex(ctx context.Context, objects []object.Object) error
}

// GitService abstracts git operations.
type GitService interface {
	Status() (modified, added, deleted []string, clean bool, err error)
	Checkpoint(ctx context.Context, message, author string) (hash string, err error)
	CreateTag(name, message string) error
}
