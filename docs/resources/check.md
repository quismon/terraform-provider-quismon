---
page_title: "quismon_check Resource - quismon"
subcategory: ""
description: |-
  Manages a Quismon health check.
---

# quismon_check (Resource)

Manages a Quismon health check.

## Configuration reference (per check type)

The `config` map carries the check-type-specific settings. Unknown keys are
**rejected** by the API (a typo like `expiration_warning_days` fails at plan
time instead of silently never applying). Required keys are marked ★.

### http / https

| Key | Type | Default | Notes |
|---|---|---|---|
| ★ `url` | string | — | Full URL; the scheme decides http vs https |
| `method` | string | `GET` | |
| `headers` | map | — | Custom `User-Agent` requires the **Adopter tier**+ |
| `body` | string | — | For POST/PUT |
| `expected_status` | array / int / comma string | `[200]` | |
| `expected_content` | string | — | With `content_match_type` |
| `content_match_type` | string | `contains` | `contains` \| `exact` \| `regex` \| `not_contains` |
| `timeout_seconds` | number | `10` | |
| `follow_redirects` | bool | `false` | |
| `verify_ssl` | bool | `false` | |

Emits: `status_code`, `response_time_ms`, `resolved_ips`, up to 32 response
headers (used by multistep header extraction), `body_preview`.

### ping

| Key | Type | Default | Notes |
|---|---|---|---|
| ★ `host` | string | — | |
| `count` / `packet_count` | number | `3` | Aliases |
| `timeout_seconds` | number | `5` | |

Emits: `packet_loss_pct`, `packets_sent/received`, `min/max/avg_time_ms`,
per-packet `rtt_ms`. Success = at least one reply.

### tcp

| Key | Type | Default | Notes |
|---|---|---|---|
| ★ `host` | string | — | |
| ★ `port` | number | — | |
| `timeout_seconds` | number | `5` | |

Emits: `connected`, `resolved_ips`. `response_time_ms` = connect latency.

### udp

| Key | Type | Default | Notes |
|---|---|---|---|
| ★ `host`, ★ `port` | — | — | |
| `payload` | string | `\n` | Hex-encoded or plain text |
| `expected_response` | string | — | Substring or regex, matched against the response |
| `timeout_seconds` | number | `5` | |

### dns

| Key | Type | Default | Notes |
|---|---|---|---|
| ★ `domain` | string | — | |
| ★ `record_type` | string | `A` | `A` \| `AAAA` \| `CNAME` \| `MX` \| `TXT` \| `NS` \| `SOA` |
| `expected_records` | array | — | Substring match on rendered records |
| `timeout_seconds` | number | `5` | |

### dnssec

| Key | Type | Default | Notes |
|---|---|---|---|
| ★ `domain` | string | — | |
| `nameservers` | array | system resolvers | IPv4/IPv6, optional port |
| `timeout_seconds` | number | `10` | |

Verifies DNSKEY **and DS** record presence plus a normal lookup of the
requested type. Note: it validates *presence of signed-zone records*, not a
full chain-of-trust validation.

### ssl

| Key | Type | Default | Notes |
|---|---|---|---|
| ★ `hostname` | string | — | SNI + connect target |
| `port` | number | `443` | |
| `warning_days` | number | `30` | The check goes **unhealthy** when fewer days remain |
| `expected_fingerprint` | string | — | SHA-256 cert pinning |
| `expected_domains` | array | — | SAN pinning |

Emits: `days_until_expiry` (also usable by the `ssl_days_remaining` alert
condition), `not_after`, `fingerprint`, `san_domains`.

### smtp-imap

Real SMTP → IMAP round trip: sends a uniquely-tagged message, polls the inbox
for it, and parses the `Authentication-Results` header. Requires
`config_json` (nested structure).

