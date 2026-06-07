// Package revision defines Version and Revision types for object versioning.
// Uses SemVer-like scheme: v[major].[minor].
package revision

import (
	"fmt"
	"strconv"
	"strings"
)

// Version represents a SemVer-like version string, e.g. "1.0".
type Version struct {
	Major int
	Minor int
}

// ParseVersion parses a version string like "1.0" or "2.3".
func ParseVersion(s string) (Version, error) {
	s = strings.TrimPrefix(s, "v")
	parts := strings.Split(s, ".")
	if len(parts) != 2 {
		return Version{}, fmt.Errorf("revision: invalid version %q: expected format 'major.minor'", s)
	}
	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return Version{}, fmt.Errorf("revision: invalid major in %q: %w", s, err)
	}
	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return Version{}, fmt.Errorf("revision: invalid minor in %q: %w", s, err)
	}
	if major < 0 || minor < 0 {
		return Version{}, fmt.Errorf("revision: version parts must be non-negative: %q", s)
	}
	return Version{Major: major, Minor: minor}, nil
}

// String returns the version string without "v" prefix.
func (v Version) String() string {
	return fmt.Sprintf("%d.%d", v.Major, v.Minor)
}

// WithVPrefix returns the version string with "v" prefix.
func (v Version) WithVPrefix() string {
	return "v" + v.String()
}

// BumpMinor returns a new Version with minor incremented.
func (v Version) BumpMinor() Version {
	return Version{Major: v.Major, Minor: v.Minor + 1}
}

// BumpMajor returns a new Version with major incremented and minor reset to 0.
func (v Version) BumpMajor() Version {
	return Version{Major: v.Major + 1, Minor: 0}
}

// IsNewer returns true if v is a later version than other.
func (v Version) IsNewer(other Version) bool {
	if v.Major != other.Major {
		return v.Major > other.Major
	}
	return v.Minor > other.Minor
}
