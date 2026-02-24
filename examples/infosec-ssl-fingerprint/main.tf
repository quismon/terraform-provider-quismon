terraform {
  required_providers {
    quismon = {
      source = "quismon/quismon"
    }
  }
}

provider "quismon" {
  # api_key  = var.quismon_api_key  # Or use QUISMON_API_KEY env var
  # base_url = "https://console.quismon.com"  # Optional
}

# =============================================================================
# Infosec Example: SSL Certificate Fingerprint Monitoring
# =============================================================================
#
# This example demonstrates how to monitor SSL/TLS certificates for:
# 1. Expiry detection - Alert before certificates expire
# 2. Fingerprint verification - Detect unauthorized certificate changes
# 3. Chain validation - Ensure proper certificate chain
#
# Integration with Let's Encrypt:
# - Use certbot or ACME client to deploy certificates
# - Extract SHA256 fingerprint from deployed certificate
# - Configure Quismon to verify the expected fingerprint
# - If certificate changes (renewal or attack), you'll be alerted
#
# =============================================================================

# -----------------------------------------------------------------------------
# Example 1: Basic SSL Certificate Monitoring
# -----------------------------------------------------------------------------
# Monitors certificate expiry for a website
resource "quismon_check" "website_ssl_expiry" {
  name             = "Website SSL Expiry"
  type             = "ssl"
  interval_seconds = 3600 # Check every hour
  enabled          = true

  regions = ["na-east-ewr", "eu-central-fra"]

  config = {
    host               = "www.example.com"
    port               = 443
    expiry_threshold_days = 14  # Alert if expires within 14 days
    timeout_seconds    = 10
  }
}

# -----------------------------------------------------------------------------
# Example 2: SSL Fingerprint Verification (Certificate Pinning)
# -----------------------------------------------------------------------------
# Verifies the exact certificate fingerprint - detects any changes
#
# HOW TO GET THE FINGERPRINT:
# Run this command on your server or using openssl:
#
#   echo | openssl s_client -connect www.example.com:443 2>/dev/null | \
#     openssl x509 -fingerprint -sha256 -noout | \
#     sed 's/.*=//; s/://g' | tr '[:upper:]' '[:lower:]'
#
# Output example: a1b2c3d4e5f6... (64 hex characters for SHA256)
#
# Or using certbot after certificate issuance:
#
#   certbot certificates 2>/dev/null | grep "SHA-256" | \
#     awk '{print $2}' | tr -d ':'
#
resource "quismon_check" "website_ssl_fingerprint" {
  name             = "Website SSL Fingerprint"
  type             = "ssl"
  interval_seconds = 300  # Check every 5 minutes
  enabled          = true

  regions = ["na-east-ewr"]

  config = {
    host                   = "www.example.com"
    port                   = 443
    expected_fingerprint   = "a1b2c3d4e5f6789012345678901234567890abcdefabcdef0123456789abcdef"
    fingerprint_algorithm  = "sha256"
    timeout_seconds        = 10
  }
}

# -----------------------------------------------------------------------------
# Example 3: Multi-domain Certificate Monitoring
# -----------------------------------------------------------------------------
# Monitor multiple domains with the same certificate (SAN certificate)
resource "quismon_check" "wildcard_ssl" {
  name             = "Wildcard Certificate (*.example.com)"
  type             = "ssl"
  interval_seconds = 3600
  enabled          = true

  regions = ["na-east-ewr", "eu-central-fra", "ap-southeast-sin"]

  config = {
    host                 = "example.com"
    port                 = 443
    expiry_threshold_days = 14
    # Verify SAN includes expected domains
    expected_san = jsonencode([
      "example.com",
      "*.example.com",
      "www.example.com",
      "api.example.com"
    ])
    timeout_seconds = 10
  }
}

# -----------------------------------------------------------------------------
# Example 4: API Endpoint Certificate Chain Validation
# -----------------------------------------------------------------------------
# Validates the entire certificate chain for an API
resource "quismon_check" "api_ssl_chain" {
  name             = "API SSL Chain Validation"
  type             = "ssl"
  interval_seconds = 1800 # Every 30 minutes
  enabled          = true

  regions = ["na-east-ewr", "eu-central-fra"]

  config = {
    host = "api.example.com"
    port = 443
    # Verify the issuer is Let's Encrypt (or your expected CA)
    expected_issuer   = "Let's Encrypt Authority X3"
    expiry_threshold_days = 7
    verify_chain      = true
    timeout_seconds   = 15
  }
}

# -----------------------------------------------------------------------------
# Example 5: Internal Service Certificate (Self-signed or Private CA)
# -----------------------------------------------------------------------------
# Monitor internal services with self-signed or private CA certificates
resource "quismon_check" "internal_api_ssl" {
  name             = "Internal API SSL"
  type             = "ssl"
  interval_seconds = 3600
  enabled          = true

  regions = ["na-east-ewr"]  # Only check from regions with VPN access

  config = {
    host               = "internal-api.example.local"
    port               = 8443
    # For self-signed certs, you might skip verification
    # But ALWAYS pin the fingerprint to detect changes
    expected_fingerprint   = "deadbeef1234567890abcdef1234567890abcdef1234567890abcdef12345678"
    skip_verify_chain  = false
    expiry_threshold_days = 30
    timeout_seconds    = 10
  }
}