| Key | Type | Default | Notes |
|---|---|---|---|
| ★ `smtp_host`, ★ `smtp_port` | — | — | |
| ★ `imap_host`, ★ `imap_port` | — | — | |
| ★ `from_address`, ★ `to_address` | — | — | |
| `smtp_username` / `smtp_password` | string | — | Sensitive |
| `smtp_tls_mode` | string | auto by port | `auto` \| `starttls` \| `tls` \| `none` (465 TLS, 587 STARTTLS, 25 plain) |
| `imap_username` / `imap_password` | string | SMTP creds | Sensitive |
| `imap_tls_mode` | string | auto | |
| `subject` / `body` | string | templated | |
| `timeout_seconds` | number | `30` | |
| `max_wait_seconds` | number | `60` | Inbox poll window |
| `require_spf_pass` | bool | `false` | Check fails unless the mail's SPF verdict is `pass` |
| `require_dkim_pass` | bool | `false` | Same for DKIM |
| `require_dmarc_pass` | bool | `false` | Same for DMARC |

Emits: `spf_result`, `dkim_result`, `dmarc_result`, `smtp_tls_mode`,
`imap_tls_mode`, `message_id`, round-trip timings.

### throughput

| Key | Type | Default | Notes |
|---|---|---|---|
| ★ `url` | string | — | |
| `max_size_mb` | number | `5` | Tier-capped: 5 free/cafe, 25 adopter, 100 pro, 250 advanced, 500 enterprise |
| `timeout_seconds` | number | `30` | |

Emits: `throughput_mbps`, `bytes_downloaded`, `time_seconds`.

### http3

| Key | Type | Default | Notes |
|---|---|---|---|
| ★ `url` | string | — | |
| `expected_status` | — | `[200]` | |
| `timeout_seconds` | number | `10` | |

Detects HTTP/3 support via `Alt-Svc`/`h3` (`supports_http3`,
`http3_available` metadata) and validates the HTTP response. Note: success is
based on the HTTP response only — a server without HTTP/3 still passes.

### traceroute

ICMP probes natively (per-hop latency data the dashboard graphs), with a
`traceroute`-binary fallback.

| Key | Type | Default | Notes |
|---|---|---|---|
| ★ `host` | string | — | |
| `max_hops` | number | `30` | |
| `probes_per_hop` | number | `3` | |
| `timeout_seconds` | number | `5` | Per probe |

Emits: `hop_latencies` (per-hop RTT arrays — rendered as graphs),
`traceroute_path`, `path_hash`, `total_hops`, `successful_hops`,
`timeout_hops`, `completed`, `avg_hop_rtt_ms`, `final_hop_rtt_ms`.

### multistep

`config_json` with a `steps` array; each step is a full check of any type
with `name`, `type`, `config`, optional `extracts` and `timeout_seconds`.
Top-level `fail_fast` (default `true`) stops at the first failure.

- **extracts**: `jsonpath` (must match exactly one node), `regex` (capture
  group 1 or full match), `header` (case-insensitive), each with an optional
  `default`
- **`{{var}}` interpolation**: every string in later step configs; forward
  references are rejected at plan time by the API
- **Unresolved variables fail the step** instead of sending literal braces
- Extracts require the **Cafe tier**+; 5 steps free, 10 cafe/adopter, 20 pro+

### playwright

`config_json` with either `actions` (low-code: goto, click, fill, type, wait,
screenshot, evaluate, select, check, uncheck, press, hover, scroll) or
`script_inline`. Supports `browser` (chromium/firefox/webkit), `viewport_*`,
`timeout_seconds`, `extracts`, `assertions`, and `config.user_agent`
(**Adopter tier**+). Tier gates: **Pro+** single-region, **Advanced+**
multi-region.

## The modifiers

Any check type combines with:

- **`inverted = true`** — alert when the *bad* state exists (open firewalls,
  public admin panels, resurrected servers)
- **`starts_broken = true`** (+ `starts_broken_arrival_notice`,
  `starts_broken_deadline_seconds`) — create the check before the fix exists;
  silent until the first success, arrival notice, graduation into permanent
  enforcement, one-time escalation if the deadline passes
- **`expires_after_seconds`** — temporary checks that auto-delete (CI/CD,
  migrations, incident windows)
- **`check_dependencies`** — run only when the named checks are healthy

## Example Usage

### HTTP/HTTPS

