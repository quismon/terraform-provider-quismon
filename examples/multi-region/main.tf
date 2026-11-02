terraform {
  required_providers {
    quismon = {
      source = "quismon/quismon"
    }
  }
}

provider "quismon" {
  api_key = var.quismon_api_key
}

variable "quismon_api_key" {
  type      = string
  sensitive = true
}

# Notification channel
resource "quismon_notification_channel" "webhook" {
  name = "Incident Webhook"
  type = "webhook"

  config = {
    url    = var.incident_webhook_url
    method = "POST"
    secret = var.incident_webhook_secret
  }
}

variable "incident_webhook_url" {
  type      = string
  sensitive = true
}

variable "incident_webhook_secret" {
  type      = string
  sensitive = true
}

# Multi-region API monitoring
resource "quismon_check" "api_multi_region" {
  name             = "Global API"
  type             = "https"
  interval_seconds = 60
  enabled          = true

  # Monitor from multiple regions
  regions = [
    "na-east-ewr",
    "na-west-lax",
    "eu-west-ams",
    "eu-central-fra",
    "ap-southeast-sin"
  ]

  config = {
    url                  = "https://api.example.com/health"
    method               = "GET"
    expected_status_code = "200"
    timeout_seconds      = "15"
  }
}

# Alert if API is down in ANY region
resource "quismon_alert_rule" "api_regional_failure" {
  check_id = quismon_check.api_multi_region.id
  name     = "API Regional Failure"
  enabled  = true

  condition = {
    failure_threshold = 2
  }

  notification_channel_ids = [
    quismon_notification_channel.webhook.id
  ]
}

# Alert on slow global response times (response_time in milliseconds)
resource "quismon_alert_rule" "api_global_latency" {
  check_id = quismon_check.api_multi_region.id
  name     = "High Global Latency"
  enabled  = true

  condition = {
    response_time_ms = 3000 # 3 seconds
  }

  notification_channel_ids = [
    quismon_notification_channel.webhook.id
  ]
}

output "monitored_regions" {
  value = quismon_check.api_multi_region.regions
}
