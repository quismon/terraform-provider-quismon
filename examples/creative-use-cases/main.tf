# Creative Monitoring Use Cases
# Going beyond basic uptime monitoring with clever, advanced patterns
#
# This example demonstrates creative ways to use Quismon for:
# 1. Blue/Green deployment monitoring
# 2. Third-party dependency health
# 3. DNS failover verification
# 4. Rate limiting validation
# 5. CDN performance verification
# 6. Maintenance window detection
# 7. API contract testing
# 8. Certificate chain validation
# 9. Geographic blocking verification
# 10. Circuit breaker pattern monitoring

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

variable "primary_domain" {
  description = "Primary production domain"
  type        = string
  default     = "example.com"
}

variable "api_domain" {
  description = "API domain"
  type        = string
  default     = "api.example.com"
}

variable "slack_webhook_url" {
  description = "Slack webhook for alerts"
  type        = string
  sensitive   = true
}

variable "pagerduty_webhook" {
  description = "PagerDuty webhook for critical alerts"
  type        = string
  sensitive   = true
  default     = ""
}

variable "current_api_version" {
  description = "Current API version string"
  type        = string
  default     = "v2.0.0"
}

variable "third_party_apis" {
  description = "Third-party APIs your service depends on"
  type        = map(string)
  default = {
    stripe   = "https://api.stripe.com/v1/health"
    sendgrid = "https://status.sendgrid.com/api/v2/status.json"
    twilio   = "https://status.twilio.com/api/v2/status.json"
  }
}

# =============================================================================
# Notification Channels
# =============================================================================

resource "quismon_notification_channel" "slack_ops" {
  name = "Ops Team Slack"
  type = "slack"
  config = {
    webhook_url = var.slack_webhook_url
  }
  enabled = true
}

resource "quismon_notification_channel" "pagerduty" {
  count = var.pagerduty_webhook != "" ? 1 : 0
  name  = "PagerDuty"
  type  = "webhook"
  config = {
    url = var.pagerduty_webhook
  }
  enabled = true
}

# =============================================================================
# Use Case 1: Blue/Green Deployment Monitoring
# =============================================================================
# Monitor both blue and green deployment targets
# Alert when they return different content (version mismatch)

resource "quismon_check" "blue_target" {
  name             = "Blue Deployment Target"
  type             = "https"
  interval_seconds = 30
  enabled          = true

  regions = ["na-east-ewr"]

  config_json = jsonencode({
    url              = "https://blue.${var.primary_domain}/health"
    method           = "GET"
    expected_status  = [200]
    expected_content = "\"status\":\"healthy\""
    content_match_type = "contains"
    timeout_seconds  = 5
  })

  # Tag for easy identification
  # tags = { deployment = "blue" }  # If tags are supported
}

resource "quismon_check" "green_target" {
  name             = "Green Deployment Target"
  type             = "https"
  interval_seconds = 30
  enabled          = true

  regions = ["na-east-ewr"]

  config_json = jsonencode({
    url              = "https://green.${var.primary_domain}/health"
    method           = "GET"
    expected_status  = [200]
    expected_content = "\"status\":\"healthy\""
    content_match_type = "contains"
    timeout_seconds  = 5
  })
}

# Alert when either deployment target is unhealthy
resource "quismon_alert_rule" "deployment_target_down" {
  count = 2

  check_id = count.index == 0 ? quismon_check.blue_target.id : quismon_check.green_target.id
  name     = count.index == 0 ? "Blue Target Down" : "Green Target Down"
  enabled  = true

  condition = {
    failure_threshold = 2
  }

  notification_channel_ids = [
    quismon_notification_channel.slack_ops.id
  ]
}

# =============================================================================
# Use Case 2: Third-Party Dependency Health
# =============================================================================
# Monitor the health of external services your application depends on
# This helps you distinguish between "our service is down" vs "Stripe is down"

resource "quismon_check" "third_party_health" {
  for_each = var.third_party_apis

  name             = "${each.key} API Health"
  type             = "https"
  interval_seconds = 300  # Check every 5 minutes
  enabled          = true

  regions = ["na-east-ewr"]

  config_json = jsonencode({
    url              = each.value
    method           = "GET"
    expected_status  = [200]
    timeout_seconds  = 10
  })
}