```terraform
# HTTP/HTTPS check - the workhorse
resource "quismon_check" "api_health" {
  name             = "api-health"
  type             = "https"
  interval_seconds = 60
  regions          = ["na-east-ewr", "eu-central-fra", "ap-southeast-sin"]

  config = {
    url              = "https://api.example.com/health"
    method           = "GET"                       # default GET
    expected_status  = [200, 201]                  # array, int, or comma string
    timeout_seconds  = 10
    follow_redirects = false
    verify_ssl       = false
    expected_content = "OK"                        # substring match by default
    # content_match_type = "contains"              # contains | exact | regex | not_contains
    # headers = { "X-Probe" = "quismon" }          # custom User-Agent requires Adopter tier+
    # body = "{}"                                  # for POST/PUT
  }
}
```

### Network checks (TCP / Ping / UDP)

```terraform
# TCP port check (also the base for inverted firewall assertions)
resource "quismon_check" "db_port" {
  name             = "db-port-reachable"
  type             = "tcp"
  interval_seconds = 300
  regions          = ["na-east-ewr"]

  config = {
    host            = "db.example.com"
    port            = 5432
    timeout_seconds = 5
  }
}

# ICMP ping - packet loss, min/max/avg RTT
resource "quismon_check" "gateway" {
  name             = "gateway-ping"
  type             = "ping"
  interval_seconds = 60
  regions          = ["eu-central-fra"]

  config = {
    host            = "gateway.example.com"
    count           = 3
    timeout_seconds = 5
  }
}

# UDP probe - send a payload, optionally match the response
resource "quismon_check" "syslog" {
  name             = "syslog-udp"
  type             = "udp"
  interval_seconds = 300
  regions          = ["eu-central-fra"]

  config = {
    host             = "syslog.example.com"
    port             = 514
    payload          = "<134>quismon probe"
    timeout_seconds  = 5
  }
}
```

### DNS / DNSSEC

```terraform
# DNS record check
resource "quismon_check" "mx" {
  name             = "mail-dns"
  type             = "dns"
  interval_seconds = 300
  regions          = ["na-east-ewr"]

  config = {
    domain           = "example.com"
    record_type      = "MX"            # A | AAAA | CNAME | MX | TXT | NS | SOA
    expected_records = ["mail.example.com"]
    timeout_seconds  = 5
  }
}

# DNSSEC - DNSKEY/DS presence plus a normal lookup of the requested type
resource "quismon_check" "signed-zone" {
  name             = "zone-signed"
  type             = "dnssec"
  interval_seconds = 3600
  regions          = ["na-east-ewr"]

  config = {
    domain          = "example.com"
    nameservers     = ["1.1.1.1", "8.8.8.8"]
    timeout_seconds = 10
  }
}
```

### SSL certificate expiry

```terraform
# Certificate expiry - goes unhealthy when fewer than warning_days remain
# (alert with a health_status: "down" rule - there is no separate ssl_expiry condition)
resource "quismon_check" "cert" {
  name             = "api-cert-expiry"
  type             = "ssl"
  interval_seconds = 3600
  regions          = ["na-east-ewr", "eu-central-fra"]

  config = {
    hostname        = "api.example.com"
    port            = 443
    warning_days    = 30
    # expected_fingerprint = "AA:BB:..."  # SHA-256, pin the cert
    # expected_domains    = ["api.example.com"] # SAN pinning
  }
}
```

### Traceroute

```terraform
# Traceroute - per-hop latency (dashboard graphs hop_latencies), path hash,
# completion and packet-loss thresholds. ICMP probes natively, with a
# traceroute-binary fallback.
resource "quismon_check" "route" {
  name             = "route-to-origin"
  type             = "traceroute"
  interval_seconds = 300
  regions          = ["na-east-ewr", "eu-central-fra"]

  config = {
    host             = "origin.example.com"
    max_hops         = 20
    probes_per_hop   = 3
    timeout_seconds  = 5
  }
}
```

### Throughput / HTTP-3

```terraform
# Throughput - streamed download up to max_size_mb (tier-capped: 5 MB free,
# 25 adopter, 100 pro, 250 advanced), reports throughput_mbps
resource "quismon_check" "cdn_download" {
  name             = "cdn-throughput"
  type             = "throughput"
  interval_seconds = 900
  regions          = ["na-east-ewr", "sa-east-sao"]

  config = {
    url        = "https://cdn.example.com/10mb.bin"
    max_size_mb = 100
    timeout_seconds = 30
  }
}

# HTTP/3 - detects Alt-Svc/h3 support and validates the HTTP response
resource "quismon_check" "quic" {
  name             = "h3-support"
  type             = "http3"
  interval_seconds = 3600
  regions          = ["na-east-ewr"]

  config = {
    url             = "https://example.com"
    expected_status = [200]
  }
}
```

