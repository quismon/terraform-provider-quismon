# Inverted Checks Example
# Inverted checks alert when something SUCCEEDS that shouldn't.
# Think of them as "security posture monitoring" or "proving negatives".
#
# Use cases:
# 1. Deployment verification - alert when new version appears
# 2. Security monitoring - alert when private endpoints become public
# 3. Firewall validation - alert when blocked ports become accessible
# 4. Compliance monitoring - alert when internal services are exposed

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

variable "production_url" {
  description = "Production website URL"
  type        = string
  default     = "https://example.com"
}

variable "api_url" {
  description = "Production API URL"
  type        = string
  default     = "https://api.example.com"
}

variable "current_version" {
  description = "Current deployed version string (update this after deployments)"
  type        = string
  default     = "v1.2.3"
}

variable "slack_webhook_url" {
  description = "Slack webhook for alerts"
  type        = string
  sensitive   = true
}

# =============================================================================
# Notification Channel
# =============================================================================

resource "quismon_notification_channel" "slack" {
  name = "Security Alerts"
  type = "slack"

  config = {
    webhook_url = var.slack_webhook_url
  }

  enabled = true
}

# =============================================================================
# Use Case 1: Deployment Verification
# =============================================================================
# Alert when a NEW version string appears in production
# This helps you verify deployments happened as expected
# Set inverted = true, expected_content = current version
# When a NEW version is deployed, the check will "fail" (not find old version)

resource "quismon_check" "deployment_version" {
  name             = "Production Version Monitor"
  type             = "https"
  interval_seconds = 60
  enabled          = true

  regions = ["na-east-ewr"]

  config_json = jsonencode({
    url               = "${var.production_url}/version"
    method            = "GET"
    expected_status   = [200]
    # We expect to see the CURRENT version
    # When deployment happens, content changes and check fails
    expected_content  = var.current_version
    content_match_type = "contains"
    timeout_seconds   = 10
  })

  # NOT inverted - we want to be alerted when content CHANGES
  # (i.e., when the old version string is no longer there)
}

resource "quismon_alert_rule" "new_version_deployed" {
  check_id = quismon_check.deployment_version.id
  name     = "New Version Deployed"
  enabled  = true

  condition = {
    health_status = "down"  # Check fails when version changes
  }

  notification_channel_ids = [
    quismon_notification_channel.slack.id
  ]
}

# =============================================================================
# Use Case 2: Admin Panel Security (True Inverted Check)
# =============================================================================
# We want the admin panel to return 401/403 (unauthorized)
# If it returns 200, something is wrong (it's publicly accessible)

resource "quismon_check" "admin_panel_not_public" {
  name             = "Admin Panel Should NOT Be Public"
  type             = "https"
  interval_seconds = 300
  enabled          = true

  regions = ["na-east-ewr", "eu-west-ams"]

  config_json = jsonencode({
    url             = "${var.production_url}/admin"
    method          = "GET"
    # We WANT 401 or 403 - these mean the admin panel is protected
    expected_status = [401, 403]
    timeout_seconds = 10
  })

  # INVERTED: Check succeeds when we get 401/403
  # If check fails (we get 200), the admin panel is exposed!
  inverted = true
}

resource "quismon_alert_rule" "admin_exposed" {
  check_id = quismon_check.admin_panel_not_public.id
  name     = "ALERT: Admin Panel May Be Exposed"
  enabled  = true

  condition = {
    health_status = "down"
  }

  notification_channel_ids = [
    quismon_notification_channel.slack.id
  ]
}

# =============================================================================
# Use Case 3: API Endpoint Security
# =============================================================================
# Internal API endpoints should require authentication

resource "quismon_check" "internal_api_protected" {
  name             = "Internal API Should Require Auth"
  type             = "https"
  interval_seconds = 300
  enabled          = true

  regions = ["na-east-ewr"]

  config_json = jsonencode({
    url             = "${var.api_url}/internal/stats"
    method          = "GET"
    # Should return 401 Unauthorized
    expected_status = [401]
    timeout_seconds = 10
  })

  inverted = true
}

resource "quismon_alert_rule" "internal_api_exposed" {
  check_id = quismon_check.internal_api_protected.id
  name     = "ALERT: Internal API May Be Exposed"
  enabled  = true

  condition = {
    health_status = "down"
  }

  notification_channel_ids = [
    quismon_notification_channel.slack.id
  ]
}

# =============================================================================
# Use Case 4: Firewall/Port Security
# =============================================================================
# These ports should NOT be accessible from the internet

resource "quismon_check" "database_port_blocked" {
  name             = "Database Port Should Be Blocked"
  type             = "tcp"
  interval_seconds = 300
  enabled          = true

  regions = ["na-east-ewr", "eu-west-ams"]

  config_json = jsonencode({
    host            = var.production_url  # Replace with your host
    port            = 5432  # PostgreSQL
    timeout_seconds = 5
  })

  # INVERTED: Connection should FAIL
  # If connection succeeds, the database port is exposed!
  inverted = true
}

resource "quismon_check" "redis_port_blocked" {
  name             = "Redis Port Should Be Blocked"
  type             = "tcp"
  interval_seconds = 300
  enabled          = true

  regions = ["na-east-ewr", "eu-west-ams"]

  config_json = jsonencode({
    host            = var.production_url
    port            = 6379  # Redis
    timeout_seconds = 5
  })

  inverted = true
}

resource "quismon_alert_rule" "sensitive_port_exposed" {
  check_id = quismon_check.database_port_blocked.id
  name     = "ALERT: Database Port May Be Exposed"
  enabled  = true

  condition = {
    health_status = "down"
  }

  notification_channel_ids = [
    quismon_notification_channel.slack.id
  ]
}

# =============================================================================
# Use Case 5: Staging/Dev Environment Isolation
# =============================================================================
# Staging should not be accessible from public internet

resource "quismon_check" "staging_not_public" {
  name             = "Staging Should Not Be Public"
  type             = "https"
  interval_seconds = 600
  enabled          = true

  regions = ["na-east-ewr"]

  config_json = jsonencode({
    url             = "https://staging.${var.production_url}"
    method          = "GET"
    # If staging is properly firewalled, connection should fail
    expected_status = [200]  # We expect to NOT get this
    timeout_seconds = 10
  })

  # INVERTED: We want this check to fail
  # If it succeeds, staging is publicly accessible
  inverted = true
}

# =============================================================================
# Outputs
# =============================================================================

output "inverted_checks" {
  value = {
    deployment_version_check = quismon_check.deployment_version.id
    admin_panel_check        = quismon_check.admin_panel_not_public.id
    internal_api_check       = quismon_check.internal_api_protected.id
    database_port_check      = quismon_check.database_port_blocked.id
    redis_port_check         = quismon_check.redis_port_blocked.id
  }
  description = "IDs of all inverted check resources"
}

output "how_inverted_works" {
  value = <<EOT
Inverted checks work by reversing the success/failure logic:

NORMAL CHECK: success = expected result, failure = alert
INVERTED CHECK: success = alert, failure = expected result

For security monitoring:
- We want things to FAIL (be inaccessible, return 403, etc.)
- When they SUCCEED, something is wrong (security breach)
- Setting inverted = true makes "success" trigger an alert

For deployment verification:
- We monitor for a specific version string
- When it changes, the check fails
- This alerts us that a deployment happened
EOT
}