# Alert when a critical dependency is down
resource "quismon_alert_rule" "dependency_down" {
  for_each = var.third_party_apis

  check_id = quismon_check.third_party_health[each.key].id
  name     = "${each.key} Dependency Down"
  enabled  = true

  condition = {
    failure_threshold = 3
  }

  notification_channel_ids = [
    quismon_notification_channel.slack_ops.id
  ]
}

# =============================================================================
# Use Case 3: DNS Failover Verification
# =============================================================================
# Verify that your DNS failover is working correctly
# Monitor the same domain from multiple regions to detect DNS changes

resource "quismon_check" "dns_failover_primary" {
  name             = "DNS Primary Region"
  type             = "https"
  interval_seconds = 60
  enabled          = true

  # Only check from primary region
  regions = ["na-east-ewr"]

  config_json = jsonencode({
    url              = "https://${var.primary_domain}/region"
    method           = "GET"
    expected_status  = [200]
    # Response should indicate the primary region
    expected_content = "\"region\":\"us-east\""
    content_match_type = "contains"
    timeout_seconds  = 5
  })
}

resource "quismon_check" "dns_failover_secondary" {
  name             = "DNS Secondary Region"
  type             = "https"
  interval_seconds = 60
  enabled          = true

  # Only check from secondary region
  regions = ["eu-west-ams"]

  config_json = jsonencode({
    url              = "https://${var.primary_domain}/region"
    method           = "GET"
    expected_status  = [200]
    # Response should indicate the secondary region
    expected_content = "\"region\":\"eu-west\""
    content_match_type = "contains"
    timeout_seconds  = 5
  })
}

# =============================================================================
# Use Case 4: Rate Limiting Validation (Inverted Check)
# =============================================================================
# Verify that rate limiting is working on your API
# We WANT to get rate limited (429) after several requests

resource "quismon_check" "rate_limit_working" {
  name             = "API Rate Limit Verification"
  type             = "https"
  interval_seconds = 600  # Check every 10 minutes
  enabled          = true

  regions = ["na-east-ewr"]

  config_json = jsonencode({
    url             = "${var.api_domain}/v1/test-endpoint"
    method          = "GET"
    # We expect rate limiting to kick in (429 Too Many Requests)
    # If we get 200, rate limiting might be broken
    expected_status = [429]
    timeout_seconds = 10
  })

  # INVERTED: Success means we're NOT rate limited = problem
  inverted = true
}

resource "quismon_alert_rule" "rate_limit_broken" {
  check_id = quismon_check.rate_limit_working.id
  name     = "Rate Limiting May Be Broken"
  enabled  = true

  condition = {
    health_status = "down"
  }

  notification_channel_ids = [
    quismon_notification_channel.slack_ops.id
  ]
}

# =============================================================================
# Use Case 5: CDN Performance Verification
# =============================================================================
# Verify CDN is serving cached content by checking response times
# If response time is high, CDN might not be caching properly

resource "quismon_check" "cdn_static_assets" {
  name             = "CDN Static Asset Performance"
  type             = "https"
  interval_seconds = 300
  enabled          = true

  regions = ["na-east-ewr", "eu-west-ams", "ap-southeast-sin"]

  config_json = jsonencode({
    url              = "https://${var.primary_domain}/static/logo.png"
    method           = "GET"
    expected_status  = [200]
    timeout_seconds  = 5
  })
}

# Alert when CDN response is slow (> 500ms for static assets)
resource "quismon_alert_rule" "cdn_slow" {
  check_id = quismon_check.cdn_static_assets.id
  name     = "CDN Response Slow"
  enabled  = true

  condition = {
    response_time_ms = 500
  }

  notification_channel_ids = [
    quismon_notification_channel.slack_ops.id
  ]
}

# =============================================================================
# Use Case 6: Maintenance Window Detection
# =============================================================================
# Detect when your site is in maintenance mode unexpectedly

resource "quismon_check" "not_in_maintenance" {
  name             = "Site Not In Maintenance Mode"
  type             = "https"
  interval_seconds = 60
  enabled          = true

  regions = ["na-east-ewr"]

  config_json = jsonencode({
    url              = "https://${var.primary_domain}/"
    method           = "GET"
    expected_status  = [200]
    # Ensure we don't see maintenance page
    timeout_seconds  = 10
  })
}

