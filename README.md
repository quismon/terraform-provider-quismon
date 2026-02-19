# Terraform Provider for Quismon

The Quismon Terraform provider allows you to manage your monitoring infrastructure as code.

## Features

- **Checks**: Create and manage HTTP/HTTPS, TCP, Ping, DNS, SSL, HTTP/3, Throughput, SMTP/IMAP, and Multi-step health checks
- **Alert Rules**: Configure alert conditions using flexible condition maps
- **Notification Channels**: Set up email, ntfy, webhook, and Slack notifications
- **Custom Templates**: Use template variables for personalized alert messages
- **Data Sources**: Query existing checks and channels
- **Multi-Region Monitoring**: Deploy checks across multiple geographic regions

## Requirements

- Terraform >= 1.0
- Go >= 1.21 (for building from source)
- Quismon API key

## Installation

### Using Terraform Registry (Once Published)

```hcl
terraform {
  required_providers {
    quismon = {
      source  = "quismon/quismon"
      version = "~> 1.0"
    }
  }
}
```

### Building From Source

```bash
git clone https://github.com/quismon/terraform-provider-quismon
cd terraform-provider-quismon
go build -o terraform-provider-quismon
mkdir -p ~/.terraform.d/plugins/registry.terraform.io/quismon/quismon/1.0.0/linux_amd64/
mv terraform-provider-quismon ~/.terraform.d/plugins/registry.terraform.io/quismon/quismon/1.0.0/linux_amd64/
```

## Authentication

Obtain an API key from the Quismon dashboard, then configure the provider:

```hcl
provider "quismon" {
  api_key  = var.quismon_api_key  # Or use QUISMON_API_KEY env var
  base_url = "https://api.quismon.com"  # Optional, defaults to this
}
```

Environment variables:
- `QUISMON_API_KEY` - API key for authentication
- `QUISMON_BASE_URL` - API base URL (optional)

## Seamless Quickstart (Self-Service Signup)

The Quismon provider supports a **zero-configuration quickstart** for new users. If you don't have an API key yet, you can create a new organization directly from Terraform:

```hcl
terraform {
  required_providers {
    quismon = {
      source  = "quismon/quismon"
      version = "~> 1.0"
    }
  }
}

# Create a new organization (no API key needed!)
resource "quismon_signup" "main" {
  email    = "your-email@example.com"
  org_name = "My Organization"
}

# Create checks - provider automatically reads API key from state
resource "quismon_check" "website" {
  name             = "My Website"
  type             = "https"
  interval_seconds = 300
  enabled          = true

  config = {
    url             = "https://example.com"
    method          = "GET"
    expected_status = "200"
  }

  regions = ["us-east-1"]
}

output "api_key" {
  value     = quismon_signup.main.api_key
  sensitive = true
}
```

### How It Works

The provider automatically detects when:

1. No `api_key` is configured in the provider block
2. No `QUISMON_API_KEY` environment variable is set
3. A `quismon_signup` resource exists in the terraform state

When these conditions are met, it reads the API key directly from the state file, allowing seamless operation after the initial signup.

### Usage Flow

```bash
# Phase 1: Create signup (first time only)
terraform apply -target=quismon_signup.main -auto-approve

# Phase 2: Create all resources (no API key needed!)
terraform apply -auto-approve
```

When reading from state, you'll see a helpful warning:

```
│ Warning: API Key Read from Terraform State
│
│ Using API key from terraform state (.../terraform.tfstate).
│
│ 💡 Tip: For future runs, export the key:
│    export QUISMON_API_KEY=$(terraform output -raw api_key)
│
│ 📧 Verify your email to increase your free tier check frequency from 20/hr to 60/hr!
```

### Best Practices for CI/CD

For automated environments, extract the API key after initial signup:

```bash
# After initial signup
export QUISMON_API_KEY=$(terraform output -raw api_key)

# Store in your secrets manager (GitHub Actions, Vault, etc.)
# Subsequent runs will use the environment variable
terraform apply -auto-approve  # No warning when key is from env
```

### Destroy Considerations

When destroying resources, export the API key first to ensure all resources can be deleted:

```bash
export QUISMON_API_KEY=$(terraform output -raw api_key)
terraform destroy
```

## Quick Start Example

