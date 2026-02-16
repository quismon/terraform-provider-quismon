## Terraform Provider Testing Guide

## Overview

The Quismon Terraform provider includes comprehensive test coverage with unit tests, acceptance tests, matrix tests, and integration tests.

## Test Structure

```
internal/provider/
├── provider_test.go                    # Provider configuration tests
├── check_resource_test.go              # Check resource CRUD tests
├── alert_rule_resource_test.go         # Alert rule resource tests
├── notification_channel_resource_test.go # Channel resource tests
├── data_source_test.go                 # Data source tests
└── integration_test.go                 # Complete stack integration tests
```

## Test Types

### 1. Unit Tests
Basic tests that don't require API access:
```bash
make test
# or
go test ./... -v -short
```

### 2. Acceptance Tests
Integration tests that create real resources via the API:
```bash
make testacc
# or
TF_ACC=1 go test ./internal/provider -v
```

**Requirements:**
- Valid API key in `QUISMON_API_KEY` environment variable
- API endpoint in `QUISMON_BASE_URL` (defaults to https://api.quismon.com)

### 3. Matrix Tests
Tests all combinations of check types, alert conditions, and notification channels:
```bash
make test-matrix
# or
./scripts/test.sh matrix
```

Covers:
- **Check Types**: HTTP, HTTPS (with headers), TCP, Ping
- **Alert Conditions**: consecutive_failures, response_time, status_code, ssl_expiry
- **Notification Channels**: Email, Webhook, Ntfy, Slack, PagerDuty

### 4. Integration Tests
Full end-to-end test creating a complete monitoring stack:
```bash
make test-integration
# or
./scripts/test.sh integration
```

Creates:
- 3 checks (API, database, gateway)
- 3 notification channels (email, webhook, ntfy)
- 3 alert rules (API down, API slow, DB down)
- Tests data sources

### 5. Quick Tests
Fast smoke tests for CI/CD:
```bash
make test-quick
# or
./scripts/test.sh quick
```

## Running Tests

### Prerequisites

1. **Set API Key:**
```bash
export QUISMON_API_KEY="qm_your_api_key_here"
```

2. **Set API URL (optional):**
```bash
export QUISMON_BASE_URL="https://api.quismon.com"
```

### Test Modes

```bash
# Run all tests (default)
./scripts/test.sh all

# Unit tests only
./scripts/test.sh unit

# All acceptance tests
./scripts/test.sh acceptance

# Matrix tests (all variations)
./scripts/test.sh matrix

# Integration test (complete stack)
./scripts/test.sh integration

# Quick smoke test
./scripts/test.sh quick
```

### Via Makefile

```bash
# Unit tests
make test

# Acceptance tests
make testacc

# All tests
make test-all

# Matrix tests
make test-matrix

# Integration test
make test-integration

# Quick test
make test-quick
```

## Test Coverage

### Check Resource Tests

**File:** `check_resource_test.go`

- ✅ Create HTTPS check
- ✅ Read check
- ✅ Update check
- ✅ Delete check
- ✅ Import check by ID
- ✅ TCP check creation
- ✅ Ping check creation
- ✅ Multi-region check

**Test Count:** 8 test cases

### Alert Rule Resource Tests

**File:** `alert_rule_resource_test.go`

- ✅ Create alert with consecutive_failures condition
- ✅ Create alert with response_time condition
- ✅ Alert with multiple notification channels
- ✅ Delete alert rule

**Test Count:** 3 test cases

### Notification Channel Resource Tests

**File:** `notification_channel_resource_test.go`

- ✅ Create email channel
- ✅ Update email channel
- ✅ Import channel by ID
- ✅ Create webhook channel
- ✅ Create ntfy channel
- ✅ Delete channel

**Test Count:** 3 test cases

### Data Source Tests

**File:** `data_source_test.go`

- ✅ Query check by name
- ✅ List all checks
- ✅ Query notification channel by name

**Test Count:** 3 test cases

### Matrix Tests

**File:** `integration_test.go`

**Check Type Matrix** (4 variations):
- HTTP basic
- HTTPS with headers
- TCP basic
- Ping basic

**Alert Condition Matrix** (4 variations):
- consecutive_failures (threshold: 3)
- response_time (threshold: 5000ms)
- status_code (threshold: 500)
- ssl_expiry (threshold: 7 days)

**Notification Channel Matrix** (5 variations):
- Email (with multiple recipients)
- Webhook (POST to URL)
- Ntfy (with topic)
- Slack (with webhook URL)
- PagerDuty (with routing key)

**Total Matrix Tests:** 13 test cases

### Integration Test

**File:** `integration_test.go`

**Complete Stack Test:**
- Creates 3 checks (HTTPS, TCP, Ping)
- Creates 3 notification channels (Email, Webhook, Ntfy)
- Creates 3 alert rules with different conditions
- Queries data sources to verify resources
- Verifies all resources are linked correctly

**Test Count:** 1 comprehensive test case

## Total Test Coverage

- **Unit Tests:** 2
- **Resource Tests:** 14
- **Data Source Tests:** 3
- **Matrix Tests:** 13
- **Integration Tests:** 1
- **Total:** 33 test cases

## Test Execution Time

Approximate durations:

| Test Type | Duration |
|-----------|----------|
| Unit | ~5 seconds |
| Quick | ~1-2 minutes |
| Resource | ~10-15 minutes |
| Matrix | ~20-30 minutes |
| Integration | ~5-10 minutes |
| All | ~40-60 minutes |

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Unit Tests
        run: make test

      - name: Acceptance Tests
        env:
          QUISMON_API_KEY: ${{ secrets.QUISMON_API_KEY }}
          QUISMON_BASE_URL: ${{ secrets.QUISMON_BASE_URL }}
        run: make test-quick
```

## Writing New Tests

### Resource Test Template

```go
func TestAccNewResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read
			{
				Config: testAccNewResourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("quismon_new.test", "name", "test-name"),
					resource.TestCheckResourceAttrSet("quismon_new.test", "id"),
				),
			},
			// ImportState
			{
				ResourceName:      "quismon_new.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update
			{
				Config: testAccNewResourceConfig_updated(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("quismon_new.test", "name", "updated-name"),
				),
			},
		},
	})
}
```

## Debugging Tests

### Verbose Output

```bash
TF_ACC=1 go test ./internal/provider -v -run TestAccCheckResource
```

### Run Single Test

```bash
TF_ACC=1 go test ./internal/provider -v -run TestAccCheckResource/TestAccCheckResource_TCP
```

### Enable Terraform Logs

```bash
TF_LOG=DEBUG TF_ACC=1 go test ./internal/provider -v -run TestAccCheckResource
```

### Test with Local Provider Build

```bash
# Build provider
make build

