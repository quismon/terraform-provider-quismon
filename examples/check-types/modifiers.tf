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
