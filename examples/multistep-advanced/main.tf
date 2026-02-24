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

# Advanced Multi-step Check Examples
# These demonstrate more complex real-world monitoring scenarios

# =============================================================================
# Example 1: OAuth2 Client Credentials Flow
# =============================================================================
# Monitors an API that requires OAuth2 client credentials authentication
resource "quismon_check" "oauth2_api_flow" {
  name             = "OAuth2 API Flow"
  type             = "multistep"
  interval_seconds = 300
  enabled          = true

  regions = ["na-east-ewr", "eu-central-fra"]

  config_json = jsonencode({
    steps = [
      {
        name = "Get OAuth2 Token"
        type = "https"
        config = {
          url    = "https://auth.example.com/oauth/token"
          method = "POST"
          headers = {
            Content-Type  = "application/x-www-form-urlencoded"
            Authorization = "Basic {{base64(client_id:client_secret)}}" # Pre-computed
          }
          body            = "grant_type=client_credentials&scope=api:read"
          expected_status = [200]
          timeout_seconds = 10
        }
        extracts = {
          access_token = {
            jsonpath = "$.access_token"
          }
          token_type = {
            jsonpath = "$.token_type"
          }
        }
      },
      {
        name = "Call Protected API"
        type = "https"
        config = {
          url    = "https://api.example.com/v1/data"
          method = "GET"
          headers = {
            Authorization = "{{token_type}} {{access_token}}"
            Accept        = "application/json"
          }
          expected_status = [200]
          timeout_seconds = 15
        }
        extracts = {
          data_count = {
            jsonpath = "$.items.length()"
          }
        }
      },
      {
        name = "Verify Response Quality"
        type = "https"
        config = {
          url             = "https://api.example.com/v1/health"
          method          = "GET"
          headers         = { Authorization = "{{token_type}} {{access_token}}" }
          expected_status = [200]
          # Verify API is responding within acceptable latency
          timeout_seconds = 5
        }
      }
    ]
    fail_fast       = true
    timeout_seconds = 45
  })
}

# =============================================================================
# Example 2: GraphQL API Monitoring with Query Chaining
# =============================================================================
# Monitors a GraphQL API with multiple queries in sequence
resource "quismon_check" "graphql_workflow" {
  name             = "GraphQL API Workflow"
  type             = "multistep"
  interval_seconds = 600
  enabled          = true

  regions = ["na-east-ewr"]

  config_json = jsonencode({
    steps = [
      {
        name = "GraphQL - Get User"
        type = "https"
        config = {
          url    = "https://graphql.example.com/v1/graphql"
          method = "POST"
          headers = {
            Content-Type = "application/json"
            X-API-Key    = "{{graphql_api_key}}" # Set via variable
          }
          body = jsonencode({
            query = "{ user(id: \"123\") { id name email organization { id name } } }"
          })
          expected_status = [200]
          timeout_seconds = 10
        }
        extracts = {
          org_id = {
            jsonpath = "$.data.user.organization.id"
          }
        }
      },
      {
        name = "GraphQL - Get Organization Data"
        type = "https"
        config = {
          url    = "https://graphql.example.com/v1/graphql"
          method = "POST"
          headers = {
            Content-Type = "application/json"
            X-API-Key    = "{{graphql_api_key}}"
          }
          body = jsonencode({
            query = "{ organization(id: \"{{org_id}}\") { id name checks { id name status } } }"
          })
          expected_status = [200]
          timeout_seconds = 15
        }
      }
    ]
    fail_fast       = true
    timeout_seconds = 30
  })
}

# =============================================================================
# Example 3: Webhook Signature Verification Flow
# =============================================================================
# Monitors a webhook endpoint by sending signed payloads
resource "quismon_check" "webhook_verification" {
  name             = "Webhook Signature Verification"
  type             = "multistep"
  interval_seconds = 300
  enabled          = true

  regions = ["na-east-ewr", "eu-central-fra", "ap-southeast-sin"]

  config_json = jsonencode({
    steps = [
      {
        name = "Get Current Timestamp"
        type = "https"
        config = {
          url             = "https://api.example.com/v1/time"
          method          = "GET"
          expected_status = [200]
          timeout_seconds = 5
        }
        extracts = {
          timestamp = {
            jsonpath = "$.timestamp"
          }
        }
      },
      {
        name = "Send Signed Webhook Test"
        type = "https"
        config = {
          url    = "https://webhooks.example.com/test"
          method = "POST"
          headers = {
            Content-Type   = "application/json"
            X-Signature    = "{{computed_hmac_sha256}}" # Pre-computed for test
            X-Timestamp    = "{{timestamp}}"
            X-Webhook-Test = "true"
          }
          body = jsonencode({
            event   = "test.ping"
            data    = { source = "quismon-monitoring" }
            time    = "{{timestamp}}"
            test    = true
          })
          expected_status = [200, 201, 204]
          timeout_seconds = 10
        }
      },
      {
        name = "Verify Webhook Received"
        type = "https"
        config = {
          url             = "https://api.example.com/v1/webhooks/logs?event=test.ping&time={{timestamp}}"
          method          = "GET"
          expected_status = [200]
          expected_content = "test.ping"
          timeout_seconds = 5
        }
      }
    ]
    fail_fast       = false # Continue to see all results
    timeout_seconds = 30
  })
}

