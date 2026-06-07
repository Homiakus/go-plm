// Package naming implements Naming Standard v10: parse, generate, validate object IDs.
package naming

import (
	"fmt"
	"regexp"
	"strconv"
	"sync"
)

// ParsedID holds the components of an object ID.
type ParsedID struct {
	Project  string
	Class    string
	Sequence string
	Major    int
	Minor    int
}

// String reconstructs the full ID string.
func (p ParsedID) String() string {
	return fmt.Sprintf("%s-%s-%s-v%d.%d", p.Project, p.Class, p.Sequence, p.Major, p.Minor)
}

var idPattern = regexp.MustCompile(
	`^([a-z][a-z0-9]*)-([a-z]+)-([a-z0-9]+)-v(\d+)\.(\d+)$`,
)

// Parse breaks an ID string into its components.
func Parse(id string) (ParsedID, error) {
	m := idPattern.FindStringSubmatch(id)
	if m == nil {
		return ParsedID{}, fmt.Errorf("naming: invalid ID %q: must match [project]-[class]-[sequence]-v[major].[minor]", id)
	}

	major, _ := strconv.Atoi(m[4])
	minor, _ := strconv.Atoi(m[5])

	return ParsedID{
		Project:  m[1],
		Class:    m[2],
		Sequence: m[3],
		Major:    major,
		Minor:    minor,
	}, nil
}

// Validate checks that an ID string conforms to Naming Standard v10.
func Validate(id string) error {
	_, err := Parse(id)
	return err
}

// MemorySequenceStore is an in-memory sequence counter for testing and MVP.
type MemorySequenceStore struct {
	mu     sync.Mutex
	seqs   map[string]int
	starts map[string]int // default starting values
}

// NewMemorySequenceStore creates a store with the given starting sequences.
func NewMemorySequenceStore(starts map[string]int) *MemorySequenceStore {
	seqs := make(map[string]int)
	for k, v := range starts {
		seqs[k] = v
	}
	return &MemorySequenceStore{seqs: seqs}
}

// DefaultStarts returns the default sequence starting values defined by the standard.
func DefaultStarts() map[string]int {
	return map[string]int{
		"prt": 0, "asm": 99, "drw": 0, "doc": 0, "cut": 0, "bnd": 0,
		"nc": 0, "ins": 0, "tpc": 0, "wi": 0, "bom": 0, "rel": 0, "cr": 0,
	}
}

// Current returns the current sequence value for a class.
func (s *MemorySequenceStore) Current(class string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.seqs[class]
}

// Next increments and returns the next sequence value for a class.
func (s *MemorySequenceStore) Next(class string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.seqs[class]++
	return s.seqs[class]
}

// Generator produces the next object ID for a given project and class.
type Generator struct {
	Project string
	Store   *MemorySequenceStore
	version string
}

// NewGenerator creates a Generator with the default "1.0" starting version.
func NewGenerator(project string, store *MemorySequenceStore) *Generator {
	return &Generator{Project: project, Store: store, version: "1.0"}
}

// NextID increments the sequence and returns a full ID string.
func (g *Generator) NextID(class string) string {
	seq := g.Store.Next(class)
	width := 4
	seqStr := fmt.Sprintf("%0*d", width, seq)
	return fmt.Sprintf("%s-%s-%s-v%s", g.Project, class, seqStr, g.version)
}
