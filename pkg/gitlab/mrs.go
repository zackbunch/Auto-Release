package gitlab

import (
	"encoding/json"
	"fmt"
)

// MergeRequestsService defines the interface for GitLab Merge Request operations.
type MergeRequestsService interface {
	GetMergeRequestDescription(mrID string) (string, error)
	// Add other methods as needed, e.g.:
	// GetMergeRequest(mrID string) (*MergeRequest, error)
	// ListMergeRequests(opts *MergeRequestListOptions) ([]*MergeRequest, error)
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
