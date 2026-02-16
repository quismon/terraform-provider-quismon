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
- **Paid tier**: 10 steps
- **Enterprise**: 20 steps
