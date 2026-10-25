# TCP port check (also the base for inverted firewall assertions)
resource "quismon_check" "db_port" {
  name             = "db-port-reachable"
  type             = "tcp"
  interval_seconds = 300
  regions          = ["na-east-ewr"]

  config = {
    host            = "db.example.com"
    port            = 5432
    timeout_seconds = 5
  }
}

# ICMP ping - packet loss, min/max/avg RTT
resource "quismon_check" "gateway" {
  name             = "gateway-ping"
  type             = "ping"
  interval_seconds = 60
  regions          = ["eu-central-fra"]

  config = {
    host            = "gateway.example.com"
    count           = 3
    timeout_seconds = 5
  }
}

# UDP probe - send a payload, optionally match the response
resource "quismon_check" "syslog" {
  name             = "syslog-udp"
  type             = "udp"
  interval_seconds = 300
  regions          = ["eu-central-fra"]

  config = {
    host             = "syslog.example.com"
    port             = 514
    payload          = "<134>quismon probe"
    timeout_seconds  = 5
  }
}
