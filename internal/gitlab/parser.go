package gitlab

import (
	"strings"
	"syac/internal/version"
)

// ParseVersionBumpHint parses the merge request description for a version bump hint.
// It returns the version type to bump, or an empty string if no hint is found.
func ParseVersionBumpHint(description string) version.VersionType {
	if strings.Contains(description, "#major") {
		return version.Major
	}
	if strings.Contains(description, "#minor") {
		return version.Minor
	}
	if strings.Contains(description, "#patch") {
		return version.Patch
	}
	return ""
}
