package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client is the Quismon API client
type Client struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
}

// APIResponse represents the standard API response wrapper
type APIResponse struct {
	Data  json.RawMessage   `json:"data,omitempty"`
	Error *string           `json:"error,omitempty"`
	Meta  map[string]string `json:"meta,omitempty"`
}

// New creates a new Quismon API client
func New(baseURL, apiKey string) (*Client, error) {
	if baseURL == "" {
		return nil, fmt.Errorf("base URL is required")
	}
	// API key is optional - some resources (like signup) don't require it
	// Individual methods that need auth will fail if no key is provided

	return &Client{
		BaseURL: baseURL,
		APIKey:  apiKey,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// DoRequest performs an HTTP request with authentication
func (c *Client) DoRequest(method, path string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, c.BaseURL+path, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set Authorization header - ensure proper format
	apiKey := strings.TrimSpace(c.APIKey)
	if apiKey != "" {
		// Don't double-prefix if the key already has Bearer
		if !strings.HasPrefix(apiKey, "Bearer ") {
			req.Header.Set("Authorization", "Bearer "+apiKey)
		} else {
			req.Header.Set("Authorization", apiKey)
		}
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "terraform-provider-quismon/1.0")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Handle error responses
	if resp.StatusCode >= 400 {
		var apiResp APIResponse
		if err := json.Unmarshal(bodyBytes, &apiResp); err == nil && apiResp.Error != nil {
			return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, *apiResp.Error)
		}
		return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, string(bodyBytes))
	}

	return bodyBytes, nil
}

// UnmarshalAPIResponse unmarshals the API response data field
func UnmarshalAPIResponse(data []byte, v interface{}) error {
	var apiResp APIResponse
	if err := json.Unmarshal(data, &apiResp); err != nil {
		return fmt.Errorf("failed to unmarshal API response: %w", err)
	}

	if apiResp.Error != nil {
		return fmt.Errorf("API error: %s", *apiResp.Error)
	}

	if apiResp.Data != nil {
		if err := json.Unmarshal(apiResp.Data, v); err != nil {
			return fmt.Errorf("failed to unmarshal response data: %w", err)
		}
	}

	return nil
}
