# Multistep with variable passing (extracts require Cafe tier+): log in,
# extract a token from the response, use it in the next step.
resource "quismon_check" "login_flow" {
  name             = "api-login-flow"
  type             = "multistep"
  interval_seconds = 300
  regions          = ["na-east-ewr"]

  config_json = jsonencode({
    steps = [
      {
        name = "login"
        type = "https"
        config = {
          url             = "https://api.example.com/login"
          method          = "POST"
          body            = jsonencode({ user = "probe", pass = "..." })
          expected_status = [200]
        }
        extracts = {
          token = { jsonpath = "$.access_token" }
        }
      },
      {
        name = "use-token"
        type = "https"
        config = {
          url     = "https://api.example.com/me"
          headers = { Authorization = "Bearer {{token}}" }
        }
      }
    ]
    fail_fast = true
  })
}
