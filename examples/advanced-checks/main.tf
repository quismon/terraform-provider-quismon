# Advanced Check Types Example
# This example demonstrates throughput and HTTP/3 checks

terraform {
  required_providers {
    quismon = {
      source = "quismon/quismon"
    }
  }
}

provider "quismon" {
  # Configure via QUISMON_API_KEY environment variable
  # or explicitly set:
  # api_key = "qm_xxx"
}

# Throughput Check - Measures download speed
# Note: Tier limits apply - Free: 5MB, Paid: 100MB, Enterprise: 500MB
resource "quismon_check" "throughput" {
  name            = "CDN Throughput Test"
  type            = "throughput"
  interval_seconds = 300  # Every 5 minutes

  config_json = jsonencode({
    url            = "https://speed.cloudflare.com/__down?bytes=10000000"  # 10MB test file
    max_size_mb    = 5     # Max download size (tier-limited)
    timeout_seconds = 30
  })

  regions = ["na-east-ewr", "eu-west-ams"]
  enabled = true
}

# HTTP/3 (QUIC) Check - Tests HTTP/3 protocol support
resource "quismon_check" "http3" {
  name            = "HTTP/3 Endpoint Check"
  type            = "http3"
  interval_seconds = 60

  config_json = jsonencode({
    url              = "https://cloudflare.com"
    method           = "GET"
    expected_status  = [200, 301, 302]
    timeout_seconds  = 10
  })

  regions = ["na-east-ewr"]
  enabled = true
}

# HTTP/3 Check with Content Validation
resource "quismon_check" "http3_with_content" {
  name            = "HTTP/3 API Health"
  type            = "http3"
  interval_seconds = 120

  config_json = jsonencode({
    url                = "https://api.example.com/health"
    method             = "GET"
    expected_status    = [200]
    expected_content   = "\"status\":\"ok\""
    content_match_type = "contains"
    timeout_seconds    = 10
  })

  regions = ["na-east-ewr", "eu-west-ams"]
  enabled = true
}

# Throughput Check for Performance Benchmarking
resource "quismon_check" "bandwidth_benchmark" {
  name            = "Bandwidth Benchmark"
  type            = "throughput"
  interval_seconds = 600  # Every 10 minutes

  config_json = jsonencode({
    url             = "https://proof.ovh.net/files/1Mb.dat"  # 1MB test file
    max_size_mb     = 1
    timeout_seconds = 20
  })

  regions = ["na-east-ewr", "eu-west-ams", "ap-southeast-sin"]
  enabled = true
}

# Outputs
output "throughput_check_id" {
  value = quismon_check.throughput.id
}

output "http3_check_id" {
  value = quismon_check.http3.id
}
