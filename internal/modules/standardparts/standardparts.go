// Package standardparts provides standard part normalization, duplicate detection, and classification.
package standardparts

import (
	"fmt"
	"strings"

	"github.com/Homiakus/go-plm/internal/core/diagnostic"
	"github.com/Homiakus/go-plm/internal/core/object"
)

// DuplicateLevel classifies how similar two standard parts are.
type DuplicateLevel string

const (
	ExactDuplicate     DuplicateLevel = "exact_duplicate"
	ProbableDuplicate  DuplicateLevel = "probable_duplicate"
	PossibleEquivalent DuplicateLevel = "possible_equivalent"
	Substitutable      DuplicateLevel = "substitutable"
	NotDuplicate       DuplicateLevel = "not_duplicate"
)

// DuplicateCandidate is a potential duplicate of a standard part.
type DuplicateCandidate struct {
	ObjectID      string         `json:"object_id"`
	Title         string         `json:"title"`
	NormalizedKey string         `json:"normalized_key"`
	Level         DuplicateLevel `json:"level"`
	Reason        string         `json:"reason"`
}

// StdClass classifies the type of standard part.
type StdClass string

const (
	Fst  StdClass = "fst"  // Fastener
	Brg  StdClass = "brg"  // Bearing
	Elc  StdClass = "elc"  // Electrical
	Pne  StdClass = "pne"  // Pneumatic
	Hyd  StdClass = "hyd"  // Hydraulic
	Mat  StdClass = "mat"  // Material
	Cab  StdClass = "cab"  // Cable
	Mot  StdClass = "mot"  // Motor
	Sen  StdClass = "sen"  // Sensor
	Prof StdClass = "prof" // Profile
	Seal StdClass = "seal" // Seal
)

// NormalizedPart holds the normalized representation of a standard part.
type NormalizedPart struct {
	Key    string            `json:"key"`
	Class  StdClass          `json:"std_class"`
	Tokens map[string]string `json:"tokens"`
}

// Normalize computes a normalized key from standard part metadata.
func Normalize(class StdClass, metadata map[string]any) (NormalizedPart, error) {
	tokens := make(map[string]string)
	tokens["std_class"] = string(class)

	switch class {
	case Fst:
		standard := getStr(metadata, "standard")
		thread := getStr(metadata, "thread")
		length := getStr(metadata, "length_mm")
		grade := getStr(metadata, "material_grade")
		tokens["standard"] = standard
		tokens["thread"] = thread
		tokens["length_mm"] = length
		tokens["material_grade"] = grade
		key := fmt.Sprintf("fst:%s:%s:%s:%s",
			normalizeToken(standard), normalizeToken(thread), normalizeToken(length), normalizeToken(grade))
		return NormalizedPart{Key: key, Class: class, Tokens: tokens}, nil

	case Brg:
		series := getStr(metadata, "bearing_series")
		seal := getStr(metadata, "seal_type")
		clearance := getStr(metadata, "clearance")
		tokens["bearing_series"] = series
		tokens["seal_type"] = seal
		tokens["clearance"] = clearance
		key := fmt.Sprintf("brg:%s:%s:%s",
			normalizeToken(series), normalizeToken(seal), normalizeToken(clearance))
		return NormalizedPart{Key: key, Class: class, Tokens: tokens}, nil

	case Elc:
		typ := getStr(metadata, "component_type")
		value := getStr(metadata, "value")
		pkg := getStr(metadata, "package")
		key := fmt.Sprintf("elc:%s:%s:%s", normalizeToken(typ), normalizeToken(value), normalizeToken(pkg))
		return NormalizedPart{Key: key, Class: class, Tokens: map[string]string{
			"std_class":      string(class),
			"component_type": typ,
			"value":          value,
			"package":        pkg,
		}}, nil

	default:
		key := fmt.Sprintf("%s:%s", class, generateSimpleKey(metadata))
		return NormalizedPart{Key: key, Class: class, Tokens: tokens}, nil
	}
}

// FindDuplicates compares a candidate normalized key against a list of existing parts.
func FindDuplicates(candidate NormalizedPart, existing []NormalizedPart) []DuplicateCandidate {
	var candidates []DuplicateCandidate
	for _, e := range existing {
		level, reason := classifyMatch(candidate, e)
		if level == NotDuplicate {
			continue
		}
		candidates = append(candidates, DuplicateCandidate{
			NormalizedKey: e.Key, Level: level, Reason: reason,
		})
	}
	return candidates
}

// Validate checks a standard part object for issues.
func Validate(obj object.Object) []diagnostic.Diagnostic {
	var diags []diagnostic.Diagnostic
	if obj.Class != object.ClassStandardPart {
		return diags
	}
	if obj.Metadata == nil {
		diags = append(diags, diagnostic.Diagnostic{
			ID: "std-no-meta", Severity: diagnostic.Blocker, Code: "STD_NO_METADATA",
			ObjectID: string(obj.ID), Message: "Standard part must have metadata",
		})
		return diags
	}
	if _, ok := obj.Metadata["std_class"]; !ok {
		diags = append(diags, diagnostic.Diagnostic{
			ID: "std-no-class", Severity: diagnostic.Blocker, Code: "STD_NO_CLASS",
			ObjectID: string(obj.ID), Message: "std_class is required",
		})
	}
	return diags
}

func classifyMatch(a, b NormalizedPart) (DuplicateLevel, string) {
	if a.Key == b.Key {
		return ExactDuplicate, "normalized keys match exactly"
	}
	if a.Class == b.Class {
		matches := 0
		total := len(a.Tokens)
		for k, v := range a.Tokens {
			if k == "std_class" {
				matches++
				continue
			}
			if bv, ok := b.Tokens[k]; ok && v == bv {
				matches++
			}
		}
		if total > 1 && float64(matches)/float64(total) >= 0.8 {
			return ProbableDuplicate, fmt.Sprintf("%d/%d tokens match", matches, total)
		}
	}
	return NotDuplicate, ""
}

func normalizeToken(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "-", "")
	s = strings.ReplaceAll(s, "_", "")
	return s
}

func getStr(m map[string]any, key string) string {
	if v, ok := m[key]; ok {
		return fmt.Sprint(v)
	}
	return ""
}

func generateSimpleKey(m map[string]any) string {
	parts := []string{}
	for k, v := range m {
		parts = append(parts, normalizeToken(k)+":"+normalizeToken(fmt.Sprint(v)))
	}
	return strings.Join(parts, ":")
}