### SMTP → IMAP round trip with deliverability assertions

```terraform
# SMTP -> IMAP round trip with email-auth assertions (Only Quismon)
# Sends a uniquely-tagged mail, polls the inbox for it, and parses the
# Authentication-Results header. Credentials are sensitive.
resource "quismon_check" "mail_pipeline" {
  name             = "mail-roundtrip"
  type             = "smtp-imap"
  interval_seconds = 900
  regions          = ["eu-central-fra"]

  config_json = jsonencode({
    smtp_host = "smtp.example.com"
    smtp_port = 587
    imap_host = "imap.example.com"
    imap_port = 993
    from_address = "probe@example.com"
    to_address   = "probe-inbox@example.com"
    smtp_username = var.mail_username
    smtp_password = var.mail_password
    smtp_tls_mode = "starttls"       # auto | starttls | tls | none
    subject       = "Quismon round-trip probe"
    # Deliverability assertions - the check fails when the mail does not pass:
    require_spf_pass   = true
    require_dkim_pass  = true
    require_dmarc_pass = true
    max_wait_seconds   = 60
  })
}

variable "mail_username" { sensitive = true }
variable "mail_password" { sensitive = true }
```

### Multistep with variable passing

```terraform
# Multistep with variable passing (extracts require Cafe tier+): log in,
# extract a token from the response, use it in the next step.
resource "quismon_check" "login_flow" {
  name             = "api-login-flow"
  type             = "multistep"
  interval_seconds = 300
  regions          = ["na-east-ewr"]

  config_json = jsonencode({
    steps = [
      {
        name = "login"
        type = "https"
        config = {
          url             = "https://api.example.com/login"
          method          = "POST"
          body            = jsonencode({ user = "probe", pass = "..." })
          expected_status = [200]
        }
        extracts = {
          token = { jsonpath = "$.access_token" }
        }
      },
      {
        name = "use-token"
        type = "https"
        config = {
          url     = "https://api.example.com/me"
          headers = { Authorization = "Bearer {{token}}" }
        }
      }
    ]
    fail_fast = true
  })
}
```

### Playwright browser check

```terraform
# Playwright browser check (Pro+ single region, Advanced+ multi-region).
# Low-code actions or a full script_inline; screenshots/traces via S3.
resource "quismon_check" "checkout" {
  name             = "checkout-flow"
  type             = "playwright"
  interval_seconds = 900
  regions          = ["na-east-ewr"]

  config_json = jsonencode({
    browser         = "chromium"
    timeout_seconds = 60
    viewport_width  = 1280
    viewport_height = 720
    actions = [
      { action = "goto",  url = "https://shop.example.com" },
      { action = "click", selector = ".product:first-child button.add" },
      { action = "click", selector = "#checkout" },
      { action = "wait",  selector = "#order-complete", timeout_ms = 10000 },
      { action = "screenshot" }
    ]
    # or a full script:
    # script_inline = file("${path.module}/checkout.spec.js")
  })
}
```

### The modifiers: inverted, starts-broken + deadline, expiring, dependencies

