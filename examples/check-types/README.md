# check-types

## Overview

Reference configurations for every check type supported by `quismon_check`,
split by protocol family, plus the check-type-agnostic modifiers that are the
product's differentiators.

## Interface

- `http.tf` — `quismon_check.api_health` (type `https`): url/method/`expected_status`/`expected_content` + `content_match_type`, `follow_redirects`, `verify_ssl`
- `tcp-ping-udp.tf` — `quismon_check.db_port` (`tcp`), `quismon_check.gateway` (`ping` with `count`), `quismon_check.syslog` (`udp` with `payload`)
- `dns.tf` — `quismon_check.mx` (`dns` with `record_type`/`expected_records`), `quismon_check.signed-zone` (`dnssec` with `nameservers`)
- `ssl.tf` — `quismon_check.cert` (`ssl` with `hostname`/`port`/`warning_days`; optional `expected_fingerprint`, `expected_domains`)
- `multistep.tf` — `quismon_check.login_flow` (`multistep` `config_json` with `extracts` + `{{token}}` interpolation, `fail_fast`)
- `smtp-imap.tf` — `quismon_check.mail_pipeline` (`smtp-imap` round-trip with `require_spf_pass`/`require_dkim_pass`/`require_dmarc_pass`, sensitive credentials)
- `throughput-http3.tf` — `quismon_check.cdn_download` (`throughput` with `max_size_mb`), `quismon_check.quic` (`http3`)
- `traceroute.tf` — `quismon_check.route` (`traceroute` with `max_hops`/`probes_per_hop`)
- `playwright.tf` — `quismon_check.checkout` (`playwright` low-code `actions` array or `script_inline`; Pro+ single region per comments)
- `modifiers.tf` — `quismon_check.db_must_be_closed` (`inverted = true`), `db_must_be_closed_by_friday` (`starts_broken` + `starts_broken_deadline_seconds`), `deploy_probe` (`expires_after_seconds`), `app_after_db` (`check_dependencies`)

## Invariants

- Complex/nested configs (multistep, playwright, smtp-imap, dns with arrays) use `config_json = jsonencode({...})`; flat string configs use the `config` map (per-file usage)
- Extracts (`jsonpath`) and `{{variable}}` interpolation in multisteps require Cafe tier+ (`multistep.tf` header comment)
- Dom script path and playwright are tier-gated per file comments (playwright: "Pro+ single region, Advanced+ multi-region")
- Modifiers compose with any check type (`modifiers.tf`: inverted + starts_broken + deadline on one tcp check)

## Side Effects

- Applying creates real Quismon checks that then run on their intervals against the example targets (example.com placeholders — replace before applying)

## Verification

- No test manifest in this directory; inherits the provider module's verification: `cd /workspace/terraform-provider-quismon && CGO_ENABLED=0 go test ./...` (example configs are exercised by the provider acceptance suite: `./scripts/test.sh matrix`)
