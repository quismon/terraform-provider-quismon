# Dom checks: browser-lite rendering assertions via lightpanda.
# Declarative (adopter tier+) or scripted (pro tier+).

terraform {
  required_providers {
    quismon = {
      source  = "quismon/quismon"
      version = "~> 1.4"
    }
  }
}

provider "quismon" {
  api_key  = var.quismon_api_key
  base_url = "https://api.quismon.com"
}

variable "quismon_api_key" {
  type      = string
  sensitive = true
}

# Declarative: did the page load AND hydrate?
resource "quismon_check" "spa_health" {
  name             = "SPA - Dashboard Hydrates"
  type             = "dom"
  interval_seconds = 60
  regions          = ["na-east-ewr"]

  config = {
    url                  = "https://app.example.com/dashboard"
    expected_status_code = "200"
    expected_selector    = "#root .main-view"
    expected_text        = "Welcome"
    timeout_seconds      = "30"
  }
}

# Scripted: authenticated journey with secrets (pro tier+).
# Secret values are write-only - masked as ***REDACTED*** on read; the
# script references them as $LP_<UPPERCASED_KEY>.
resource "quismon_check" "authenticated_journey" {
  name             = "Dom - Login and Dashboard Journey"
  type             = "dom"
  interval_seconds = 300
  regions          = ["na-east-ewr"]

  # nested values (secrets) require config_json; flat string values can use config
  config_json = jsonencode({
    timeout_seconds = 45
    secrets = {
      APP_EMAIL    = "ops@example.com"
      APP_PASSWORD = var.app_password
    }
    script = <<EOT
const page = new Page();
await page.goto('https://app.example.com/login', { waitUntil: 'load' });
page.fill({ selector: '#email', value: '$LP_APP_EMAIL' });
page.fill({ selector: '#password', value: '$LP_APP_PASSWORD' });
page.click({ selector: 'button[type=submit]' });
page.waitForSelector('[data-testid=dashboard-table]', { timeout: 15000 });
const data = page.extract({ rows: { selector: 'tr[data-row]', limit: 5, fields: { id: { attr: 'data-id' } } } });
if (!data.rows || data.rows.length < 1) throw new Error('dashboard empty after login');
return { rows: data.rows.length };
EOT
  })
}

# Multistep: http state feeds a dom assertion
resource "quismon_check" "build_deployment" {
  name             = "Dom - Deployed Build Is Live"
  type             = "multistep"
  interval_seconds = 300
  regions          = ["na-east-ewr"]

  config_json = jsonencode({
    steps = [
      {
        name   = "api: build id"
        type   = "http"
        config = { url = "https://app.example.com/api/health", expected_status = 200, timeout_seconds = 10 }
        extracts = { build = { jsonpath = "$.build" } }
      },
      {
        name   = "dom: deployed build is live"
        type   = "dom"
        config = { url = "https://app.example.com/", timeout_seconds = 30 }
      }
    ]
  })
}

# Mixed multistep: dom extraction feeds an HTTP POST, the response token
# rides a Bearer header, and both vars land in a final dom assertion.
# (Validated live against firingrange.quismon.com - see internal-docs.)
resource "quismon_check" "dom_mixed_flow" {
  name             = "Journey - Dom Extract Feeds POST Auth Chain"
  type             = "multistep"
  interval_seconds = 300
  regions          = ["na-east-ewr", "eu-west-ams"]

  config_json = jsonencode({
    timeout_seconds = 120
    steps = [
      # 1. Pull a value OUT of a rendered page (CSRF-token pattern)
      {
        name   = "dom: extract page title"
        type   = "dom"
        config = { url = "https://example.com/", timeout_seconds = 20 }
        extracts = { site_title = { selector = "title", attribute = "text" } }
      },
      # 2. POST: the extracted var rides the JSON body (interpolation is
      #    generic - any string field of any step config)
      {
        name   = "http: POST login (body uses {{site_title}})"
        type   = "http"
        config = {
          url             = "https://firingrange.quismon.com/auth/login"
          method          = "POST"
          headers         = { "Content-Type" = "application/json" }
          body            = "{\"username\": \"{{site_title}}\", \"password\": \"ci-password\"}"
          expected_status = 200
          timeout_seconds = 15
        }
        extracts = {
          token  = { jsonpath = "$.access_token" }
          whoami = { jsonpath = "$.user.username" }
        }
      },
      # 3. The extracted token rides an Authorization header
      {
        name   = "http: Bearer GET protected"
        type   = "http"
        config = {
          url             = "https://firingrange.quismon.com/auth/protected"
          headers         = { "Authorization" = "Bearer {{token}}" }
          expected_status = 200
          timeout_seconds = 15
        }
        extracts = { grant = { jsonpath = "$.message" } }
      },
      # 4. Dom assertion with the chain's state injected into the JS
      #    (dom steps cannot set request headers - use query params or the
      #    scripted dom check's page.fill flow for header-auth journeys)
      {
        name   = "dom: assert chain state"
        type   = "dom"
        config = {
          url           = "https://example.com/"
          timeout_seconds = 20
          expected_js   = "document.readyState === 'complete' && '{{token}}'.startsWith('bearer-') && '{{whoami}}'.length > 0"
        }
      }
    ]
  })
}
