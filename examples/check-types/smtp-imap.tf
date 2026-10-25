# SMTP -> IMAP round trip with email-auth assertions (Only Quismon)
# Sends a uniquely-tagged mail, polls the inbox for it, and parses the
# Authentication-Results header. Credentials are sensitive.
resource "quismon_check" "mail_pipeline" {
  name             = "mail-roundtrip"
  type             = "smtp-imap"
  interval_seconds = 900
  regions          = ["eu-central-fra"]

  config_json = jsonencode({
    smtp_host = "smtp.example.com"
    smtp_port = 587
    imap_host = "imap.example.com"
    imap_port = 993
    from_address = "probe@example.com"
    to_address   = "probe-inbox@example.com"
    smtp_username = var.mail_username
    smtp_password = var.mail_password
    smtp_tls_mode = "starttls"       # auto | starttls | tls | none
    subject       = "Quismon round-trip probe"
    # Deliverability assertions - the check fails when the mail does not pass:
    require_spf_pass   = true
    require_dkim_pass  = true
    require_dmarc_pass = true
    max_wait_seconds   = 60
  })
}

variable "mail_username" { sensitive = true }
variable "mail_password" { sensitive = true }
