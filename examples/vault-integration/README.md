# HashiCorp Vault Integration Example

This example demonstrates how to use Quismon multistep checks to fetch secrets from HashiCorp Vault and use them in your monitoring workflows.

## Use Case

You don't want to store sensitive credentials (API keys, database passwords, etc.) in your monitoring configuration. Instead:

1. Store secrets in HashiCorp Vault
2. Use Quismon multistep checks to authenticate and fetch secrets at check time
3. Use the fetched secrets in subsequent API calls

This approach provides:
- **Zero credential storage** in monitoring config
- **Short-lived tokens** that expire automatically
- **Audit trail** of secret access in Vault
- **Centralized secret management** across your infrastructure

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Quismon Multistep Check                    │
├─────────────────────────────────────────────────────────────┤
│  Step 1: Login to Vault                                      │
│  POST /v1/auth/userpass/login/{username}                     │
│  Body: {"password": "***"}                                   │
│  ────────────────────────────────────────────────►           │
│                         Vault                                │
│  ◄────────────────────────────────────────────────           │
│  Response: {"auth": {"client_token": "hvs.xxx..."}}          │
│                                                              │
│  Step 2: Read Database Secret                                │
│  GET /v1/secret/data/database/production                     │
│  Header: X-Vault-Token: hvs.xxx...                           │
│  ────────────────────────────────────────────────►           │
│                         Vault                                │
│  ◄────────────────────────────────────────────────           │
│  Response: {"data": {"db_host": "...", "db_password": "..."}}│
│                                                              │
│  Step 3: Read API Secret                                     │
│  GET /v1/secret/data/api/external-service                    │
│  Header: X-Vault-Token: hvs.xxx...                           │
│  ────────────────────────────────────────────────►           │
│                         Vault                                │
│  ◄────────────────────────────────────────────────           │
│  Response: {"data": {"api_key": "..."}}                      │
│                                                              │
│  Step 4: Use Secrets                                         │
│  GET https://api.example.com/health                          │
│  Header: Authorization: Bearer {api_key}                     │
│  ────────────────────────────────────────────────►           │
│                    Your Service                              │
│  ◄────────────────────────────────────────────────           │
│  Response: {"healthy": true}                                 │
└─────────────────────────────────────────────────────────────┘
```

## Prerequisites

### HashiCorp Vault Setup

1. **Enable userpass authentication**:
   ```bash
   vault auth enable userpass
   ```

2. **Create a policy for the monitoring user**:
   ```hcl
   # monitoring-policy.hcl
   path "secret/data/database/*" {
     capabilities = ["read"]
   }
   path "secret/data/api/*" {
     capabilities = ["read"]
   }
   ```
   ```bash
   vault policy write monitoring monitoring-policy.hcl
   ```

3. **Create the monitoring user**:
   ```bash
   vault write auth/userpass/users/quismon-monitor \
     password="your-secure-password" \
     policies="monitoring"
   ```

4. **Store your secrets**:
   ```bash
   vault kv put secret/database/production \
     db_host="prod-db.example.com" \
     db_port="5432" \
     db_user="app_user" \
     db_password="super-secret-password"

   vault kv put secret/api/external-service \
     api_key="sk-your-api-key" \
     api_url="https://api.example.com"
   ```

### Quismon Setup

1. Get your API key from [Quismon Console](https://console.quismon.com/api-keys)
2. Note your preferred check region (e.g., `na-east-ewr`)

## Usage

1. Create a `terraform.tfvars` file:
   ```hcl
   quismon_api_key = "qm_your_api_key"
   vault_base_url  = "https://vault.yourcompany.com"
   vault_username  = "quismon-monitor"
   vault_password  = "your-secure-password"
   ```

2. Initialize and apply:
   ```bash
   terraform init
   terraform apply
   ```

3. View the check in Quismon Console at the URL output by Terraform

## Security Considerations

1. **Use least-privilege policies**: The monitoring user should only have read access to the specific secrets it needs.

2. **Consider AppRole instead of userpass**: For production, AppRole provides better security than username/password:
   ```hcl
   # Step 1 would change to:
   {
     name = "vault-login"
     type = "https"
     config = {
       url    = "${var.vault_base_url}/v1/auth/approle/login"
       method = "POST"
       body   = jsonencode({
         role_id   = var.role_id
         secret_id = var.secret_id
       })
       expected_status = [200]
     }
     extracts = {
       vault_token = { jsonpath = "$.auth.client_token" }
     }
   }
   ```

3. **Use short TTLs**: Configure Vault to issue tokens with short TTLs (e.g., 1 hour).

4. **Never commit secrets**: Use environment variables or a secrets manager for sensitive variables.

5. **Audit access**: Monitor Vault audit logs for unauthorized access attempts.

## Troubleshooting

### Check fails at Step 1 (Login)
- Verify username and password are correct
- Check that userpass auth is enabled
- Ensure the user exists and has valid policies

### Check fails at Step 2/3 (Read Secret)
- Verify the token is being extracted correctly (check step results)
- Ensure the policy allows read access to the secret path
- For KV v2, use `secret/data/{path}` not `secret/{path}`

### Check fails at Step 4 (Use Secrets)
- Verify variables are being extracted correctly from previous steps
- Check that the API endpoint is reachable from Quismon's check regions
- Validate the Authorization header format matches what your API expects

## Related Examples

- [Multistep Basic](../multistep/) - Basic multistep check patterns
- [Multistep Advanced](../multistep-advanced/) - Complex workflows with conditionals

---

## EGM Contract

### Interface (main.tf)

- `quismon_check.vault_secret_fetch` — 4-step multistep: `vault-login` (POST `/v1/auth/userpass/login/{user}`, extracts `vault_token` + `lease_duration`), `read-db-secret` (GET KV v2 `secret/data/database/production`, extracts db_host/port/user/password), `read-api-secret`, `use-secrets` (GET `{{api_url}}/health` with `Bearer {{api_key}}`); `fail_fast = false`, total `timeout_seconds = 60`
- `quismon_alert_rule.vault_check_failure` — `health_status = "down"`; `notification_channel_ids` is commented out (fill in before applying)
- Outputs: `check_id`, `check_name`, `alert_rule_id`
- Variables: `quismon_api_key`, `vault_base_url`, `vault_username`, `vault_password` (sensitive)

### Invariants

- Vault tokens ride the `X-Vault-Token` header via `{{vault_token}}` interpolation; secrets never land in Terraform state beyond the variable values themselves (`main.tf` steps 2–3)
- KV v2 paths use `secret/data/{path}` (`main.tf` URLs)
- Provider pin `quismon >= 1.0.0` (`main.tf` terraform block)

### Side Effects

- Applying creates a check that authenticates to Vault every 300s and reads production secrets (ensure the monitoring user has least-privilege read policies)

### Verification

- No test manifest in this directory; inherits the provider module's verification: `cd /workspace/terraform-provider-quismon && CGO_ENABLED=0 go test ./...`