```terraform
# The modifiers apply to ANY check type and are the Quismon USP.

# 1. Inverted - alert when the BAD state exists (firewall enforcement)
resource "quismon_check" "db_must_be_closed" {
  name     = "db-port-must-be-closed"
  type     = "tcp"
  inverted = true
  interval_seconds = 300
  regions  = ["na-east-ewr", "eu-central-fra"]

  config = { host = "db.example.com", port = 5432 }
}

# 2. Starts-broken + deadline - register the problem before the fix exists.
# Silent while the issue is open, arrival notice when it lands, one-time
# escalation if the deadline passes, permanent enforcement afterwards.
resource "quismon_check" "db_must_be_closed_by_friday" {
  name     = "db-port-must-be-closed"
  type     = "tcp"
  inverted = true
  starts_broken                  = true
  starts_broken_arrival_notice   = true
  starts_broken_deadline_seconds = 604800 # one week to close it
  interval_seconds = 300
  regions  = ["na-east-ewr"]

  config = { host = "db.example.com", port = 5432 }
}

# 3. Expiring - temporary checks that auto-delete (CI/CD, migrations)
resource "quismon_check" "deploy_probe" {
  name     = "deploy-verification"
  type     = "https"
  expires_after_seconds = 86400
  interval_seconds = 60
  regions  = ["na-east-ewr"]

  config = { url = "https://staging.example.com/health", expected_status = [200] }
}

# 4. Dependencies - a check only runs when its dependencies are healthy
resource "quismon_check" "app_after_db" {
  name     = "app-health"
  type     = "https"
  check_dependencies = [quismon_check.db_port.id]
  interval_seconds = 60
  regions  = ["na-east-ewr"]

  config = { url = "https://app.example.com/health", expected_status = [200] }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `interval_seconds` (Number) Check interval in seconds (minimum 60).
- `name` (String) Check name.
- `type` (String) Check type: http, https, tcp, ping, udp, dns, dnssec, ssl, multistep, smtp-imap, throughput, http3, spf, dkim, dmarc, playwright, traceroute.

### Optional

- `alert_after_first_success` (Boolean, Deprecated) Deprecated alias of starts_broken. Either attribute sets the same stored flag.
- `check_dependencies` (Set of String) List of check IDs that must be healthy before this check runs. If any dependency is unhealthy, this check is skipped with 'dependency_failed' status.
- `config` (Map of String, Sensitive) Check-specific configuration (for simple types). Use config_json for complex nested configs like multistep. Password fields (smtp_password, imap_password, password) are sensitive and cannot be re-read from the API.
- `config_json` (String, Sensitive) Check configuration as JSON string (required for multistep, smtp-imap, and other complex configs). Use jsonencode() to create this. Password fields are sensitive and cannot be re-read from the API.
- `enabled` (Boolean) Whether the check is enabled.
- `expires_after_seconds` (Number) Check auto-deletes after this many seconds. NULL or 0 means no expiration. Note: expiring checks are typically created via API for temporary monitoring, not via Terraform.
- `iac_locked` (Boolean) If true, this check can only be modified via API (prevents web UI changes).
- `inverted` (Boolean) If true, alerts on success instead of failure. Useful for firewall validation - alert when a blocked port opens.
- `recheck_on_failure` (Boolean) If true, failed checks trigger an immediate recheck from a different region to verify the failure before alerting.
- `regions` (Set of String) Monitoring regions (set - order does not matter, duplicates not allowed).
- `show_on_status_page` (Boolean) If true, this check contributes to the public status page. Default is false (opt-in).
- `simultaneous_regions` (Boolean) If true, all regional checks execute simultaneously. If false (default), regional checks are staggered to avoid rate limiting.
- `starts_broken` (Boolean) Starts-broken checks (TDD workflow): when true, the check does not alert (no incidents, no notifications) until it has succeeded at least once; an arrival notice is sent through the check's alert-rule channels on that first success. Create-only: once the check first succeeds, the platform silently graduates the check to a normal configuration (the stored flag clears) and this attribute is ignored on updates - remove it from your configuration after the first successful apply to keep plans clean.
- `starts_broken_arrival_notice` (Boolean) When true (default), the first success of a starts-broken check sends the arrival notice through the check's alert-rule channels. Has no effect unless starts_broken is true.
- `starts_broken_deadline_seconds` (Number) Optional create-only "never came up" guard: if the check's first success hasn't happened within this many seconds of creation, starts-broken suppression lifts, a one-time deadline-breach escalation fires, and failure alerts flow. Only meaningful with starts_broken. Create-only.

### Read-Only

- `config_hash` (String) Hash of sensitive config fields for drift detection. Use this to detect if passwords have changed externally.
- `created_at` (String) Creation timestamp.
- `health_status` (String) Current health status: healthy, unhealthy, or unknown.
- `id` (String) Check ID.
- `last_checked` (String) Last check timestamp.
- `org_id` (String) Organization ID.
- `updated_at` (String) Last update timestamp.