# Install locally
make install

# Run tests against local build
TF_ACC=1 go test ./internal/provider -v
```

## Cleanup

Tests automatically clean up resources after execution. If tests fail mid-execution, orphaned resources may remain in your account.

### Manual Cleanup

Check your Quismon dashboard for test resources (usually named "test-*") and delete them manually if needed.

## Best Practices

1. **Unique Names:** Use unique resource names to avoid conflicts
2. **Cleanup:** Always ensure tests clean up resources
3. **Idempotent:** Tests should be runnable multiple times
4. **Fast:** Keep unit tests fast, acceptance tests comprehensive
5. **Documented:** Comment test intent and expected behavior

## Troubleshooting

### "QUISMON_API_KEY must be set"

Set your API key:
```bash
export QUISMON_API_KEY="qm_your_key"
```

### "Connection refused"

Check your `QUISMON_BASE_URL`:
```bash
export QUISMON_BASE_URL="https://api.quismon.com"
```

### "Resource not found" errors

Ensure your API key has proper permissions for your organization.

### Tests timeout

Increase timeout:
```bash
TF_ACC=1 go test ./internal/provider -v -timeout 60m
```

## Contributing

When adding new features, please add corresponding tests:

1. **Unit tests** for provider logic
2. **Resource tests** for CRUD operations
3. **Matrix tests** for variations
4. **Integration tests** for complete flows

Ensure all tests pass before submitting PRs:
```bash
make test
make testacc
```
