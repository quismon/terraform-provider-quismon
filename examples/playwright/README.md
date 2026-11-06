# playwright

## Overview

Playwright browser-check examples in five flavors: low-code actions, a full
script with injected variables, an API-endpoint script via `page.request`, a
multi-step e-commerce flow, and a disabled cross-browser (Firefox) variant.

## Interface

- `main.tf` resources (all `type = "playwright"`, `config_json = jsonencode({...})`):
  - `quismon_check.low_code_form_test` — `actions` array (goto/wait/fill/click) + `extracts` with `transform = "trim"`
  - `quismon_check.full_script_test` — `script_inline` heredoc with `{{base_url}}`/`{{username}}`/`{{password}}` variable interpolation, `sensitive_variables = ["password"]`
  - `quismon_check.api_endpoint_test` — script using `context.setExtraHTTPHeaders` + `page.request.get` for API testing
  - `quismon_check.ecommerce_flow_test` — 2-region 180s-timeout checkout journey
  - `quismon_check.cross_browser_test` — `browser = "firefox"`, `enabled = false` by default
  - Outputs: one check ID per resource
- Variables: `quismon_api_key`, `target_base_url`, `test_username`, `test_password`, `test_api_token` (sensitive)

## Invariants

- Playwright is tier-gated: Pro+ single-region, Advanced+ multi-region (`check-types/playwright.tf` comment; the e-commerce flow here uses 2 regions)
- `screenshot_on_failure`/`trace_enabled`/`video_enabled`/`headless` flags configure artifact capture (`main.tf`)
- Secrets passed via `variables` + `sensitive_variables` masking (`main.tf` full_script_test)
- Provider requirement `quismon >= 1.0.0` (`main.tf` terraform block)

## Side Effects

- Applying creates browser checks that drive a real headless browser against the target on each interval (form submissions, cart operations, screenshots on failure)

## Verification

- No test manifest in this directory; inherits the provider module's verification: `cd /workspace/terraform-provider-quismon && CGO_ENABLED=0 go test ./...`

## Script grammar

`script_inline` uses the versioned **body-v1** grammar: top-level statements only, with `page`, `context`, `browser`, `step`, `endStep`, `secrets`, and `console` in scope. Function/arrow wrappers, CommonJS/ESM, and `this` are rejected by the API. Use `secrets` (configured via `config.secrets`) instead of inlining credentials. `GET /v1/check-types` exposes `data.playwright_script_grammar`.
