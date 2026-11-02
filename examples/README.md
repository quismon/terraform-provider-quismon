# examples

## Overview

Reference Terraform configurations for the Quismon provider, organized by
scenario: quick start, per-check-type references, security/inverted
patterns, multistep workflows, browser checks, and full-stack setups. Each
subdirectory is a self-contained root module (`terraform init && terraform
apply`).

## Subdirectories & Boundaries

- advanced-checks: [./advanced-checks/README.md](./advanced-checks/README.md) - throughput and HTTP/3 (QUIC) checks
- basic: [./basic/README.md](./basic/README.md) - minimal stack: one HTTPS check + email channel + failure-threshold rule
- check-types: [./check-types/README.md](./check-types/README.md) - per-type reference configs (http, tcp/ping/udp, dns/dnssec, ssl, multistep, playwright, smtp-imap, throughput/http3, traceroute) + modifiers (inverted, starts-broken, expiring, dependencies)
- complete: [./complete/README.md](./complete/README.md) - production-shaped stack: 3 channels, 3 checks, 4 alert rules incl. `ssl_days_remaining`
- creative-use-cases: [./creative-use-cases/README.md](./creative-use-cases/README.md) - 13 patterns (blue/green, dependency health, DNS failover, inverted rate-limit, CDN, API contract, circuit breaker, ...)
- dns-ssl: [./dns-ssl/README.md](./dns-ssl/README.md) - DNS/DNSSEC record checks + SSL expiry/fingerprint/SAN pinning
- dom: [./dom/README.md](./dom/README.md) - dom checks: declarative assertions, scripted journeys with `$LP_*` secrets, dom-in-multistep variable chains
- full-stack: [./full-stack/README.md](./full-stack/README.md) - layered stack: DNS/SSL/website/API/database + inverted security checks
- infosec-ssl-fingerprint: [./infosec-ssl-fingerprint/README.md](./infosec-ssl-fingerprint/README.md) - security-focused SSL monitoring with Let's Encrypt/certbot integration
- inverted-checks: [./inverted-checks/README.md](./inverted-checks/README.md) - alert-on-success patterns (admin exposure, blocked ports, version watching)
- multi-region: [./multi-region/README.md](./multi-region/README.md) - one check from five regions + PagerDuty + latency rule
- multistep: [./multistep/README.md](./multistep/README.md) - multistep basics: login flows, variable extraction, fail_fast semantics
- multistep-advanced: [./multistep-advanced/README.md](./multistep-advanced/README.md) - OAuth2, GraphQL chaining, webhook signatures, pagination, circuit breaker, HTTP/3 workflows
- playwright: [./playwright/README.md](./playwright/README.md) - browser checks: low-code actions, full scripts with variables, API testing, e-commerce flow, cross-browser
- vault-integration: [./vault-integration/README.md](./vault-integration/README.md) - fetch secrets from HashiCorp Vault at check time via multistep

## Invariants

- All examples pin `source = "quismon/quismon"` in `required_providers` (versions pinned only in dom `~> 1.4`, playwright/vault `>= 1.0.0`)
- Credentials enter via sensitive variables or `QUISMON_API_KEY` env var; placeholder URLs/webhooks/fingerprints across examples must be replaced before apply
- Complex/nested check configs use `config_json = jsonencode({...})`; flat string configs use the `config` map

## Side Effects

- Every example, when applied, creates real Quismon resources (checks, channels, alert rules) that start executing immediately

## Verification

- No test manifest in this directory; inherits the provider module's verification: `cd /workspace/terraform-provider-quismon && CGO_ENABLED=0 go test ./...` (acceptance suite via `./scripts/test.sh`)
