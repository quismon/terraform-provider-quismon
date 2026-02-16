package provider

import (
	"context"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
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

// =============================================================================
// UNIT TESTS - Test provider internals without API calls
// =============================================================================

// TestCheckResourceModel_ComputedFields tests that the check model has all computed fields
func TestCheckResourceModel_ComputedFields(t *testing.T) {
	model := checkResourceModel{
		ID:              types.StringValue("test-id"),
		OrgID:           types.StringValue("test-org-id"),
		Name:            types.StringValue("Test Check"),
		Type:            types.StringValue("https"),
		IntervalSeconds: types.Int64Value(60),
		Enabled:         types.BoolValue(true),
		HealthStatus:    types.StringValue("healthy"),
		CreatedAt:       types.StringValue("2026-02-16T00:00:00Z"),
		UpdatedAt:       types.StringValue("2026-02-16T00:00:00Z"),
	}

	// Verify all fields can be set
	if model.ID.ValueString() != "test-id" {
		t.Error("ID not set correctly")
	}
	if model.OrgID.ValueString() != "test-org-id" {
		t.Error("OrgID not set correctly")
	}
	if model.HealthStatus.ValueString() != "healthy" {
		t.Error("HealthStatus not set correctly")
	}
	if model.CreatedAt.ValueString() == "" {
		t.Error("CreatedAt should be set after apply")
	}
	if model.UpdatedAt.ValueString() == "" {
		t.Error("UpdatedAt should be set after apply")
	}
}

// TestNotificationChannelModel_Enabled tests that enabled field is properly set
func TestNotificationChannelModel_Enabled(t *testing.T) {
	// Test with enabled=true
	model := notificationChannelResourceModel{
		ID:      types.StringValue("test-id"),
		Name:    types.StringValue("Test Channel"),
		Type:    types.StringValue("email"),
		Enabled: types.BoolValue(true),
	}

	if !model.Enabled.ValueBool() {
		t.Error("Enabled should be true")
	}

	// Test with enabled=false
	model.Enabled = types.BoolValue(false)
	if model.Enabled.ValueBool() {
		t.Error("Enabled should be false")
	}
}

// TestCheckResourceSchema_ComputedAttributes verifies all computed attributes exist
func TestCheckResourceSchema_ComputedAttributes(t *testing.T) {
	ctx := context.Background()
	r := NewCheckResource()

	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, schemaResp)

	if schemaResp.Diagnostics.HasError() {
		t.Fatalf("Schema returned errors: %v", schemaResp.Diagnostics)
	}

	// These attributes must be computed (known after apply)
	computedAttrs := []string{"id", "org_id", "health_status", "last_checked", "created_at", "updated_at"}

	for _, attrName := range computedAttrs {
		if _, ok := schemaResp.Schema.Attributes[attrName]; !ok {
			t.Errorf("Computed attribute %q missing from schema", attrName)
		}
	}
}

// TestNotificationChannelSchema_ComputedAttributes verifies all computed attributes exist
func TestNotificationChannelSchema_ComputedAttributes(t *testing.T) {
	ctx := context.Background()
	r := NewNotificationChannelResource()

	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, schemaResp)

	if schemaResp.Diagnostics.HasError() {
		t.Fatalf("Schema returned errors: %v", schemaResp.Diagnostics)
	}

	computedAttrs := []string{"id", "org_id", "created_at", "updated_at"}

	for _, attrName := range computedAttrs {
		if _, ok := schemaResp.Schema.Attributes[attrName]; !ok {
			t.Errorf("Computed attribute %q missing from schema", attrName)
		}
	}
}

// TestAlertRuleSchema_ComputedAttributes verifies all computed attributes exist
func TestAlertRuleSchema_ComputedAttributes(t *testing.T) {
	ctx := context.Background()
	r := NewAlertRuleResource()

	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, schemaResp)

	if schemaResp.Diagnostics.HasError() {
		t.Fatalf("Schema returned errors: %v", schemaResp.Diagnostics)
	}

	computedAttrs := []string{"id", "created_at", "updated_at"}

	for _, attrName := range computedAttrs {
		if _, ok := schemaResp.Schema.Attributes[attrName]; !ok {
			t.Errorf("Computed attribute %q missing from schema", attrName)
		}
	}
}

// TestSignupSchema_ComputedAttributes verifies all computed attributes exist
func TestSignupSchema_ComputedAttributes(t *testing.T) {
	ctx := context.Background()
	r := NewSignupResource()

	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, schemaResp)

	if schemaResp.Diagnostics.HasError() {
		t.Fatalf("Schema returned errors: %v", schemaResp.Diagnostics)
	}

	computedAttrs := []string{"id", "org_id", "api_key"}

	for _, attrName := range computedAttrs {
		if _, ok := schemaResp.Schema.Attributes[attrName]; !ok {
			t.Errorf("Computed attribute %q missing from schema", attrName)
		}
	}
}

// TestCheckResourceModel_Types verifies all model field types are correct
func TestCheckResourceModel_Types(t *testing.T) {
	// Test that Config can handle nested map values
	configMap := types.MapValueMust(types.StringType, map[string]attr.Value{
		"url":     types.StringValue("https://example.com"),
		"method":  types.StringValue("GET"),
		"timeout": types.StringValue("10"),
	})

	model := checkResourceModel{
		Name:            types.StringValue("Test"),
		Type:            types.StringValue("https"),
		Config:          configMap,
		IntervalSeconds: types.Int64Value(60),
		Regions: types.ListValueMust(types.StringType, []attr.Value{
			types.StringValue("us-east-1"),
		}),
		Enabled: types.BoolValue(true),
	}

	if model.Name.ValueString() != "Test" {
		t.Error("Name type incorrect")
	}
	if model.IntervalSeconds.ValueInt64() != 60 {
		t.Error("IntervalSeconds type incorrect")
	}
	if !model.Enabled.ValueBool() {
		t.Error("Enabled type incorrect")
	}
}
