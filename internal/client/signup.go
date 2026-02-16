package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// QuickSignupRequest represents the quick signup request
type QuickSignupRequest struct {
	Email   string `json:"email"`
	OrgName string `json:"org_name,omitempty"`
}

// QuickSignupResponse represents the quick signup response
type QuickSignupResponse struct {
	OrgID                string `json:"org_id"`
	APIKey               string `json:"api_key"`
	Email                string `json:"email"`
	VerificationRequired bool   `json:"verification_required"`
}

// QuickSignup performs a quick signup without requiring existing credentials
func QuickSignup(ctx context.Context, baseURL string, req QuickSignupRequest) (*QuickSignupResponse, error) {
	if baseURL == "" {
		return nil, fmt.Errorf("base URL is required")
	}

	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", baseURL+"/v1/auth/quick-signup", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", "terraform-provider-quismon/1.0")

	resp, err := httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		var apiResp APIResponse
		if err := json.Unmarshal(bodyBytes, &apiResp); err == nil && apiResp.Error != nil {
			return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, *apiResp.Error)
		}
		return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, string(bodyBytes))
	}

	var signupResp QuickSignupResponse
	if err := UnmarshalAPIResponse(bodyBytes, &signupResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &signupResp, nil
}
