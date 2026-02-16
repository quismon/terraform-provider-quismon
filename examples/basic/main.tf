terraform {
  required_providers {
    quismon = {
      source = "quismon/quismon"
    }
  }
}

provider "quismon" {
  # api_key  = var.quismon_api_key  # Or use QUISMON_API_KEY env var
  # base_url = "https://app.quismon.com"  # Optional
}

# Create an email notification channel
resource "quismon_notification_channel" "email" {
  name = "Team Alerts"
  type = "email"

  config = {
    to = jsonencode(["ops@example.com"])
  }

  enabled = true
}

# Create a simple HTTPS check
resource "quismon_check" "website" {
  name             = "Company Website"
  type             = "https"
  interval_seconds = 60
  enabled          = true

  regions = ["us-east-1"]

  config = {
    url                  = "https://www.example.com"
    method               = "GET"
    expected_status_code = "200"
    timeout_seconds      = "10"
  }
}

# Alert on 3 consecutive failures
resource "quismon_alert_rule" "website_down" {
  check_id = quismon_check.website.id
  name     = "Website Down"
  enabled  = true

  condition = {
    failure_threshold = 3
  }

  notification_channel_ids = [
    quismon_notification_channel.email.id
  ]
}

# Output the check ID
output "check_id" {
  value       = quismon_check.website.id
  description = "The ID of the created check"
}

output "health_status" {
  value       = quismon_check.website.health_status
  description = "Current health status of the website"
}