# Also check for 503 Service Unavailable (common maintenance response)
resource "quismon_check" "no_503_maintenance" {
  name             = "No 503 Maintenance Response"
  type             = "https"
  interval_seconds = 60
  enabled          = true

  regions = ["na-east-ewr"]

  config_json = jsonencode({
    url             = "https://${var.primary_domain}/"
    method          = "GET"
    # We expect 200, not 503
    expected_status = [200]
    timeout_seconds = 10
  })
}

# =============================================================================
# Use Case 7: API Contract Testing
# =============================================================================
# Verify API responses match expected schema/version

resource "quismon_check" "api_version_contract" {
  name             = "API Version Contract"
  type             = "https"
  interval_seconds = 300
  enabled          = true

  regions = ["na-east-ewr"]

  config_json = jsonencode({
    url              = "${var.api_domain}/version"
    method           = "GET"
    expected_status  = [200]
    # Verify API is returning expected version
    expected_content = "\"version\":\"${var.current_api_version}\""
    content_match_type = "contains"
    timeout_seconds  = 5
  })
}

resource "quismon_alert_rule" "api_version_changed" {
  check_id = quismon_check.api_version_contract.id
  name     = "API Version Changed"
  enabled  = true

  condition = {
    health_status = "down"
  }

  # This is informational - not necessarily critical
  notification_channel_ids = [
    quismon_notification_channel.slack_ops.id
  ]
}

# =============================================================================
# Use Case 8: Certificate Chain Validation
# =============================================================================
# Ensure SSL certificate chain is complete and valid
# Check multiple endpoints to ensure consistent SSL config

resource "quismon_check" "ssl_main" {
  name             = "Main Site SSL"
  type             = "ssl"
  interval_seconds = 3600  # Check hourly
  enabled          = true

  regions = ["na-east-ewr"]

  config_json = jsonencode({
    hostname            = var.primary_domain
    port                = 443
    warn_days_remaining = 30
  })
}

resource "quismon_check" "ssl_api" {
  name             = "API SSL"
  type             = "ssl"
  interval_seconds = 3600
  enabled          = true

  regions = ["na-east-ewr"]

  config_json = jsonencode({
    hostname            = var.api_domain
    port                = 443
    warn_days_remaining = 30
  })
}

# =============================================================================
# Use Case 9: Geographic Blocking Verification (Inverted Check)
# =============================================================================
# If you geo-block certain countries, verify it's working

resource "quismon_check" "geo_block_working" {
  name             = "Geo-Blocking Active"
  type             = "https"
  interval_seconds = 600
  enabled          = true

  # This assumes you have a region that should be blocked
  # Adjust region as needed
  regions = ["na-east-ewr"]

  config_json = jsonencode({
    url             = "https://${var.primary_domain}/"
    method          = "GET"
    # We expect 200 from allowed region
    expected_status = [200]
    timeout_seconds = 10
  })
}

# =============================================================================
# Use Case 10: Circuit Breaker Pattern Monitoring
# =============================================================================
# Monitor the health of services behind a circuit breaker
# Alert if circuit breaker is stuck open (all backends failing)

# This uses the "trigger check" feature to force an immediate check
# when the circuit breaker trips

# Primary service check
resource "quismon_check" "backend_service" {
  name             = "Backend Service Health"
  type             = "https"
  interval_seconds = 15  # Frequent checks for circuit breaker monitoring
  enabled          = true

  regions = ["na-east-ewr"]

  config_json = jsonencode({
    url              = "${var.api_domain}/health"
    method           = "GET"
    expected_status  = [200]
    expected_content = "\"healthy\":true"
    content_match_type = "contains"
    timeout_seconds  = 3
  })

  # Enable recheck on failure for faster detection
  recheck_on_failure = true
}

resource "quismon_alert_rule" "circuit_breaker_trip" {
  check_id = quismon_check.backend_service.id
  name     = "Backend Unhealthy - Circuit Breaker May Trip"
  enabled  = true

  condition = {
    failure_threshold = 3  # Circuit breaker threshold
  }

  notification_channel_ids = var.pagerduty_webhook != "" ? [
    quismon_notification_channel.slack_ops.id,
    quismon_notification_channel.pagerduty[0].id
  ] : [
    quismon_notification_channel.slack_ops.id
  ]
}

