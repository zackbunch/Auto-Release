package gitlab

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sort"

	"syac/internal/version"
)

// Client is a simple GitLab API client.
// It is not a full-featured client, but it provides the functionality needed by SYAC.
type Client struct {
	Token      string
	ProjectID  string
	BaseURL    string
	HTTPClient *http.Client
}

// NewClient creates a new GitLab client.
func NewClient() (*Client, error) {
	token := os.Getenv("GITLAB_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("GITLAB_TOKEN environment variable not set")
	}

	projectID := os.Getenv("CI_PROJECT_ID")
	if projectID == "" {
		return nil, fmt.Errorf("CI_PROJECT_ID environment variable not set")
	}

	baseURL := os.Getenv("CI_API_V4_URL")
	if baseURL == "" {
		return nil, fmt.Errorf("CI_API_V4_URL environment variable not set")
	}

	return &Client{
		Token:      token,
		ProjectID:  projectID,
		BaseURL:    baseURL,
		HTTPClient: &http.Client{},
	}, nil
}

// MergeRequest represents a GitLab merge request.
// It only contains the fields needed by SYAC.
type MergeRequest struct {
	Description string `json:"description"`
}

// GetMergeRequestDescription fetches the description of a merge request.
func (c *Client) GetMergeRequestDescription(mrID string) (string, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/projects/%s/merge_requests/%s", c.BaseURL, c.ProjectID, mrID), nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Add("PRIVATE-TOKEN", c.Token)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get merge request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get merge request: %s", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	var mr MergeRequest
	if err := json.Unmarshal(body, &mr); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return mr.Description, nil
}

// Tag represents a GitLab tag.
// It only contains the fields needed by SYAC.
type Tag struct {
	Name string `json:"name"`
}

// GetLatestTag fetches the latest semantic version tag from the project.
func (c *Client) GetLatestTag() (version.Version, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/projects/%s/repository/tags", c.BaseURL, c.ProjectID), nil)
	if err != nil {
		return version.Version{}, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Add("PRIVATE-TOKEN", c.Token)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return version.Version{}, fmt.Errorf("failed to get tags: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return version.Version{}, fmt.Errorf("failed to get tags: %s", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return version.Version{}, fmt.Errorf("failed to read response body: %w", err)
	}

	var tags []Tag
	if err := json.Unmarshal(body, &tags); err != nil {
		return version.Version{}, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	var versions []version.Version
	for _, tag := range tags {
		v, err := version.Parse(tag.Name)
		if err == nil {
			versions = append(versions, v)
		}
	}

	if len(versions) == 0 {
		return version.Version{}, fmt.Errorf("no semantic version tags found")
	}

	// Sort versions descending
	sort.Slice(versions, func(i, j int) bool {
		if versions[i].Major != versions[j].Major {
			return versions[i].Major > versions[j].Major
		}
		if versions[i].Minor != versions[j].Minor {
			return versions[i].Minor > versions[j].Minor
		}
		return versions[i].Patch > versions[j].Patch
	})

	return versions[0], nil
}
