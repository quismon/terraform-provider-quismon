---
page_title: "Provider: Quismon"
description: |-
  The Quismon provider is used to interact with Quismon's external monitoring platform. It allows you to manage checks, alerts, and notifications as Infrastructure as Code.
---

# Quismon Provider

The Quismon provider is used to interact with [Quismon](https://quismon.com), an API-first external monitoring platform with 31 global regions. The provider allows you to manage checks, alert rules, and notification channels as Infrastructure as Code.

Use the navigation to the left to read about the available resources and data sources.

## Example Usage

### With Existing API Key

```terraform
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

# Monitor a website
resource "quismon_check" "website" {
  name             = "Production Website"
  type             = "https"
  interval_seconds = 300
  enabled          = true

  regions = ["na-east-ewr", "eu-central-fra"]

  config = {
    url             = "https://example.com"
    method          = "GET"
    expected_status = "200"
  }
}
```

### Seamless Quickstart (Self-Service Signup)

No API key? Create one directly from Terraform:

```terraform
terraform {
  required_providers {
    quismon = {
      source  = "quismon/quismon"
      version = "~> 1.0"
    }
  }
}

# Create a new organization - no API key needed!
resource "quismon_signup" "main" {
  email    = "your-email@example.com"
  org_name = "My Organization"
}

# Provider reads API key from state automatically
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

  regions = ["na-east-ewr"]
}

output "api_key" {
  value     = quismon_signup.main.api_key
  sensitive = true
}
```

Apply in two phases:

```bash
# Phase 1: Create signup (first time only)
terraform apply -target=quismon_signup.main -auto-approve

# Phase 2: Create all resources
terraform apply -auto-approve
```

## Requirements

- Terraform >= 1.0
- A Quismon API key (or use the [signup resource](#seamless-quickstart-self-service-signup))

## Authentication

The Quismon provider supports three ways to authenticate:

### 1. Provider Configuration

Explicitly set the API key in the provider block:

```terraform
provider "quismon" {
  api_key  = "your-api-key"
  base_url = "https://api.quismon.com"  # Optional
}
```

### 2. Environment Variables

Set credentials via environment variables:

```bash
export QUISMON_API_KEY="your-api-key"
export QUISMON_BASE_URL="https://api.quismon.com"  # Optional
```

```terraform
provider "quismon" {}  # Uses environment variables
```

### 3. Signup Resource (State-Based)

When no API key is configured and a `quismon_signup` resource exists, the provider automatically reads the key from Terraform state. See [Seamless Quickstart](#seamless-quickstart-self-service-signup).

~> **Important:** For CI/CD environments, extract the API key after initial signup and store it in your secrets manager:

```bash
export QUISMON_API_KEY=$(terraform output -raw api_key)
```

## Check Types

Quismon supports the following check types:

| Type | Description | Use Case |
|------|-------------|----------|
| `http` / `https` | HTTP/HTTPS requests | Web endpoints, APIs |
| `http3` | HTTP/3 over QUIC | Modern APIs, CDNs |
| `ping` | ICMP echo requests | Host availability |
| `tcp` | TCP port connectivity | Databases, cache servers |
| `dns` | DNS record queries | DNS configuration, propagation |
| `dnssec` | DNSSEC validation | DNS security verification |
| `ssl` | SSL/TLS certificate checks | Certificate expiry |
| `multistep` | Sequential workflows | User journeys, API chains |
| `throughput` | Download speed tests | Bandwidth monitoring |
| `smtp-imap` | Email delivery validation | Mail server health |

## Global Regions

Quismon operates 31 monitoring regions worldwide. Common regions include:

| Region Code | Location |
|-------------|----------|
| `na-east-ewr` | New Jersey, US |
| `na-west-lax` | Los Angeles, US |
| `eu-central-fra` | Frankfurt, DE |
| `eu-west-lhr` | London, UK |
| `ap-southeast-sin` | Singapore |
| `ap-northeast-nrt` | Tokyo, JP |

See the [regions data source](data-sources/regions.md) for the complete list.

## Resources

- [quismon_check](resources/check.md) - Create and manage health checks
- [quismon_alert_rule](resources/alert_rule.md) - Configure alert conditions
- [quismon_notification_channel](resources/notification_channel.md) - Set up notification channels
- [quismon_organization_otlp](resources/organization_otlp.md) - Configure OTLP metrics export
- [quismon_signup](resources/signup.md) - Create a new organization (self-service)

## Data Sources

- [quismon_check](data-sources/check.md) - Read a single check
- [quismon_checks](data-sources/checks.md) - List all checks
- [quismon_notification_channel](data-sources/notification_channel.md) - Read a notification channel
- [quismon_regions](data-sources/regions.md) - List available monitoring regions

## Examples

For complete working examples, see the [examples directory](https://github.com/quismon/terraform-provider-quismon/tree/main/examples):

- **basic** - Simple website monitoring
- **multi-region** - Checks from multiple locations
- **multistep-advanced** - Complex API workflows with token extraction
- **ssl-fingerprint** - Certificate pinning and monitoring
- **dns-custom-nameservers** - DNS checks with custom resolvers
- **full-stack** - Layered monitoring (DNS, SSL, health, API, database)

## Budget Calculation

Each regional check execution counts toward your hourly check budget:

**Formula:** `(3600 / interval_seconds) x number_of_regions = checks_per_hour`

| Regions | Interval | Checks/Hour |
|---------|----------|-------------|
| 1 | 60s | 60 |
| 3 | 60s | 180 |
| 5 | 60s | 300 |
| 3 | 300s | 36 |

Multi-step checks count each step separately. A 5-step check with 3 regions = 15 checks per execution.

## Argument Reference

The following arguments are supported in the provider block:

- `api_key` (String, Sensitive) - Quismon API key. Can also be set via the `QUISMON_API_KEY` environment variable.
- `base_url` (String) - Quismon API base URL. Defaults to `https://api.quismon.com`. Can also be set via the `QUISMON_BASE_URL` environment variable.

## Getting Help

- [Documentation](https://quismon.com/docs)
- [API Reference](https://quismon.com/swagger)
- [GitHub Repository](https://github.com/quismon/terraform-provider-quismon)
- [Terraform Registry](https://registry.terraform.io/providers/quismon/quismon)