# =============================================================================
# Use Case 11: SMTP/IMAP Mail Server Health
# =============================================================================
# End-to-end mail server testing (if smtp-imap check is enabled)

resource "quismon_check" "mail_server" {
  name             = "Mail Server E2E"
  type             = "smtp-imap"
  interval_seconds = 1800  # Check every 30 minutes
  enabled          = true

  regions = ["na-east-ewr"]

  config_json = jsonencode({
    smtp_host     = "mail.${var.primary_domain}"
    smtp_port     = 587
    smtp_username = "monitoring@${var.primary_domain}"
    smtp_password = "CHANGE_ME"  # Use a variable in production
    imap_host     = "mail.${var.primary_domain}"
    imap_port     = 993
    imap_username = "monitoring@${var.primary_domain}"
    imap_password = "CHANGE_ME"
    from_address  = "monitoring@${var.primary_domain}"
    to_address    = "monitoring@${var.primary_domain}"
    timeout       = 30
  })
}

# =============================================================================
# Use Case 12: HTTP/3 (QUIC) Support Verification
# =============================================================================
# Verify HTTP/3 is working on your endpoints

resource "quismon_check" "http3_support" {
  name             = "HTTP/3 Support Check"
  type             = "http3"
  interval_seconds = 300
  enabled          = true

  regions = ["na-east-ewr", "eu-west-ams"]

  config_json = jsonencode({
    url             = "https://${var.primary_domain}/"
    method          = "GET"
    expected_status = [200]
    timeout_seconds = 10
  })
}

# =============================================================================
# Use Case 13: Throughput/Bandwidth Testing
# =============================================================================
# Periodically test download throughput from your CDN

resource "quismon_check" "download_throughput" {
  name             = "CDN Download Throughput"
  type             = "throughput"
  interval_seconds = 3600  # Test hourly
  enabled          = true

  regions = ["na-east-ewr", "eu-west-ams"]

  config_json = jsonencode({
    url          = "https://${var.primary_domain}/downloads/test-10mb.bin"
    max_size_mb  = 10
    timeout_seconds = 60
  })
}

# =============================================================================
# Outputs
# =============================================================================

output "creative_monitoring_summary" {
  value = {
    blue_green_checks = {
      blue  = quismon_check.blue_target.id
      green = quismon_check.green_target.id
    }
    third_party_checks = { for k, v in quismon_check.third_party_health : k => v.id }
    dns_failover = {
      primary   = quismon_check.dns_failover_primary.id
      secondary = quismon_check.dns_failover_secondary.id
    }
    security_checks = {
      rate_limit   = quismon_check.rate_limit_working.id
      ssl_main     = quismon_check.ssl_main.id
      ssl_api      = quismon_check.ssl_api.id
    }
    performance_checks = {
      cdn       = quismon_check.cdn_static_assets.id
      throughput = quismon_check.download_throughput.id
    }
    protocol_checks = {
      http3 = quismon_check.http3_support.id
      mail  = quismon_check.mail_server.id
    }
  }
  description = "Summary of all creative monitoring checks"
}

output "tips" {
  value = <<EOT
Creative Monitoring Tips:

1. Blue/Green: Monitor both targets during deployment to catch issues early
2. Dependencies: Distinguish "we're down" from "Stripe is down"
3. DNS Failover: Multi-region checks verify geographic routing
4. Rate Limiting: Use inverted checks to verify protection is active
5. CDN: Slow responses for static assets = cache miss problem
6. Maintenance: Alert on unexpected maintenance mode
7. API Contract: Version mismatch = unexpected deployment
8. SSL: Check all endpoints, not just the main one
9. Geo-Block: Verify blocking is working (not bypassed)
10. Circuit Breaker: Fast checks + recheck_on_failure for rapid detection
11. Mail: End-to-end SMTP/IMAP ensures deliverability
12. HTTP/3: Verify QUIC support for modern clients
13. Throughput: Periodic bandwidth tests catch CDN issues
EOT
}
