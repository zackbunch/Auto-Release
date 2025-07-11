package gitlab

import (
	"fmt"
)

// ReleasesService defines the interface for GitLab Release operations.
type ReleasesService interface {
	CreateRelease(tagName, ref, name, description string) error
}

// releasesService is a concrete implementation of ReleasesService.
type releasesService struct {
	client *Client
}

// CreateRelease creates a new release in the project.
func (s *releasesService) CreateRelease(tagName, ref, name, description string) error {
	path := fmt.Sprintf("/projects/%s/releases", urlEncode(s.client.projectID))

	payload := map[string]string{
		"tag_name":    tagName,
		"ref":         ref,
		"name":        name,
		"description": description,
	}

	_, err := s.client.DoRequest("POST", path, payload)
	if err != nil {
		return fmt.Errorf("failed to create release %q: %w", tagName, err)
	}
	return nil
}
