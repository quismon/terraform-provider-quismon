# Complete Example

Production-ready monitoring setup demonstrating all Quismon features:

## Resources Created

### Notification Channels
- Email (multiple recipients)
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

The API check runs from both `us-east-1` and `eu-west-1` regions for global coverage.
