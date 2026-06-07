// Package index provides SQLite-backed index and projections for PLM objects.
// The index is rebuildable from Source of Truth at any time.
package index

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Homiakus/go-plm/internal/core/artifact"
	"github.com/Homiakus/go-plm/internal/core/object"
	"github.com/Homiakus/go-plm/internal/core/relation"
	_ "modernc.org/sqlite"
)

// DB wraps the SQLite connection for the PLM index.
type DB struct {
	db *sql.DB
}

// Open creates or opens the index database.
func Open(path string) (*DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("index: open: %w", err)
	}
	db.SetMaxOpenConns(4) // SQLite: allow concurrent readers in WAL mode

	idx := &DB{db: db}
	if err := idx.migrate(context.Background()); err != nil {
		db.Close()
		return nil, err
	}
	return idx, nil
}

// Close shuts down the database.
func (d *DB) Close() error {
	return d.db.Close()
}

// migrate creates the schema if it doesn't exist.
func (d *DB) migrate(ctx context.Context) error {
	schema := `
	PRAGMA journal_mode=WAL;
	PRAGMA user_version=1;

	CREATE TABLE IF NOT EXISTS objects (
		id          TEXT PRIMARY KEY,
		project     TEXT NOT NULL,
		class       TEXT NOT NULL,
		sequence    TEXT,
		version     TEXT,
		revision    TEXT,
		state       TEXT NOT NULL DEFAULT 'draft',
		title       TEXT NOT NULL,
		metadata    TEXT,
		body        TEXT DEFAULT '',
		created_at  TEXT,
		updated_at  TEXT
	);

	CREATE TABLE IF NOT EXISTS relations (
		id        INTEGER PRIMARY KEY AUTOINCREMENT,
		from_id   TEXT NOT NULL,
		to_id     TEXT NOT NULL,
		type      TEXT NOT NULL,
		quantity  REAL,
		unit      TEXT,
		metadata  TEXT
	);

	CREATE TABLE IF NOT EXISTS artifacts (
		id           TEXT PRIMARY KEY,
		object_id    TEXT NOT NULL,
		kind         TEXT,
		role         TEXT,
		path         TEXT,
		original_name TEXT,
		checksum     TEXT,
		size_bytes   INTEGER,
		generated    INTEGER DEFAULT 0,
		required     INTEGER DEFAULT 0,
		status       TEXT
	);

	CREATE VIRTUAL TABLE IF NOT EXISTS fts_objects USING fts5(
		id, title, class, state, metadata, body,
		content='objects', content_rowid='rowid'
	);
	`
	if _, err := d.db.ExecContext(ctx, schema); err != nil {
		return err
	}
	_, _ = d.db.ExecContext(ctx, "ALTER TABLE objects ADD COLUMN body TEXT DEFAULT ''")
	return nil
}

// UpsertObject inserts or updates an object in the index.
func (d *DB) UpsertObject(ctx context.Context, obj object.Object) error {
	now := time.Now().UTC().Format(time.RFC3339)
	metaJSON, _ := json.Marshal(obj.Metadata)

	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `
		INSERT INTO objects (id, project, class, sequence, version, revision, state, title, metadata, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			class=excluded.class, state=excluded.state, title=excluded.title,
			metadata=excluded.metadata, updated_at=excluded.updated_at
	`, string(obj.ID), obj.Project, string(obj.Class), obj.Sequence,
		obj.Version, obj.Revision, string(obj.State), obj.Title,
		string(metaJSON), now, now)
	if err != nil {
		return err
	}
	if err := rebuildFTS(ctx, tx); err != nil {
		return err
	}
	return tx.Commit()
}

// DeleteObject removes an object and its relations/artifacts from the index.
func (d *DB) DeleteObject(ctx context.Context, id object.ID) error {
	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	tx.ExecContext(ctx, "DELETE FROM relations WHERE from_id=? OR to_id=?", string(id), string(id))
	tx.ExecContext(ctx, "DELETE FROM artifacts WHERE object_id=?", string(id))
	tx.ExecContext(ctx, "DELETE FROM objects WHERE id=?", string(id))
	if err := rebuildFTS(ctx, tx); err != nil {
		return err
	}

	return tx.Commit()
}

// UpsertRelation inserts a relation into the index.
func (d *DB) UpsertRelation(ctx context.Context, r relation.Relation) error {
	var qty *float64
	if r.HasQuantity() {
		qty = r.Quantity
	}
	metaJSON, _ := json.Marshal(r.Metadata)
	_, err := d.db.ExecContext(ctx,
		"INSERT INTO relations (from_id, to_id, type, quantity, unit, metadata) VALUES (?, ?, ?, ?, ?, ?)",
		r.FromID, r.ToID, string(r.Type), qty, r.Unit, string(metaJSON))
	return err
}

