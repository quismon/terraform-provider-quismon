# =============================================================================
# Quismon + HashiCorp Vault Integration Example
# =============================================================================
# This example demonstrates how to use Quismon multistep checks to fetch secrets
# from HashiCorp Vault and use them in your monitoring workflows.
#
# Use Case: You don't want to store sensitive credentials in your monitoring
# configuration. Instead, store them in Vault and fetch them at check time.
#
# Prerequisites:
# - A running HashiCorp Vault server
# - Userpass auth method enabled
# - KV secrets engine v2 enabled
# - A user with read access to the secrets
# =============================================================================

terraform {
  required_providers {
    quismon = {
      source  = "quismon/quismon"
      version = ">= 1.0.0"
    }
  }
}

provider "quismon" {
  # Get your API key from https://console.quismon.com/api-keys
  api_key = var.quismon_api_key
}

# =============================================================================
# VARIABLES
# =============================================================================

variable "quismon_api_key" {
  description = "Quismon API key"
  type        = string
  sensitive   = true
}

variable "vault_base_url" {
  description = "Base URL for HashiCorp Vault (e.g., https://vault.example.com)"
  type        = string
}

variable "vault_username" {
  description = "Vault username for userpass authentication"
  type        = string
  sensitive   = true
}

variable "vault_password" {
  description = "Vault password for userpass authentication"
  type        = string
  sensitive   = true
}

# =============================================================================
# MULTISTEP CHECK: Fetch Secrets from Vault
# =============================================================================
# This check demonstrates the complete workflow:
# 1. Authenticate to Vault using username/password
# 2. Receive a short-lived token
# 3. Use the token to read secrets
# 4. Use the secrets in a subsequent API call

resource "quismon_check" "vault_secret_fetch" {
  name             = "Vault-Secret-Fetch-Example"
  type             = "multistep"
  enabled          = true
  interval_seconds = 300
  regions          = ["na-east-ewr"]

  config_json = jsonencode({
    fail_fast       = false
    timeout_seconds = 60

    steps = [
      # Step 1: Authenticate to Vault
      # POST to /v1/auth/userpass/login/{username} with password
      # Response contains auth.client_token (the Vault token)
      {
        name = "vault-login"
        type = "https"
        config = {
          url             = "${var.vault_base_url}/v1/auth/userpass/login/${var.vault_username}"
          method          = "POST"
          body            = jsonencode({ password = var.vault_password })
          expected_status = [200]
          headers = {
            "Content-Type" = "application/json"
          }
        }
        extracts = {
          # Extract the Vault token from the response
          vault_token = {
            jsonpath = "$.auth.client_token"
          }
          # Also capture lease duration for monitoring
          lease_duration = {
            jsonpath = "$.auth.lease_duration"
          }
        }
      },

      # Step 2: Read database credentials from Vault
      # GET /v1/secret/data/{path} with X-Vault-Token header
      {
        name = "read-db-secret"
        type = "https"
        config = {
          url             = "${var.vault_base_url}/v1/secret/data/database/production"
          method          = "GET"
          expected_status = [200]
          headers = {
            "X-Vault-Token" = "{{vault_token}}"
          }
        }
        extracts = {
          db_host = {
            jsonpath = "$.data.data.db_host"
          }
          db_port = {
            jsonpath = "$.data.data.db_port"
          }
          db_user = {
            jsonpath = "$.data.data.db_user"
          }
          db_password = {
            jsonpath = "$.data.data.db_password"
          }
        }
      },

      # Step 3: Read API credentials from Vault
      {
        name = "read-api-secret"
        type = "https"
        config = {
          url             = "${var.vault_base_url}/v1/secret/data/api/external-service"
          method          = "GET"
          expected_status = [200]
          headers = {
            "X-Vault-Token" = "{{vault_token}}"
          }
        }
        extracts = {
          api_key = {
            jsonpath = "$.data.data.api_key"
          }
          api_url = {
            jsonpath = "$.data.data.api_url"
          }
        }
      },

      # Step 4: Use the fetched secrets
      # Now you can use the secrets in your API calls
      {
        name = "use-secrets"
        type = "https"
        config = {
          url             = "{{api_url}}/health"
          method          = "GET"
          expected_status = [200]
          headers = {
            "Authorization" = "Bearer {{api_key}}"
            "X-DB-Host"     = "{{db_host}}"
          }
        }
        extracts = {
          service_healthy = {
            jsonpath = "$.healthy"
          }
        }
      }
    ]
  })
}

# =============================================================================
# ALERT RULE
# =============================================================================

resource "quismon_alert_rule" "vault_check_failure" {
  name    = "Vault-Secret-Fetch-Failure"
  enabled = true
  check_id = quismon_check.vault_secret_fetch.id

  condition = {
    health_status = "down"
  }

  # Add your notification channel IDs here
  # notification_channel_ids = ["your-channel-id"]
}

# =============================================================================
# OUTPUTS
# =============================================================================

output "check_id" {
  description = "ID of the Vault secret fetch check"
  value       = quismon_check.vault_secret_fetch.id
}

output "check_name" {
  description = "Name of the check"
  value       = quismon_check.vault_secret_fetch.name
}

output "alert_rule_id" {
  description = "ID of the alert rule"
  value       = quismon_alert_rule.vault_check_failure.id
}
