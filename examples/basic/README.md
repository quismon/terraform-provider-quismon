# Basic Example

This example creates a simple monitoring setup with:
- One HTTPS check monitoring a website
- One email notification channel
- One alert rule (3 consecutive failures trigger)

## Usage

```bash
# Set your API key
export QUISMON_API_KEY="your-api-key-here"

# Initialize Terraform
terraform init

# Plan the changes
terraform plan

# Apply the changes
terraform apply
```

## Requirements

- Terraform >= 1.0
- Quismon API key with write permissions

---

## EGM Contract

### Interface (main.tf)

- `quismon_notification_channel.email` (type `email`, config key `email` = recipient address string)
- `quismon_check.website` (type `https`, region `na-east-ewr`, `config` map with url/method/expected_status_code/timeout_seconds)
- `quismon_alert_rule.website_down` (`failure_threshold = 3`)
- Outputs: `check_id`, `health_status`

### Invariants

- Simple flat configs use the `config` map with string values; `expected_status_code = "200"` is a string (`main.tf`)
- Region keys are the platform's current region codes (`GET /v1/regions`)

### Side Effects

- Applying creates a real check (60s interval), channel, and alert rule against `www.example.com` (placeholder — replace before applying)

### Verification

- No test manifest in this directory; inherits the provider module's verification: `cd /workspace/terraform-provider-quismon && CGO_ENABLED=0 go test ./...`
