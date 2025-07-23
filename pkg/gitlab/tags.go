package gitlab

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"syac/internal/version"
)

// TagsService defines the contract for interacting with GitLab tags.
// It enables listing tags, creating tags, and calculating semantic version bumps.
// This abstraction makes unit testing easier and allows future replacement or mocking of GitLab.
type TagsService interface {
	// ListProjectTags returns all tags in the current GitLab project.
	ListProjectTags() ([]Tag, error)

	// GetLatestTag parses the existing Git tags and returns the highest semantic version tag.
	// If no SemVer-compatible tag is found, it returns v0.0.0.
	GetLatestTag() (version.Version, error)

	// CreateTag creates a new tag from the provided ref (commit SHA or branch).
	// Optionally includes a message (annotated tag).
	// Also sets the SYAC_TAG environment variable to the created tag name.
	CreateTag(tagName, ref, message string) error

	// GetNextVersion returns the current version and the next bumped version (major, minor, or patch).
	GetNextVersion(bump version.VersionType) (version.Version, version.Version, error)
}

// tagsService is the default implementation of TagsService using a GitLab API client.
type tagsService struct {
	client *Client // Reference to GitLab client with API credentials and project context.
}

// ListProjectTags retrieves all tags in the current project by calling the GitLab Tags API.
// Returns all tags in raw string form (may include non-SemVer tags).
func (s *tagsService) ListProjectTags() ([]Tag, error) {
	path := fmt.Sprintf("/projects/%s/repository/tags", urlEncode(s.client.projectID))
	respData, err := s.client.DoRequest("GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tags: %w", err)
	}

	var tags []Tag
	if err := json.Unmarshal(respData, &tags); err != nil {
		return nil, fmt.Errorf("failed to parse tag list: %w", err)
	}

	return tags, nil
}

// GetLatestTag finds the highest SemVer-compliant tag from the list of GitLab tags.
// It skips any tags that are not valid semantic versions.
// If no SemVer tags exist, it returns version 0.0.0.
func (s *tagsService) GetLatestTag() (version.Version, error) {
	tags, err := s.ListProjectTags()
	if err != nil {
		return version.Version{}, fmt.Errorf("failed to list tags: %w", err)
	}

	var parsedVersions []version.Version
	for _, tag := range tags {
		v, err := version.Parse(tag.Name)
		if err != nil {
			// Ignore non-SemVer tags (e.g., "latest", "release-foo")
			continue
		}
		parsedVersions = append(parsedVersions, v)
	}

	if len(parsedVersions) == 0 {
		// Fallback for empty repo or initial release
		return version.Version{Major: 0, Minor: 0, Patch: 0}, nil
	}

	// Sort versions in ascending order, then return the last item (latest version)
	sort.Slice(parsedVersions, func(i, j int) bool {
		return parsedVersions[i].LessThan(parsedVersions[j])
	})

	return parsedVersions[len(parsedVersions)-1], nil
}

// CreateTag creates a Git tag for the specified ref (commit SHA or branch).
// If a message is provided, the tag is annotated.
// This method also sets the SYAC_TAG environment variable for downstream use.
func (s *tagsService) CreateTag(tagName, ref, message string) error {
	path := fmt.Sprintf("/projects/%s/repository/tags", urlEncode(s.client.projectID))

	payload := map[string]string{
		"tag_name": tagName,
		"ref":      ref,
	}
	if message != "" {
		payload["message"] = message
	}

	_, err := s.client.DoRequest("POST", path, payload)
	if err != nil {
		return fmt.Errorf("failed to create tag %q on ref %q: %w", tagName, ref, err)
	}

	// Set environment variable for other parts of the program
	if err := os.Setenv("SYAC_TAG", tagName); err != nil {
		return fmt.Errorf("failed to set environment variable SYAC_TAG: %w", err)
	}

	return nil
}

// GetNextVersion returns the current version and the version that would result
// from bumping by the given type (major, minor, patch).
func (s *tagsService) GetNextVersion(bump version.VersionType) (version.Version, version.Version, error) {
	current, err := s.GetLatestTag()
	if err != nil {
		return version.Version{}, version.Version{}, fmt.Errorf("failed to get latest tag: %w", err)
	}

	next := current.Increment(bump)
	return current, next, nil
}
