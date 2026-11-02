# internal

## Overview

Private packages of the Quismon Terraform provider: the API client layer and
the Plugin Framework provider implementation.

## Subdirectories & Boundaries

- client: [./client/README.md](./client/README.md) - HTTP client for the Quismon REST API (`DoRequest`, envelope decoding, typed CRUD for checks/channels/alert-rules, quick-signup helper)
- provider: [./provider/README.md](./provider/README.md) - provider schema, resources (`quismon_check`, `quismon_alert_rule`, `quismon_notification_channel`, `quismon_signup`, `quismon_organization_otlp`, `quismon_status_page`), data sources, plan-time validation
