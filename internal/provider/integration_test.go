package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccCompleteStack tests a complete monitoring stack with all resource types
func TestAccCompleteStack(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCompleteStackConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify checks
					resource.TestCheckResourceAttr("quismon_check.api", "name", "Production API"),
					resource.TestCheckResourceAttr("quismon_check.database", "name", "Database"),
					resource.TestCheckResourceAttr("quismon_check.gateway", "name", "Gateway"),

					// Verify channels
					resource.TestCheckResourceAttr("quismon_notification_channel.email", "name", "Ops Team"),
					resource.TestCheckResourceAttr("quismon_notification_channel.webhook", "name", "Slack"),
					resource.TestCheckResourceAttr("quismon_notification_channel.ntfy", "name", "Mobile"),

					// Verify alert rules
					resource.TestCheckResourceAttr("quismon_alert_rule.api_down", "name", "API Down"),
					resource.TestCheckResourceAttr("quismon_alert_rule.api_slow", "name", "API Slow"),
					resource.TestCheckResourceAttr("quismon_alert_rule.db_down", "name", "DB Down"),

					// Verify data sources work
					resource.TestCheckResourceAttrSet("data.quismon_checks.all", "checks.#"),
				),
			},
		},
	})
}

// TestAccCheckTypeMatrix tests all check types with various configurations
func TestAccCheckTypeMatrix(t *testing.T) {
	testCases := []struct {
		name   string
		config string
	}{
		{
			name:   "http_basic",
			config: testAccCheckTypeMatrixConfig_http(),
		},
		{
			name:   "https_with_headers",
			config: testAccCheckTypeMatrixConfig_https_headers(),
		},
		{
			name:   "tcp_basic",
			config: testAccCheckTypeMatrixConfig_tcp(),
		},
		{
			name:   "ping_basic",
			config: testAccCheckTypeMatrixConfig_ping(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resource.Test(t, resource.TestCase{
				PreCheck:                 func() { testAccPreCheck(t) },
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: tc.config,
						Check: resource.ComposeAggregateTestCheckFunc(
							resource.TestCheckResourceAttrSet("quismon_check.test", "id"),
							resource.TestCheckResourceAttrSet("quismon_check.test", "health_status"),
						),
					},
				},
			})
		})
	}
}

// TestAccAlertConditionMatrix tests all alert condition types
func TestAccAlertConditionMatrix(t *testing.T) {
	testCases := []struct {
		name          string
		conditionType string
		threshold     float64
	}{
		{"consecutive_failures", "consecutive_failures", 3},
		{"response_time", "response_time", 5000},
		{"status_code", "status_code", 500},
		{"ssl_expiry", "ssl_expiry", 7},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resource.Test(t, resource.TestCase{
				PreCheck:                 func() { testAccPreCheck(t) },
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: testAccAlertConditionMatrixConfig(tc.conditionType, tc.threshold),
						Check: resource.ComposeAggregateTestCheckFunc(
							resource.TestCheckResourceAttr("quismon_alert_rule.test", "condition_type", tc.conditionType),
						),
					},
				},
			})
		})
	}
}

// TestAccNotificationChannelMatrix tests all channel types
func TestAccNotificationChannelMatrix(t *testing.T) {
	testCases := []struct {
		name        string
		channelType string
		config      string
	}{
		{"email", "email", testAccNotificationChannelMatrixConfig_email()},
		{"webhook", "webhook", testAccNotificationChannelMatrixConfig_webhook()},
		{"ntfy", "ntfy", testAccNotificationChannelMatrixConfig_ntfy()},
		{"slack", "slack", testAccNotificationChannelMatrixConfig_slack()},
		{"pagerduty", "pagerduty", testAccNotificationChannelMatrixConfig_pagerduty()},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resource.Test(t, resource.TestCase{
				PreCheck:                 func() { testAccPreCheck(t) },
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: tc.config,
						Check: resource.ComposeAggregateTestCheckFunc(
							resource.TestCheckResourceAttr("quismon_notification_channel.test", "type", tc.channelType),
						),
					},
				},
			})
		})
	}
}

func testAccCompleteStackConfig() string {
	return `
# Notification Channels
resource "quismon_notification_channel" "email" {
  name    = "Ops Team"
  type    = "email"
  enabled = true
  config = {
    to = jsonencode(["ops@example.com"])
  }
}

resource "quismon_notification_channel" "webhook" {
  name    = "Slack"
  type    = "webhook"
  enabled = true
  config = {
    url    = "https://hooks.slack.com/services/test"
    method = "POST"
  }
}

resource "quismon_notification_channel" "ntfy" {
  name    = "Mobile"
  type    = "ntfy"
  enabled = true
  config = {
    topic  = "quismon-test"
    server = "https://ntfy.sh"
  }
}

# Checks
resource "quismon_check" "api" {
  name             = "Production API"
  type             = "https"
  interval_seconds = 60
  enabled          = true
  regions          = ["us-east-1", "eu-west-1"]

  config = {
    url                  = "https://api.example.com/health"
    method               = "GET"
    expected_status_code = "200"
    timeout_seconds      = "10"
  }
}

resource "quismon_check" "database" {
  name             = "Database"
  type             = "tcp"
  interval_seconds = 120
  enabled          = true
  regions          = ["us-east-1"]

  config = {
    host            = "db.example.com"
    port            = "5432"
    timeout_seconds = "5"
  }
}

resource "quismon_check" "gateway" {
  name             = "Gateway"
  type             = "ping"
  interval_seconds = 300
  enabled          = true
  regions          = ["us-east-1"]

  config = {
    host            = "192.168.1.1"
    timeout_seconds = "3"
    packet_count    = "4"
  }
}

# Alert Rules
resource "quismon_alert_rule" "api_down" {
  check_id       = quismon_check.api.id
  name           = "API Down"
  condition_type = "consecutive_failures"
  threshold      = 3
  enabled        = true

  notification_channel_ids = [
    quismon_notification_channel.email.id,
    quismon_notification_channel.webhook.id,
    quismon_notification_channel.ntfy.id
  ]
}

resource "quismon_alert_rule" "api_slow" {
  check_id       = quismon_check.api.id
  name           = "API Slow"
  condition_type = "response_time"
  threshold      = 5000
  enabled        = true

  notification_channel_ids = [
    quismon_notification_channel.webhook.id
  ]
}

resource "quismon_alert_rule" "db_down" {
  check_id       = quismon_check.database.id
  name           = "DB Down"
  condition_type = "consecutive_failures"
  threshold      = 2
  enabled        = true

  notification_channel_ids = [
    quismon_notification_channel.email.id,
    quismon_notification_channel.ntfy.id
  ]
}

# Data Sources
data "quismon_checks" "all" {
  depends_on = [
    quismon_check.api,
    quismon_check.database,
    quismon_check.gateway
  ]
}

data "quismon_check" "api_lookup" {
  name = "Production API"
  depends_on = [quismon_check.api]
}
`
}

