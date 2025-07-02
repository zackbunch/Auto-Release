package semver

// VersionType represents a semantic version bump level
type VersionType string

const (
	Patch VersionType = "Patch"
	Minor VersionType = "Minor"
	Major VersionType = "Major"
)
