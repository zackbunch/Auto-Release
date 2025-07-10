package gitlab

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"syac/internal/version"
)

// TagsService defines the interface for GitLab tag operations
type TagsService interface {
	ListProjectTags() ([]Tag, error)
	GetLatestTag() (version.Version, error)
	CreateTag(tagName, ref, message string) error
	GetNextVersion(bump version.VersionType) (version.Version, version.Version, error)
}

// tagsService is a concrete implementation of TagsService.
type tagsService struct {
	client *Client // A reference to the base GitLab client
}

// ListProjectTags grabs all tags from the project
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

// GetLatestTag fetches all Git tags for a project and returns the latest semantic version
func (s *tagsService) GetLatestTag() (version.Version, error) {
	tags, err := s.ListProjectTags()
	if err != nil {
		return version.Version{}, fmt.Errorf("failed to list tags: %w", err)
	}

	var parsedVersions []version.Version
	for _, tag := range tags {
		v, err := version.Parse(tag.Name)
		if err != nil {
			// skip non-semver tags
			continue
		}
		parsedVersions = append(parsedVersions, v)
	}

	if len(parsedVersions) == 0 {
		return version.Version{}, fmt.Errorf("no valid semantic version tags found")
	}

	sort.Slice(parsedVersions, func(i, j int) bool {
		return parsedVersions[i].LessThan(parsedVersions[j])
	})

	return parsedVersions[len(parsedVersions)-1], nil
}

// CreateTag creates a new tag in the specified project.
// tagName: the new tag to create (e.g., "1.2.4")
// ref: the commit SHA or branch to tag from (e.g., "main" or "abc123")
// message: annotated tag message (optional)
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

	if err := os.Setenv("SYAC_TAG", tagName); err != nil {
		return fmt.Errorf("failed to set environment variable SYAC_TAG: %w", err)
	}

	return nil
}

func (s *tagsService) GetNextVersion(bump version.VersionType) (version.Version, version.Version, error) {
	current, err := s.GetLatestTag()
	if err != nil {
		return version.Version{}, version.Version{}, fmt.Errorf("failed to get latest tag: %w", err)
	}

	next := current.Increment(bump)
	return current, next, nil
}
