# Certificate expiry - goes unhealthy when fewer than warning_days remain
# (alert with a health_status: "down" rule - there is no separate ssl_expiry condition)
resource "quismon_check" "cert" {
  name             = "api-cert-expiry"
  type             = "ssl"
  interval_seconds = 3600
  regions          = ["na-east-ewr", "eu-central-fra"]

  config = {
    hostname        = "api.example.com"
    port            = 443
    warning_days    = 30
    # expected_fingerprint = "AA:BB:..."  # SHA-256, pin the cert
    # expected_domains    = ["api.example.com"] # SAN pinning
  }
}
