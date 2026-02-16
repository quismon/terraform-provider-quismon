package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccCheckDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create a check, then query it via data source
			{
				Config: testAccCheckDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Check resource assertions
					resource.TestCheckResourceAttr("quismon_check.test", "name", "test-datasource-check"),
					// Data source assertions
					resource.TestCheckResourceAttr("data.quismon_check.test", "name", "test-datasource-check"),
					resource.TestCheckResourceAttr("data.quismon_check.test", "type", "https"),
					resource.TestCheckResourceAttrSet("data.quismon_check.test", "id"),
					resource.TestCheckResourceAttrSet("data.quismon_check.test", "health_status"),
				),
			},
		},
	})
}

func TestAccChecksDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create multiple checks, then query all via data source
			{
				Config: testAccChecksDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify we have at least 2 checks
					resource.TestCheckResourceAttr("data.quismon_checks.all", "checks.#", "2"),
				),
			},
		},
	})
}

func TestAccNotificationChannelDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationChannelDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.quismon_notification_channel.test", "name", "test-datasource-channel"),
					resource.TestCheckResourceAttr("data.quismon_notification_channel.test", "type", "email"),
					resource.TestCheckResourceAttrSet("data.quismon_notification_channel.test", "id"),
				),
			},
		},
	})
}

func testAccCheckDataSourceConfig() string {
	return `
resource "quismon_check" "test" {
  name             = "test-datasource-check"
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

data "quismon_check" "test" {
  name = quismon_check.test.name
}
`
}

func testAccChecksDataSourceConfig() string {
	return `
resource "quismon_check" "test1" {
  name             = "test-datasource-check-1"
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

resource "quismon_check" "test2" {
  name             = "test-datasource-check-2"
  type             = "tcp"
  interval_seconds = 120
  enabled          = true

  regions = ["us-east-1"]

  config = {
    host            = "localhost"
    port            = "5432"
    timeout_seconds = "5"
  }
}

data "quismon_checks" "all" {
  depends_on = [
    quismon_check.test1,
    quismon_check.test2
  ]
}
`
}

func testAccNotificationChannelDataSourceConfig() string {
	return `
resource "quismon_notification_channel" "test" {
  name    = "test-datasource-channel"
  type    = "email"
  enabled = true

  config = {
    to = jsonencode(["test@example.com"])
  }
}

data "quismon_notification_channel" "test" {
  name = quismon_notification_channel.test.name
}
`
}
