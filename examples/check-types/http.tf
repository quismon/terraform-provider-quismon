# HTTP/HTTPS check - the workhorse
resource "quismon_check" "api_health" {
  name             = "api-health"
  type             = "https"
  interval_seconds = 60
  regions          = ["na-east-ewr", "eu-central-fra", "ap-southeast-sin"]

  config = {
    url              = "https://api.example.com/health"
    method           = "GET"                       # default GET
    expected_status  = [200, 201]                  # array, int, or comma string
    timeout_seconds  = 10
    follow_redirects = false
    verify_ssl       = false
    expected_content = "OK"                        # substring match by default
    # content_match_type = "contains"              # contains | exact | regex | not_contains
    # headers = { "X-Probe" = "quismon" }          # custom User-Agent requires Adopter tier+
    # body = "{}"                                  # for POST/PUT
  }
}