```hcl
terraform {
  required_providers {
    quismon = {
      source  = "quismon/quismon"
      version = "~> 1.0"
    }
  }
}

provider "quismon" {
  api_key = var.quismon_api_key
}

# Create a Slack notification channel
resource "quismon_notification_channel" "slack" {
  name = "Slack Alerts"
  type = "slack"

  config = {
    webhook_url = var.slack_webhook_url
  }
}

# Monitor production API
resource "quismon_check" "production_api" {
  name             = "Production API Health"
  type             = "https"
  interval_seconds = 60
  enabled          = true

  regions = ["us-east-1", "eu-west-1"]

  config = {
    url                  = "https://api.example.com/health"
    method               = "GET"
    expected_status_code = "200"
    timeout_seconds      = "10"
  }
}

# Alert when API is down
resource "quismon_alert_rule" "api_down" {
  check_id = quismon_check.production_api.id
  name     = "Production API Down"
  enabled  = true

  condition = {
    health_status = "down"
  }

  notification_channel_ids = [
    quismon_notification_channel.slack.id
  ]
}

# Alert on 3 consecutive failures
resource "quismon_alert_rule" "consecutive_failures" {
  check_id = quismon_check.production_api.id
  name     = "API Consecutive Failures"
  enabled  = true

  condition = {
    failure_threshold = 3
  }

  notification_channel_ids = [
    quismon_notification_channel.slack.id
  ]
}
```

## Exporting Existing Resources

If you've created checks through the dashboard or API, you can export them to Terraform configuration for version control and infrastructure-as-code management.

### Using the Export API

```bash
# Export your organization's configuration as Terraform HCL
curl -H "Authorization: Bearer YOUR_API_KEY" \
  https://api.quismon.com/v1/exports/terraform > main.tf

# Or export as JSON for processing
curl -H "Authorization: Bearer YOUR_API_KEY" \
  https://api.quismon.com/v1/exports/json > config.json
```

### Importing Existing Resources

The export includes import blocks, making it easy to bring existing resources under Terraform management:

```bash
# 1. Export your configuration
curl -H "Authorization: Bearer $QUISMON_API_KEY" \
  https://api.quismon.com/v1/exports/terraform > main.tf

# 2. Create a variables file
cat > terraform.tfvars << EOF
base_url = "https://api.quismon.com"
api_key  = "YOUR_API_KEY"
EOF

# 3. Initialize and import
terraform init
terraform apply  # Imports all existing resources

# 4. Verify - should show "No changes"
terraform plan
```

### What's Included in the Export

| Resource | Details |
|----------|---------|
| **Checks** | All check types with full configuration |
| **Notification Channels** | Email, Slack, ntfy, webhook configs |
| **Alert Rules** | Conditions and channel associations |
| **Outputs** | Resource IDs for reference |
| **Import Blocks** | Ready for Terraform 1.5+ import |

### Transitioning from Dashboard to IaC

The export enables a smooth transition from manual/dashboard management to Infrastructure as Code:

1. **Export** - Capture your current setup
2. **Version Control** - Commit the generated Terraform files
3. **Iterate** - Make changes through Terraform, not the dashboard
4. **Collaborate** - Use pull requests for monitoring changes

See [docs/VIBECODING-TO-IAC.md](../docs/VIBECODING-TO-IAC.md) for a complete guide.

## Check Types

### HTTP/HTTPS Check

```hcl
resource "quismon_check" "web_app" {
  name             = "Web Application"
  type             = "https"
  interval_seconds = 60
  enabled          = true

  regions = ["us-east-1", "eu-west-1", "ap-southeast-1"]

  config = {
    url                  = "https://www.example.com"
    method               = "GET"
    expected_status_code = "200"
    timeout_seconds      = "15"
    headers              = jsonencode({
      "User-Agent" = "Quismon-Monitor/1.0"
      "X-API-Key"  = var.api_key
    })
  }
}
```

#### HTTP Check with Body Assertions

Validate response body content:

