# Complete Example

Production-ready monitoring setup demonstrating all Quismon features:

## Resources Created

### Notification Channels
- Email (single recipient address)
- Slack incoming webhook
- ntfy.sh push notifications

### Checks
- **Production API** - HTTPS check with custom headers, multi-region
- **PostgreSQL Database** - TCP port check
- **Network Gateway** - ICMP ping check

### Alert Rules
- Health status alerts (triggers when service goes down)
- Failure threshold alerts (triggers after N consecutive failures)

## Usage

```bash
# Create terraform.tfvars with your credentials
cat > terraform.tfvars << EOF
quismon_api_key   = "your-api-key-here"
slack_webhook_url = "https://hooks.slack.com/services/XXX/YYY/ZZZ"
EOF

# Initialize and apply
terraform init
terraform plan
terraform apply
```

## Multi-Region

The API check runs from both `na-east-ewr` and `eu-west-ams` regions for global coverage.

---

## EGM Contract

### Interface (main.tf)

- Channels: `quismon_notification_channel.ops_email` (email, `email = "ops@example.com"`), `.slack`, `.ntfy`
- Checks: `quismon_check.api` (https with `headers` jsonencode, 2 regions), `.database` (tcp 5432), `.gateway` (ping)
- Alert rules: `api_down` (health_status), `api_consecutive_failures` (`failure_threshold = 3` + `severity = "critical"`), `db_down` (threshold 2), `cert_expiring_soon` (`ssl_days_remaining = "14"`, `severity = "warning"`, `notify_on = "unhealthy"`)
- Outputs: `api_check_id`, `api_health_status`, `all_checks`

### Invariants

- `ssl_days_remaining` fires while the check is still healthy (custom earlier threshold vs the check's own 30-day window) (`main.tf` comment)
- Variables `quismon_api_key` and `slack_webhook_url` are sensitive (`main.tf`)

### Side Effects

- Applying creates 3 channels, 3 checks, and 4 alert rules against example.com placeholders

### Verification

- No test manifest in this directory; inherits the provider module's verification: `cd /workspace/terraform-provider-quismon && CGO_ENABLED=0 go test ./...`
