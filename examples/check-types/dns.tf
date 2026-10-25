# DNS record check
resource "quismon_check" "mx" {
  name             = "mail-dns"
  type             = "dns"
  interval_seconds = 300
  regions          = ["na-east-ewr"]

  config = {
    domain           = "example.com"
    record_type      = "MX"            # A | AAAA | CNAME | MX | TXT | NS | SOA
    expected_records = ["mail.example.com"]
    timeout_seconds  = 5
  }
}

# DNSSEC - DNSKEY/DS presence plus a normal lookup of the requested type
resource "quismon_check" "signed-zone" {
  name             = "zone-signed"
  type             = "dnssec"
  interval_seconds = 3600
  regions          = ["na-east-ewr"]

  config = {
    domain          = "example.com"
    nameservers     = ["1.1.1.1", "8.8.8.8"]
    timeout_seconds = 10
  }
}
