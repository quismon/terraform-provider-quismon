# Publishing the Terraform Provider to the Registry

This guide explains how to publish the Quismon Terraform Provider to the Terraform Registry.

## Prerequisites

1. GitHub account
2. GPG key for signing releases (required by Terraform Registry)
3. Access to the Quismon GitHub organization

## Step 1: Create GitHub Organization

1. Go to https://github.com/organizations/new
2. Create organization named `quismon`
3. Set up the organization profile

## Step 2: Create the Provider Repository

1. Create a new repository: `github.com/quismon/terraform-provider-quismon`
2. Make it public (required for Terraform Registry)

## Step 3: Push the Provider Code

From your local machine:

```bash
# Navigate to the provider directory
cd terraform-provider-quismon

# Add the new remote (if not already configured)
git remote add origin https://github.com/quismon/terraform-provider-quismon.git

# Push to the new repository
git push -u origin main
```

Alternatively, if you want to keep the provider as a submodule:

```bash
# From the main quismon repo
cd /quismon
git submodule add https://github.com/quismon/terraform-provider-quismon.git terraform-provider-quismon-external
```

## Step 4: Generate GPG Key for Signing

The Terraform Registry requires releases to be signed with GPG.

```bash
# Generate a new GPG key (use RSA, 4096 bits)
gpg --full-generate-key

# List your keys to get the fingerprint
gpg --list-secret-keys --keyid-format=long

# Export the private key (for GitHub Actions)
gpg --armor --export-secret-keys YOUR_KEY_ID > private-key.asc

# Export the public key (for Terraform Registry)
gpg --armor --export YOUR_KEY_ID > public-key.asc
```

## Step 5: Configure GitHub Secrets

In the GitHub repository settings (Settings > Secrets and variables > Actions), add:

| Secret Name | Value |
|-------------|-------|
| `GPG_PRIVATE_KEY` | The contents of `private-key.asc` |
| `PASSPHRASE` | Your GPG key passphrase |

## Step 6: Create First Release

```bash
# Tag the first release
git tag v1.0.0
git push origin v1.0.0
```

GitHub Actions will automatically:
1. Build the provider for all platforms
2. Sign the release with GPG
3. Create a GitHub release with all binaries

## Step 7: Register with Terraform Registry

1. Go to https://registry.terraform.io/
2. Sign in with your GitHub account
3. Navigate to "Publish" > "Provider"
4. Select the `quismon/terraform-provider-quismon` repository
5. Add your GPG public key
6. Click "Publish Provider"

## Step 8: Verify Publication

After publishing, users can use the provider:

```hcl
terraform {
  required_providers {
    quismon = {
      source  = "quismon/quismon"
      version = "~> 1.0"
    }
  }
}

provider "quismon" {
  api_key = var.quismon_api_key
}
```

## Release Process

For future releases:

1. Update the code and merge to `main`
2. Update `CHANGELOG.md` with the changes
3. Tag a new version: `git tag v1.1.0`
4. Push the tag: `git push origin v1.1.0`
5. GitHub Actions handles the rest

## Files Created for Publishing

- `.github/workflows/release.yml` - GitHub Actions release workflow
- `.github/workflows/ci.yml` - CI workflow for PRs
- `.goreleaser.yml` - GoReleaser configuration
- `terraform-registry-manifest.json` - Registry metadata
- `LICENSE` - MIT license
- `CONTRIBUTING.md` - Contribution guidelines
- `CHANGELOG.md` - Version history

## Troubleshooting

### Release fails with GPG error

Ensure the GPG key is properly exported and the passphrase is correct.

### Provider not showing in registry

1. Ensure the repository is public
2. Check that the release has all required assets
3. Verify the GPG public key matches the private key

### Users can't install the provider

1. Check the Terraform Registry logs
2. Verify the provider binary is built for their platform
3. Ensure the version exists in the registry

## References

- [Terraform Registry Providers](https://www.terraform.io/registry/providers/publishing)
- [GoReleaser Documentation](https://goreleaser.com/)
- [GitHub Actions for Terraform Providers](https://github.com/hashicorp/terraform-provider-scaffolding)
