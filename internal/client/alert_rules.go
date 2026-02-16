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

// DeleteAlertRule deletes an alert rule
func (c *Client) DeleteAlertRule(checkID, ruleID string) error {
	_, err := c.DoRequest(http.MethodDelete, fmt.Sprintf("/v1/checks/%s/alerts/%s", checkID, ruleID), nil)
	return err
}
