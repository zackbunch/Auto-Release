package version

import (
	"fmt"
	"strconv"
	"strings"
)

// VersionType represents a semantic version bump level
type VersionType string

const (
	Patch VersionType = "Patch"
	Minor VersionType = "Minor"
	Major VersionType = "Major"
)

type Version struct {
	Major int
	Minor int
	Patch int
}

func (v Version) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// Parse parses a version string in the format "X.Y.Z"
func Parse(versionStr string) (Version, error) {
	parts := strings.Split(versionStr, ".")
	if len(parts) != 3 {
		return Version{}, fmt.Errorf("invalid version format: expected X.Y.Z, got %s", versionStr)
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return Version{}, fmt.Errorf("invalid major version: %w", err)
	}
	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return Version{}, fmt.Errorf("invalid minor version: %w", err)
	}
	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return Version{}, fmt.Errorf("invalid patch version: %w", err)
	}

	return Version{Major: major, Minor: minor, Patch: patch}, nil
}

// Inc returns a new Version incremented based on the given semantic version type
func (v Version) Inc(bump VersionType) Version {
	switch bump {
	case Major:
		return Version{Major: v.Major + 1, Minor: 0, Patch: 0}
	case Minor:
		return Version{Major: v.Major, Minor: v.Minor + 1, Patch: 0}
	case Patch:
		return Version{Major: v.Major, Minor: v.Minor, Patch: v.Patch + 1}
	default:
		return v // no change if invalid bump type
	}
}
