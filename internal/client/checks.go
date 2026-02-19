package client

import (
	"fmt"
	"net/http"
)

// Check represents a Quismon health check
type Check struct {
	ID                  string                 `json:"id"`
	OrgID               string                 `json:"org_id"`
	Name                string                 `json:"name"`
	Type                string                 `json:"type"`
	Config              map[string]interface{} `json:"config"`
	ConfigHash          string                 `json:"config_hash,omitempty"` // Hash of sensitive fields for drift detection
	IntervalSeconds     int                    `json:"interval_seconds"`
	Regions             []string               `json:"regions"`
	Enabled             bool                   `json:"enabled"`
	Inverted            bool                   `json:"inverted"` // Alert on success instead of failure
	SimultaneousRegions bool                   `json:"simultaneous_regions"`
	RecheckOnFailure    bool                   `json:"recheck_on_failure"`
	ExpiresAfterSeconds *int                   `json:"expires_after_seconds,omitempty"` // Check auto-deletes after this many seconds
	HealthStatus        string                 `json:"health_status,omitempty"`
	LastChecked         *string                `json:"last_checked,omitempty"`
	CreatedAt           string                 `json:"created_at"`
	UpdatedAt           string                 `json:"updated_at"`
}

// CreateCheckRequest represents a request to create a check
type CreateCheckRequest struct {
	Name                string                 `json:"name"`
	Type                string                 `json:"type"`
	Config              map[string]interface{} `json:"config"`
	IntervalSeconds     int                    `json:"interval_seconds"`
	Regions             []string               `json:"regions,omitempty"`
	Enabled             bool                   `json:"enabled"`
	Inverted            *bool                  `json:"inverted,omitempty"` // Alert on success instead of failure
	SimultaneousRegions *bool                  `json:"simultaneous_regions,omitempty"`
	RecheckOnFailure    *bool                  `json:"recheck_on_failure,omitempty"`
	ExpiresAfterSeconds *int                   `json:"expires_after_seconds,omitempty"` // Check auto-deletes after this many seconds
}

// UpdateCheckRequest represents a request to update a check
type UpdateCheckRequest struct {
	Name                *string                 `json:"name,omitempty"`
	Type                *string                 `json:"type,omitempty"`
	Config              *map[string]interface{} `json:"config,omitempty"`
	IntervalSeconds     *int                    `json:"interval_seconds,omitempty"`
	Regions             *[]string               `json:"regions,omitempty"`
	Enabled             *bool                   `json:"enabled,omitempty"`
	Inverted            *bool                   `json:"inverted,omitempty"` // Alert on success instead of failure
	SimultaneousRegions *bool                   `json:"simultaneous_regions,omitempty"`
	RecheckOnFailure    *bool                   `json:"recheck_on_failure,omitempty"`
	ExpiresAfterSeconds *int                    `json:"expires_after_seconds,omitempty"` // Check auto-deletes after this many seconds
}

// ListChecks retrieves all checks
func (c *Client) ListChecks() ([]Check, error) {
	data, err := c.DoRequest(http.MethodGet, "/v1/checks", nil)
	if err != nil {
		return nil, err
	}

	var checks []Check
	if err := UnmarshalAPIResponse(data, &checks); err != nil {
		return nil, err
	}

	return checks, nil
}

// GetCheck retrieves a specific check by ID
func (c *Client) GetCheck(id string) (*Check, error) {
	data, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/v1/checks/%s", id), nil)
	if err != nil {
		return nil, err
	}

	var check Check
	if err := UnmarshalAPIResponse(data, &check); err != nil {
		return nil, err
	}

	return &check, nil
}

// CreateCheck creates a new check
func (c *Client) CreateCheck(req CreateCheckRequest) (*Check, error) {
	data, err := c.DoRequest(http.MethodPost, "/v1/checks", req)
	if err != nil {
		return nil, err
	}

	var check Check
	if err := UnmarshalAPIResponse(data, &check); err != nil {
		return nil, err
	}

	return &check, nil
}

// UpdateCheck updates an existing check
func (c *Client) UpdateCheck(id string, req UpdateCheckRequest) (*Check, error) {
	data, err := c.DoRequest(http.MethodPut, fmt.Sprintf("/v1/checks/%s", id), req)
	if err != nil {
		return nil, err
	}

	var check Check
	if err := UnmarshalAPIResponse(data, &check); err != nil {
		return nil, err
	}

	return &check, nil
}

// DeleteCheck deletes a check
func (c *Client) DeleteCheck(id string) error {
	_, err := c.DoRequest(http.MethodDelete, fmt.Sprintf("/v1/checks/%s", id), nil)
	return err
}

// GetCheckByName retrieves a check by name
func (c *Client) GetCheckByName(name string) (*Check, error) {
	checks, err := c.ListChecks()
	if err != nil {
		return nil, err
	}

	for _, check := range checks {
		if check.Name == name {
			return &check, nil
		}
	}

	return nil, fmt.Errorf("check not found: %s", name)
}
