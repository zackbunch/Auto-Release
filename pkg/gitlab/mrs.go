package gitlab

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"syac/internal/assets"
	"syac/internal/version"
)

// MergeRequestsService defines the interface for GitLab Merge Request operations.
type MergeRequestsService interface {
	GetMergeRequestDescription(mrID string) (string, error)
	UpdateMergeRequestDescription(mrID string, newDescription string) error
	CreateMergeRequestComment(mrID string) error
	GetVersionBump(mrID string) (version.VersionType, error)
	GetMergeRequestForCommit(sha string) (MergeRequest, error)
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

func (s *mrsService) CreateMergeRequestComment(mrID string) error {
	contentBytes, err := assets.MrCommentContent.ReadFile("mr_comment.md")
	if err != nil {
		return fmt.Errorf("failed to read embedded mr_comment.md: %w", err)
	}
	content := string(contentBytes)

	path := fmt.Sprintf("/projects/%s/merge_requests/%s/notes", urlEncode(s.client.projectID), mrID)
	payload := map[string]string{"body": content}

	_, err = s.client.DoRequest("POST", path, payload)
	if err != nil {
		return fmt.Errorf("failed to post MR note: %w", err)
	}

	return nil
}

// ParseVersionBump scans the string for a SYAC version checkbox
func ParseVersionBump(description string) (version.VersionType, bool) {
	checkboxRe := regexp.MustCompile(`- \[x\] \*\*(Patch|Minor|Major)\*\*`)
	lines := strings.Split(description, "\n")
	for _, line := range lines {
		if matches := checkboxRe.FindStringSubmatch(line); len(matches) > 1 {
			return version.VersionType(matches[1]), true
		}
	}

	return "", false
}

func (s *mrsService) GetVersionBump(mrID string) (version.VersionType, error) {
	// Try to get bump type from MR description first
	description, err := s.GetMergeRequestDescription(mrID)
	if err != nil {
		return "", fmt.Errorf("failed to get MR description: %w", err)
	}

	if bumpType, found := ParseVersionBump(description); found {
		return bumpType, nil
	}

	fmt.Println("WARNING: No version type checkbox checked in MR description. Defaulting to Patch.")
	return version.Patch, nil
}

func (s *mrsService) GetMergeRequestForCommit(sha string) (MergeRequest, error) {
	path := fmt.Sprintf("/projects/%s/repository/commits/%s/merge_requests", urlEncode(s.client.projectID), sha)
	respData, err := s.client.DoRequest("GET", path, nil)
	if err != nil {
		return MergeRequest{}, fmt.Errorf("failed to get merge requests for commit %s: %w", sha, err)
	}

	var mrs []MergeRequest
	if err := json.Unmarshal(respData, &mrs); err != nil {
		return MergeRequest{}, fmt.Errorf("failed to unmarshal merge requests: %w", err)
	}

	if len(mrs) == 0 {
		return MergeRequest{}, fmt.Errorf("no merge request found for commit %s", sha)
	}

	return mrs[0], nil
}

func (s *mrsService) UpdateMergeRequestDescription(mrID string, newDescription string) error {
	path := fmt.Sprintf("/projects/%s/merge_requests/%s", urlEncode(s.client.projectID), mrID)

	payload := map[string]string{
		"description": newDescription,
	}

	_, err := s.client.DoRequest("PUT", path, payload)
	if err != nil {
		return fmt.Errorf("failed to update merge request description: %w", err)
	}

	return nil
}
