# internal/provider

## Overview

Terraform Plugin Framework implementation of the `quismon` provider: provider
schema/configuration, six resources and four data sources, plus plan-time
validation and helpers. All resources talk to the Quismon API through
`internal/client`.

## Interface

Resources (registered in `provider.go:Resources`):
- `quismon_check` (`check_resource.go:NewCheckResource`) — full CRUD + import by ID; config via `config` map or `config_json` string
- `quismon_alert_rule` (`alert_rule_resource.go:NewAlertRuleResource`) — CRUD + import with ID format `check_id:rule_id`
- `quismon_notification_channel` (`notification_channel_resource.go:NewNotificationChannelResource`) — CRUD + import by ID
- `quismon_signup` (`signup_resource.go:NewSignupResource`) — bootstrap org creation via unauthenticated quick-signup; `api_key` computed/sensitive
- `quismon_organization_otlp` (`organization_otlp_resource.go:NewOrganizationOTLPResource`) — singleton OTLP export config via `PUT/GET /v1/org/otlp`
- `quismon_status_page` (`status_page_resource.go:NewStatusPageResource`) — singleton public status page via `/v1/status-pages`

Data sources (registered in `provider.go:DataSources`):
- `quismon_check` (`check_data_source.go:NewCheckDataSource`) — fetch one check by `name`
- `quismon_checks` (`checks_data_source.go:NewChecksDataSource`) — list all checks
- `quismon_notification_channel` (`notification_channel_data_source.go:NewNotificationChannelDataSource`) — fetch one channel by `name`
- `quismon_regions` (`regions_data_source.go:NewRegionsDataSource`) — static region catalog (no API call); `ValidateRegion(code)` exported helper

Provider wiring:
- `New(version string) func() provider.Provider` (`provider.go`) — provider type name `quismon`; schema attributes `api_key` (sensitive) and `base_url`
- `state_reader.go` — `readAPIKeyFromState()` reads a `quismon_signup` API key from `terraform.tfstate` when no key is configured
- `validation.go` — `validatePlaywrightConfig`, `validatePlaywrightConfigJSON`, `validateDomConfig`
- `starts_broken.go` — `mergeStartsBrokenAlias(startsBroken, alias)` resolves `starts_broken` vs deprecated `alert_after_first_success`
- `expires_after_seconds_plan_modifier.go` — `ExpiresAfterSecondsModifier()` preserves the API value unless explicitly configured

## Invariants

- Provider config precedence: Terraform attribute > env `QUISMON_API_KEY`/`QUISMON_BASE_URL` > API key read from terraform state; base URL defaults to `https://api.quismon.com` (`provider.go:Configure`)
- An unknown `api_key` is a warning, not an error (bootstrap pattern `api_key = quismon_signup.foo.api_key`); unknown `base_url` is an error (`provider.go:Configure`)
- Check config: exactly one of `config` or `config_json` must be set; `config_json` wins when both present; must parse as a JSON object (`check_resource.go:Create`)
- Playwright and dom configs are validated at plan time on create and update (`check_resource.go:Create/Update` calling `validation.go`): playwright needs one of `script_id`/`script_inline`/`script_url`; dom needs `url` or `script` (script ≤ 16KB); browsers limited to chromium/firefox/webkit
- `alert_after_first_success` is a deprecated alias of `starts_broken`; `starts_broken` wins when both set (`starts_broken.go:mergeStartsBrokenAlias`); the flag is create-only — Update never sends it (platform clears it after first success; `check_resource.go:Update`)
- Alert-rule condition values for numeric keys (`failure_threshold`, `response_time_ms`, `ssl_days_remaining`, ...) are converted to JSON numbers, not strings (`alert_rule_resource.go:conditionValue`) — string thresholds silently never fire
- Check Read never overwrites `config`/`config_json` from the API (sensitive fields come back redacted); external drift is detected via `config_hash` comparison with a warning (`check_resource.go:Read`)
- Volatile telemetry (`health_status`, `last_checked`, `created_at`, `updated_at`) is only populated when absent (import), never refreshed, to avoid perpetual plan drift (`check_resource.go:Read`)
- `expires_after_seconds` is never sent on check update; the plan modifier preserves the API value unless explicitly set (`check_resource.go:Update`, `expires_after_seconds_plan_modifier.go`)
- Status page: `slug` is immutable (RequiresReplace); `access_token`/`access_password` are write-only (API stores hashes); the status-pages API returns `{"status_page": ...}` without the `{data: ...}` envelope and is decoded directly (`status_page_resource.go:getStatusPage`); delete treats 404 as success
- Signup `Read`/`Update` are no-ops and `Delete` does not delete the org (`signup_resource.go`)
- OTLP `Delete` disables export (`{"enabled": false, "endpoint": ""}`) rather than deleting anything (`organization_otlp_resource.go:Delete`); `export_interval_seconds` ≥ 10 (schema validator)
- Alert-rule import ID must be `check_id:rule_id` (`alert_rule_resource.go:ImportState`)

## Side Effects

- All resources/data sources perform authenticated HTTP calls to the Quismon API via `internal/client` (reads and writes)
- `quismon_signup` creates a real organization and API key (unauthenticated endpoint)
- `state_reader.go` reads `terraform.tfstate` from the working directory (`findStateFile`: `terraform.tfstate`, `.terraform/terraform.tfstate`)
- Provider diagnostics surface warnings (drift detection, state-read API key, unknown API key) into Terraform CLI output

## Verification

- Run: `cd /workspace/terraform-provider-quismon && CGO_ENABLED=0 go test ./...`
- Acceptance tests (require `TF_ACC=1` and `QUISMON_API_KEY`): `TF_ACC=1 go test ./internal/provider/... -v`
