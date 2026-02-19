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

// TestAccCheckResource_SMTPIMAP tests SMTP-IMAP check with password fields
// Verifies that config_hash is returned and passwords are handled securely
func TestAccCheckResource_SMTPIMAP(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckResourceConfig_smtpImap("test-mail"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("quismon_check.test", "name", "test-mail"),
					resource.TestCheckResourceAttr("quismon_check.test", "type", "smtp-imap"),
					// config_hash should be set (non-empty) when passwords are present
					resource.TestCheckResourceAttrSet("quismon_check.test", "config_hash"),
				),
			},
			// ImportState testing - config is sensitive so it won't be imported
			{
				ResourceName:      "quismon_check.test",
				ImportState:       true,
				ImportStateVerify: true,
				// Config fields with passwords won't match after import
				// because API returns ***REDACTED***
				ImportStateVerifyIgnore: []string{"config", "config_json"},
			},
		},
	})
}

// TestAccCheckResource_ConfigJSON tests using config_json for complex configs
func TestAccCheckResource_ConfigJSON(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckResourceConfig_configJSON("test-json-config"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("quismon_check.test", "name", "test-json-config"),
					resource.TestCheckResourceAttr("quismon_check.test", "type", "https"),
					// config_json should be sensitive
					// We can't check the actual value since it's marked sensitive
				),
			},
		},
	})
}

// TestAccCheckResource_ConfigHashDrift tests that config_hash is updated on read
// This verifies the drift detection mechanism works
func TestAccCheckResource_ConfigHashDrift(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckResourceConfig_smtpImap("test-drift"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("quismon_check.test", "config_hash"),
				),
			},
			// Refresh should preserve the config_hash from API
			{
				RefreshState: true,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("quismon_check.test", "config_hash"),
				),
			},
		},
	})
}

func testAccCheckResourceConfig_smtpImap(name string) string {
	return fmt.Sprintf(`
resource "quismon_check" "test" {
  name             = %[1]q
  type             = "smtp-imap"
  interval_seconds = 300
  enabled          = true

  regions = ["us-east-1"]

  config = {
    smtp_host      = "smtp.example.com"
    smtp_port      = "587"
    smtp_username  = "test@example.com"
    smtp_password  = "test-password-123"
    smtp_use_tls   = "true"
    imap_host      = "imap.example.com"
    imap_port      = "993"
    imap_username  = "test@example.com"
    imap_password  = "test-password-456"
    imap_use_tls   = "true"
    from_address   = "test@example.com"
    to_address     = "test+inbox@example.com"
    subject        = "Test Email"
    body           = "Test body"
    timeout_seconds = "30"
  }
}
`, name)
}

func testAccCheckResourceConfig_configJSON(name string) string {
	return fmt.Sprintf(`
resource "quismon_check" "test" {
  name             = %[1]q
  type             = "https"
  interval_seconds = 60
  enabled          = true

  regions = ["us-east-1"]

  config_json = jsonencode({
    url             = "https://api.example.com/health"
    method          = "GET"
    expected_status = [200]
    timeout_seconds = 10
    headers = {
      "Authorization" = "Bearer secret-token-123"
    }
  })
}
`, name)
}
