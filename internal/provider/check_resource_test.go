package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccCheckResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccCheckResourceConfig_https("test-api", "https://api.example.com/health"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("quismon_check.test", "name", "test-api"),
					resource.TestCheckResourceAttr("quismon_check.test", "type", "https"),
					resource.TestCheckResourceAttr("quismon_check.test", "interval_seconds", "60"),
					resource.TestCheckResourceAttr("quismon_check.test", "enabled", "true"),
					resource.TestCheckResourceAttrSet("quismon_check.test", "id"),
					resource.TestCheckResourceAttrSet("quismon_check.test", "org_id"),
					resource.TestCheckResourceAttrSet("quismon_check.test", "created_at"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "quismon_check.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccCheckResourceConfig_https("test-api-updated", "https://api.example.com/status"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("quismon_check.test", "name", "test-api-updated"),
					resource.TestCheckResourceAttr("quismon_check.test", "config.url", "https://api.example.com/status"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccCheckResource_TCP(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckResourceConfig_tcp("test-db", "localhost", 5432),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("quismon_check.test", "name", "test-db"),
					resource.TestCheckResourceAttr("quismon_check.test", "type", "tcp"),
					resource.TestCheckResourceAttr("quismon_check.test", "config.host", "localhost"),
					resource.TestCheckResourceAttr("quismon_check.test", "config.port", "5432"),
				),
			},
		},
	})
}

func TestAccCheckResource_Ping(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckResourceConfig_ping("test-gateway", "192.168.1.1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("quismon_check.test", "name", "test-gateway"),
					resource.TestCheckResourceAttr("quismon_check.test", "type", "ping"),
					resource.TestCheckResourceAttr("quismon_check.test", "config.host", "192.168.1.1"),
				),
			},
		},
	})
}

func TestAccCheckResource_MultiRegion(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckResourceConfig_multiRegion("global-api"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("quismon_check.test", "name", "global-api"),
					resource.TestCheckResourceAttr("quismon_check.test", "regions.#", "3"),
					resource.TestCheckResourceAttr("quismon_check.test", "regions.0", "us-east-1"),
					resource.TestCheckResourceAttr("quismon_check.test", "regions.1", "eu-west-1"),
					resource.TestCheckResourceAttr("quismon_check.test", "regions.2", "ap-southeast-1"),
				),
			},
		},
	})
}

func testAccCheckResourceConfig_https(name, url string) string {
	return fmt.Sprintf(`
resource "quismon_check" "test" {
  name             = %[1]q
  type             = "https"
  interval_seconds = 60
  enabled          = true

  regions = ["us-east-1"]

  config = {
    url                  = %[2]q
    method               = "GET"
    expected_status_code = "200"
    timeout_seconds      = "10"
  }
}
`, name, url)
}

func testAccCheckResourceConfig_tcp(name, host string, port int) string {
	return fmt.Sprintf(`
resource "quismon_check" "test" {
  name             = %[1]q
  type             = "tcp"
  interval_seconds = 120
  enabled          = true

  regions = ["us-east-1"]

  config = {
    host            = %[2]q
    port            = "%[3]d"
    timeout_seconds = "5"
  }
}
`, name, host, port)
}

func testAccCheckResourceConfig_ping(name, host string) string {
	return fmt.Sprintf(`
resource "quismon_check" "test" {
  name             = %[1]q
  type             = "ping"
  interval_seconds = 300
  enabled          = true

  regions = ["us-east-1"]

  config = {
    host            = %[2]q
    timeout_seconds = "3"
    packet_count    = "4"
  }
}
`, name, host)
}

func testAccCheckResourceConfig_multiRegion(name string) string {
	return fmt.Sprintf(`
resource "quismon_check" "test" {
  name             = %[1]q
  type             = "https"
  interval_seconds = 60
  enabled          = true

  regions = ["us-east-1", "eu-west-1", "ap-southeast-1"]

  config = {
    url                  = "https://api.example.com/health"
    method               = "GET"
    expected_status_code = "200"
    timeout_seconds      = "10"
  }
}
`, name)
}
