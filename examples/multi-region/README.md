# multi-region

## Overview

Minimal multi-region example: one HTTPS health check executed from five
regions, a webhook channel, and two alert rules (regional failure +
latency).

## Interface

- `main.tf` resources:
  - `quismon_notification_channel.webhook` (type `webhook`, config keys `url`, `method = "POST"`, `secret`)
  - `quismon_check.api_multi_region` — https check, `regions = ["na-east-ewr", "na-west-lax", "eu-west-ams", "eu-central-fra", "ap-southeast-sin"]`
  - `quismon_alert_rule.api_regional_failure` — `failure_threshold = 2`
  - `quismon_alert_rule.api_global_latency` — `response_time_ms = 3000`
  - Output: `monitored_regions` (from `quismon_check.api_multi_region.regions`)
- Variables: `quismon_api_key`, `incident_webhook_url`, `incident_webhook_secret` (all sensitive)

## Invariants

- Region keys are the platform's current region codes (`GET /v1/regions`)
- Latency condition uses numeric `response_time_ms = 3000` (`main.tf`)

## Side Effects

- Applying creates a check running from five regions every 60s

## Verification

- No test manifest in this directory; inherits the provider module's verification: `cd /workspace/terraform-provider-quismon && CGO_ENABLED=0 go test ./...`
