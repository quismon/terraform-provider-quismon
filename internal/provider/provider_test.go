package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"quismon": providerserver.NewProtocol6WithError(New("test")()),
}

func testAccPreCheck(t *testing.T) {
	// Verify required environment variables are set
	if v := os.Getenv("QUISMON_API_KEY"); v == "" {
		t.Fatal("QUISMON_API_KEY must be set for acceptance tests")
	}

	if v := os.Getenv("QUISMON_BASE_URL"); v == "" {
		t.Log("QUISMON_BASE_URL not set, using default")
	}
}

func TestProvider(t *testing.T) {
	// Basic provider instantiation test
	provider := New("test")()
	if provider == nil {
		t.Fatal("Provider is nil")
	}
}

func TestProvider_Schema(t *testing.T) {
	provider := New("test")()

	// Ensure the provider schema is valid
	schemaReq := provider.Schema
	if schemaReq == nil {
		t.Fatal("Provider schema method is nil")
	}
}
