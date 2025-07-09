package gitlab

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// Client represents a GitLab API client.
// It holds the base URL, authentication token, HTTP client, and the project ID
// for project-scoped API calls. It also exposes various GitLab API services.
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
	projectID  string

	// Services
	MergeRequests MergeRequestsService
	// Add other services here as they are implemented, e.g., Tags, Projects, etc.
}

// GitLabError represents an error response from the GitLab API.
type GitLabError struct {
	StatusCode int
	Message    string
	Body       []byte
}

// Error returns a string representation of the GitLabError.
func (e *GitLabError) Error() string {
	return fmt.Sprintf("GitLab API error (%d): %s -- %s", e.StatusCode, e.Message, string(e.Body))
}

// NewClient creates a new GitLab client using environment variables for configuration.
// It supports both GitLab CI pipeline mode and local development mode.
// Required environment variables:
//   - CI mode: GITLAB_CI=true, CI_API_V4_URL, CI_JOB_TOKEN (or SYAC_TOKEN), CI_PROJECT_ID
//   - Local mode: GITLAB_BASE_URL, GITLAB_API_TOKEN, GITLAB_PROJECT_ID
// An optional GITLAB_CLIENT_TIMEOUT_SECONDS can be set to configure the HTTP client timeout.
func NewClient() (*Client, error) {
	var baseURL, token, projectID string

	isPipeline := os.Getenv("GITLAB_CI") == "true"

	if isPipeline {
		baseURL = strings.TrimSuffix(os.Getenv("CI_API_V4_URL"), "/api/v4")
		token = os.Getenv("CI_JOB_TOKEN")
		if token == "" {
			// Optional fallback for impersonation token
			token = os.Getenv("SYAC_TOKEN")
		}
		projectID = os.Getenv("CI_PROJECT_ID")
	} else {
		baseURL = os.Getenv("GITLAB_BASE_URL")
		token = os.Getenv("GITLAB_API_TOKEN")
		projectID = os.Getenv("GITLAB_PROJECT_ID")
	}

	if token == "" {
		if isPipeline {
			return nil, errors.New("CI_JOB_TOKEN or SYAC_TOKEN must be set in CI mode")
		}
		return nil, errors.New("GITLAB_API_TOKEN must be set in local mode")
	}

	if projectID == "" {
		return nil, errors.New("CI_PROJECT_ID or GITLAB_PROJECT_ID must be set")
	}

	if _, err := url.ParseRequestURI(baseURL); err != nil {
		return nil, errors.New("invalid GitLab base URL: " + err.Error())
	}

	// Optional timeout override
	timeout := 10 * time.Second
	if timeoutStr := os.Getenv("GITLAB_CLIENT_TIMEOUT_SECONDS"); timeoutStr != "" {
		if seconds, err := strconv.Atoi(timeoutStr); err == nil && seconds > 0 {
			timeout = time.Duration(seconds) * time.Second
		}
	}

	c := &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		token:   token,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		projectID: projectID,
	}

	// Initialize services
	c.MergeRequests = &mrsService{client: c}

	return c, nil
}

// DoRequest sends an HTTP request to the GitLab API and returns the response body.
// It handles request creation, authentication, execution, and error parsing.
// The 'path' should be relative to the /api/v4 endpoint (e.g., "/projects/123/merge_requests/456").
func (c *Client) DoRequest(method, path string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBytes, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBytes)
	}

	fullURL := fmt.Sprintf("%s/api/v4%s", c.baseURL, path)
	req, err := http.NewRequest(method, fullURL, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request [%s %s]: %w", method, fullURL, err)
	}

	req.Header.Set("PRIVATE-TOKEN", c.token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed [%s %s]: %w", method, fullURL, err)
	}
	defer resp.Body.Close()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, &GitLabError{
			StatusCode: resp.StatusCode,
			Message:    http.StatusText(resp.StatusCode),
			Body:       respData,
		}
	}

	return respData, nil
}

// urlEncode safely encodes a GitLab project path (e.g., "group/project" -> "group%2Fproject").
func urlEncode(s string) string {
	return url.PathEscape(s)
}