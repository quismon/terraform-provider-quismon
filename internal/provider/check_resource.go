package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/quismon/terraform-provider-quismon/internal/client"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                   = &checkResource{}
	_ resource.ResourceWithConfigure      = &checkResource{}
	_ resource.ResourceWithImportState    = &checkResource{}
	_ resource.ResourceWithValidateConfig = &checkResource{}
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
	ID                        types.String `tfsdk:"id"`
	OrgID                     types.String `tfsdk:"org_id"`
	Name                      types.String `tfsdk:"name"`
	Type                      types.String `tfsdk:"type"`
	Config                    types.Map    `tfsdk:"config"`
	ConfigJSON                types.String `tfsdk:"config_json"`
	ConfigHash                types.String `tfsdk:"config_hash"`
	IntervalSeconds           types.Int64  `tfsdk:"interval_seconds"`
	Regions                   types.Set    `tfsdk:"regions"`
	RegionThreshold           types.Int64  `tfsdk:"region_threshold"`
	Enabled                   types.Bool   `tfsdk:"enabled"`
	Inverted                  types.Bool   `tfsdk:"inverted"`
	SimultaneousRegions       types.Bool   `tfsdk:"simultaneous_regions"`
	RecheckOnFailure          types.Bool   `tfsdk:"recheck_on_failure"`
	AlertAfterFirstSuccess    types.Bool   `tfsdk:"alert_after_first_success"`
	StartsBroken              types.Bool   `tfsdk:"starts_broken"`
	StartsBrokenArrivalNotice types.Bool   `tfsdk:"starts_broken_arrival_notice"`
	StartsBrokenDeadline      types.Int64  `tfsdk:"starts_broken_deadline_seconds"`
	ShowOnStatusPage          types.Bool   `tfsdk:"show_on_status_page"`
	ExpiresAfterSeconds       types.Int64  `tfsdk:"expires_after_seconds"`
	ResultWebhook             types.String `tfsdk:"result_webhook"`
	DependsOn                 types.Set    `tfsdk:"check_dependencies"`
	IaCLocked                 types.Bool   `tfsdk:"iac_locked"`
	HealthStatus              types.String `tfsdk:"health_status"`
	LastChecked               types.String `tfsdk:"last_checked"`
	CreatedAt                 types.String `tfsdk:"created_at"`
	UpdatedAt                 types.String `tfsdk:"updated_at"`
}

