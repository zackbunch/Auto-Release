package gitlab

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"syac/internal/version"
)

// MergeRequestsService defines the interface for GitLab Merge Request operations.
type MergeRequestsService interface {
	GetMergeRequestDescription(mrID string) (string, error)
	CreateMergeRequestComment(mrID string) error
	GetMergeRequestNotes(mrID string) ([]MergeRequestNote, error)
	GetVersionBump(mrID string) (version.VersionType, error)
}

// mrsService is a concrete implementation of MergeRequestsService.
// It holds a reference to the base Client to make API calls.
type mrsService struct {
	client *Client // A reference to the base GitLab client
}

// GetMergeRequestDescription implements the MergeRequestsService interface.
// It fetches the description of a specific merge request within the client's project.
func (s *mrsService) GetMergeRequestDescription(mrID string) (string, error) {
	path := fmt.Sprintf("/projects/%s/merge_requests/%s", urlEncode(s.client.projectID), mrID)
	respData, err := s.client.DoRequest("GET", path, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get merge request description: %w", err)
	}

	// Define a minimal struct to unmarshal only the description
	var mr struct {
		Description string `json:"description"`
	}

	if err := json.Unmarshal(respData, &mr); err != nil {
		return "", fmt.Errorf("failed to unmarshal merge request description: %w", err)
	}

	return mr.Description, nil
}

// CreateMergeRequestComment creates a new comment on a Merge Request
func (s *mrsService) CreateMergeRequestComment(mrID string) error {
	content, err := os.ReadFile("mr_comment.md")
	if err != nil {
		return fmt.Errorf("failed to read mr_comment.md: %w", err)
	}

	path := fmt.Sprintf("/projects/%s/merge_requests/%s/notes", urlEncode(s.client.projectID), mrID)
	payload := map[string]string{
		"body": string(content),
	}

	_, err = s.client.DoRequest("POST", path, payload)
	if err != nil {
		return fmt.Errorf("failed to post MR note: %w", err)
	}

	return nil
}

// GetMergeRequestNotes returns all user comments on a given merge request
func (s *mrsService) GetMergeRequestNotes(mrID string) ([]MergeRequestNote, error) {
	path := fmt.Sprintf("/projects/%s/merge_requests/%s/notes", urlEncode(s.client.projectID), mrID)

	respData, err := s.client.DoRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var notes []MergeRequestNote
	if err := json.Unmarshal(respData, &notes); err != nil {
		return nil, fmt.Errorf("failed to parse merge request notes: %w", err)
	}

	return notes, nil
}

// ParseVersionFromDescription scans the MR description for a SYAC version checkbox
func ParseVersionFromDescription(description string) (version.VersionType, error) {
	checkboxRe := regexp.MustCompile(`- \[x\] \*\*(Patch|Minor|Major)\*\*`)
	lines := strings.Split(description, "\n")
	for _, line := range lines {
		if matches := checkboxRe.FindStringSubmatch(line); len(matches) > 1 {
			return version.VersionType(matches[1]), nil
		}
	}

	fmt.Println("WARNING: No version type checkbox checked in MR description. Defaulting to Patch.")
	return "Patch", nil
}

func (s *mrsService) GetVersionBump(mrID string) (version.VersionType, error) {
	description, err := s.GetMergeRequestDescription(mrID)
	if err != nil {
		return "", fmt.Errorf("failed to get MR description: %w", err)
	}

	bumpType, err := ParseVersionFromDescription(description)
	if err != nil {
		return "", fmt.Errorf("failed to parse version from description: %w", err)
	}

	return bumpType, nil
}