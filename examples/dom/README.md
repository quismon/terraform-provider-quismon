# dom

## Overview

Dom-check examples: browser-lite rendering assertions via lightpanda.
Declarative selector/text assertions (Adopter tier+), scripted journeys with
secrets (Paid tier+), and multisteps where dom steps interleave with http
steps and share extracted variables.

## Interface

- `main.tf` resources:
  - `quismon_check.spa_health` — declarative dom: `url`, `expected_status_code`, `expected_selector`, `expected_text` (string or array of strings)
  - `quismon_check.authenticated_journey` — scripted dom via `config_json`: `secrets` map referenced in the script as `$LP_APP_EMAIL`/`$LP_APP_PASSWORD`, `script` heredoc using `page.fill`/`page.click`/`page.waitForSelector`/`page.extract`
  - `quismon_check.build_deployment` — multistep: http step extracts `build`, followed by a dom step
  - `quismon_check.dom_mixed_flow` — 4-step multistep: dom extract (`selector`+`attribute`) → http POST with `{{site_title}}` in body → Bearer GET with `{{token}}` → dom `expected_js` assertion using `{{token}}`/`{{whoami}}`
- `variables.tf` — `app_password` (sensitive)

## Invariants

- Nested values (the `secrets` map) require `config_json`; flat string values can use the `config` map (`main.tf` authenticated_journey comment)
- Dom secrets are write-only — masked `***REDACTED***` on read; scripts reference them as `$LP_<UPPERCASED_KEY>` (`main.tf` comments)
- Dom steps cannot set request headers — use query params, or the scripted dom check's `page.fill` login flow for header-auth journeys (`main.tf` mixed-flow comment)
- Provider pin `quismon ~> 1.4` and explicit `base_url = "https://api.quismon.com"` (`main.tf` terraform block)

## Side Effects

- Applying creates dom/multistep checks; scripted checks render real pages in a headless browser on each run

## Verification

- No test manifest in this directory; inherits the provider module's verification: `cd /workspace/terraform-provider-quismon && CGO_ENABLED=0 go test ./...`
