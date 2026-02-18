package client

import (
	"fmt"
	"net/http"
)

// AlertRule represents an alert rule
type AlertRule struct {
	ID                     string                 `json:"id"`
	CheckID                string                 `json:"check_id"`
	Name                   string                 `json:"name"`
	Condition              map[string]interface{} `json:"condition"`
	NotificationChannelIDs []string               `json:"notification_channel_ids"`
	Enabled                bool                   `json:"enabled"`
	CreatedAt              string                 `json:"created_at"`
	UpdatedAt              string                 `json:"updated_at"`
}

// CreateAlertRuleRequest represents a request to create an alert rule
type CreateAlertRuleRequest struct {
	Name                   string                 `json:"name"`
	Condition              map[string]interface{} `json:"condition"`
	NotificationChannelIDs []string               `json:"notification_channel_ids"`
	Enabled                bool                   `json:"enabled"`
}

// UpdateAlertRuleRequest represents a request to update an alert rule
type UpdateAlertRuleRequest struct {
	Name                   *string                 `json:"name,omitempty"`
	Condition              *map[string]interface{} `json:"condition,omitempty"`
	NotificationChannelIDs *[]string               `json:"notification_channel_ids,omitempty"`
	Enabled                *bool                   `json:"enabled,omitempty"`
}

// ListAlertRules retrieves all alert rules for a check
func (c *Client) ListAlertRules(checkID string) ([]AlertRule, error) {
	data, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/v1/checks/%s/alerts", checkID), nil)
	if err != nil {
		return nil, err
	}

	var rules []AlertRule
	if err := UnmarshalAPIResponse(data, &rules); err != nil {
		return nil, err
	}

	return rules, nil
}

// GetAlertRule retrieves a specific alert rule
func (c *Client) GetAlertRule(checkID, ruleID string) (*AlertRule, error) {
	data, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/v1/checks/%s/alerts/%s", checkID, ruleID), nil)
	if err != nil {
		return nil, err
	}

	var rule AlertRule
	if err := UnmarshalAPIResponse(data, &rule); err != nil {
		return nil, err
	}

	return &rule, nil
}

// CreateAlertRule creates a new alert rule
func (c *Client) CreateAlertRule(checkID string, req CreateAlertRuleRequest) (*AlertRule, error) {
	data, err := c.DoRequest(http.MethodPost, fmt.Sprintf("/v1/checks/%s/alerts", checkID), req)
	if err != nil {
		return nil, err
	}

	var rule AlertRule
	if err := UnmarshalAPIResponse(data, &rule); err != nil {
		return nil, err
	}

	return &rule, nil
}

// UpdateAlertRule updates an existing alert rule
func (c *Client) UpdateAlertRule(checkID, ruleID string, req UpdateAlertRuleRequest) (*AlertRule, error) {
	data, err := c.DoRequest(http.MethodPut, fmt.Sprintf("/v1/checks/%s/alerts/%s", checkID, ruleID), req)
	if err != nil {
		return nil, err
	}

	var rule AlertRule
	if err := UnmarshalAPIResponse(data, &rule); err != nil {
		return nil, err
	}

	return &rule, nil
}

// DeleteAlertRule deletes an alert rule
func (c *Client) DeleteAlertRule(checkID, ruleID string) error {
	_, err := c.DoRequest(http.MethodDelete, fmt.Sprintf("/v1/checks/%s/alerts/%s", checkID, ruleID), nil)
	return err
}
