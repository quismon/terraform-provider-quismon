#!/bin/bash

# Terraform Provider Test Script
# Tests the Quismon Terraform provider with comprehensive acceptance tests

set -e

echo "=== Quismon Terraform Provider Test Suite ==="
echo ""

# Check required environment variables
if [ -z "$QUISMON_API_KEY" ]; then
    echo "❌ Error: QUISMON_API_KEY environment variable not set"
    echo "   Please set your Quismon API key:"
    echo "   export QUISMON_API_KEY=\"qm_your_api_key_here\""
    exit 1
fi

if [ -z "$QUISMON_BASE_URL" ]; then
    echo "⚠️  Warning: QUISMON_BASE_URL not set, using default (https://api.quismon.com)"
    export QUISMON_BASE_URL="https://api.quismon.com"
fi

echo "✓ Environment configured"
echo "  API URL: $QUISMON_BASE_URL"
echo ""

# Test mode selection
TEST_MODE="${1:-all}"

case "$TEST_MODE" in
    unit)
        echo "[1/1] Running unit tests..."
        go test ./... -v -short
        ;;

    acceptance)
        echo "[1/1] Running acceptance tests..."
        TF_ACC=1 go test ./internal/provider -v -timeout 30m
        ;;

    matrix)
        echo "[1/3] Running check type matrix tests..."
        TF_ACC=1 go test ./internal/provider -v -run TestAccCheckTypeMatrix -timeout 20m

        echo ""
        echo "[2/3] Running alert condition matrix tests..."
        TF_ACC=1 go test ./internal/provider -v -run TestAccAlertConditionMatrix -timeout 20m

        echo ""
        echo "[3/3] Running notification channel matrix tests..."
        TF_ACC=1 go test ./internal/provider -v -run TestAccNotificationChannelMatrix -timeout 20m
        ;;

    integration)
        echo "[1/1] Running integration tests..."
        TF_ACC=1 go test ./internal/provider -v -run TestAccCompleteStack -timeout 30m
        ;;

    quick)
        echo "[1/2] Running quick unit tests..."
        go test ./... -v -short -timeout 5m

        echo ""
        echo "[2/2] Running quick acceptance test..."
        TF_ACC=1 go test ./internal/provider -v -run TestAccCheckResource$ -timeout 10m
        ;;

    all|*)
        echo "[1/4] Running unit tests..."
        go test ./... -v -short

        echo ""
        echo "[2/4] Running resource tests..."
        TF_ACC=1 go test ./internal/provider -v -run "TestAccCheckResource|TestAccAlertRuleResource|TestAccNotificationChannelResource" -timeout 20m

        echo ""
        echo "[3/4] Running data source tests..."
        TF_ACC=1 go test ./internal/provider -v -run "TestAccCheckDataSource|TestAccChecksDataSource|TestAccNotificationChannelDataSource" -timeout 15m

        echo ""
        echo "[4/4] Running integration tests..."
        TF_ACC=1 go test ./internal/provider -v -run TestAccCompleteStack -timeout 30m
        ;;
esac

echo ""
echo "===================================="
echo "✅ Tests completed successfully!"
echo "===================================="
echo ""

# Display test summary
echo "Test Summary:"
echo "  Mode: $TEST_MODE"
echo "  API URL: $QUISMON_BASE_URL"
echo ""
echo "Available test modes:"
echo "  all         - Run all tests (default)"
echo "  unit        - Run only unit tests"
echo "  acceptance  - Run all acceptance tests"
echo "  matrix      - Run matrix tests (all variations)"
echo "  integration - Run complete stack integration test"
echo "  quick       - Run quick smoke tests"
echo ""
echo "Usage: ./scripts/test.sh [mode]"
