package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/quismon/terraform-provider-quismon/internal/client"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &checkResource{}
	_ resource.ResourceWithConfigure   = &checkResource{}
	_ resource.ResourceWithImportState = &checkResource{}
)

// NewCheckResource is a helper function to simplify the provider implementation.
func NewCheckResource() resource.Resource {
	return &checkResource{}
}

// checkResource is the resource implementation.
type checkResource struct {
	client *client.Client
}

// checkResourceModel maps the resource schema data.
type checkResourceModel struct {
	ID                  types.String `tfsdk:"id"`
	OrgID               types.String `tfsdk:"org_id"`
	Name                types.String `tfsdk:"name"`
	Type                types.String `tfsdk:"type"`
	Config              types.Map    `tfsdk:"config"`
	ConfigJSON          types.String `tfsdk:"config_json"`
	IntervalSeconds     types.Int64  `tfsdk:"interval_seconds"`
	Regions             types.List   `tfsdk:"regions"`
	Enabled             types.Bool   `tfsdk:"enabled"`
	SimultaneousRegions types.Bool   `tfsdk:"simultaneous_regions"`
	RecheckOnFailure    types.Bool   `tfsdk:"recheck_on_failure"`
	IaCLocked           types.Bool   `tfsdk:"iac_locked"`
	HealthStatus        types.String `tfsdk:"health_status"`
	LastChecked         types.String `tfsdk:"last_checked"`
	CreatedAt           types.String `tfsdk:"created_at"`
	UpdatedAt           types.String `tfsdk:"updated_at"`
}

// Metadata returns the resource type name.
func (r *checkResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_check"
}

