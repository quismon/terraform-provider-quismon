# Traceroute - per-hop latency (dashboard graphs hop_latencies), path hash,
# completion and packet-loss thresholds. ICMP probes natively, with a
# traceroute-binary fallback.
resource "quismon_check" "route" {
  name             = "route-to-origin"
  type             = "traceroute"
  interval_seconds = 300
  regions          = ["na-east-ewr", "eu-central-fra"]

  config = {
    host             = "origin.example.com"
    max_hops         = 20
    probes_per_hop   = 3
    timeout_seconds  = 5
  }
}
