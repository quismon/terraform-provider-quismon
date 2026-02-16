# DNS & SSL Example

Monitor DNS records and SSL certificates:

## Resources Created

### DNS Checks
- **A Record** - Verify domain resolves to expected IP
- **MX Record** - Monitor mail exchange records

### SSL Checks
- **Certificate Expiry** - Alert when cert expires within 30 days
- **Fingerprint Validation** - Verify certificate hasn't changed
- **SAN Validation** - Ensure multi-domain certs cover all expected domains

## Usage

```bash
# Create terraform.tfvars
cat > terraform.tfvars << EOF
quismon_api_key   = "your-api-key-here"
slack_webhook_url = "https://hooks.slack.com/services/XXX/YYY/ZZZ"
EOF

terraform init
terraform apply
```

## SSL Check Options

| Option | Description |
|--------|-------------|
| `warn_days_remaining` | Days before expiry to trigger warning |
| `expected_fingerprint_sha256` | Verify exact certificate |
| `expected_domains` | Validate Subject Alternative Names |