func testAccCheckTypeMatrixConfig_http() string {
	return `
resource "quismon_check" "test" {
  name             = "HTTP Check"
  type             = "http"
  interval_seconds = 60
  enabled          = true
  regions          = ["us-east-1"]

  config = {
    url                  = "http://example.com"
    method               = "GET"
    expected_status_code = "200"
    timeout_seconds      = "10"
  }
}
`
}

func testAccCheckTypeMatrixConfig_https_headers() string {
	return `
resource "quismon_check" "test" {
  name             = "HTTPS with Headers"
  type             = "https"
  interval_seconds = 60
  enabled          = true
  regions          = ["us-east-1"]

  config = {
    url                  = "https://api.example.com"
    method               = "GET"
    expected_status_code = "200"
    timeout_seconds      = "15"
    headers              = jsonencode({
      "User-Agent"    = "Quismon/1.0"
      "Authorization" = "Bearer test-token"
    })
  }
}
`
}

func testAccCheckTypeMatrixConfig_tcp() string {
	return `
resource "quismon_check" "test" {
  name             = "TCP Check"
  type             = "tcp"
  interval_seconds = 90
  enabled          = true
  regions          = ["us-east-1"]

  config = {
    host            = "localhost"
    port            = "3306"
    timeout_seconds = "5"
  }
}
`
}

func testAccCheckTypeMatrixConfig_ping() string {
	return `
resource "quismon_check" "test" {
  name             = "Ping Check"
  type             = "ping"
  interval_seconds = 180
  enabled          = true
  regions          = ["us-east-1"]

  config = {
    host            = "8.8.8.8"
    timeout_seconds = "2"
    packet_count    = "3"
  }
}
`
}

func testAccAlertConditionMatrixConfig(conditionType string, threshold float64) string {
	return `
resource "quismon_check" "test" {
  name             = "Alert Matrix Check"
  type             = "https"
  interval_seconds = 60
  enabled          = true
  regions          = ["us-east-1"]

  config = {
    url                  = "https://api.example.com"
    method               = "GET"
    expected_status_code = "200"
    timeout_seconds      = "10"
  }
}

resource "quismon_notification_channel" "test" {
  name    = "Test Channel"
  type    = "email"
  enabled = true

  config = {
    to = jsonencode(["test@example.com"])
  }
}

resource "quismon_alert_rule" "test" {
  check_id       = quismon_check.test.id
  name           = "` + conditionType + ` Alert"
  condition_type = "` + conditionType + `"
  threshold      = ` + fmt.Sprintf("%.0f", threshold) + `
  enabled        = true

  notification_channel_ids = [
    quismon_notification_channel.test.id
  ]
}
`
}

func testAccNotificationChannelMatrixConfig_email() string {
	return `
resource "quismon_notification_channel" "test" {
  name    = "Email Channel"
  type    = "email"
  enabled = true
  config = {
    to = jsonencode(["test@example.com", "ops@example.com"])
  }
}
`
}

func testAccNotificationChannelMatrixConfig_webhook() string {
	return `
resource "quismon_notification_channel" "test" {
  name    = "Webhook Channel"
  type    = "webhook"
  enabled = true
  config = {
    url    = "https://hooks.example.com/webhook"
    method = "POST"
  }
}
`
}

func testAccNotificationChannelMatrixConfig_ntfy() string {
	return `
resource "quismon_notification_channel" "test" {
  name    = "Ntfy Channel"
  type    = "ntfy"
  enabled = true
  config = {
    topic  = "quismon-matrix-test"
    server = "https://ntfy.sh"
  }
}
`
}

func testAccNotificationChannelMatrixConfig_slack() string {
	return `
resource "quismon_notification_channel" "test" {
  name    = "Slack Channel"
  type    = "slack"
  enabled = true
  config = {
    webhook_url = "https://hooks.slack.com/services/TEST/WEBHOOK/URL"
  }
}
`
}

func testAccNotificationChannelMatrixConfig_pagerduty() string {
	return `
resource "quismon_notification_channel" "test" {
  name    = "PagerDuty Channel"
  type    = "pagerduty"
  enabled = true
  config = {
    routing_key = "test-routing-key-12345"
  }
}
`
}