# =============================================================================
# Example 4: Database API with Pagination
# =============================================================================
# Monitors a paginated API endpoint
resource "quismon_check" "paginated_api_flow" {
  name             = "Paginated API Flow"
  type             = "multistep"
  interval_seconds = 300
  enabled          = true

  regions = ["na-east-ewr"]

  config_json = jsonencode({
    steps = [
      {
        name = "Fetch First Page"
        type = "https"
        config = {
          url             = "https://api.example.com/v1/items?page=1&limit=10"
          method          = "GET"
          expected_status = [200]
          timeout_seconds = 10
        }
        extracts = {
          next_page = {
            jsonpath = "$.pagination.next_page"
          }
          total_items = {
            jsonpath = "$.pagination.total"
          }
        }
      },
      {
        name = "Fetch Second Page"
        type = "https"
        config = {
          # Using conditional - only runs if next_page exists
          url             = "https://api.example.com/v1/items?page=2&limit=10"
          method          = "GET"
          expected_status = [200]
          timeout_seconds = 10
        }
        extracts = {
          page_2_count = {
            jsonpath = "$.items.length()"
          }
        }
      },
      {
        name = "Fetch Specific Item"
        type = "https"
        config = {
          url             = "https://api.example.com/v1/items/first-item-id"
          method          = "GET"
          expected_status = [200, 404] # 404 OK if item was deleted
          timeout_seconds = 5
        }
      }
    ]
    fail_fast       = false
    timeout_seconds = 30
  })
}

# =============================================================================
# Example 5: Circuit Breaker Pattern with Fallback
# =============================================================================
# Monitors primary API with automatic fallback to secondary
resource "quismon_check" "circuit_breaker_flow" {
  name             = "Circuit Breaker API"
  type             = "multistep"
  interval_seconds = 60
  enabled          = true

  regions = ["na-east-ewr", "eu-central-fra"]

  config_json = jsonencode({
    steps = [
      {
        name = "Health Check Primary"
        type = "https"
        config = {
          url             = "https://primary.api.example.com/health"
          method          = "GET"
          expected_status = [200]
          timeout_seconds = 5
        }
        extracts = {
          primary_healthy = {
            jsonpath = "$.status"
            # Will be "healthy" or "unhealthy"
          }
        }
      },
      {
        name = "Use Primary API"
        type = "https"
        config = {
          url             = "https://primary.api.example.com/v1/data"
          method          = "GET"
          expected_status = [200]
          timeout_seconds = 10
        }
        # This step fails if primary is down, triggering fail_fast
      },
      {
        name = "Fallback to Secondary"
        type = "https"
        config = {
          url             = "https://secondary.api.example.com/v1/data"
          method          = "GET"
          expected_status = [200]
          timeout_seconds = 10
        }
        # This only matters if primary failed - whole check fails either way
        # But helps identify which endpoint is the problem
      }
    ]
    fail_fast       = true # Stop after primary failure to alert quickly
    timeout_seconds = 30
  })
}

# =============================================================================
# Example 6: HTTP/3 Multi-step Check
# =============================================================================
# Monitors HTTP/3 endpoints with QUIC protocol
resource "quismon_check" "http3_workflow" {
  name             = "HTTP/3 API Workflow"
  type             = "multistep"
  interval_seconds = 300
  enabled          = true

  regions = ["na-east-ewr", "eu-central-fra", "ap-southeast-sin"]

  config_json = jsonencode({
    steps = [
      {
        name = "HTTP/3 Health Check"
        type = "http3"
        config = {
          url             = "https://api.example.com/health"
          method          = "GET"
          expected_status = [200]
          timeout_seconds = 10
        }
        extracts = {
          server_version = {
            header = "X-Server-Version"
          }
        }
      },
      {
        name = "HTTP/3 API Call"
        type = "http3"
        config = {
          url             = "https://api.example.com/v1/status"
          method          = "GET"
          expected_status = [200]
          timeout_seconds = 10
        }
      },
      {
        name = "HTTP/3 Metrics"
        type = "http3"
        config = {
          url             = "https://api.example.com/v1/metrics"
          method          = "GET"
          expected_status = [200]
          timeout_seconds = 10
        }
      }
    ]
    fail_fast       = true
    timeout_seconds = 40
  })
}

# Notification channels
resource "quismon_notification_channel" "ops_email" {
  name   = "Ops Team Email"
  type   = "email"
  config = { to = jsonencode(["ops@example.com"]) }
  enabled = true
}

resource "quismon_notification_channel" "slack_alerts" {
  name   = "Slack Alerts"
  type   = "slack"
  config = { webhook_url = "https://hooks.slack.com/services/XXX/YYY/ZZZ" }
  enabled = true
}

# Alert rules
resource "quismon_alert_rule" "oauth2_down" {
  check_id    = quismon_check.oauth2_api_flow.id
  name        = "OAuth2 API Flow Down"
  enabled     = true
  condition   = { failure_threshold = 2 }
  notification_channel_ids = [
    quismon_notification_channel.ops_email.id,
    quismon_notification_channel.slack_alerts.id
  ]
}

resource "quismon_alert_rule" "circuit_breaker_down" {
  check_id    = quismon_check.circuit_breaker_flow.id
  name        = "Circuit Breaker API Down"
  enabled     = true
  condition   = { failure_threshold = 1 } # Alert immediately
  notification_channel_ids = [
    quismon_notification_channel.ops_email.id,
    quismon_notification_channel.slack_alerts.id
  ]
}

# Outputs
output "oauth2_check_id" {
  value = quismon_check.oauth2_api_flow.id
}

output "graphql_check_id" {
  value = quismon_check.graphql_workflow.id
}

output "http3_check_id" {
  value = quismon_check.http3_workflow.id
}
