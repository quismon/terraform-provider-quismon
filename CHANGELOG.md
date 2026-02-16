# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Initial release preparation for Terraform Registry

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