```hcl
resource "quismon_check" "api_health" {
  name             = "API Health Check"
  type             = "https"
  interval_seconds = 60

  config = {
    url                  = "https://api.example.com/health"
    method               = "GET"
    expected_status_code = "200"
    expected_content     = "status\":\"healthy"  # String that must be in response
    content_match_type   = "contains"             # contains, exact, regex, or not_contains
  }

  regions = ["us-east-1"]
}

# Regex validation example
resource "quismon_check" "api_version" {
  name             = "API Version Check"
  type             = "https"
  interval_seconds = 300

  config = {
    url                = "https://api.example.com/version"
    expected_content   = "\"version\":\"v[0-9]+\\.[0-9]+\\.[0-9]+\""
    content_match_type = "regex"
  }

  regions = ["us-east-1"]
}

# Inverted check - alert if private endpoint becomes public
resource "quismon_check" "private_endpoint_exposure" {
  name             = "Admin Panel Should Not Be Public"
  type             = "https"
  interval_seconds = 300

  config = {
    url                = "https://internal.example.com/admin"
    expected_status    = "401,403,404"  # Any of these is good
    expected_content   = "Not Found"
    content_match_type = "not_contains"  # Fail if content IS found
  }

  regions = ["us-east-1"]
}
```

#### Multi-Step Checks

Chain multiple HTTP requests in a single check (for auth flows, API workflows):

```hcl
resource "quismon_check" "login_flow" {
  name             = "Login and API Access Flow"
  type             = "multistep"
  interval_seconds = 300

  config_json = jsonencode({
    steps = [
      {
        name   = "Get CSRF Token"
        type   = "http"
        url    = "https://api.example.com/auth/csrf"
        method = "GET"
        extract = {
          csrf_token = "$.token"
        }
      },
      {
        name    = "Login"
        type    = "http"
        url     = "https://api.example.com/auth/login"
        method  = "POST"
        headers = {
          "Content-Type" = "application/json"
        }
        body = jsonencode({
          email    = "test@example.com"
          password = "test-password"
          csrf     = "${steps.0.extract.csrf_token}"
        })
        extract = {
          access_token = "$.access_token"
        }
      },
      {
        name    = "Access Protected Resource"
        type    = "http"
        url     = "https://api.example.com/user/profile"
        method  = "GET"
        headers = {
          "Authorization" = "Bearer ${steps.1.extract.access_token}"
        }
        expected_status = [200]
      }
    ]
    fail_fast        = true
    timeout_seconds  = 30
  })

  regions = ["us-east-1"]
}
```

### TCP Port Check

```hcl
resource "quismon_check" "redis_server" {
  name             = "Redis Cache"
  type             = "tcp"
  interval_seconds = 90

  config = {
    host            = "cache.example.com"
    port            = "6379"
    timeout_seconds = "5"
  }

  regions = ["us-east-1"]
}
```

### Ping Check

```hcl
resource "quismon_check" "dns_server" {
  name             = "Primary DNS"
  type             = "ping"
  interval_seconds = 180

  config = {
    host            = "8.8.8.8"
    timeout_seconds = "3"
    packet_count    = "4"
  }

  regions = ["us-east-1"]
}
```

### DNS Check

```hcl
resource "quismon_check" "dns_a_record" {
  name             = "Main Domain A Record"
  type             = "dns"
  interval_seconds = 300

  config = {
    domain       = "example.com"
    record_type  = "A"
    expected_ips = jsonencode(["93.184.216.34"])
  }

  regions = ["us-east-1"]
}

# DNS MX Record Check
resource "quismon_check" "dns_mx_record" {
  name             = "Mail Exchange Records"
  type             = "dns"
  interval_seconds = 300

  config = {
    domain      = "example.com"
    record_type = "MX"
  }

  regions = ["us-east-1"]
}

# DNS TXT Record Check
resource "quismon_check" "dns_spf" {
  name             = "SPF Record"
  type             = "dns"
  interval_seconds = 3600

  config = {
    domain      = "example.com"
    record_type = "TXT"
  }

  regions = ["us-east-1"]
}
```

**DNS Record Types**: `A`, `AAAA`, `CNAME`, `MX`, `TXT`, `NS`, `SOA`

### HTTP/3 (QUIC) Check

Monitor endpoints that support HTTP/3 protocol:

```hcl
resource "quismon_check" "http3" {
  name             = "HTTP/3 Endpoint"
  type             = "http3"
  interval_seconds = 60

  config_json = jsonencode({
    url              = "https://cloudflare.com"
    method           = "GET"
    expected_status  = [200, 301, 302]
    timeout_seconds  = 10
  })

  regions = ["na-east-ewr"]
  enabled = true
}

# HTTP/3 with content validation
resource "quismon_check" "http3_api" {
  name             = "HTTP/3 API Health"
  type             = "http3"
  interval_seconds = 120

  config_json = jsonencode({
    url                = "https://api.example.com/health"
    method             = "GET"
    expected_status    = [200]
    expected_content   = "\"status\":\"ok\""
    content_match_type = "contains"
    timeout_seconds    = 10
  })

  regions = ["na-east-ewr", "eu-west-ams"]
  enabled = true
}
```

