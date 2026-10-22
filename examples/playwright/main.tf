# =============================================================================
# Quismon Playwright Check Example
# =============================================================================
# This example demonstrates how to use Playwright browser automation
# checks for comprehensive end-to-end testing.
#
# Use Cases:
# - Browser automation testing
# - Low-code actions (no coding required)
# - Full TypeScript/JavaScript scripts
# - Form validation and submission
# - Screenshot capture on failure
# =============================================================================

terraform {
  required_providers {
    quismon = {
      source  = "quismon/quismon"
      version = ">= 1.0.0"
    }
  }
}

provider "quismon" {
  # Get your API key from https://console.quismon.com/api-keys
  api_key = var.quismon_api_key
}

# =============================================================================
# VARIABLES
# =============================================================================

variable "quismon_api_key" {
  description = "Quismon API key"
  type        = string
  sensitive   = true
}

variable "target_base_url" {
  description = "Base URL of the target application to test"
  type        = string
  default     = "https://example.com"
}

variable "test_username" {
  description = "Test username for authenticated endpoints"
  type        = string
  default     = "testuser"
  sensitive   = true
}

variable "test_password" {
  description = "Test password for authenticated endpoints"
  type        = string
  default     = "testpass123"
  sensitive   = true
}

variable "test_api_token" {
  description = "Test API token for API endpoints"
  type        = string
  sensitive   = true
}

# =============================================================================
# PLAYWRIGHT CHECK: Low-Code Actions
# =============================================================================
# This check uses low-code actions instead of a full script.
# Perfect for simple workflows that don't need custom scripting.

resource "quismon_check" "low_code_form_test" {
  name             = "Playwright-Low-Code-Form-Test"
  type             = "playwright"
  enabled          = true
  interval_seconds = 300
  regions          = ["na-east-ewr"]

  config_json = jsonencode({
    browser              = "chromium"
    viewport_width       = 1280
    viewport_height      = 720
    timeout_seconds      = 60
    screenshot_on_failure = true
    trace_enabled        = true

    actions = [
      {
        name     = "Navigate to form"
        action   = "goto"
        url      = "${var.target_base_url}/contact"
      },
      {
        name     = "Wait for form"
        action   = "wait"
        wait_for = "networkidle"
      },
      {
        name     = "Fill Name"
        action   = "fill"
        selector = "#name"
        value    = "Test User"
      },
      {
        name     = "Fill Email"
        action   = "fill"
        selector = "#email"
        value    = "test@example.com"
      },
      {
        name     = "Fill Message"
        action   = "fill"
        selector = "#message"
        value    = "This is a test message from Quismon Playwright"
      },
      {
        name     = "Submit Form"
        action   = "click"
        selector = "#submit-btn"
      },
      {
        name     = "Verify Success"
        action   = "wait"
        wait_for = ".success-message"
        timeout_ms = 5000
      }
    ]

    extracts = {
      success_text = {
        selector  = ".success-message"
        transform = "trim"
      }
    }
  })
}

# =============================================================================
# PLAYWRIGHT CHECK: Full Script with Variables
# =============================================================================
# This check uses a full TypeScript script with injected variables.
# Perfect for complex workflows that need custom logic.

resource "quismon_check" "full_script_test" {
  name             = "Playwright-Full-Script-Test"
  type             = "playwright"
  enabled          = true
  interval_seconds = 300
  regions          = ["na-east-ewr"]

  config_json = jsonencode({
    browser              = "chromium"
    viewport_width       = 1920
    viewport_height      = 1080
    timeout_seconds      = 120
    screenshot_on_failure = true
    screenshot_on_success = false
    trace_enabled        = false
    video_enabled        = false
    headless              = true

    variables = {
      base_url      = var.target_base_url
      username     = var.test_username
      password     = var.test_password
      expected_title = "Dashboard"
    }

    sensitive_variables = ["password"]

    script_inline = <<SCRIPT
const { page } = this;

// Navigate to login page
await page.goto('{{base_url}}/login');
await page.waitForLoadState('networkidle');

// Fill login form
await page.fill('#username', '{{username}}');
await page.fill('#password', '{{password}}');

// Submit login
await page.click('#login-btn');

// Wait for redirect to dashboard
await page.waitForURL('**/dashboard**', { timeout: 10000 });

// Verify we're on the dashboard
const title = await page.title();
if (!title.includes('{{expected_title}}')) {
  throw new Error('Expected to be on dashboard, but title is: ' + title);
}

// Take a screenshot of the dashboard
await page.screenshot({ path: 'dashboard.png', fullPage: true });

// Extract user info
const userName = await page.textContent('.user-name');
const userEmail = await page.textContent('.user-email');

return {
  success: true,
  userName: userName,
  userEmail: userEmail
};
SCRIPT
  })
}

# =============================================================================
# PLAYWRIGHT CHECK: API Endpoint Test
# =============================================================================
# This check tests API endpoints with authentication headers.

