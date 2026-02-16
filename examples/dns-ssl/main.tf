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

variable "slack_webhook_url" {
  type      = string
  sensitive = true
}

# Notification channel
resource "quismon_notification_channel" "slack" {
  name = "Slack #alerts"
  type = "slack"

  config = {
    webhook_url = var.slack_webhook_url
  }

  enabled = true
}

# DNS A Record Check
resource "quismon_check" "dns_a_record" {
  name             = "Main Domain A Record"
  type             = "dns"
  interval_seconds = 300
  enabled          = true

  regions = ["us-east-1"]

  config = {
    domain       = "example.com"
    record_type  = "A"
    expected_ips = jsonencode(["93.184.216.34"])
  }
}

# DNS MX Record Check
resource "quismon_check" "dns_mx_record" {
  name             = "Mail Exchange Records"
  type             = "dns"
  interval_seconds = 300
  enabled          = true

  regions = ["us-east-1"]

  config = {
    domain      = "example.com"
    record_type = "MX"
  }
}

# SSL Certificate Check
resource "quismon_check" "ssl_cert" {
  name             = "API Certificate Expiry"
  type             = "ssl"
  interval_seconds = 3600  # Check every hour
  enabled          = true

  regions = ["us-east-1"]

  config = {
    domain              = "api.example.com"
    port                = 443
    warn_days_remaining = 30
  }
}

# SSL Certificate with Fingerprint Validation
resource "quismon_check" "ssl_cert_fingerprint" {
  name             = "Critical API Certificate"
  type             = "ssl"
  interval_seconds = 3600
  enabled          = true

  regions = ["us-east-1"]

  config = {
    domain                    = "critical.example.com"
    port                      = 443
    warn_days_remaining       = 30
    expected_fingerprint_sha256 = "abc123def456..."  # Actual SHA-256 fingerprint
  }
}

# SSL Certificate with SAN Domain Validation
resource "quismon_check" "ssl_cert_san" {
  name             = "Multi-Domain Certificate"
  type             = "ssl"
  interval_seconds = 3600
  enabled          = true

  regions = ["us-east-1"]

  config = {
    domain              = "secure.example.com"
    port                = 443
    warn_days_remaining = 30
    expected_domains    = jsonencode([
      "secure.example.com",
      "www.secure.example.com",
      "api.secure.example.com"
    ])
  }
}

# Alert on DNS failure
resource "quismon_alert_rule" "dns_failure" {
  check_id = quismon_check.dns_a_record.id
  name     = "DNS Resolution Failed"
  enabled  = true

  condition = {
    health_status = "down"
  }

  notification_channel_ids = [
    quismon_notification_channel.slack.id
  ]
}

# Alert on SSL certificate expiry
resource "quismon_alert_rule" "ssl_expiring" {
  check_id = quismon_check.ssl_cert.id
  name     = "SSL Certificate Expiring Soon"
  enabled  = true

  condition = {
    health_status = "down"
  }

  notification_channel_ids = [
    quismon_notification_channel.slack.id
  ]
}

output "dns_check_id" {
  value = quismon_check.dns_a_record.id
}

output "ssl_check_id" {
  value = quismon_check.ssl_cert.id
}

output "ssl_cert_health" {
  value = quismon_check.ssl_cert.health_status
}
