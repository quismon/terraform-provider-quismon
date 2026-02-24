# Infosec Example: SSL Certificate Fingerprint Monitoring

This example demonstrates how to use Quismon for security-focused SSL/TLS certificate monitoring, including integration with Let's Encrypt certificate management.

## Use Cases

1. **Certificate Expiry Detection** - Alert before certificates expire
2. **Fingerprint Verification** (Certificate Pinning) - Detect unauthorized certificate changes
3. **Chain Validation** - Ensure proper certificate chain
4. **Multi-domain Monitoring** - Monitor SAN certificates

## Getting Certificate Fingerprints

### Using OpenSSL
```bash
# Get SHA-256 fingerprint from a live server
echo | openssl s_client -connect www.example.com:443 2>/dev/null | \
  openssl x509 -fingerprint -sha256 -noout | \
  sed 's/.*=//; s/://g' | tr '[:upper:]' '[:lower:]'
```

### Using Certbot (Let's Encrypt)
```bash
# List certificates with fingerprints
certbot certificates

# Get specific certificate fingerprint
openssl x509 -in /etc/letsencrypt/live/example.com/cert.pem \
  -fingerprint -sha256 -noout | sed 's/.*=//; s/://g' | tr '[:upper:]' '[:lower:]'
```

## Let's Encrypt Integration

### Automated Fingerprint Updates

Create a post-renewal hook for certbot:

```bash
#!/bin/bash
# /etc/letsencrypt/renewal-hooks/deploy/quismom-update.sh

DOMAIN="example.com"
CHECK_ID="your-check-uuid-here"

# Get new fingerprint
FINGERPRINT=$(openssl x509 -in /etc/letsencrypt/live/$DOMAIN/cert.pem \
  -fingerprint -sha256 -noout | sed 's/.*=//; s/://g' | tr '[:upper:]' '[:lower:]')

# Update Quismon check
curl -X PATCH "https://api.quismon.com/v1/checks/$CHECK_ID" \
  -H "Authorization: Bearer $QUISMON_API_KEY" \
  -H "Content-Type: application/json" \
  -d "{\"config\": {\"expected_fingerprint\": \"$FINGERPRINT\"}}"
```

### Terraform Automation

Store the fingerprint in a Terraform variable that can be updated:

```hcl
variable "ssl_fingerprint" {
  description = "Current SSL certificate SHA-256 fingerprint"
  type        = string
  default     = "a1b2c3d4e5f6..."  # Update after each renewal
}

resource "quismon_check" "ssl_fingerprint" {
  # ...
  config = {
    expected_fingerprint = var.ssl_fingerprint
    # ...
  }
}
```

## Security Best Practices

### 1. Monitor from Multiple Regions
Certificate changes should be detected from any region. A compromised CDN edge node might serve a different certificate.

### 2. Alert Immediately on Fingerprint Changes
A fingerprint change could indicate:
- Legitimate certificate renewal (expected)
- Man-in-the-middle attack (unexpected)
- CDN/proxy misconfiguration

### 3. Combine SSL with HTTPS Checks
Verify both the certificate AND the endpoint response in a multi-step check.

### 4. Set Appropriate Expiry Thresholds
- Production: 14-30 days before expiry
- Staging: 7 days
- Development: 3 days

### 5. Use Certificate Pinning for Critical Services
For banking, healthcare, or other sensitive applications, pin the exact certificate fingerprint.

## Alert Severity

| Condition | Severity | Action |
|-----------|----------|--------|
| Fingerprint changed | CRITICAL | Investigate immediately - possible attack |
| Expiring < 7 days | HIGH | Renew immediately |
| Expiring < 14 days | MEDIUM | Schedule renewal |
| Chain invalid | HIGH | Check intermediate certificates |
| SAN mismatch | MEDIUM | Certificate doesn't cover all domains |

## Files

- `main.tf` - Complete Terraform configuration with examples
- All check types: basic SSL, fingerprint pinning, multi-domain, chain validation
- Multi-step combined SSL + HTTPS check
- Alert rules and notification channels