// Schema defines the schema for the resource.
func (r *checkResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Quismon health check.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Check ID.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"org_id": schema.StringAttribute{
				Description: "Organization ID.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Check name.",
				Required:    true,
			},
			"type": schema.StringAttribute{
				Description: "Check type: http, https, tcp, ping, udp, dns, ssl, multistep, smtp-imap, throughput, or http3.",
				Required:    true,
			},
			"config": schema.MapAttribute{
				Description: "Check-specific configuration (for simple types). Use config_json for complex nested configs like multistep.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"config_json": schema.StringAttribute{
				Description: "Check configuration as JSON string (required for multistep, smtp-imap, and other complex configs). Use jsonencode() to create this.",
				Optional:    true,
			},
			"interval_seconds": schema.Int64Attribute{
				Description: "Check interval in seconds (minimum 60).",
				Required:    true,
			},
			"regions": schema.ListAttribute{
				Description: "Monitoring regions.",
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Default:     listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{types.StringValue("us-east-1")})),
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the check is enabled.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"simultaneous_regions": schema.BoolAttribute{
				Description: "If true, all regional checks execute simultaneously. If false (default), regional checks are staggered to avoid rate limiting.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"recheck_on_failure": schema.BoolAttribute{
				Description: "If true, failed checks trigger an immediate recheck from a different region to verify the failure before alerting.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"iac_locked": schema.BoolAttribute{
				Description: "If true, this check can only be modified via API (prevents web UI changes).",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"health_status": schema.StringAttribute{
				Description: "Current health status: healthy, unhealthy, or unknown.",
				Computed:    true,
			},
			"last_checked": schema.StringAttribute{
				Description: "Last check timestamp.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "Creation timestamp.",
				Computed:    true,
			},
			"updated_at": schema.StringAttribute{
				Description: "Last update timestamp.",
				Computed:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *checkResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create creates the resource and sets the initial Terraform state.
func (r *checkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan checkResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build config from either config map or config_json
	var configMap map[string]interface{}

	if !plan.ConfigJSON.IsNull() && plan.ConfigJSON.ValueString() != "" {
		// Use config_json (for complex configs like multistep)
		if err := json.Unmarshal([]byte(plan.ConfigJSON.ValueString()), &configMap); err != nil {
			resp.Diagnostics.AddError(
				"Error Parsing config_json",
				"Could not parse config_json as JSON: "+err.Error(),
			)
			return
		}
	} else if !plan.Config.IsNull() {
		// Use config map (for simple configs)
		configMap = make(map[string]interface{})
		for key, value := range plan.Config.Elements() {
			if strVal, ok := value.(types.String); ok {
				configMap[key] = strVal.ValueString()
			}
		}
	} else {
		resp.Diagnostics.AddError(
			"Missing Configuration",
			"Either 'config' or 'config_json' must be specified",
		)
		return
	}

	// Convert regions list to []string
	var regions []string
	diags = plan.Regions.ElementsAs(ctx, &regions, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create the check
	createReq := client.CreateCheckRequest{
		Name:                plan.Name.ValueString(),
		Type:                plan.Type.ValueString(),
		Config:              configMap,
		IntervalSeconds:     int(plan.IntervalSeconds.ValueInt64()),
		Regions:             regions,
		Enabled:             plan.Enabled.ValueBool(),
		SimultaneousRegions: plan.SimultaneousRegions.ValueBoolPointer(),
		RecheckOnFailure:    plan.RecheckOnFailure.ValueBoolPointer(),
	}

	check, err := r.client.CreateCheck(createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Check",
			"Could not create check, unexpected error: "+err.Error(),
		)
		return
	}

	// Map response to state
	plan.ID = types.StringValue(check.ID)
	plan.OrgID = types.StringValue(check.OrgID)
	plan.SimultaneousRegions = types.BoolValue(check.SimultaneousRegions)
	plan.RecheckOnFailure = types.BoolValue(check.RecheckOnFailure)
	plan.HealthStatus = types.StringValue(check.HealthStatus)
	if check.LastChecked != nil {
		plan.LastChecked = types.StringValue(*check.LastChecked)
	} else {
		plan.LastChecked = types.StringNull()
	}
	plan.CreatedAt = types.StringValue(check.CreatedAt)
	plan.UpdatedAt = types.StringValue(check.UpdatedAt)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Read refreshes the Terraform state with the latest data.
func (r *checkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state checkResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	check, err := r.client.GetCheck(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Check",
			"Could not read check ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Map response to state
	state.OrgID = types.StringValue(check.OrgID)
	state.Name = types.StringValue(check.Name)
	state.Type = types.StringValue(check.Type)
	state.IntervalSeconds = types.Int64Value(int64(check.IntervalSeconds))
	state.Enabled = types.BoolValue(check.Enabled)
	state.SimultaneousRegions = types.BoolValue(check.SimultaneousRegions)
	state.RecheckOnFailure = types.BoolValue(check.RecheckOnFailure)
	state.HealthStatus = types.StringValue(check.HealthStatus)
	if check.LastChecked != nil {
		state.LastChecked = types.StringValue(*check.LastChecked)
	} else {
		state.LastChecked = types.StringNull()
	}
	state.CreatedAt = types.StringValue(check.CreatedAt)
	state.UpdatedAt = types.StringValue(check.UpdatedAt)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *checkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan checkResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build config from either config map or config_json
	var configMap map[string]interface{}

	if !plan.ConfigJSON.IsNull() && plan.ConfigJSON.ValueString() != "" {
		// Use config_json (for complex configs like multistep)
		if err := json.Unmarshal([]byte(plan.ConfigJSON.ValueString()), &configMap); err != nil {
			resp.Diagnostics.AddError(
				"Error Parsing config_json",
				"Could not parse config_json as JSON: "+err.Error(),
			)
			return
		}
	} else if !plan.Config.IsNull() {
		// Use config map (for simple configs)
		configMap = make(map[string]interface{})
		for key, value := range plan.Config.Elements() {
			if strVal, ok := value.(types.String); ok {
				configMap[key] = strVal.ValueString()
			}
		}
	}

	var regions []string
	diags = plan.Regions.ElementsAs(ctx, &regions, false)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Update the check
	name := plan.Name.ValueString()
	checkType := plan.Type.ValueString()
	intervalSeconds := int(plan.IntervalSeconds.ValueInt64())
	enabled := plan.Enabled.ValueBool()
	simultaneousRegions := plan.SimultaneousRegions.ValueBool()
	recheckOnFailure := plan.RecheckOnFailure.ValueBool()

	updateReq := client.UpdateCheckRequest{
		Name:                &name,
		Type:                &checkType,
		Config:              &configMap,
		IntervalSeconds:     &intervalSeconds,
		Regions:             &regions,
		Enabled:             &enabled,
		SimultaneousRegions: &simultaneousRegions,
		RecheckOnFailure:    &recheckOnFailure,
	}

	check, err := r.client.UpdateCheck(plan.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Check",
			"Could not update check, unexpected error: "+err.Error(),
		)
		return
	}

	// Map response to state
	plan.OrgID = types.StringValue(check.OrgID)
	plan.SimultaneousRegions = types.BoolValue(check.SimultaneousRegions)
	plan.RecheckOnFailure = types.BoolValue(check.RecheckOnFailure)
	plan.HealthStatus = types.StringValue(check.HealthStatus)
	if check.LastChecked != nil {
		plan.LastChecked = types.StringValue(*check.LastChecked)
	} else {
		plan.LastChecked = types.StringNull()
	}
	plan.CreatedAt = types.StringValue(check.CreatedAt)
	plan.UpdatedAt = types.StringValue(check.UpdatedAt)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *checkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state checkResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteCheck(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Check",
			"Could not delete check, unexpected error: "+err.Error(),
		)
		return
	}
}

// ImportState imports the resource state.
func (r *checkResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
