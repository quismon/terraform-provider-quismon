package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccNotificationChannelResource_Email(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccNotificationChannelConfig_email("test-email", "ops@example.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("quismon_notification_channel.test", "name", "test-email"),
					resource.TestCheckResourceAttr("quismon_notification_channel.test", "type", "email"),
					resource.TestCheckResourceAttr("quismon_notification_channel.test", "enabled", "true"),
					resource.TestCheckResourceAttrSet("quismon_notification_channel.test", "id"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "quismon_notification_channel.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccNotificationChannelConfig_email("test-email-updated", "alerts@example.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("quismon_notification_channel.test", "name", "test-email-updated"),
				),
			},
		},
	})
}

func TestAccNotificationChannelResource_Webhook(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationChannelConfig_webhook("test-webhook", "https://hooks.example.com/alerts"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("quismon_notification_channel.test", "name", "test-webhook"),
					resource.TestCheckResourceAttr("quismon_notification_channel.test", "type", "webhook"),
					resource.TestCheckResourceAttr("quismon_notification_channel.test", "config.url", "https://hooks.example.com/alerts"),
				),
			},
		},
	})
}

func TestAccNotificationChannelResource_Ntfy(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationChannelConfig_ntfy("test-ntfy", "quismon-test"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("quismon_notification_channel.test", "name", "test-ntfy"),
					resource.TestCheckResourceAttr("quismon_notification_channel.test", "type", "ntfy"),
					resource.TestCheckResourceAttr("quismon_notification_channel.test", "config.topic", "quismon-test"),
				),
			},
		},
	})
}

func testAccNotificationChannelConfig_email(name, email string) string {
	return fmt.Sprintf(`
resource "quismon_notification_channel" "test" {
  name    = %[1]q
  type    = "email"
  enabled = true

  config = {
    to = jsonencode([%[2]q])
  }
}
`, name, email)
}

func testAccNotificationChannelConfig_webhook(name, url string) string {
	return fmt.Sprintf(`
resource "quismon_notification_channel" "test" {
  name    = %[1]q
  type    = "webhook"
  enabled = true

  config = {
    url    = %[2]q
    method = "POST"
  }
}
`, name, url)
}

func testAccNotificationChannelConfig_ntfy(name, topic string) string {
	return fmt.Sprintf(`
resource "quismon_notification_channel" "test" {
  name    = %[1]q
  type    = "ntfy"
  enabled = true

  config = {
    topic  = %[2]q
    server = "https://ntfy.sh"
  }
}
`, name, topic)
}
