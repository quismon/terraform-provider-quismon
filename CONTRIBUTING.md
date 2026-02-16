# Contributing to the Quismon Terraform Provider

Thank you for your interest in contributing to the Quismon Terraform Provider!

## Development Setup

### Prerequisites

- Go 1.21 or later
- Terraform 1.0 or later
- A Quismon API key (for acceptance tests)

### Building

```bash
go build -o terraform-provider-quismon
```

### Running Tests

```bash
# Unit tests
go test ./...

# Acceptance tests (requires API key)
TF_ACC=1 QUISMON_API_KEY=qm_your_key go test ./... -v -timeout 30m
```

### Local Development

Create `~/.terraformrc`:

```hcl
provider_installation {
  dev_overrides {
    "quismon/quismon" = "/path/to/terraform-provider-quismon"
  }
  direct {}
}
```

Then run Terraform commands normally - it will use your local build.

## Making Changes

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/my-feature`)
3. Make your changes
4. Add tests for new functionality
5. Run tests to ensure everything passes
6. Commit your changes (`git commit -am 'Add some feature'`)
7. Push to the branch (`git push origin feature/my-feature`)
8. Open a Pull Request

## Code Style

- Run `go fmt` before committing
- Run `go vet` to catch issues
- Add documentation for new resources/data sources

## Release Process

Releases are automated via GitHub Actions. To create a new release:

1. Tag a new version: `git tag v1.0.0`
2. Push the tag: `git push origin v1.0.0`
3. GitHub Actions will build and publish the release

The version number should follow [Semantic Versioning](https://semver.org/).

## Questions?

Open an issue at https://github.com/quismon/terraform-provider-quismon/issues