// UpsertArtifact inserts an artifact into the index.
func (d *DB) UpsertArtifact(ctx context.Context, objectID string, a artifact.Artifact) error {
	_, err := d.db.ExecContext(ctx,
		`INSERT INTO artifacts (id, object_id, kind, role, path, original_name, checksum, size_bytes, generated, required, status)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT(id) DO UPDATE SET path=excluded.path, checksum=excluded.checksum, status=excluded.status`,
		a.ID, objectID, string(a.Kind), string(a.Role), a.Path, a.OriginalName,
		a.Checksum, a.SizeBytes, boolToInt(a.Generated), boolToInt(a.Required), string(a.Status))
	return err
}

// GetObject retrieves an object from the index.
func (d *DB) GetObject(ctx context.Context, id object.ID) (object.Object, error) {
	row := d.db.QueryRowContext(ctx, "SELECT id, project, class, sequence, version, revision, state, title, metadata FROM objects WHERE id=?", string(id))
	var obj object.Object
	var metaJSON string
	err := row.Scan(&obj.ID, &obj.Project, &obj.Class, &obj.Sequence, &obj.Version, &obj.Revision, &obj.State, &obj.Title, &metaJSON)
	if err != nil {
		return obj, err
	}
	if metaJSON != "" {
		json.Unmarshal([]byte(metaJSON), &obj.Metadata)
	}
	return obj, nil
}

// sanitizeFTS5Query escapes FTS5 special characters and wraps the query safely.
// FTS5 syntax: *, ", AND, OR, NOT, NEAR(), ( ) have special meaning.
// We wrap the user query in double-quotes to treat it as a phrase,
// escaping any embedded double-quotes per FTS5 rules.
func sanitizeFTS5Query(q string) string {
	if q == "" {
		return `""`
	}
	// Escape double-quotes: " → ""
	escaped := strings.ReplaceAll(q, `"`, `""`)
	return `"` + escaped + `"`
}

// SearchObjects performs FTS5 search over objects.
func (d *DB) SearchObjects(ctx context.Context, query string, limit int) ([]object.ID, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := d.db.QueryContext(ctx,
		"SELECT id FROM fts_objects WHERE fts_objects MATCH ? LIMIT ?", sanitizeFTS5Query(query), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []object.ID
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			continue
		}
		ids = append(ids, object.ID(id))
	}
	return ids, nil
}

// RebuildIndex drops and recreates all index data from a list of objects.
// Uses batch INSERT (500 rows/batch) for performance.
func (d *DB) RebuildIndex(ctx context.Context, objects []object.Object) error {
	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	tx.ExecContext(ctx, "DELETE FROM artifacts")
	tx.ExecContext(ctx, "DELETE FROM relations")
	tx.ExecContext(ctx, "DELETE FROM objects")

	if len(objects) == 0 {
		if err := rebuildFTS(ctx, tx); err != nil {
			return err
		}
		return tx.Commit()
	}

	now := time.Now().UTC().Format(time.RFC3339)
	const batchSize = 500

	for i := 0; i < len(objects); i += batchSize {
		end := i + batchSize
		if end > len(objects) {
			end = len(objects)
		}
		batch := objects[i:end]

		var sb strings.Builder
		sb.WriteString("INSERT INTO objects (id, project, class, sequence, version, revision, state, title, metadata, created_at, updated_at) VALUES ")
		args := make([]interface{}, 0, len(batch)*11)

		for j, obj := range batch {
			if j > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString("(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
			metaJSON, _ := json.Marshal(obj.Metadata)
			args = append(args, string(obj.ID), obj.Project, string(obj.Class), obj.Sequence,
				obj.Version, obj.Revision, string(obj.State), obj.Title,
				string(metaJSON), now, now)
		}

		if _, err := tx.ExecContext(ctx, sb.String(), args...); err != nil {
			return fmt.Errorf("index: batch insert: %w", err)
		}
	}

	if err := rebuildFTS(ctx, tx); err != nil {
		return err
	}

	return tx.Commit()
}

// Stats returns row counts for diagnostics.
func (d *DB) Stats(ctx context.Context) (objects, relations, artifacts int, _ error) {
	d.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM objects").Scan(&objects)
	d.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM relations").Scan(&relations)
	d.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM artifacts").Scan(&artifacts)
	return
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

type ftsRebuilder interface {
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
}

func rebuildFTS(ctx context.Context, exec ftsRebuilder) error {
	_, err := exec.ExecContext(ctx, "INSERT INTO fts_objects(fts_objects) VALUES('rebuild')")
	return err
}
