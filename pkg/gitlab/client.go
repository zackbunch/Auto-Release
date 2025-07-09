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

// Client represents a GitLab API client
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
	projectID  string // New field
}

// GitLabError represents an error response from the GitLab API
type GitLabError struct {
	StatusCode int
	Message    string
	Body       []byte
}

func (e *GitLabError) Error() string {
	return fmt.Sprintf("GitLab API error (%d): %s -- %s", e.StatusCode, e.Message, string(e.Body))
}

// NewClient creates a new GitLab client using environment variables.
func NewClient() (*Client, error) {
	var baseURL, token, projectID string // Add projectID

	isPipeline := os.Getenv("GITLAB_CI") == "true"

	if isPipeline {
		baseURL = strings.TrimSuffix(os.Getenv("CI_API_V4_URL"), "/api/v4")
		token = os.Getenv("CI_JOB_TOKEN")
		if token == "" {
			// Optional fallback for impersonation token
			token = os.Getenv("SYAC_TOKEN")
		}
		projectID = os.Getenv("CI_PROJECT_ID") // Get project ID
	} else {
		baseURL = os.Getenv("GITLAB_BASE_URL")
		token = os.Getenv("GITLAB_API_TOKEN")
		projectID = os.Getenv("GITLAB_PROJECT_ID") // Get project ID for local
	}

	if token == "" {
		if isPipeline {
			return nil, errors.New("CI_JOB_TOKEN or SYAC_TOKEN must be set in CI mode")
		}
		return nil, errors.New("GITLAB_API_TOKEN must be set in local mode")
	}

	if projectID == "" { // New check
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

	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		token:   token,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		projectID: projectID, // Set project ID
	}, nil
}

// DoRequest sends an HTTP request to the GitLab API and returns the response body.
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

// urlEncode safely encodes a GitLab project path (e.g., "group/project" -> "group%2Fproject")
func urlEncode(s string) string {
	return url.PathEscape(s)
}