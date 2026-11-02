# Multi-Step Check Example

This example demonstrates Quismon's multi-step check feature for monitoring complex workflows.

## What are Multi-Step Checks?

Multi-step checks allow you to execute multiple checks in sequence, with the ability to:

- **Extract values** from one step's response (using JSONPath, regex, or headers)
- **Interpolate variables** into subsequent steps using `{{variable_name}}` syntax
- **Chain API calls** to simulate real user workflows
- **Aggregate results** across all steps

## Use Cases

1. **Login Flows** - Authenticate, then access protected resources
2. **API Chains** - Get data from one endpoint, use it in another
3. **E-commerce** - Add to cart, checkout, verify order
4. **Health Chains** - Check multiple endpoints in sequence

## Examples in this File

### Example 1: API Login Flow
A typical authentication flow:
1. POST to `/auth/login` to get a token
2. Extract `token` from response using `$.token` JSONPath
3. Use `{{auth_token}}` in Authorization header for subsequent requests

### Example 2: E-commerce Checkout Flow
Simulates a shopping workflow:
1. Get product details and extract price
2. Add product to cart with extracted price
3. Verify cart contents

### Example 3: API Health Chain (Continue-on-Failure)
Checks multiple health endpoints:
- Uses `fail_fast = false` to continue even if one step fails
- Useful for checking all endpoints and reporting all failures

## Key Configuration Options

```hcl
config_json = jsonencode({
  steps = [
    {
      name = "Step Name"
      type = "https"  # Can be: http, https, ping, tcp, udp, dns, ssl

      config = {
        url             = "https://api.example.com/endpoint"
        method          = "GET"
        headers         = { Authorization = "Bearer {{token}}" }
        expected_status = [200]
      }

      extracts = {
        variable_name = {
          jsonpath = "$.path.to.value"  # Extract via JSONPath
          # OR
          regex    = "token: ([a-z0-9]+)"  # Extract via regex
          # OR
          header   = "X-Auth-Token"  # Extract from response header
          default  = "fallback"  # Optional default if extraction fails
        }
      }
    }
  ]

  fail_fast       = true   # Stop on first failure (default: true)
  timeout_seconds = 30     # Total timeout for all steps (default: 30)
})
```

## Usage

```bash
# Set your API key
export QUISMON_API_KEY="your-api-key"

# Initialize and apply
terraform init
terraform apply
```

## Tier Limits

Maximum steps per multi-step check:
- **Free tier**: 5 steps
- **Paid tier**: 20 steps
- **Enterprise**: 20 steps

---

## EGM Contract

### Interface (main.tf)

- `quismon_notification_channel.ops_team` (email)
- Checks (all `type = "multistep"`, `config_json`): `api_login_flow` (login → extract `auth_token` via `$.token` → Bearer profile GET), `checkout_flow` (product → extract `product_price` → cart POST with `{{product_price}}` → verify), `api_health_chain` (3 endpoints, `fail_fast = false`)
- `quismon_alert_rule.api_flow_down` (`failure_threshold = 2`)
- Outputs: `login_flow_check_id`, `checkout_flow_check_id`

### Invariants

- Per-step `timeout_seconds` (max time per step) vs top-level `timeout_seconds` (total budget) are distinct fields (`main.tf` comments)
- `fail_fast = false` continues past failures to report all endpoint results (`main.tf` Example 3)
- Login body embeds placeholder credentials (`monitoring@example.com` / `secret`) — replace before real use (`main.tf` Example 1)

### Side Effects

- Applying creates checks that POST logins and cart operations against example.com endpoints every 5–10 minutes

### Verification

- No test manifest in this directory; inherits the provider module's verification: `cd /workspace/terraform-provider-quismon && CGO_ENABLED=0 go test ./...`
