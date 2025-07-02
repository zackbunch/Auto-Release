package semver

import (
	"fmt"
	"regexp"
	"strings"
)

// ParseVersionNote scans a comment body for a SYAC MR note and returns the selected version bump
func ParseVersionNote(body string) VersionType {
	if !strings.HasPrefix(body, "[SYAC]") {
		return Patch // Default if no valid SYAC comment
	}

	// (?i) makes the match case-insensitive
	checkboxRe := regexp.MustCompile(`(?mi)^\s*-\s*\[x\]\s*\*\*(patch|minor|major)\*\*`)
	matches := checkboxRe.FindStringSubmatch(body)
	if len(matches) > 1 {
		switch strings.ToLower(matches[1]) {
		case "minor":
			return Minor
		case "major":
			return Major
		default:
			return Patch
		}
	}

	fmt.Println("WARNING: No version type checkbox checked. Defaulting to Patch.")
	return Patch
}