### Throughput Check

Measure download bandwidth:

```hcl
resource "quismon_check" "throughput" {
  name             = "CDN Throughput"
  type             = "throughput"
  interval_seconds = 300

  config_json = jsonencode({
    url             = "https://speed.cloudflare.com/__down?bytes=10000000"
    max_size_mb     = 5  # Tier-limited: Free=5, Paid=100, Enterprise=500
    timeout_seconds = 30
  })

  regions = ["na-east-ewr", "eu-west-ams"]
  enabled = true
}
```

**Tier Limits for Throughput**:
- Free tier: 5 MB max download
- Paid tier: 100 MB max download
- Enterprise: 500 MB max download

### SMTP/IMAP Check

End-to-end email delivery testing:

```hcl
resource "quismon_check" "smtp_imap" {
  name             = "Email Delivery Test"
  type             = "smtp-imap"
  interval_seconds = 300

  config_json = jsonencode({
    smtp_host        = "smtp.example.com"
    smtp_port        = 587
    smtp_username    = "monitoring@example.com"
    smtp_password    = var.smtp_password
    smtp_use_tls     = true

    imap_host        = "imap.example.com"
    imap_port        = 993
    imap_use_tls     = true

    from_address     = "monitoring@example.com"
    to_address       = "inbox@example.com"
    subject          = "Quismon Email Test - {{message_id}}"
    body             = "This is an automated test email."

    timeout_seconds  = 30
    max_wait_seconds = 60
  })

  regions = ["na-east-ewr"]
  enabled = true
}
```

### SSL Certificate Check

```hcl
resource "quismon_check" "ssl_cert" {
  name             = "API Certificate Expiry"
  type             = "ssl"
  interval_seconds = 3600  # Check every hour

  config = {
    domain              = "api.example.com"
    port                = 443
    warn_days_remaining = 30
  }

  regions = ["us-east-1"]
}

# SSL Certificate with Fingerprint Validation
resource "quismon_check" "ssl_cert_fingerprint" {
  name             = "Critical API Certificate"
  type             = "ssl"
  interval_seconds = 3600

  config = {
    domain                    = "critical.example.com"
    port                      = 443
    warn_days_remaining       = 30
    expected_fingerprint_sha256 = "abc123def456..."  # SHA-256 fingerprint
  }

  regions = ["us-east-1"]
}

# SSL Certificate with SAN Domain Validation
resource "quismon_check" "ssl_cert_san" {
  name             = "Multi-Domain Certificate"
  type             = "ssl"
  interval_seconds = 3600

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

  regions = ["us-east-1"]
}
```

## Inverted Checks (Security Monitoring)

Inverted checks alert when something that **shouldn't** be accessible **is**. Perfect for security posture monitoring.

### Alert When Port is Open

```hcl
# Alert if database port becomes publicly accessible
resource "quismon_check" "db_not_public" {
  name             = "Database Should Not Be Public"
  type             = "tcp"
  interval_seconds = 300
  enabled          = true

  config = {
    host            = "db.example.com"
    port            = "5432"
    timeout_seconds = "5"
    invert          = "true"  # Fail if connection succeeds
  }

  regions = ["us-east-1"]
}
```

### Alert When Private Page Becomes Public

```hcl
# Alert if admin page returns 200 (should return 401/403)
resource "quismon_check" "admin_not_public" {
  name             = "Admin Panel Should Not Be Public"
  type             = "https"
  interval_seconds = 300

  config = {
    url             = "https://app.example.com/admin"
    expected_status = "401,403,404"  # Any auth error or not found is good
  }

  regions = ["us-east-1"]
}
```

## Monitoring Regions

Quismon monitors from 31 regions across 6 continents via Vultr:

**North America:** `na-east-ewr` (NYC), `na-west-sjc` (San Jose), `na-west-lax` (LA), `na-west-sea` (Seattle), `na-central-dfw` (Dallas), `na-central-ord` (Chicago), `na-east-mia` (Miami), `na-east-atl` (Atlanta), `na-east-yto` (Toronto), `na-central-mex` (Mexico City)

**Europe:** `eu-west-ams` (Amsterdam), `eu-west-lhr` (London), `eu-west-man` (Manchester), `eu-central-fra` (Frankfurt), `eu-west-cdg` (Paris), `eu-south-mad` (Madrid), `eu-north-waw` (Warsaw), `eu-north-sto` (Stockholm)