# -----------------------------------------------------------------------------
# Example 6: Multi-step Check - SSL + HTTPS Response
# -----------------------------------------------------------------------------
# Combined check: verify SSL certificate AND test HTTPS endpoint
resource "quismon_check" "ssl_and_health" {
  name             = "SSL Certificate + Health Check"
  type             = "multistep"
  interval_seconds = 300
  enabled          = true

  regions = ["na-east-ewr", "eu-central-fra"]

  config_json = jsonencode({
    steps = [
      {
        name = "Verify SSL Certificate"
        type = "ssl"
        config = {
          host                 = "api.example.com"
          port                 = 443
          expected_fingerprint = "a1b2c3d4e5f6789012345678901234567890abcdefabcdef0123456789abcdef"
          expiry_threshold_days = 14
          timeout_seconds      = 10
        }
      },
      {
        name = "Test HTTPS Endpoint"
        type = "https"
        config = {
          url             = "https://api.example.com/v1/health"
          method          = "GET"
          expected_status = [200]
          timeout_seconds = 10
        }
      },
      {
        name = "Verify API Response"
        type = "https"
        config = {
          url             = "https://api.example.com/v1/status"
          method          = "GET"
          expected_status = [200]
          expected_content = "operational"
          timeout_seconds = 10
        }
      }
    ]
    fail_fast       = true
    timeout_seconds = 45
  })
}

# =============================================================================
# LET'S ENCRYPT INTEGRATION EXAMPLE
# =============================================================================
#
# When using certbot or another ACME client with Let's Encrypt:
#
# 1. Deploy certificate using certbot:
#    certbot certonly --nginx -d example.com -d www.example.com
#
# 2. Get the SHA-256 fingerprint:
#    openssl x509 -in /etc/letsencrypt/live/example.com/cert.pem \
#      -fingerprint -sha256 -noout | sed 's/.*=//; s/://g' | tr '[:upper:]' '[:lower:]'
#
# 3. Update the expected_fingerprint in your Terraform config
#
# 4. When certbot renews the certificate, update the fingerprint:
#    - This can be automated with a post-renewal hook
#    - Store the fingerprint in a secure location (e.g., Terraform Cloud variables)
#
# Example certbot renewal hook (/etc/letsencrypt/renewal-hooks/post/quismom-update.sh):
#
#    #!/bin/bash
#    DOMAIN="example.com"
#    FINGERPRINT=$(openssl x509 -in /etc/letsencrypt/live/$DOMAIN/cert.pem \
#      -fingerprint -sha256 -noout | sed 's/.*=//; s/://g' | tr '[:upper:]' '[:lower:]')
#
#    # Update via API or Terraform
#    curl -X PATCH "https://api.quismon.com/v1/checks/CHECK_ID" \
#      -H "Authorization: Bearer $QUISMON_API_KEY" \
#      -H "Content-Type: application/json" \
#      -d "{\"config\": {\"expected_fingerprint\": \"$FINGERPRINT\"}}"
#
# =============================================================================

# Notification channels
resource "quismon_notification_channel" "security_team_email" {
  name   = "Security Team"
  type   = "email"
  config = { to = jsonencode(["security@example.com"]) }
  enabled = true
}

resource "quismon_notification_channel" "security_slack" {
  name   = "Security Alerts Slack"
  type   = "slack"
  config = { webhook_url = "https://hooks.slack.com/services/XXX/YYY/ZZZ" }
  enabled = true
}

# Alert: Certificate Expiring Soon
resource "quismon_alert_rule" "cert_expiring" {
  check_id  = quismon_check.website_ssl_expiry.id
  name      = "SSL Certificate Expiring Soon"
  enabled   = true
  condition = {
    failure_threshold = 1  # Alert immediately
  }
  notification_channel_ids = [
    quismon_notification_channel.security_team_email.id,
    quismon_notification_channel.security_slack.id
  ]
}

# Alert: Certificate Fingerprint Changed (CRITICAL - Possible Attack)
resource "quismon_alert_rule" "cert_fingerprint_changed" {
  check_id  = quismon_check.website_ssl_fingerprint.id
  name      = "CRITICAL: SSL Certificate Changed"
  enabled   = true
  condition = {
    failure_threshold = 1  # Alert immediately
  }
  notification_channel_ids = [
    quismon_notification_channel.security_team_email.id,
    quismon_notification_channel.security_slack.id
  ]
}

# Alert: API SSL Chain Invalid
resource "quismon_alert_rule" "api_ssl_invalid" {
  check_id  = quismon_check.api_ssl_chain.id
  name      = "API SSL Chain Validation Failed"
  enabled   = true
  condition = {
    failure_threshold = 1
  }
  notification_channel_ids = [
    quismon_notification_channel.security_team_email.id,
    quismon_notification_channel.security_slack.id
  ]
}

# Outputs
output "ssl_expiry_check_id" {
  value       = quismon_check.website_ssl_expiry.id
  description = "SSL expiry monitoring check ID"
}

output "ssl_fingerprint_check_id" {
  value       = quismon_check.website_ssl_fingerprint.id
  description = "SSL fingerprint verification check ID"
}

output "ssl_multistep_check_id" {
  value       = quismon_check.ssl_and_health.id
  description = "Combined SSL + health check ID"
}
