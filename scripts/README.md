# scripts

## Overview

Developer utility scripts for the Quismon Terraform provider: the acceptance
test runner and the "TMI" public-docs leak checker.

## Interface

- `test.sh [mode]` — provider test runner; modes: `all` (default), `unit`, `acceptance`, `matrix`, `integration`, `quick` (`test.sh` case statement)
- `tmi_check.py [repo_root]` — scans public documentation surfaces for implementation-internals leakage; exits 1 with a violation list, 0 when clean (`tmi_check.py:main`)

## Invariants

- `test.sh` requires `QUISMON_API_KEY` and aborts without it; `QUISMON_BASE_URL` defaults to `https://api.quismon.com` (`test.sh` env checks); it runs with `set -e`
- Acceptance modes set `TF_ACC=1` and target `./internal/provider` with per-mode `-run` filters and timeouts (`test.sh` case statement)
- `tmi_check.py` scans different surfaces by checkout mode: monorepo (when `website/` exists → `openapi.yaml`, `API_DOCUMENTATION.md`, `website/`, `docs/`, provider `templates/`+`docs/`, `infrastructure/status-page`) vs standalone provider checkout (`README.md`, `templates/`, `docs/`) (`tmi_check.py:PUBLIC_FILES/PUBLIC_DIRS`)
- Violations can be excused only by an inline `tmi-allow: <pattern-name>` marker on the same line or up to 3 lines above (YAML multi-line scalars), or the built-in `openapi.yaml` allowlist entries (`tmi_check.py:line_allowed`, `BUILTIN_ALLOWED`)
- Only `.html`, `.md`, `.yaml`, `.yml`, `.tmpl` files are scanned (`tmi_check.py:public_files`)
- `TODO`/`FIXME` matched case-insensitively; `XXX` case-sensitively (lowercase `xxx`/`yyy` placeholders allowed) (`tmi_check.py:main`)

## Side Effects

- `test.sh` executes `go test` (unit mode: local only; other modes create/modify real Quismon resources via acceptance tests against `QUISMON_BASE_URL`)
- `tmi_check.py` is read-only

## Verification

- Run: `cd /workspace/terraform-provider-quismon && CGO_ENABLED=0 go test ./...` (Go tests); `./scripts/test.sh unit` for the script-driven suite; `python3 scripts/tmi_check.py` for the docs check
