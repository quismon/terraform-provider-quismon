# Basic Example

This example creates a simple monitoring setup with:
- One HTTPS check monitoring a website
- One email notification channel
- One alert rule (3 consecutive failures trigger)

## Usage

```bash
# Set your API key
export QUISMON_API_KEY="your-api-key-here"

# Initialize Terraform
terraform init

# Plan the changes
terraform plan

# Apply the changes
terraform apply
```

## Requirements

- Terraform >= 1.0
- Quismon API key with write permissions
