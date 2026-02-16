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

# Variables
variable "quismon_api_key" {
  description = "Quismon API key"
  type        = string
  sensitive   = true
}

variable "slack_webhook_url" {
  description = "Slack webhook URL"
  type        = string
  sensitive   = true
}

# Notification Channels
resource "quismon_notification_channel" "ops_email" {
  name = "Operations Team"
  type = "email"

  config = {
    to = jsonencode(["ops@example.com", "oncall@example.com"])
  }

  enabled = true
}

resource "quismon_notification_channel" "slack" {
  name = "Slack #alerts"
  type = "slack"

  config = {
    webhook_url = var.slack_webhook_url
  }

  enabled = true
}

resource "quismon_notification_channel" "ntfy" {
  name = "Mobile Push"
  type = "ntfy"

  config = {
    topic  = "quismon-prod-alerts"
    server = "https://ntfy.sh"
  }

  enabled = true
}

# Production API Check
resource "quismon_check" "api" {
  name             = "Production API"
  type             = "https"
  interval_seconds = 60
  enabled          = true

  regions = ["us-east-1", "eu-west-1"]

  config = {
    url                  = "https://api.example.com/health"
    method               = "GET"
    expected_status_code = "200"
    timeout_seconds      = "10"
    headers              = jsonencode({
      "User-Agent" = "Quismon-Monitor/1.0"
    })
  }
}

# API Down Alert - triggers when health status becomes "down"
resource "quismon_alert_rule" "api_down" {
  check_id = quismon_check.api.id
  name     = "Production API Down"
  enabled  = true

  condition = {
    health_status = "down"
  }

  notification_channel_ids = [
    quismon_notification_channel.ops_email.id,
    quismon_notification_channel.slack.id,
    quismon_notification_channel.ntfy.id
  ]
}

# API Consecutive Failures Alert - triggers after 3 failures in a row
resource "quismon_alert_rule" "api_consecutive_failures" {
  check_id = quismon_check.api.id
  name     = "API Consecutive Failures"
  enabled  = true

  condition = {
    failure_threshold = 3
  }

  notification_channel_ids = [
    quismon_notification_channel.slack.id,
    quismon_notification_channel.ntfy.id
  ]
}

# Database TCP Check
resource "quismon_check" "database" {
  name             = "PostgreSQL Database"
  type             = "tcp"
  interval_seconds = 120
  enabled          = true

  regions = ["us-east-1"]

  config = {
    host            = "db.example.com"
    port            = "5432"
    timeout_seconds = "5"
  }
}

# Database Down Alert - triggers after 2 consecutive failures
resource "quismon_alert_rule" "db_down" {
  check_id = quismon_check.database.id
  name     = "Database Unreachable"
  enabled  = true

  condition = {
    failure_threshold = 2
  }

  notification_channel_ids = [
    quismon_notification_channel.ops_email.id,
    quismon_notification_channel.ntfy.id
  ]
}

# Gateway Ping Check
resource "quismon_check" "gateway" {
  name             = "Network Gateway"
  type             = "ping"
  interval_seconds = 300
  enabled          = true

  regions = ["us-east-1"]

  config = {
    host            = "192.168.1.1"
    timeout_seconds = "3"
    packet_count    = "4"
  }
}

# Outputs
output "api_check_id" {
  value = quismon_check.api.id
}

output "api_health_status" {
  value = quismon_check.api.health_status
}

output "all_checks" {
  value = {
    api      = quismon_check.api.id
    database = quismon_check.database.id
    gateway  = quismon_check.gateway.id
  }
}
