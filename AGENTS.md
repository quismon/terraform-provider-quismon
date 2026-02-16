# Terraform Provider for Quismon - AI Agent Guide

This document provides essential context for AI agents working on the Quismon Terraform provider.

---

## Component Overview

The **Terraform Provider** enables Infrastructure as Code management of Quismon resources. It:
- Implements `quismon_check`, `quismon_notification_channel`, and `quismon_alert_rule` resources
- Provides data sources for querying existing resources
- Integrates with Terraform CLI via plugin framework
- Communicates with Quismon API via HTTP client

**Language:** Go 1.24
**Framework:** terraform-plugin-framework v1.4.2

---

## Related Software Components

### Core Dependencies (go.mod)

| Library | Version | Purpose |
|---------|---------|---------|
| **terraform-plugin-framework** | v1.4.2 | Terraform plugin SDK |
| **terraform-plugin-go** | v0.29.0 | Terraform gRPC protocol |
| **terraform-plugin-testing** | v1.14.0 | Acceptance testing |

---

## Directory Structure

```
terraform-provider-quismon/
├── main.go                     # Provider entry point
├── go.mod / go.sum             # Go dependencies
├── internal/
│   ├── provider/
│   │   ├── provider.go         # Provider definition, schema
│   │   ├── check_resource.go           # quismon_check resource
│   │   ├── check_data_source.go        # quismon_check data source
│   │   ├── checks_data_source.go       # quismon_checks data source
│   │   ├── notification_channel_resource.go  # quismon_notification_channel
│   │   ├── notification_channel_data_source.go
│   │   ├── alert_rule_resource.go      # quismon_alert_rule resource
│   │   └── *_test.go           # Acceptance tests
│   └── client/
│       ├── client.go           # API client implementation
│       ├── checks.go           # Check API methods
│       ├── notification_channels.go
│       └── alert_rules.go
├── examples/                   # Example Terraform configurations
│   ├── basic/
│   ├── complete/
│   ├── multi-region/
│   └── dns-ssl/
├── README.md                   # Full documentation
├── Makefile                    # Build automation
└── scripts/                    # Utility scripts
```

---

## Architecture

```
┌─────────────────────────────────┐
│      Terraform CLI              │
│                                 │
│  terraform apply/plan/destroy   │
└───────────┬─────────────────────┘
            │ gRPC
            ↓
┌─────────────────────────────────┐
│   Terraform Provider            │
│   (Plugin Framework)            │
│                                 │
│  ┌──────────────────────────┐   │
│  │ Resources                │   │
│  │ • quismon_check          │   │
│  │ • quismon_notification_  │   │
│  │   channel                │   │
│  │ • quismon_alert_rule     │   │
│  └────────┬─────────────────┘   │
│           │                     │
│  ┌────────▼─────────────────┐   │
│  │ Client Layer             │   │
│  │ (HTTP API communication) │   │
│  └────────┬─────────────────┘   │
└───────────┼─────────────────────┘
            │ HTTPS
            ↓
┌─────────────────────────────────┐
│      Quismon API                │
│   https://api.quismon.com/v1    │
└─────────────────────────────────┘
```

---

## Provider Configuration

```go
type quismonProvider struct {
    version string
}

type quismonProviderModel struct {
    APIKey  types.String `tfsdk:"api_key"`
    BaseURL types.String `tfsdk:"base_url"`
}
```

### Provider Schema

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `api_key` | String | No* | API key (can use env var) |
| `base_url` | String | No | API base URL (defaults to api.quismon.com) |

*Required if `QUISMON_API_KEY` env var not set

### Environment Variables

```bash
export QUISMON_API_KEY="qm_xxxxx"
export QUISMON_BASE_URL="https://api.quismon.com"  # Optional
```

---

## Resources

### quismon_check

**File:** `internal/provider/check_resource.go`

