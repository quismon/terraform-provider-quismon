package client

import (
	"fmt"
	"net/http"
)

// NotificationChannel represents a notification channel
type NotificationChannel struct {
	ID        string                 `json:"id"`
	OrgID     string                 `json:"org_id"`
	Name      string                 `json:"name"`
	Type      string                 `json:"type"`
	Config    map[string]interface{} `json:"config"`
	Enabled   bool                   `json:"enabled"`
	CreatedAt string                 `json:"created_at"`
	UpdatedAt string                 `json:"updated_at"`
}

// CreateNotificationChannelRequest represents a request to create a channel
type CreateNotificationChannelRequest struct {
	Name    string                 `json:"name"`
	Type    string                 `json:"type"`
	Config  map[string]interface{} `json:"config"`
	Enabled bool                   `json:"enabled"`
}

// UpdateNotificationChannelRequest represents a request to update a channel
type UpdateNotificationChannelRequest struct {
	Name    *string                 `json:"name,omitempty"`
	Config  *map[string]interface{} `json:"config,omitempty"`
	Enabled *bool                   `json:"enabled,omitempty"`
}

// ListNotificationChannels retrieves all notification channels
func (c *Client) ListNotificationChannels() ([]NotificationChannel, error) {
	data, err := c.DoRequest(http.MethodGet, "/v1/notification-channels", nil)
	if err != nil {
		return nil, err
	}

	var channels []NotificationChannel
	if err := UnmarshalAPIResponse(data, &channels); err != nil {
		return nil, err
	}

	return channels, nil
}

// GetNotificationChannel retrieves a specific notification channel
func (c *Client) GetNotificationChannel(id string) (*NotificationChannel, error) {
	data, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/v1/notification-channels/%s", id), nil)
	if err != nil {
		return nil, err
	}

	var channel NotificationChannel
	if err := UnmarshalAPIResponse(data, &channel); err != nil {
		return nil, err
	}

	return &channel, nil
}

// CreateNotificationChannel creates a new notification channel
func (c *Client) CreateNotificationChannel(req CreateNotificationChannelRequest) (*NotificationChannel, error) {
	data, err := c.DoRequest(http.MethodPost, "/v1/notification-channels", req)
	if err != nil {
		return nil, err
	}

	var channel NotificationChannel
	if err := UnmarshalAPIResponse(data, &channel); err != nil {
		return nil, err
	}

	return &channel, nil
}

// UpdateNotificationChannel updates an existing notification channel
func (c *Client) UpdateNotificationChannel(id string, req UpdateNotificationChannelRequest) (*NotificationChannel, error) {
	data, err := c.DoRequest(http.MethodPut, fmt.Sprintf("/v1/notification-channels/%s", id), req)
	if err != nil {
		return nil, err
	}

	var channel NotificationChannel
	if err := UnmarshalAPIResponse(data, &channel); err != nil {
		return nil, err
	}

	return &channel, nil
}

// DeleteNotificationChannel deletes a notification channel
func (c *Client) DeleteNotificationChannel(id string) error {
	_, err := c.DoRequest(http.MethodDelete, fmt.Sprintf("/v1/notification-channels/%s", id), nil)
	return err
}

// GetNotificationChannelByName retrieves a channel by name
func (c *Client) GetNotificationChannelByName(name string) (*NotificationChannel, error) {
	channels, err := c.ListNotificationChannels()
	if err != nil {
		return nil, err
	}

	for _, channel := range channels {
		if channel.Name == name {
			return &channel, nil
		}
	}

	return nil, fmt.Errorf("notification channel not found: %s", name)
}
