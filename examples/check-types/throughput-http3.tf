# Throughput - streamed download up to max_size_mb (tier-capped: 5 MB free,
# 25 adopter, 100 pro, 250 advanced), reports throughput_mbps
resource "quismon_check" "cdn_download" {
  name             = "cdn-throughput"
  type             = "throughput"
  interval_seconds = 900
  regions          = ["na-east-ewr", "sa-east-sao"]

  config = {
    url        = "https://cdn.example.com/10mb.bin"
    max_size_mb = 100
    timeout_seconds = 30
  }
}

# HTTP/3 - detects Alt-Svc/h3 support and validates the HTTP response
resource "quismon_check" "quic" {
  name             = "h3-support"
  type             = "http3"
  interval_seconds = 3600
  regions          = ["na-east-ewr"]

  config = {
    url             = "https://example.com"
    expected_status = [200]
  }
}