#### Schema

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | String | Yes | Check name |
| `type` | String | Yes | http, https, tcp, ping, dns, ssl |
| `config` | Map | Yes | Check-specific configuration |
| `interval_seconds` | Number | Yes | Check interval (min 60) |
| `regions` | List | No | Monitoring regions (default: ["us-east-1"]) |
| `enabled` | Boolean | No | Whether check is enabled (default: true) |
| `id` | String | Computed | Check ID |
| `health_status` | String | Computed | Current health status |
| `last_checked` | String | Computed | Last check timestamp |

#### Example

```hcl
resource "quismon_check" "website" {
  name             = "Company Website"
  type             = "https"
  interval_seconds = 60
  enabled          = true

  regions = ["us-east-1", "eu-west-1"]

  config = {
    url                  = "https://www.example.com"
    method               = "GET"
    expected_status_code = "200"
    timeout_seconds      = "10"
  }
}
```

### quismon_notification_channel

**File:** `internal/provider/notification_channel_resource.go`

#### Schema

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | String | Yes | Channel name |
| `type` | String | Yes | email, webhook, ntfy, slack |
| `config` | Map | Yes | Channel-specific configuration |
| `enabled` | Boolean | No | Whether channel is enabled (default: true) |
| `id` | String | Computed | Channel ID |

#### Example

```hcl
resource "quismon_notification_channel" "slack" {
  name = "Slack Alerts"
  type = "slack"

  config = {
    webhook_url = var.slack_webhook_url
  }

  enabled = true
}
```

### quismon_alert_rule

**File:** `internal/provider/alert_rule_resource.go`

#### Schema

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `check_id` | String | Yes | ID of check to monitor |
| `name` | String | Yes | Alert rule name |
| `condition` | Map | Yes | Alert condition (see below) |
| `notification_channel_ids` | List | Yes | List of channel IDs |
| `enabled` | Boolean | No | Whether rule is enabled (default: true) |
| `id` | String | Computed | Alert rule ID |

#### Condition Map

| Condition Type | Example | Description |
|----------------|---------|-------------|
| `health_status` | `{health_status = "down"}` | Triggers on status change |
| `failure_threshold` | `{failure_threshold = 3}` | Triggers after N failures |
| `response_time_ms` | `{response_time_ms = 2000}` | Triggers on slow response |

#### Example

```hcl
resource "quismon_alert_rule" "api_down" {
  check_id = quismon_check.website.id
  name     = "Website Down"
  enabled  = true

  condition = {
    health_status = "down"
  }

  notification_channel_ids = [
    quismon_notification_channel.slack.id
  ]
}
```

---

## Data Sources

### quismon_check

Query a single check by name:

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
```

### quismon_notification_channel

Query a single channel by name:

```hcl
data "quismon_notification_channel" "slack" {
  name = "Slack Alerts"
}
```

---

## Client Layer

**File:** `internal/client/client.go`

### Client Structure

```go
type Client struct {
    baseURL    string
    apiKey     string
    httpClient *http.Client
}

func New(baseURL, apiKey string) (*Client, error) {
    // Validate and create client
}