// Metadata returns the resource type name.
// ValidateConfig is the PLAN-TIME vehicle for the cross-attribute bound
// region_threshold <= length(regions) (Q-059, review A-6/M1: in-handler
// validators run at apply time and cannot back a plan-time guarantee).
// Never clamps - clamping is a console courtesy, not the contract. When
// regions are unknown at plan time the check defers and the API's
// apply-time 400 surfaces verbatim.
func (r *checkResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data checkResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if data.RegionThreshold.IsNull() || data.RegionThreshold.IsUnknown() {
		return
	}
	threshold := data.RegionThreshold.ValueInt64()
	if data.Regions.IsUnknown() || data.Regions.IsNull() {
		return // defer to apply-time validation
	}
	regionCount := int64(len(data.Regions.Elements()))
	if regionCount == 0 {
		regionCount = 1 // the schema default resolves one region
	}
	if threshold > regionCount {
		resp.Diagnostics.AddError(
			"Invalid region_threshold",
			fmt.Sprintf("region_threshold (%d) cannot exceed the number of regions (%d).", threshold, regionCount),
		)
	}
}

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
				Description: "Check type: http, https, tcp, ping, udp, dns, dnssec, ssl, multistep, smtp-imap, throughput, http3, spf, dkim, dmarc, playwright, traceroute, dom (alias: render). CREATE-ONLY: the API ignores type on update — changing it forces provider replacement.",
				Required:    true,
			},
			"config": schema.MapAttribute{
				Description: "Check-specific configuration (for simple types). Use config_json for complex nested configs like multistep. Password fields (smtp_password, imap_password, password) are sensitive and cannot be re-read from the API.",
				Optional:    true,
				Sensitive:   true,
				ElementType: types.StringType,
			},
			"config_json": schema.StringAttribute{
				Description: "Check configuration as JSON string (required for multistep, smtp-imap, and other complex configs). Use jsonencode() to create this. Password fields are sensitive and cannot be re-read from the API.",
				Optional:    true,
				Sensitive:   true,
			},
			"config_hash": schema.StringAttribute{
				Description: "Hash of sensitive config fields for drift detection. Use this to detect if passwords have changed externally.",
				Computed:    true,
			},
			"interval_seconds": schema.Int64Attribute{
				Description: "Check interval in seconds (minimum 60).",
				Required:    true,
			},
			"regions": schema.SetAttribute{
				Description: "Monitoring regions (set - order does not matter, duplicates not allowed).",
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Default:     setdefault.StaticValue(types.SetValueMust(types.StringType, []attr.Value{types.StringValue("na-east-ewr")})),
			},
			"region_threshold": schema.Int64Attribute{
				Description: "Multi-region consensus: the check is unhealthy only when at least N of its M regions " +
					"report failure. Default (and omitted) is 1 - any region fails the check. Must be >= 1 and " +
					"<= the number of regions (validated against the FINAL region set: free/cafe/adopter premium " +
					"regions resolve to shared primary regions, which can shrink the count - such shapes surface " +
					"the API's apply-time error). ALWAYS sent on update: removing the attribute from " +
					"configuration resets the threshold to 1 (codify your threshold first when upgrading).",
				Optional:   true,
				Computed:   true,
				Default:    int64default.StaticInt64(1),
				Validators: []validator.Int64{int64validator.AtLeast(1)},
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the check is enabled.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"inverted": schema.BoolAttribute{
				Description: "If true, alerts on success instead of failure. Useful for firewall validation - alert when a blocked port opens.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
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
			"alert_after_first_success": schema.BoolAttribute{
				Description:        "Deprecated alias of starts_broken. Either attribute sets the same stored flag.",
				Optional:           true,
				Computed:           true,
				Default:            booldefault.StaticBool(false),
				DeprecationMessage: "Use starts_broken instead.",
			},
			"starts_broken": schema.BoolAttribute{
				Description: "Starts-broken checks (TDD workflow): when true, the check does not alert (no incidents, no notifications) until it has succeeded at least once; an arrival notice is sent through the check's alert-rule channels on that first success. Create-only: once the check first succeeds, the platform silently graduates the check to a normal configuration (the stored flag clears) and this attribute is ignored on updates - remove it from your configuration after the first successful apply to keep plans clean.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"starts_broken_arrival_notice": schema.BoolAttribute{
				Description: "When true (default), the first success of a starts-broken check sends the arrival notice through the check's alert-rule channels. Has no effect unless starts_broken is true.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"starts_broken_deadline_seconds": schema.Int64Attribute{
				Description: "Optional create-only \"never came up\" guard: if the check's first success hasn't happened within this many seconds of creation, starts-broken suppression lifts, a one-time deadline-breach escalation fires, and failure alerts flow. Only meaningful with starts_broken. Create-only.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(0),
			},
			"show_on_status_page": schema.BoolAttribute{
				Description: "If true, this check contributes to the public status page. Default is false (opt-in).",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"check_dependencies": schema.SetAttribute{
				Description: "List of check IDs that must be healthy before this check runs. If any dependency is unhealthy, this check is skipped with 'dependency_failed' status.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"expires_after_seconds": schema.Int64Attribute{
				Description: "Check auto-deletes after this many seconds. NULL or 0 means no expiration. Note: expiring checks are typically created via API for temporary monitoring, not via Terraform.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					ExpiresAfterSecondsModifier(),
				},
			},
			"result_webhook": schema.StringAttribute{
				Description: "Per-check result webhook: every result (success and failure, per region) POSTs " +
					"to this URL. Delivery is fire-once per result with no signature and no redelivery; " +
					"responses >= 400 are log-only. Absolute http(s) URL (tier-dependent private-address " +
					"rules are enforced server-side). May embed a token - treat as secret. REMOVING this " +
					"attribute clears the stored webhook (an explicit empty string is sent).",
				Optional:   true,
				Sensitive:  true,
				Validators: []validator.String{resultWebhookValidator{}},
			},
			"iac_locked": schema.BoolAttribute{
				DeprecationMessage: "This attribute never had any effect. Remove it from your configuration; it will be dropped in a future major release.",
				Description:        "Deprecated: never had any effect (nothing was transmitted or enforced). Remove it from your configuration.",
				Optional:           true,
				Computed:           true,
				Default:            booldefault.StaticBool(false),
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

	// Validate Playwright configuration at plan time
	if plan.Type.ValueString() == "playwright" {
		if err := validatePlaywrightConfig(configMap); err != nil {
			resp.Diagnostics.AddError(
				"Invalid Playwright Configuration",
				err.Error(),
			)
			return
		}
	}

	// Validate dom configuration at plan time (script size + script/url presence)
	if plan.Type.ValueString() == "dom" || plan.Type.ValueString() == "render" {
		if err := validateDomConfig(configMap); err != nil {
			resp.Diagnostics.AddError(
				"Invalid Dom Configuration",
				err.Error(),
			)
			return
		}
	}

	// Convert regions list to []string
	var regions []string
	diags = plan.Regions.ElementsAs(ctx, &regions, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert depends_on list to []string
	var dependsOn []string
	if !plan.DependsOn.IsNull() {
		diags = plan.DependsOn.ElementsAs(ctx, &dependsOn, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Create the check
	createReq := client.CreateCheckRequest{
		Name:                      plan.Name.ValueString(),
		Type:                      plan.Type.ValueString(),
		Config:                    configMap,
		IntervalSeconds:           int(plan.IntervalSeconds.ValueInt64()),
		Regions:                   regions,
		RegionThreshold:           regionThresholdPtr(plan.RegionThreshold),
		ResultWebhook:             plan.ResultWebhook.ValueString(),
		Enabled:                   plan.Enabled.ValueBool(),
		Inverted:                  plan.Inverted.ValueBoolPointer(),
		SimultaneousRegions:       plan.SimultaneousRegions.ValueBoolPointer(),
		RecheckOnFailure:          plan.RecheckOnFailure.ValueBoolPointer(),
		AlertAfterFirstSuccess:    mergeStartsBrokenAlias(plan.StartsBroken, plan.AlertAfterFirstSuccess),
		StartsBrokenArrivalNotice: plan.StartsBrokenArrivalNotice.ValueBoolPointer(),
		ShowOnStatusPage:          plan.ShowOnStatusPage.ValueBoolPointer(),
		DependsOn:                 dependsOn,
	}

	// Only set starts_broken_deadline_seconds if it's explicitly set (non-zero)
	if !plan.StartsBrokenDeadline.IsNull() && plan.StartsBrokenDeadline.ValueInt64() > 0 {
		deadline := int(plan.StartsBrokenDeadline.ValueInt64())
		createReq.StartsBrokenDeadlineSeconds = &deadline
	}

	// Only set expires_after_seconds if it's explicitly set (non-zero)
	if !plan.ExpiresAfterSeconds.IsNull() && plan.ExpiresAfterSeconds.ValueInt64() > 0 {
		expiresAfter := int(plan.ExpiresAfterSeconds.ValueInt64())
		createReq.ExpiresAfterSeconds = &expiresAfter
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
	plan.ConfigHash = types.StringValue(check.ConfigHash)
	plan.Inverted = types.BoolValue(check.Inverted)
	plan.SimultaneousRegions = types.BoolValue(check.SimultaneousRegions)
	plan.RecheckOnFailure = types.BoolValue(check.RecheckOnFailure)
	plan.RegionThreshold = types.Int64Value(int64(check.RegionThreshold))
	if check.ResultWebhook != "" {
		plan.ResultWebhook = types.StringValue(check.ResultWebhook)
	} else {
		plan.ResultWebhook = types.StringNull()
	}
	plan.AlertAfterFirstSuccess = types.BoolValue(check.AlertAfterFirstSuccess)
	plan.StartsBroken = types.BoolValue(check.StartsBroken)
	plan.StartsBrokenArrivalNotice = types.BoolValue(check.StartsBrokenArrivalNotice)
	plan.ShowOnStatusPage = types.BoolValue(check.ShowOnStatusPage)
	if len(check.DependsOn) > 0 {
		dependsOnSet, _ := types.SetValueFrom(ctx, types.StringType, check.DependsOn)
		plan.DependsOn = dependsOnSet
	} else {
		plan.DependsOn = types.SetNull(types.StringType)
	}
	if check.ExpiresAfterSeconds != nil {
		plan.ExpiresAfterSeconds = types.Int64Value(int64(*check.ExpiresAfterSeconds))
	} else {
		plan.ExpiresAfterSeconds = types.Int64Null()
	}
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

	// Store the previous config_hash for drift detection
	previousConfigHash := state.ConfigHash.ValueString()

	check, err := r.client.GetCheck(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Check",
			"Could not read check ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Check for config drift via hash comparison
	// If the hash changed externally, the sensitive config fields have been modified
	// outside of Terraform (e.g., via API or dashboard)
	if previousConfigHash != "" && check.ConfigHash != "" && previousConfigHash != check.ConfigHash {
		resp.Diagnostics.AddWarning(
			"Configuration Drift Detected",
			"The check's sensitive configuration (passwords) has been modified externally (via API or dashboard). "+
				"The config_hash changed from "+previousConfigHash+" to "+check.ConfigHash+". "+
				"To reconcile, either update your Terraform configuration with the new values, "+
				"or re-apply to reset the check to your Terraform-defined configuration.",
		)
	}

	// Map response to state
	state.OrgID = types.StringValue(check.OrgID)
	state.Name = types.StringValue(check.Name)
	state.Type = types.StringValue(check.Type)
	state.ConfigHash = types.StringValue(check.ConfigHash)
	state.IntervalSeconds = types.Int64Value(int64(check.IntervalSeconds))
	state.Enabled = types.BoolValue(check.Enabled)
	state.Inverted = types.BoolValue(check.Inverted)
	state.SimultaneousRegions = types.BoolValue(check.SimultaneousRegions)
	state.RecheckOnFailure = types.BoolValue(check.RecheckOnFailure)
	state.RegionThreshold = types.Int64Value(int64(check.RegionThreshold))
	state.AlertAfterFirstSuccess = types.BoolValue(check.AlertAfterFirstSuccess)
	state.StartsBroken = types.BoolValue(check.StartsBroken)
	state.StartsBrokenArrivalNotice = types.BoolValue(check.StartsBrokenArrivalNotice)
	state.ShowOnStatusPage = types.BoolValue(check.ShowOnStatusPage)
	if len(check.DependsOn) > 0 {
		dependsOnSet, _ := types.SetValueFrom(ctx, types.StringType, check.DependsOn)
		state.DependsOn = dependsOnSet
	} else {
		state.DependsOn = types.SetNull(types.StringType)
	}
	if check.ExpiresAfterSeconds != nil {
		state.ExpiresAfterSeconds = types.Int64Value(int64(*check.ExpiresAfterSeconds))
	} else {
		state.ExpiresAfterSeconds = types.Int64Null()
	}
	if check.ResultWebhook != "" {
		state.ResultWebhook = types.StringValue(check.ResultWebhook)
	} else {
		state.ResultWebhook = types.StringNull()
	}

	// Regions are non-sensitive and must be tracked for drift detection
	if len(check.Regions) > 0 {
		regionsSet, _ := types.SetValueFrom(ctx, types.StringType, check.Regions)
		state.Regions = regionsSet
	}

	// Volatile telemetry (health_status, last_checked, updated_at, created_at)
	// changes on every check run - only populate when absent (import), never
	// overwrite on refresh, otherwise every plan shows perpetual drift.
	if state.HealthStatus.IsNull() {
		state.HealthStatus = types.StringValue(check.HealthStatus)
	}
	if state.LastChecked.IsNull() && check.LastChecked != nil {
		state.LastChecked = types.StringValue(*check.LastChecked)
	}
	if state.UpdatedAt.IsNull() {
		state.UpdatedAt = types.StringValue(check.UpdatedAt)
	}
	if state.CreatedAt.IsNull() {
		state.CreatedAt = types.StringValue(check.CreatedAt)
	}

	// Note: We intentionally do NOT update state.Config or state.ConfigJSON from the API
	// because sensitive fields (passwords) are returned as ***REDACTED***.
	// We preserve the user's Terraform-defined configuration in state.
	// Drift detection is handled via config_hash comparison above.

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

	// Validate Playwright configuration at plan time (if config is being updated)
	if plan.Type.ValueString() == "playwright" && len(configMap) > 0 {
		if err := validatePlaywrightConfig(configMap); err != nil {
			resp.Diagnostics.AddError(
				"Invalid Playwright Configuration",
				err.Error(),
			)
			return
		}
	}

	// Validate dom configuration at plan time (if config is being updated)
	if (plan.Type.ValueString() == "dom" || plan.Type.ValueString() == "render") && len(configMap) > 0 {
		if err := validateDomConfig(configMap); err != nil {
			resp.Diagnostics.AddError(
				"Invalid Dom Configuration",
				err.Error(),
			)
			return
		}
	}

	var regions []string
	diags = plan.Regions.ElementsAs(ctx, &regions, false)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Convert depends_on list to []string
	var dependsOn []string
	if !plan.DependsOn.IsNull() {
		diags = plan.DependsOn.ElementsAs(ctx, &dependsOn, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Update the check
	name := plan.Name.ValueString()
	intervalSeconds := int(plan.IntervalSeconds.ValueInt64())
	enabled := plan.Enabled.ValueBool()
	inverted := plan.Inverted.ValueBool()
	simultaneousRegions := plan.SimultaneousRegions.ValueBool()
	recheckOnFailure := plan.RecheckOnFailure.ValueBool()
	showOnStatusPage := plan.ShowOnStatusPage.ValueBool()

	// Q-061: type is CREATE-ONLY — the API's UpdateCheckRequest drops it, so
	// sending it produced perpetual refresh-revert diffs on type changes.
	updateReq := client.UpdateCheckRequest{
		Name:            &name,
		Config:          &configMap,
		IntervalSeconds: &intervalSeconds,
		Regions:         &regions,
		// ALWAYS sent (Q-059 anti-bricking): a concrete pointer keeps the
		// server's threshold-vs-final-regions guard on every update path;
		// omitting it would preserve a stored threshold that a regions
		// shrink just invalidated (the 2026-10-27 bricking class).
		RegionThreshold: regionThresholdPtr(plan.RegionThreshold),
		// null CLEARS: pointer-to-empty is the API's documented clear
		// semantics; nil would preserve and make the webhook unremovable.
		ResultWebhook:       resultWebhookPtr(plan.ResultWebhook),
		Enabled:             &enabled,
		Inverted:            &inverted,
		SimultaneousRegions: &simultaneousRegions,
		RecheckOnFailure:    &recheckOnFailure,
		// alert_after_first_success is intentionally NOT sent on update:
		// it is create-only. The platform clears the stored flag after the
		// check's first success (graduation to a normal check) - sending it
		// on update would flip it back and fight the graduation
		// (2026-10-22 starts-broken graduation).
		ShowOnStatusPage: &showOnStatusPage,
		DependsOn:        &dependsOn,
		// Note: We deliberately do NOT set ExpiresAfterSeconds here.
		// Expiring checks are typically temporary and managed via API,
		// not Terraform. We don't want to tamper with them.
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
	plan.ConfigHash = types.StringValue(check.ConfigHash)
	plan.Inverted = types.BoolValue(check.Inverted)
	plan.SimultaneousRegions = types.BoolValue(check.SimultaneousRegions)
	plan.RecheckOnFailure = types.BoolValue(check.RecheckOnFailure)
	plan.AlertAfterFirstSuccess = types.BoolValue(check.AlertAfterFirstSuccess)
	plan.StartsBroken = types.BoolValue(check.StartsBroken)
	plan.StartsBrokenArrivalNotice = types.BoolValue(check.StartsBrokenArrivalNotice)
	plan.ShowOnStatusPage = types.BoolValue(check.ShowOnStatusPage)
	// Only update DependsOn from API if the plan had a non-null value
	// This prevents inconsistency when the API preserves old dependencies
	// but the plan explicitly set check_dependencies to null/empty
	if !plan.DependsOn.IsNull() {
		if len(check.DependsOn) > 0 {
			dependsOnSet, _ := types.SetValueFrom(ctx, types.StringType, check.DependsOn)
			plan.DependsOn = dependsOnSet
		} else {
			plan.DependsOn = types.SetNull(types.StringType)
		}
	}
	// Preserve the expires_after_seconds from API response (read-only for updates)
	if check.ExpiresAfterSeconds != nil {
		plan.ExpiresAfterSeconds = types.Int64Value(int64(*check.ExpiresAfterSeconds))
	} else {
		plan.ExpiresAfterSeconds = types.Int64Null()
	}
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
