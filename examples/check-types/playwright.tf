# Playwright browser check (Pro+ single region, Advanced+ multi-region).
# Low-code actions or a full script_inline; screenshots/traces via S3.
resource "quismon_check" "checkout" {
  name             = "checkout-flow"
  type             = "playwright"
  interval_seconds = 900
  regions          = ["na-east-ewr"]

  config_json = jsonencode({
    browser         = "chromium"
    timeout_seconds = 60
    viewport_width  = 1280
    viewport_height = 720
    actions = [
      { action = "goto",  url = "https://shop.example.com" },
      { action = "click", selector = ".product:first-child button.add" },
      { action = "click", selector = "#checkout" },
      { action = "wait",  selector = "#order-complete", timeout_ms = 10000 },
      { action = "screenshot" }
    ]
    # or a full script:
    # script_inline = file("${path.module}/checkout.spec.js")
  })
}
