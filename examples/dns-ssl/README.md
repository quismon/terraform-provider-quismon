# DNS & SSL Example

Monitor DNS records, DNSSEC validation, and SSL certificates:

## Resources Created

### DNS Checks
- **A Record** - Verify domain resolves to expected IP
- **MX Record** - Monitor mail exchange records

### DNSSEC Checks
- **Signed Domain Validation** - Verify domain is DNSSEC-signed
- **Custom Nameservers** - Query specific DNS servers for DNSSEC

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

## DNSSEC Check Options

| Option | Description |
|--------|-------------|
| `domain` | Domain to check for DNSSEC |
| `record_type` | DNS record type to verify (default: A) |
| `require_signed` | Fail if domain is not DNSSEC-signed (default: true) |
| `nameservers` | Optional list of DNS servers to query (e.g., ["8.8.8.8"]) |
| `timeout_seconds` | Query timeout (default: 10) |

## SSL Check Options

| Option | Description |
|--------|-------------|
| `warn_days_remaining` | Days before expiry to trigger warning |
| `expected_fingerprint_sha256` | Verify exact certificate |
| `expected_domains` | Validate Subject Alternative Names |
