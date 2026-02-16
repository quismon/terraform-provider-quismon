package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAlertRuleResource_ConsecutiveFailures(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create check and channel first, then alert rule
			{
				Config: testAccAlertRuleConfig_consecutiveFailures("test-check", "test-channel", "test-alert", 3),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Check assertions
					resource.TestCheckResourceAttr("quismon_check.test", "name", "test-check"),
					// Channel assertions
					resource.TestCheckResourceAttr("quismon_notification_channel.test", "name", "test-channel"),
					// Alert rule assertions
					resource.TestCheckResourceAttr("quismon_alert_rule.test", "name", "test-alert"),
					resource.TestCheckResourceAttr("quismon_alert_rule.test", "condition_type", "consecutive_failures"),
					resource.TestCheckResourceAttr("quismon_alert_rule.test", "threshold", "3"),
					resource.TestCheckResourceAttrSet("quismon_alert_rule.test", "id"),
					resource.TestCheckResourceAttrSet("quismon_alert_rule.test", "check_id"),
				),
			},
		},
	})
}

func TestAccAlertRuleResource_ResponseTime(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAlertRuleConfig_responseTime("test-check-rt", "test-channel-rt", "test-alert-rt", 5000),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("quismon_alert_rule.test", "name", "test-alert-rt"),
					resource.TestCheckResourceAttr("quismon_alert_rule.test", "condition_type", "response_time"),
					resource.TestCheckResourceAttr("quismon_alert_rule.test", "threshold", "5000"),
				),
			},
		},
	})
}

func TestAccAlertRuleResource_MultipleChannels(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAlertRuleConfig_multipleChannels(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("quismon_alert_rule.test", "notification_channel_ids.#", "2"),
				),
			},
		},
	})
}

func testAccAlertRuleConfig_consecutiveFailures(checkName, channelName, alertName string, threshold int) string {
	return fmt.Sprintf(`
resource "quismon_check" "test" {
  name             = %[1]q
  type             = "https"
  interval_seconds = 60
  enabled          = true

  regions = ["us-east-1"]

  config = {
    url                  = "https://api.example.com/health"
    method               = "GET"
    expected_status_code = "200"
    timeout_seconds      = "10"
  }
}

resource "quismon_notification_channel" "test" {
  name    = %[2]q
  type    = "email"
  enabled = true

  config = {
    to = jsonencode(["test@example.com"])
  }
}

resource "quismon_alert_rule" "test" {
  check_id       = quismon_check.test.id
  name           = %[3]q
  condition_type = "consecutive_failures"
  threshold      = %[4]d
  enabled        = true

  notification_channel_ids = [
    quismon_notification_channel.test.id
  ]
}
`, checkName, channelName, alertName, threshold)
}

func testAccAlertRuleConfig_responseTime(checkName, channelName, alertName string, threshold int) string {
	return fmt.Sprintf(`
resource "quismon_check" "test" {
  name             = %[1]q
  type             = "https"
  interval_seconds = 60
  enabled          = true

  regions = ["us-east-1"]

  config = {
    url                  = "https://api.example.com/health"
    method               = "GET"
    expected_status_code = "200"
    timeout_seconds      = "10"
  }
}

resource "quismon_notification_channel" "test" {
  name    = %[2]q
  type    = "email"
  enabled = true

  config = {
    to = jsonencode(["test@example.com"])
  }
}

resource "quismon_alert_rule" "test" {
  check_id       = quismon_check.test.id
  name           = %[3]q
  condition_type = "response_time"
  threshold      = %[4]d
  enabled        = true

  notification_channel_ids = [
    quismon_notification_channel.test.id
  ]
}
`, checkName, channelName, alertName, threshold)
}

func testAccAlertRuleConfig_multipleChannels() string {
	return `
resource "quismon_check" "test" {
  name             = "test-check-multi"
  type             = "https"
  interval_seconds = 60
  enabled          = true

  regions = ["us-east-1"]

  config = {
    url                  = "https://api.example.com/health"
    method               = "GET"
    expected_status_code = "200"
    timeout_seconds      = "10"
  }
}

resource "quismon_notification_channel" "email" {
  name    = "test-email"
  type    = "email"
  enabled = true

  config = {
    to = jsonencode(["ops@example.com"])
  }
}

resource "quismon_notification_channel" "webhook" {
  name    = "test-webhook"
  type    = "webhook"
  enabled = true

  config = {
    url    = "https://hooks.example.com/alerts"
    method = "POST"
  }
}

resource "quismon_alert_rule" "test" {
  check_id       = quismon_check.test.id
  name           = "test-multi-channel-alert"
  condition_type = "consecutive_failures"
  threshold      = 3
  enabled        = true

  notification_channel_ids = [
    quismon_notification_channel.email.id,
    quismon_notification_channel.webhook.id
  ]
}
`
}