func (c *Client) Do(req *http.Request, v interface{}) error {
    // Add auth headers, parse response
}
```

### API Methods

**checks.go:**
- `CreateCheck(req CreateCheckRequest) (*Check, error)`
- `GetCheck(id string) (*Check, error)`
- `ListChecks() ([]Check, error)`
- `UpdateCheck(id string, req UpdateCheckRequest) (*Check, error)`
- `DeleteCheck(id string) error`

**notification_channels.go:**
- `CreateNotificationChannel(req CreateNotificationChannelRequest) (*NotificationChannel, error)`
- `GetNotificationChannel(id string) (*NotificationChannel, error)`
- `ListNotificationChannels() ([]NotificationChannel, error)`
- `UpdateNotificationChannel(id string, req UpdateNotificationChannelRequest) (*NotificationChannel, error)`
- `DeleteNotificationChannel(id string) error`

**alert_rules.go:**
- `CreateAlertRule(checkID string, req CreateAlertRuleRequest) (*AlertRule, error)`
- `GetAlertRule(checkID, ruleID string) (*AlertRule, error)`
- `DeleteAlertRule(checkID, ruleID string) error`

---

## Resource Lifecycle

Each resource implements the `resource.Resource` interface:

```go
type Resource interface {
    Metadata(context.Context, MetadataRequest, *MetadataResponse)
    Schema(context.Context, SchemaRequest, *SchemaResponse)
    Configure(context.Context, ConfigureRequest, *ConfigureResponse)
    Create(context.Context, CreateRequest, *CreateResponse)
    Read(context.Context, ReadRequest, *ReadResponse)
    Update(context.Context, UpdateRequest, *UpdateResponse)
    Delete(context.Context, DeleteRequest, *DeleteResponse)
}
```

### Create Flow

```go
func (r *checkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    // 1. Get plan from Terraform
    var plan checkResourceModel
    req.Plan.Get(ctx, &plan)

    // 2. Convert to API request
    createReq := client.CreateCheckRequest{
        Name: plan.Name.ValueString(),
        // ...
    }

    // 3. Call API
    check, err := r.client.CreateCheck(createReq)
    if err != nil {
        resp.Diagnostics.AddError("Error Creating Check", err.Error())
        return
    }

    // 4. Update state with response
    plan.ID = types.StringValue(check.ID)
    plan.HealthStatus = types.StringValue(check.HealthStatus)

    resp.State.Set(ctx, plan)
}
```

---

## Import Functionality

Resources support importing existing resources:

```bash
terraform import quismon_check.example <check-id>
terraform import quismon_notification_channel.slack <channel-id>
```

```go
func (r *checkResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
    resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
```

---

## Testing

### Unit Tests

```bash
cd terraform-provider-quismon
go test ./...
```

### Acceptance Tests

```bash
# Set API credentials
export QUISMON_API_KEY="qm_xxxxx"

# Run acceptance tests
TF_ACC=1 go test ./internal/provider/... -v
```

### Test Structure

```go
func TestAccCheckResource(t *testing.T) {
    resource.Test(t, resource.TestCase{
        ProtoV5ProviderFactories: map[string]func() (proto.Provider, error){
            "quismon": func() (proto.Provider, error) {
                return NewProvider("test")(), nil
            },
        },
        Steps: []resource.TestStep{
            {
                Config: `
                    resource "quismon_check" "test" {
                        name = "Test Check"
                        type = "https"
                        config = { url = "https://example.com" }
                        interval_seconds = 60
                    }
                `,
                Check: resource.ComposeAggregateTestCheckFunc(
                    // Verify state...
                ),
            },
        },
    })
}
```

---

## Build & Install

### Build

```bash
make build
# or
go build -o terraform-provider-quismon
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

### Install for Terraform

```bash
mkdir -p ~/.terraform.d/plugins/registry.terraform.io/quismon/quismon/1.0.0/linux_amd64/
cp terraform-provider-quismon ~/.terraform.d/plugins/registry.terraform.io/quismon/quismon/1.0.0/linux_amd64/
```

---

## Common Tasks

### Add a New Check Type

1. Update `check_resource.go` type description
2. Add example in `examples/`
3. Update `README.md`

### Add a New Notification Channel Type

1. Update `notification_channel_resource.go` type description
2. Add example in `examples/`
3. Update `README.md`

### Add a New Alert Condition

1. Update `alert_rule_resource.go` condition description
2. Update examples

### Debug Provider Issues

Enable logging:

```bash
export TF_LOG=DEBUG
terraform apply
```

---

## Error Handling

Use diagnostics for errors:

```go
resp.Diagnostics.AddError(
    "Error Creating Check",
    fmt.Sprintf("Could not create check: %s", err.Error()),
)
```

Use warnings for non-fatal issues:

```go
resp.Diagnostics.AddWarning(
    "Deprecated field",
    "The 'foo' field is deprecated, use 'bar' instead",
)
```

---

## Related Components

- **quismon-api** - Backend API
- **Terraform Registry** - Where provider will be published
- **examples/** - Reference configurations
