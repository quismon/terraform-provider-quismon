# full-stack

## Overview

Layered monitoring of a complete web application stack in one `main.tf`:
DNS → SSL → website → API → database → security (inverted checks), with
three notification channels and per-layer alert rules.

## Interface

- `main.tf` resources:
  - Channels: `quismon_notification_channel.slack`, `.email`, `.webhook` (generic incident webhook)
  - Checks: `dns_a` (dns, 2 regions), `ssl_cert` (ssl, `warning_days = 30`), `website` (https + `expected_content` "Welcome", 3 regions), `api_health` (https + JSON status content match, 30s interval), `database` (tcp 5432), `admin_panel_security` (https expecting `[401, 403]` with `inverted = true`), `dev_ports_closed` (tcp 3000, inverted)
  - Alert rules: `dns_change`, `ssl_expiry`, `website_down` (health_status), `website_slow` (`response_time_ms = 3000`), `api_down` (`failure_threshold = 2`), `database_down`, `admin_exposed`
  - Outputs: `monitoring_summary`, `notification_channels`
- Variables: `domain`, `api_endpoint`, `database_host`, `slack_webhook_url`, `oncall_email`

## Invariants

- The two security checks are inverted: raw success (200 OK from `/admin`, open port 3000) means exposure (`main.tf` Layer 6)
- Complex configs use `config_json = jsonencode({...})` throughout (`main.tf`)
- The webhook channel embeds a placeholder integration URL (`events.example.com/integration/YOUR_KEY/enqueue`) that must be replaced (`main.tf`)

## Side Effects

- Applying creates 7 checks and 7 alert rules running continuously against the configured domain

## Verification

- No test manifest in this directory; inherits the provider module's verification: `cd /workspace/terraform-provider-quismon && CGO_ENABLED=0 go test ./...`
