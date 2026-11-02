# Advanced Check Types

This example demonstrates the advanced check types available in Quismon:

- **HTTP/3 (QUIC)**: Tests endpoints supporting the HTTP/3 protocol
- **Throughput**: Measures download bandwidth

## Types in this Example

### HTTP/3 Check

HTTP/3 uses the QUIC protocol instead of TCP. This check verifies that your endpoints properly support HTTP/3 and can optionally validate response content.

Key configuration options:
- `url`: The HTTPS URL to check (HTTP/3 requires HTTPS)
- `method`: HTTP method (GET, POST, etc.)
- `expected_status`: List of acceptable status codes
- `expected_content`: Optional content to validate in response
- `content_match_type`: How to match content (contains, exact, regex, not_contains)

### Throughput Check

Measures download bandwidth from a URL. This is useful for:
- CDN performance monitoring
- Bandwidth benchmarking
- Network capacity verification

Key configuration options:
- `url`: URL to download for testing
- `max_size_mb`: Maximum download size (tier-limited)
- `timeout_seconds`: Download timeout

**Tier Limits**:
- Free: 5 MB maximum
- Paid: 100 MB maximum
- Enterprise: 500 MB maximum

## Usage

```bash
# Initialize
terraform init

# Plan
terraform plan

# Apply
terraform apply
```

## Requirements

| Name | Version |
|------|---------|
| terraform | >= 1.0 |
| quismon | ~> 1.0 |

## Providers

| Name | Version |
|------|---------|
| quismon | ~> 1.0 |

---

## EGM Contract

### Interface (main.tf)

- `quismon_check.throughput` — throughput check via `config_json` (10MB Cloudflare speed file, `max_size_mb = 5`, 2 regions, 300s interval)
- `quismon_check.http3` / `quismon_check.http3_with_content` — HTTP/3 checks with `expected_status` arrays and `expected_content` + `content_match_type = "contains"`
- `quismon_check.bandwidth_benchmark` — 1MB OVH test file, 3 regions, 600s interval
- Outputs: `throughput_check_id`, `http3_check_id`

### Invariants

- All configs use `config_json = jsonencode({...})` (numeric/array values) (`main.tf`)
- `max_size_mb` is tier-limited (comment: Free 5MB / Paid 100MB / Enterprise 500MB) (`main.tf`)

### Side Effects

- Applying creates checks that repeatedly download test files (bandwidth consumption on both ends)

### Verification

- No test manifest in this directory; inherits the provider module's verification: `cd /workspace/terraform-provider-quismon && CGO_ENABLED=0 go test ./...`
