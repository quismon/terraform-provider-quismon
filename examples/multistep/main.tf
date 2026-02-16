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

# Multi-step checks allow you to monitor complex workflows with variable
# extraction and interpolation between steps.
#
# Use cases:
# - Login flows: Authenticate, then access protected resources
# - API chains: Get data from one endpoint, use it in another
# - User journeys: Simulate multi-page workflows

# Example 1: Login flow with token extraction
resource "quismon_check" "api_login_flow" {
  name             = "API Login Flow"
  type             = "multistep"
  interval_seconds = 300
  enabled          = true

  regions = ["us-east-1"]

  # Use config_json for complex nested configs like multistep
  config_json = jsonencode({
    steps = [
      {
        name = "Login"
        type = "https"
        config = {
          url             = "https://api.example.com/auth/login"
          method          = "POST"
          body            = jsonencode({ email = "monitoring@example.com", password = "secret" })
          expected_status = [200]
          timeout_seconds = 10
        }
        extracts = {
          auth_token = {
            jsonpath = "$.token"
          }
        }
      },
      {
        name = "Get Profile"
        type = "https"
        config = {
          url     = "https://api.example.com/user/profile"
          method  = "GET"
          headers = { Authorization = "Bearer {{auth_token}}" }
          expected_status = [200]
          timeout_seconds = 10
        }
      }
    ]
    fail_fast       = true
    timeout_seconds = 30
  })
}

# Example 2: E-commerce checkout flow
resource "quismon_check" "checkout_flow" {
  name             = "E-commerce Checkout Flow"
  type             = "multistep"
  interval_seconds = 600
  enabled          = true

  regions = ["us-east-1", "eu-west-1"]

  config_json = jsonencode({
    steps = [
      {
        name = "Get Product"
        type = "https"
        config = {
          url             = "https://shop.example.com/api/products/123"
          method          = "GET"
          expected_status = [200]
        }
        extracts = {
          product_price = {
            jsonpath = "$.price"
          }
        }
      },
      {
        name = "Add to Cart"
        type = "https"
        config = {
          url             = "https://shop.example.com/api/cart"
          method          = "POST"
          body            = jsonencode({ product_id = "123", price = "{{product_price}}" })
          expected_status = [201]
        }
      },
      {
        name = "Verify Cart"
        type = "https"
        config = {
          url             = "https://shop.example.com/api/cart"
          method          = "GET"
          expected_status = [200]
          expected_content = "123"
        }
      }
    ]
    fail_fast       = true
    timeout_seconds = 45
  })
}

# Example 3: API health chain with continue-on-failure
resource "quismon_check" "api_health_chain" {
  name             = "API Health Chain"
  type             = "multistep"
  interval_seconds = 60
  enabled          = true

  regions = ["us-east-1"]

  config_json = jsonencode({
    steps = [
      {
        name = "Health Check"
        type = "https"
        config = {
          url             = "https://api.example.com/health"
          method          = "GET"
          expected_status = [200]
        }
      },
      {
        name = "API Status"
        type = "https"
        config = {
          url             = "https://api.example.com/status"
          method          = "GET"
          expected_status = [200]
        }
      },
      {
        name = "Metrics Endpoint"
        type = "https"
        config = {
          url             = "https://api.example.com/metrics"
          method          = "GET"
          expected_status = [200]
        }
      }
    ]
    # Continue even if one step fails - useful for checking all endpoints
    fail_fast       = false
    timeout_seconds = 30
  })
}

# Notification channel for alerts
resource "quismon_notification_channel" "ops_team" {
  name = "Ops Team"
  type = "email"

  config = {
    to = jsonencode(["ops@example.com"])
  }

  enabled = true
}

# Alert on multi-step check failure
resource "quismon_alert_rule" "api_flow_down" {
  check_id = quismon_check.api_login_flow.id
  name     = "API Login Flow Down"
  enabled  = true

  condition = {
    failure_threshold = 2
  }

  notification_channel_ids = [
    quismon_notification_channel.ops_team.id
  ]
}

# Outputs
output "login_flow_check_id" {
  value       = quismon_check.api_login_flow.id
  description = "The ID of the login flow check"
}

output "checkout_flow_check_id" {
  value       = quismon_check.checkout_flow.id
  description = "The ID of the checkout flow check"
}