**Asia Pacific:** `ap-northeast-nrt` (Tokyo), `ap-northeast-itm` (Osaka), `ap-northeast-icn` (Seoul), `ap-southeast-sin` (Singapore), `ap-south-bom` (Mumbai), `ap-south-del` (Delhi), `ap-south-blr` (Bangalore), `ap-west-tlv` (Tel Aviv)

**Australia:** `au-southeast-syd` (Sydney), `au-south-mel` (Melbourne)

**South America:** `sa-east-sao` (São Paulo), `sa-west-scl` (Santiago)

**Africa:** `af-south-jnb` (Johannesburg)

### Query Available Regions

```hcl
data "quismon_regions" "all" {}

output "available_regions" {
  value = [for r in data.quismon_regions.all.regions : "${r.code} - ${r.display_name}"]
}
```

## Alert Rule Conditions

Alert rules use a flexible `condition` map that supports different trigger types:

### Health Status Condition

Triggers when the health status changes:

```hcl
resource "quismon_alert_rule" "downtime" {
  check_id = quismon_check.production_api.id
  name     = "Service Down"
  enabled  = true

  condition = {
    health_status = "down"
  }

  notification_channel_ids = [quismon_notification_channel.slack.id]
}
```

### Failure Threshold Condition

Triggers after N consecutive failures:

```hcl
resource "quismon_alert_rule" "failures" {
  check_id = quismon_check.production_api.id
  name     = "Multiple Failures"
  enabled  = true

  condition = {
    failure_threshold = 3
  }

  notification_channel_ids = [quismon_notification_channel.slack.id]
}
```

### Response Time Condition

Triggers when response time exceeds threshold (milliseconds):

```hcl
resource "quismon_alert_rule" "latency" {
  check_id = quismon_check.production_api.id
  name     = "High Latency"
  enabled  = true

  condition = {
    response_time_ms = 2000  # 2 seconds
  }

  notification_channel_ids = [quismon_notification_channel.slack.id]
}
```

## Notification Channels

### Email

```hcl
resource "quismon_notification_channel" "email" {
  name = "Engineering Team"
  type = "email"

  config = {
    to = jsonencode(["dev@example.com", "ops@example.com"])
  }

  enabled = true
}
```

### ntfy

```hcl
resource "quismon_notification_channel" "ntfy" {
  name = "Mobile Notifications"
  type = "ntfy"

  config = {
    topic  = "quismon-alerts"
    server = "https://ntfy.sh"  # Optional, defaults to ntfy.sh
  }

  enabled = true
}
```

### Webhook

```hcl
resource "quismon_notification_channel" "webhook" {
  name = "Custom Webhook"
  type = "webhook"

  config = {
    url    = "https://your-service.com/webhooks/alerts"
    method = "POST"
  }

  enabled = true
}
```

### Slack

```hcl
resource "quismon_notification_channel" "slack" {
  name = "Slack #alerts"
  type = "slack"

  config = {
    webhook_url = var.slack_webhook_url
  }

  enabled = true
}
```

## Resource Reference

### quismon_signup

Creates a new Quismon organization via self-service signup. No API key is required for this resource.

#### Arguments

| Argument | Type | Required | Description |
|----------|------|----------|-------------|
| `email` | String | Yes | Email address for the organization |
| `org_name` | String | Yes | Name of the organization |

#### Attributes

| Attribute | Type | Description |
|-----------|------|-------------|
| `id` | String | Organization ID |
| `org_id` | String | Organization ID (same as id) |
| `org_name` | String | Organization name |
| `email` | String | Email address |
| `api_key` | String | API key for the organization (sensitive) |
| `verification_required` | Boolean | Whether email verification is required |

#### Example

```hcl
resource "quismon_signup" "main" {
  email    = "admin@company.com"
  org_name = "Acme Corp"
}

output "api_key" {
  value     = quismon_signup.main.api_key
  sensitive = true
}
```

### quismon_check

#### Arguments

| Argument | Type | Required | Description |
|----------|------|----------|-------------|
| `name` | String | Yes | Check name |
| `type` | String | Yes | Check type: `http`, `https`, `tcp`, `ping`, `dns`, or `ssl` |
| `config` | Map | Yes | Check-specific configuration (see examples above) |
| `interval_seconds` | Number | Yes | Check interval in seconds (minimum 60) |
| `regions` | List | No | Monitoring regions (default: `["us-east-1"]`) |
| `enabled` | Boolean | No | Whether check is enabled (default: `true`) |

