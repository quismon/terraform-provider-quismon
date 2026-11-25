# Temporary checks: deploy-window monitoring that cleans itself up.
#
# expires_after_seconds makes the check auto-expire: it is DISABLED at
# expiry and stays visible (disabled) for a 7-day grace window before the
# platform removes it (Q-050). To keep a check permanently, clear the
# expiry: expires_after_seconds = 0 on update makes it permanent (and
# re-enables an expired check).
#
# NOTE: unverified free-tier orgs cannot create permanent checks — every
# check is forced to expire until the org contact email is verified, and
# expires_after_seconds = 0 is rejected there too.

terraform {
  required_providers {
    quismon = {
      source  = "quismon/quismon"
      version = ">= 1.0.0"
    }
  }
}

provider "quismon" {
  # api_key via QUISMON_API_KEY env var
}

# Monitor the NEW deployment for the length of the deploy window only.
variable "deploy_version" {
  type    = string
  default = "unset"
}

resource "quismon_check" "deploy_window" {
  name             = "deploy ${var.deploy_version} - canary"
  type             = "https"
  interval_seconds = 60

  # 3 hours of monitoring, then gone (7-day disabled grace before removal).
  expires_after_seconds = 3 * 60 * 60

  regions = ["na-east-ewr"]

  config = {
    url                  = "https://staging.example.com/health"
    method               = "GET"
    expected_status_code = "200"
    timeout_seconds      = "10"
  }
}

output "check_id" {
  value = quismon_check.deploy_window.id
}
