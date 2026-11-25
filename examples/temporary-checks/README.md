# Temporary Checks (deploy-window monitoring)

Temporary checks monitor a deploy window and clean themselves up:
`expires_after_seconds` disables the check when the window passes and the
platform removes it after a 7-day grace window (during which it stays
visible in the console, disabled, with an "Expired · removes <date>"
badge).

## Lifecycle

| Phase | What happens |
|---|---|
| Running | check executes normally; console shows the countdown badge |
| Expired | check is DISABLED (no executions) and stays visible for 7 days |
| Grace passed | check is deleted by the platform |

## Making one permanent

Set `expires_after_seconds = 0` on an update: the expiry is cleared and an
expired (disabled) check is re-enabled. Unverified free-tier orgs cannot
opt out of expiry (forced until the org email is verified).

## Notes

- `expires_after_seconds = 0` is only meaningful on UPDATE; on CREATE use a
  positive duration (or omit the attribute for permanent checks on verified
  orgs).
- Example sets 3 hours; any positive duration works.
