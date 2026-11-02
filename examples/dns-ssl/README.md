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
| `hostname` | Domain to fetch the certificate from (required) |
| `port` | Port to connect to (default: 443) |
| `warning_days` | Days before expiry to trigger warning |
| `expected_fingerprint` | Verify exact certificate (SHA-256) |
| `expected_domains` | Validate Subject Alternative Names |

---

## EGM Contract

### Interface (main.tf)

- `quismon_notification_channel.slack`
- Checks: `dns_a_record` (dns A + `expected_ips` jsonencode), `dns_mx_record` (dns MX), `dnssec_signed` (dnssec, `require_signed = true`), `dnssec_custom_ns` (dnssec + `nameservers`), `ssl_cert` (ssl, `warning_days = 30`), `ssl_cert_fingerprint` (`expected_fingerprint`), `ssl_cert_san` (`expected_domains` jsonencode)
- Alert rules: `dns_failure`, `ssl_expiring` (both `health_status = "down"`)
- Outputs: `dns_check_id`, `ssl_check_id`, `ssl_cert_health`

### Invariants

- dnssec and array-valued configs use `config_json`; flat ssl configs use the `config` map (`main.tf`)
- Fingerprint value is a placeholder (`abc123def456...`) — must be replaced with a real SHA-256 (`main.tf`)

### Side Effects

- Applying creates DNS/DNSSEC/SSL checks (hourly SSL cadence) and 2 alert rules

### Verification

- No test manifest in this directory; inherits the provider module's verification: `cd /workspace/terraform-provider-quismon && CGO_ENABLED=0 go test ./...`
