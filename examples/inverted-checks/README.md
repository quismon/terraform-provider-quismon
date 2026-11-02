# inverted-checks

## Overview

Inverted-check patterns: checks that alert when something *succeeds* that
shouldn't — deployment version watching, admin-panel/internal-API exposure,
firewall port validation, and staging isolation — plus an explanation output
of the inverted semantics.

## Interface

- `main.tf` resources:
  - `quismon_notification_channel.slack`
  - Checks: `deployment_version` (NOT inverted — https content match on `var.current_version`; fails when a new version deploys), `admin_panel_not_public` (https expecting `[401, 403]`, `inverted = true`), `internal_api_protected` (https expecting `[401]`, inverted), `database_port_blocked` (tcp 5432, inverted), `redis_port_blocked` (tcp 6379, inverted), `staging_not_public` (https expecting `[200]`, inverted)
  - Alert rules: `new_version_deployed`, `admin_exposed`, `internal_api_exposed`, `sensitive_port_exposed` (all `health_status = "down"`)
  - Outputs: `inverted_checks` (IDs), `how_inverted_works` (heredoc explaining normal vs inverted logic)
- Variables: `production_url`, `api_url`, `current_version`, `slack_webhook_url`

## Invariants

- `deployment_version` deliberately uses a NON-inverted content check: the alert fires when the old version string disappears (`main.tf` Use Case 1 comments)
- All security-posture checks set `inverted = true`; for inverted checks `health_status = "down"` means the problem is NOT detected (`main.tf` alert comments)
- `database_port_blocked`/`redis_port_blocked` use `var.production_url` as the host — a URL placeholder the comments say to replace with the actual host (`main.tf` Use Case 4)

## Side Effects

- Applying creates exposure-detection checks probing `/admin`, `/internal/stats`, and ports 5432/6379/3000 against the configured target

## Verification

- No test manifest in this directory; inherits the provider module's verification: `cd /workspace/terraform-provider-quismon && CGO_ENABLED=0 go test ./...`
