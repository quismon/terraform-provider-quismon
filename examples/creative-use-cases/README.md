# creative-use-cases

## Overview

Single `main.tf` with thirteen monitoring patterns that go beyond basic
uptime: blue/green deployment watching, third-party dependency health, DNS
failover verification, inverted rate-limit/geo-block assertions, CDN
performance, maintenance detection, API contract testing, SSL coverage,
circuit-breaker monitoring, mail round-trip, HTTP/3 and throughput.

## Interface

- `main.tf` resources:
  - Channels: `quismon_notification_channel.slack_ops`, `quismon_notification_channel.incident_webhook` (`count`-gated webhook)
  - Checks: `blue_target`/`green_target` (https + content match), `third_party_health` (`for_each` over `var.third_party_apis`), `dns_failover_primary`/`dns_failover_secondary` (region-pinned), `rate_limit_working` (`inverted = true`, expects 429), `cdn_static_assets`, `not_in_maintenance`/`no_503_maintenance`, `api_version_contract`, `ssl_main`/`ssl_api`, `geo_block_working`, `backend_service` (15s interval + `recheck_on_failure`), `mail_server` (smtp-imap), `http3_support`, `download_throughput`
  - Alert rules keyed to the above with `failure_threshold`, `health_status`, or `response_time_ms` conditions
  - Outputs: `creative_monitoring_summary` (map of check IDs), `tips` (heredoc playbook)
- Variables: `primary_domain`, `api_domain`, `slack_webhook_url`, `incident_webhook`, `current_api_version`, `third_party_apis` (map)

## Invariants

- `rate_limit_working` is inverted and expects `[429]` — success of the raw request means rate limiting is broken (`main.tf` Use Case 4)
- `incident_webhook != ""` gates both the incident-webhook channel (`count`) and its inclusion in `circuit_breaker_trip`'s channel list (`main.tf`)
- `third_party_health` and `dependency_down` iterate `for_each = var.third_party_apis` (stripe/sendgrid/twilio status endpoints by default)
- `mail_server` embeds placeholder `"CHANGE_ME"` passwords — must be replaced with variables before real use (`main.tf` Use Case 11)
- `backend_service` runs every 15 seconds with `recheck_on_failure = true` for fast circuit-breaker detection (`main.tf` Use Case 10)

## Side Effects

- Applying creates ~20 real checks, channels, and alert rules against the configured domains; the smtp-imap check sends real email

## Verification

- No test manifest in this directory; inherits the provider module's verification: `cd /workspace/terraform-provider-quismon && CGO_ENABLED=0 go test ./...`
