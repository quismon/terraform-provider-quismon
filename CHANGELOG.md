# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Initial release preparation for Terraform Registry

## [1.0.4] - 2026-02-17

### Fixed

- **CRITICAL**: Fixed `created_at` remaining unknown after update operations
  - This was causing "Provider returned invalid result object after apply" errors
  - All computed fields are now properly set in the Update function

- **CRITICAL**: Fixed `notification_channel.enabled` remaining unknown after create
  - The enabled field is now properly set in the state after creation

### Changed

- API now normalizes `expected_status` to array format automatically
  - Accepts string ("200"), comma-separated ("200,201"), or array ([200])
  - This prevents type mismatch errors in the Rust checker

- Rust checker now has flexible `expected_status` deserializer
  - Gracefully handles string, integer, or array formats
  - Prevents "invalid type: string, expected a sequence" errors

## [1.0.3] - 2026-02-16

### Added

- Documentation for multi-step checks with variable extraction
- Documentation for body assertions (expected_content, content_match_type)
- Documentation for inverted checks
- Full list of 31 monitoring regions

### Fixed

- Improved error messages for API key auto-detection

## [1.0.2] - 2026-02-16

### Fixed

- **CRITICAL**: Fixed signup/provider cycle by changing error to warning
  - Provider now warns instead of errors when reading API key from state
  - Allows two-phase apply pattern to work

- **CRITICAL**: Fixed state auto-detection 401 error
  - Added proper Bearer prefix handling in client
  - Fixed whitespace trimming in API key

- **CRITICAL**: Fixed `notification_channel.enabled` unknown after apply
  - Added SetAttribute for enabled field in Create function

## [1.0.0] - 2026-02-16

### Added

- **Resources**
  - `quismon_check` - Manage health checks (HTTP/HTTPS, TCP, Ping, DNS, SSL)
  - `quismon_alert_rule` - Configure alert conditions
  - `quismon_notification_channel` - Set up email, ntfy, webhook, and Slack notifications
  - `quismon_signup` - Self-service organization signup

- **Data Sources**
  - `quismon_check` - Query a specific check
  - `quismon_checks` - Query all checks
  - `quismon_notification_channel` - Query a specific notification channel

- **Features**
  - Multi-region monitoring support
  - Flexible alert conditions (health status, failure threshold, response time)
  - Seamless quickstart with automatic API key detection from state
  - Import support for existing resources

### Documentation

- Comprehensive README with examples for all check types
- Examples directory with basic, multi-region, complete, and DNS/SSL examples

### Testing

- Unit tests for all resources and data sources
- Acceptance tests for integration validation
