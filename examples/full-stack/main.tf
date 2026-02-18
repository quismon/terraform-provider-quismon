# Full-Stack Monitoring Example
# This example demonstrates monitoring a complete web application stack
# from DNS to SSL to HTTP endpoints to API functionality.

# Features demonstrated:
# - DNS change detection (alert when DNS records change)
# - SSL certificate monitoring (alert before expiry)
# - HTTP health checks with content validation
# - TCP port checks for databases
# - Inverted checks for security monitoring
# - Multi-step checks for user flows
# - Alert rules with multiple notification channels

terraform {
  required_providers {
    quismon = {
      source = "quismon/quismon"
    }
  }
}

provider "quismon" {
  # Configure via QUISMON_API_KEY environment variable
}

# =============================================================================
# Variables
# =============================================================================

variable "domain" {
  description = "Your application domain"
  type        = string
  default     = "example.com"
}

variable "api_endpoint" {
  description = "Your API endpoint"
  type        = string
  default     = "https://api.example.com"
}

variable "database_host" {
  description = "Database host for TCP check"
  type        = string
  default     = "db.example.com"
}

variable "slack_webhook_url" {
  description = "Slack webhook URL for alerts"
  type        = string
  sensitive   = true
}

variable "oncall_email" {
  description = "On-call team email"
  type        = string
  default     = "oncall@example.com"
}

# =============================================================================
# Notification Channels
# =============================================================================

resource "quismon_notification_channel" "slack" {
  name = "Slack #alerts"
  type = "slack"

  config = {
    webhook_url = var.slack_webhook_url
  }

  enabled = true
}

resource "quismon_notification_channel" "email" {
  name = "On-Call Team"
  type = "email"

  config = {
    to = jsonencode([var.oncall_email])
  }

  enabled = true
}

resource "quismon_notification_channel" "webhook" {
  name = "PagerDuty"
  type = "webhook"

  config = {
    url = "https://events.pagerduty.com/integration/YOUR_KEY/enqueue"
  }

  enabled = true
}

# =============================================================================
# Layer 1: DNS Monitoring
# =============================================================================

# Monitor DNS A records - alerts if they change unexpectedly
resource "quismon_check" "dns_a" {
  name             = "${var.domain} DNS A Record"
  type             = "dns"
  interval_seconds = 300
  enabled          = true

  regions = ["na-east-ewr", "eu-west-ams"]

  config_json = jsonencode({
    domain       = var.domain
    record_type  = "A"
    timeout_seconds = 5
  })
}

# Alert when DNS records change (potential hijacking)
resource "quismon_alert_rule" "dns_change" {
  check_id = quismon_check.dns_a.id
  name     = "DNS Record Changed"
  enabled  = true

  condition = {
    health_status = "down"
  }

  notification_channel_ids = [
    quismon_notification_channel.slack.id,
    quismon_notification_channel.email.id
  ]
}

# =============================================================================
# Layer 2: SSL Certificate Monitoring
# =============================================================================

# Monitor SSL certificate expiry
resource "quismon_check" "ssl_cert" {
  name             = "${var.domain} SSL Certificate"
  type             = "ssl"
  interval_seconds = 3600  # Check hourly
  enabled          = true

  regions = ["na-east-ewr"]

  config_json = jsonencode({
    hostname            = var.domain
    port                = 443
    warn_days_remaining = 30  # Warn 30 days before expiry
  })
}

# Alert when SSL certificate is about to expire
resource "quismon_alert_rule" "ssl_expiry" {
  check_id = quismon_check.ssl_cert.id
  name     = "SSL Certificate Expiring Soon"
  enabled  = true

  condition = {
    health_status = "down"
  }

  notification_channel_ids = [
    quismon_notification_channel.email.id
  ]
}

# =============================================================================
# Layer 3: Website Health Check
# =============================================================================

# Monitor main website
resource "quismon_check" "website" {
  name             = "${var.domain} Website"
  type             = "https"
  interval_seconds = 60
  enabled          = true

  regions = ["na-east-ewr", "eu-west-ams", "ap-southeast-sin"]

  config_json = jsonencode({
    url               = "https://${var.domain}"
    method            = "GET"
    expected_status   = [200]
    expected_content  = "Welcome"
    content_match_type = "contains"
    timeout_seconds   = 10
  })
}

# Alert immediately when website is down
resource "quismon_alert_rule" "website_down" {
  check_id = quismon_check.website.id
  name     = "Website Down"
  enabled  = true

  condition = {
    health_status = "down"
  }

  notification_channel_ids = [
    quismon_notification_channel.slack.id,
    quismon_notification_channel.webhook.id
  ]
}