resource "quismon_check" "api_endpoint_test" {
  name             = "Playwright-API-Endpoint-Test"
  type             = "playwright"
  enabled          = true
  interval_seconds = 300
  regions          = ["na-east-ewr"]

  config_json = jsonencode({
    browser              = "chromium"
    timeout_seconds      = 30
    screenshot_on_failure = true

    variables = {
      api_url    = "${var.target_base_url}/api"
      api_token  = var.test_api_token
    }

    sensitive_variables = ["api_token"]

    script_inline = <<SCRIPT
const { page, context } = this;

// Set authentication header
await context.setExtraHTTPHeaders({
  'Authorization': 'Bearer {{api_token}}',
  'Content-Type': 'application/json'
});

// Test health endpoint
const healthResponse = await page.request.get('{{api_url}}/health');

if (healthResponse.status() !== 200) {
  throw new Error('API health check failed with status ' + healthResponse.status());
}

const healthData = await healthResponse.json();
if (healthData.status !== 'ok') {
  throw new Error('API status is not ok: ' + healthData.status);
}

// Test authenticated endpoint
const userResponse = await page.request.get('{{api_url}}/user/profile');

if (userResponse.status() !== 200) {
  throw new Error('Failed to get user profile: ' + userResponse.status());
}

const userData = await userResponse.json();

return {
  success: true,
  healthStatus: healthData.status,
  userId: userData.id,
  userName: userData.name
};
SCRIPT
  })
}

# =============================================================================
# PLAYWRIGHT CHECK: Multi-Step E-commerce Flow
# =============================================================================
# This check tests a complete e-commerce checkout flow.

resource "quismon_check" "ecommerce_flow_test" {
  name             = "Playwright-Ecommerce-Flow"
  type             = "playwright"
  enabled          = true
  interval_seconds = 600
  regions          = ["na-east-ewr", "eu-central-fra"]

  config_json = jsonencode({
    browser              = "chromium"
    viewport_width       = 1366
    viewport_height      = 768
    timeout_seconds      = 180
    screenshot_on_failure = true
    trace_enabled        = true

    variables = {
      base_url = var.target_base_url
    }

    script_inline = <<SCRIPT
const { page } = this;

// Step 1: Navigate to products page
await page.goto('{{base_url}}/products');
await page.waitForSelector('.product-list');

// Step 2: Click on first product
const firstProduct = await page.waitForSelector('.product-item');
await firstProduct.click();

// Step 3: Wait for product details
await page.waitForSelector('.product-details');

// Step 4: Add to cart
await page.click('#add-to-cart');
await page.waitForSelector('.cart-notification');

// Step 5: View cart
await page.click('#view-cart');
await page.waitForSelector('.cart-items');

// Step 6: Verify item in cart
const cartItems = await page.$$('.cart-item');
if (cartItems.length === 0) {
  throw new Error('No items in cart');
}

// Step 7: Proceed to checkout
await page.click('#proceed-checkout');
await page.waitForSelector('#checkout-form');

// Step 8: Fill shipping info (optional)
try {
  await page.fill('#shipping-name', 'Test User');
  await page.fill('#shipping-address', '123 Test St');
  await page.fill('#shipping-city', 'Test City');
  await page.fill('#shipping-zip', '12345');
} catch (e) {
  // Form fields might not be required
  console.log('Shipping info fields not found, skipping');
}

// Step 9: Verify we can continue
const continueBtn = await page.$('#continue-btn');
if (continueBtn) {
  // Don't actually submit, order, just verify we flow works
  console.log('Checkout flow verified successfully');
} else {
  throw new Error('Continue button not found');
}

return {
  success: true,
  cartItemCount: cartItems.length
  flowCompleted: true
};
SCRIPT
  })
}

# =============================================================================
# PLAYWRIGHT CHECK: Cross-Browser Testing
# =============================================================================
# This check runs the same test across multiple browsers to ensure compatibility.

resource "quismon_check" "cross_browser_test" {
  name             = "Playwright-Cross-Browser"
  type             = "playwright"
  enabled          = false  # Disabled by default - enable for testing
  interval_seconds = 300
  regions          = ["na-east-ewr"]

  config_json = jsonencode({
    # Test with Firefox instead of Chromium
    browser              = "firefox"
    viewport_width       = 1280
    viewport_height      = 720
    timeout_seconds      = 60
    screenshot_on_failure = true

    variables = {
      target_url = var.target_base_url
    }

    script_inline = <<SCRIPT
const { page } = this;

await page.goto('{{target_url}}');
await page.waitForLoadState('networkidle');

const title = await page.title();
const userAgent = await page.evaluate(() => navigator.userAgent);

return {
  success: true,
  title: title,
  browser: 'Firefox',
  userAgent: userAgent
};
SCRIPT
  })
}

# =============================================================================
# OUTPUTS
# =============================================================================

output "low_code_check_id" {
  value = quismon_check.low_code_form_test.id
}

output "full_script_check_id" {
  value = quismon_check.full_script_test.id
}

output "api_endpoint_check_id" {
  value = quismon_check.api_endpoint_test.id
}

output "ecommerce_flow_check_id" {
  value = quismon_check.ecommerce_flow_test.id
}

output "cross_browser_check_id" {
  value = quismon_check.cross_browser_test.id
}
