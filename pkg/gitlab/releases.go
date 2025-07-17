package gitlab

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"
)

// ReleasesService defines the interface for GitLab Release operations.
type ReleasesService interface {
	CreateRelease(tagName, ref, name, description string) error
	GetLatestRelease() (Release, error)
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

// GetLatestRelease fetches all releases and returns the latest one based on creation date.
func (s *releasesService) GetLatestRelease() (Release, error) {
	path := fmt.Sprintf("/projects/%s/releases", urlEncode(s.client.projectID))
	respData, err := s.client.DoRequest("GET", path, nil)
	if err != nil {
		return Release{}, fmt.Errorf("failed to fetch releases: %w", err)
	}

	var releases []Release
	if err := json.Unmarshal(respData, &releases); err != nil {
		return Release{}, fmt.Errorf("failed to unmarshal releases data: %w", err)
	}

	if len(releases) == 0 {
		return Release{}, fmt.Errorf("no releases found for project %s", s.client.projectID)
	}

	sort.Slice(releases, func(i, j int) bool {
		timeI, _ := time.Parse(time.RFC3339, releases[i].CreatedAt)
		timeJ, _ := time.Parse(time.RFC3339, releases[j].CreatedAt)
		return timeI.After(timeJ)
	})

	return releases[0], nil
}
