# internal/client

## Overview

HTTP client layer for the Quismon REST API, shared by all provider resources
and data sources. `client.go` provides the transport (`DoRequest`) and
envelope decoding (`UnmarshalAPIResponse`); the per-domain files add typed
CRUD methods. `signup.go` adds a standalone unauthenticated quick-signup
helper used by the `quismon_signup` resource.

## Interface

- `client.go`:
  - `Client` struct (`BaseURL`, `APIKey`, `HTTPClient`); `New(baseURL, apiKey string) (*Client, error)`
  - `(*Client).DoRequest(method, path string, body interface{}) ([]byte, error)`
  - `APIResponse` envelope type; `UnmarshalAPIResponse(data []byte, v interface{}) error`
- `checks.go`: `Check`, `CreateCheckRequest`, `UpdateCheckRequest`; methods `ListChecks`, `GetCheck(id)`, `GetCheckByName(name)`, `CreateCheck(req)`, `UpdateCheck(id, req)`, `DeleteCheck(id)` against `/v1/checks`
- `notification_channels.go`: `NotificationChannel`, `Create/UpdateNotificationChannelRequest`; methods `List/Get/Create/Update/DeleteNotificationChannel`, `GetNotificationChannelByName` against `/v1/notification-channels`
- `alert_rules.go`: `AlertRule`, `Create/UpdateAlertRuleRequest`; methods `ListAlertRules(checkID)`, `GetAlertRule(checkID, ruleID)`, `CreateAlertRule(checkID, req)`, `UpdateAlertRule(checkID, ruleID, req)`, `DeleteAlertRule(checkID, ruleID)` against `/v1/checks/{checkID}/alerts`
- `signup.go`: `QuickSignupRequest`, `QuickSignupResponse`; package function `QuickSignup(ctx, baseURL, req)` against `POST /v1/auth/quick-signup`

## Invariants

- `New` requires a non-empty base URL; the API key is intentionally optional (signup flows) — auth-needing calls fail server-side without it (`client.go:New`)
- HTTP timeout is fixed at 30s (`client.go:New`, `signup.go:QuickSignup`)
- Auth header: `Bearer <key>` unless the key already starts with `Bearer ` (no double-prefix) (`client.go:DoRequest`)
- All responses are expected in the `{data, error, meta}` envelope; HTTP >= 400 or a non-nil `error` field returns a Go error (`client.go:DoRequest`, `client.go:UnmarshalAPIResponse`)
- Requests carry `User-Agent: terraform-provider-quismon/1.0` (`client.go:DoRequest`, `signup.go:QuickSignup`)
- `GetCheckByName` / `GetNotificationChannelByName` are client-side linear scans over the full list (`checks.go`, `notification_channels.go`)

## Side Effects

- Outbound HTTPS calls to the configured Quismon base URL; create/update/delete methods mutate Quismon state
- `QuickSignup` creates a new org + API key (unauthenticated, server-side rate limited)

## Verification

- Run: `cd /workspace/terraform-provider-quismon && CGO_ENABLED=0 go test ./...`
