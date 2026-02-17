package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/quismon/terraform-provider-quismon/internal/client"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &organizationOTLPResource{}
	_ resource.ResourceWithConfigure   = &organizationOTLPResource{}
	_ resource.ResourceWithImportState = &organizationOTLPResource{}
)

// NewOrganizationOTLPResource is a helper function to simplify the provider implementation.
func NewOrganizationOTLPResource() resource.Resource {
	return &organizationOTLPResource{}
}

// organizationOTLPResource is the resource implementation.
type organizationOTLPResource struct {
	client *client.Client
}

// organizationOTLPResourceModel maps the resource schema data.
type organizationOTLPResourceModel struct {
	Enabled                types.Bool   `tfsdk:"enabled"`
	Endpoint               types.String `tfsdk:"endpoint"`
	Headers                types.Map    `tfsdk:"headers"`
	ExportIntervalSeconds  types.Int64  `tfsdk:"export_interval_seconds"`
}

// Metadata returns the resource type name.
func (r *organizationOTLPResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_otlp"
}

// Schema defines the schema for the resource.
func (r *organizationOTLPResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Configure OpenTelemetry OTLP metrics export for the organization. " +
			"This is a PAID feature and requires a 'paid' or 'enterprise' tier subscription.",
		Attributes: map[string]schema.Attribute{
			"enabled": schema.BoolAttribute{
				Description: "Enable OTLP metrics export",
				Required:    true,
			},
			"endpoint": schema.StringAttribute{
				Description: "OTLP HTTP endpoint URL (e.g., https://otlp.example.com:4318/v1/metrics)",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"headers": schema.MapAttribute{
				Description: "HTTP headers to include in OTLP requests (e.g., Authorization)",
				Optional:    true,
				ElementType: types.StringType,
			},
			"export_interval_seconds": schema.Int64Attribute{
				Description: "How often to export metrics in seconds (minimum 10, default 60)",
				Optional:    true,
				Validators: []validator.Int64{
					int64validator.AtLeast(10),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *organizationOTLPResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

// Create creates the resource and sets initial Terraform state.
func (r *organizationOTLPResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan organizationOTLPResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	updateReq := map[string]interface{}{
		"enabled": plan.Enabled.ValueBool(),
	}

	if !plan.Endpoint.IsNull() {
		updateReq["endpoint"] = plan.Endpoint.ValueString()
	}

	if !plan.Headers.IsNull() {
		var headers map[string]string
		diags := plan.Headers.ElementsAs(ctx, &headers, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		// Convert to map[string]interface{}
		headersInterface := make(map[string]interface{})
		for k, v := range headers {
			headersInterface[k] = v
		}
		updateReq["headers"] = headersInterface
	}

	if !plan.ExportIntervalSeconds.IsNull() {
		interval := int(plan.ExportIntervalSeconds.ValueInt64())
		updateReq["export_interval_seconds"] = interval
	}

	// Update OTLP config via API
	_, err := r.client.DoRequest("PUT", "/v1/org/otlp", updateReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating organization OTLP config",
			"Could not create OTLP config, unexpected error: "+err.Error(),
		)
		return
	}

	// Get the current config to populate state
	config, err := r.getOTLPConfig(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading organization OTLP config",
			"Could not read OTLP config after create: "+err.Error(),
		)
		return
	}

	// Map response to state
	plan.Enabled = types.BoolValue(config.Enabled)
	if config.Endpoint != nil {
		plan.Endpoint = types.StringValue(*config.Endpoint)
	} else {
		plan.Endpoint = types.StringNull()
	}

	// Set state
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *organizationOTLPResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state organizationOTLPResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed OTLP config from API
	config, err := r.getOTLPConfig(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading organization OTLP config",
			"Could not read OTLP config: "+err.Error(),
		)
		return
	}

	// Update state with refreshed values
	state.Enabled = types.BoolValue(config.Enabled)
	if config.Endpoint != nil {
		state.Endpoint = types.StringValue(*config.Endpoint)
	} else {
		state.Endpoint = types.StringNull()
	}

	if config.ExportIntervalSeconds != nil {
		state.ExportIntervalSeconds = types.Int64Value(int64(*config.ExportIntervalSeconds))
	} else {
		state.ExportIntervalSeconds = types.Int64Null()
	}

	// Note: Headers are not returned in full for security, so we keep the plan value
	// unless it's been explicitly changed

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *organizationOTLPResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan organizationOTLPResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	updateReq := map[string]interface{}{
		"enabled": plan.Enabled.ValueBool(),
	}

	if !plan.Endpoint.IsNull() {
		updateReq["endpoint"] = plan.Endpoint.ValueString()
	}

	if !plan.Headers.IsNull() {
		var headers map[string]string
		diags := plan.Headers.ElementsAs(ctx, &headers, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		headersInterface := make(map[string]interface{})
		for k, v := range headers {
			headersInterface[k] = v
		}
		updateReq["headers"] = headersInterface
	}

	if !plan.ExportIntervalSeconds.IsNull() {
		interval := int(plan.ExportIntervalSeconds.ValueInt64())
		updateReq["export_interval_seconds"] = interval
	}

	// Update OTLP config via API
	_, err := r.client.DoRequest("PUT", "/v1/org/otlp", updateReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating organization OTLP config",
			"Could not update OTLP config, unexpected error: "+err.Error(),
		)
		return
	}

	// Get the current config to populate state
	config, err := r.getOTLPConfig(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading organization OTLP config",
			"Could not read OTLP config after update: "+err.Error(),
		)
		return
	}

	// Map response to state
	plan.Enabled = types.BoolValue(config.Enabled)
	if config.Endpoint != nil {
		plan.Endpoint = types.StringValue(*config.Endpoint)
	} else {
		plan.Endpoint = types.StringNull()
	}

	if config.ExportIntervalSeconds != nil {
		plan.ExportIntervalSeconds = types.Int64Value(int64(*config.ExportIntervalSeconds))
	} else {
		plan.ExportIntervalSeconds = types.Int64Null()
	}

	// Set state
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *organizationOTLPResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Disable OTLP export
	updateReq := map[string]interface{}{
		"enabled":  false,
		"endpoint": "",
	}

	_, err := r.client.DoRequest("PUT", "/v1/org/otlp", updateReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting organization OTLP config",
			"Could not disable OTLP config, unexpected error: "+err.Error(),
		)
		return
	}
}

// ImportState imports an existing resource into Terraform.
func (r *organizationOTLPResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// This is a singleton resource, so we just import it
	// Retrieve the current OTLP config
	config, err := r.getOTLPConfig(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error importing organization OTLP config",
			"Could not read OTLP config: "+err.Error(),
		)
		return
	}

	// Create state from config
	state := organizationOTLPResourceModel{
		Enabled: types.BoolValue(config.Enabled),
	}

	if config.Endpoint != nil {
		state.Endpoint = types.StringValue(*config.Endpoint)
	}

	if config.ExportIntervalSeconds != nil {
		state.ExportIntervalSeconds = types.Int64Value(int64(*config.ExportIntervalSeconds))
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// OTLPConfigResponse represents the OTLP config API response
type OTLPConfigResponse struct {
	Enabled                bool                    `json:"otlp_enabled"`
	Endpoint               *string                 `json:"otlp_endpoint"`
	Headers                map[string]interface{}  `json:"otlp_headers"`
	ExportIntervalSeconds  *int                    `json:"otlp_export_interval_seconds"`
}

// getOTLPConfig fetches the current OTLP configuration
func (r *organizationOTLPResource) getOTLPConfig(ctx context.Context) (*OTLPConfigResponse, error) {
	data, err := r.client.DoRequest("GET", "/v1/org/otlp", nil)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Data OTLPConfigResponse `json:"data"`
	}
	if err := client.UnmarshalAPIResponse(data, &resp); err != nil {
		return nil, err
	}

	return &resp.Data, nil
}