#### Attributes

| Attribute | Type | Description |
|-----------|------|-------------|
| `id` | String | Check ID |
| `org_id` | String | Organization ID |
| `health_status` | String | Current health: `healthy`, `unhealthy`, or `unknown` |
| `last_checked` | String | Last check timestamp |
| `created_at` | String | Creation timestamp |
| `updated_at` | String | Last update timestamp |

### quismon_alert_rule

#### Arguments

| Argument | Type | Required | Description |
|----------|------|----------|-------------|
| `check_id` | String | Yes | ID of the check to monitor |
| `name` | String | Yes | Alert rule name |
| `condition` | Map | Yes | Condition that triggers the alert (see conditions above) |
| `notification_channel_ids` | List | Yes | List of notification channel IDs |
| `enabled` | Boolean | No | Whether rule is enabled (default: `true`) |

#### Attributes

| Attribute | Type | Description |
|-----------|------|-------------|
| `id` | String | Alert rule ID |
| `created_at` | String | Creation timestamp |
| `updated_at` | String | Last update timestamp |

### quismon_notification_channel

#### Arguments

| Argument | Type | Required | Description |
|----------|------|----------|-------------|
| `name` | String | Yes | Channel name |
| `type` | String | Yes | Channel type: `email`, `webhook`, `ntfy`, or `slack` |
| `config` | Map | Yes | Channel-specific configuration (see examples above) |
| `enabled` | Boolean | No | Whether channel is enabled (default: `true`) |

#### Attributes

| Attribute | Type | Description |
|-----------|------|-------------|
| `id` | String | Channel ID |
| `org_id` | String | Organization ID |
| `created_at` | String | Creation timestamp |
| `updated_at` | String | Last update timestamp |

## Data Sources

### quismon_check

Query a specific check by name:

```hcl
data "quismon_check" "prod_api" {
  name = "Production API"
}

output "api_health" {
  value = data.quismon_check.prod_api.health_status
}
```

### quismon_checks

Query all checks:

```hcl
data "quismon_checks" "all" {}

output "check_count" {
  value = length(data.quismon_checks.all.checks)
}

output "all_check_ids" {
  value = [for c in data.quismon_checks.all.checks : c.id]
}
```

### quismon_notification_channel

Query a specific channel by name:

```hcl
data "quismon_notification_channel" "slack" {
  name = "Slack Alerts"
}

output "slack_channel_id" {
  value = data.quismon_notification_channel.slack.id
}
```

## Import

Existing resources can be imported using their ID:

```bash
terraform import quismon_check.prod_api 550e8400-e29b-41d4-a716-446655440000
terraform import quismon_alert_rule.api_down 660e8400-e29b-41d4-a716-446655440000
terraform import quismon_notification_channel.email 770e8400-e29b-41d4-a716-446655440000
```

## Examples

See the [examples/](examples/) directory for complete working examples:

- [examples/basic/](examples/basic/) - Simple HTTPS check with email alerts
- [examples/multi-region/](examples/multi-region/) - Multi-region monitoring setup
- [examples/complete/](examples/complete/) - Full stack with checks, alerts, and notifications
- [examples/dns-ssl/](examples/dns-ssl/) - DNS and SSL certificate monitoring
- [examples/multistep/](examples/multistep/) - Multi-step workflow testing
- [examples/advanced-checks/](examples/advanced-checks/) - HTTP/3, Throughput, and advanced check types

## Development

### Building

```bash
go build -o terraform-provider-quismon
```

### Testing

```bash
# Unit tests
go test ./...

# Acceptance tests (requires API key)
TF_ACC=1 QUISMON_API_KEY=qm_your_key go test ./... -v
```

### Local Development

Create `~/.terraformrc`:

```hcl
provider_installation {
  dev_overrides {
    "quismon/quismon" = "/path/to/terraform-provider-quismon"
  }
  direct {}
}
```

Then run Terraform commands normally - it will use your local build.

## Support

- **Documentation**: https://registry.terraform.io/providers/quismon/quismon/latest/docs
- **Issues**: https://github.com/quismon/terraform-provider-quismon/issues
- **API Docs**: https://api.quismon.com/docs

## License

MIT License - see [LICENSE](LICENSE) for details.