# Alert when response time is slow (> 3 seconds)
resource "quismon_alert_rule" "website_slow" {
  check_id = quismon_check.website.id
  name     = "Website Slow Response"
  enabled  = true

  condition = {
    response_time_ms = 3000
  }

  notification_channel_ids = [
    quismon_notification_channel.slack.id
  ]
}

# =============================================================================
# Layer 4: API Health Check
# =============================================================================

# Monitor API health endpoint
resource "quismon_check" "api_health" {
  name             = "API Health Check"
  type             = "https"
  interval_seconds = 30
  enabled          = true

  regions = ["na-east-ewr", "eu-west-ams"]

  config_json = jsonencode({
    url              = "${var.api_endpoint}/health"
    method           = "GET"
    expected_status  = [200]
    expected_content = "\"status\":\"ok\""
    content_match_type = "contains"
    timeout_seconds  = 5
  })
}

# Alert after 2 consecutive API failures
resource "quismon_alert_rule" "api_down" {
  check_id = quismon_check.api_health.id
  name     = "API Down"
  enabled  = true

  condition = {
    failure_threshold = 2
  }

  notification_channel_ids = [
    quismon_notification_channel.slack.id,
    quismon_notification_channel.email.id,
    quismon_notification_channel.webhook.id
  ]
}

# =============================================================================
# Layer 5: Database Connectivity
# =============================================================================

# Monitor PostgreSQL database port
resource "quismon_check" "database" {
  name             = "PostgreSQL Database"
  type             = "tcp"
  interval_seconds = 60
  enabled          = true

  regions = ["na-east-ewr"]

  config_json = jsonencode({
    host            = var.database_host
    port            = 5432
    timeout_seconds = 5
  })
}

# Alert when database is unreachable
resource "quismon_alert_rule" "database_down" {
  check_id = quismon_check.database.id
  name     = "Database Unreachable"
  enabled  = true

  condition = {
    health_status = "down"
  }

  notification_channel_ids = [
    quismon_notification_channel.slack.id,
    quismon_notification_channel.webhook.id
  ]
}

# =============================================================================
# Layer 6: Security (Inverted Checks)
# =============================================================================

# Inverted check: Alert if admin panel becomes publicly accessible
# This check should FAIL (403/401) - if it succeeds, something is wrong
resource "quismon_check" "admin_panel_security" {
  name             = "Admin Panel Security Check"
  type             = "https"
  interval_seconds = 300
  enabled          = true

  regions = ["na-east-ewr"]

  # This is an inverted check - we expect 403 Forbidden
  # If we get 200 OK, the admin panel is exposed!
  config_json = jsonencode({
    url             = "https://${var.domain}/admin"
    method          = "GET"
    expected_status = [401, 403]  # Should be unauthorized
    timeout_seconds = 10
  })

  # Inverted: Success = problem detected
  inverted = true
}

# Alert when admin panel is accessible (inverted check succeeds = problem)
resource "quismon_alert_rule" "admin_exposed" {
  check_id = quismon_check.admin_panel_security.id
  name     = "Admin Panel Possibly Exposed"
  enabled  = true

  condition = {
    health_status = "down"  # For inverted checks, down means the problem is NOT detected
  }

  notification_channel_ids = [
    quismon_notification_channel.slack.id,
    quismon_notification_channel.email.id
  ]
}

# Inverted check: Alert if staging/development environment ports are open
resource "quismon_check" "dev_ports_closed" {
  name             = "Dev Ports Should Be Closed"
  type             = "tcp"
  interval_seconds = 600
  enabled          = true

  regions = ["na-east-ewr"]

  config_json = jsonencode({
    host            = var.domain
    port            = 3000  # Common dev port
    timeout_seconds = 5
  })

  # Inverted: We want this to fail (port should be closed)
  inverted = true
}

# =============================================================================
# Outputs
# =============================================================================

output "monitoring_summary" {
  value = {
    dns_check_id     = quismon_check.dns_a.id
    ssl_check_id     = quismon_check.ssl_cert.id
    website_check_id = quismon_check.website.id
    api_check_id     = quismon_check.api_health.id
    db_check_id      = quismon_check.database.id
  }
}

output "notification_channels" {
  value = {
    slack  = quismon_notification_channel.slack.id
    email  = quismon_notification_channel.email.id
    webhook = quismon_notification_channel.webhook.id
  }
}
