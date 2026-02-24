# Advanced Multi-step Check Examples

This directory contains advanced multi-step check examples for real-world monitoring scenarios.

## Examples

### 1. OAuth2 Client Credentials Flow
Monitors an API that requires OAuth2 authentication:
- Gets an access token using client credentials
- Uses the token to call a protected API
- Verifies the API health

### 2. GraphQL API Workflow
Monitors GraphQL APIs with query chaining:
- Executes a GraphQL query to get user data
- Uses extracted data to query related organization data
- Demonstrates JSON body construction for GraphQL

### 3. Webhook Signature Verification
Tests webhook endpoints with signed payloads:
- Gets a timestamp from the server
- Sends a signed test webhook
- Verifies the webhook was received

### 4. Paginated API Flow
Monitors paginated REST APIs:
- Fetches the first page of results
- Extracts pagination info
- Fetches subsequent pages

### 5. Circuit Breaker Pattern
Implements a circuit breaker pattern:
- Checks primary API health
- Falls back to secondary if primary fails
- Uses `fail_fast: true` for quick alerts

### 6. HTTP/3 API Workflow
Monitors HTTP/3 (QUIC) endpoints:
- All steps use HTTP/3 protocol
- Tests the same workflow over QUIC
- Useful for CDNs and modern APIs

## Key Concepts

### Variable Extraction
Extract values from responses using JSONPath:
```hcl
extracts = {
  auth_token = {
    jsonpath = "$.access_token"
  }
}
```

### Header Extraction
Extract response headers:
```hcl
extracts = {
  rate_limit = {
    header = "X-RateLimit-Remaining"
  }
}
```

### Variable Interpolation
Use extracted variables in subsequent steps:
```hcl
headers = {
  Authorization = "Bearer {{auth_token}}"
}
```

### Fail Fast
Stop execution on first failure:
```hcl
fail_fast = true  # Stop immediately on failure
fail_fast = false # Continue all steps, report all failures
```

## Usage

```bash
# Initialize
terraform init

# Plan
terraform plan

# Apply
terraform apply
```

## Best Practices

1. **Use `fail_fast: true`** for critical flows that should alert immediately
2. **Use `fail_fast: false`** when you want to see all step failures
3. **Set appropriate timeouts** - sum of step timeouts should be less than total
4. **Extract and verify** - don't just check status codes, verify content
5. **Use multiple regions** for critical checks to detect regional issues
