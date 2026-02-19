package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/quismon/terraform-provider-quismon/internal/client"
)

var (
	_ resource.Resource                = &alertRuleResource{}
	_ resource.ResourceWithConfigure   = &alertRuleResource{}
	_ resource.ResourceWithImportState = &alertRuleResource{}
)

func NewAlertRuleResource() resource.Resource {
	return &alertRuleResource{}
}

type alertRuleResource struct {
	client *client.Client
}

type alertRuleResourceModel struct {
	ID                     types.String `tfsdk:"id"`
	CheckID                types.String `tfsdk:"check_id"`
	Name                   types.String `tfsdk:"name"`
	Condition              types.Map    `tfsdk:"condition"`
	NotificationChannelIDs types.List   `tfsdk:"notification_channel_ids"`
	Enabled                types.Bool   `tfsdk:"enabled"`
	CreatedAt              types.String `tfsdk:"created_at"`
	UpdatedAt              types.String `tfsdk:"updated_at"`
}

func (r *alertRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_alert_rule"
}

func (r *alertRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Quismon alert rule.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Alert rule ID.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"check_id": schema.StringAttribute{
				Description: "ID of the check to monitor. Changing this will force recreation of the alert rule.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Alert rule name.",
				Required:    true,
			},
			"condition": schema.MapAttribute{
				Description: "Condition that triggers the alert. Examples: {\"health_status\": \"down\"}, {\"failure_threshold\": \"3\"}, {\"response_time_ms\": \"5000\"}. Values must be strings.",
				Required:    true,
				ElementType: types.StringType,
			},
			"notification_channel_ids": schema.ListAttribute{
				Description: "List of notification channel IDs.",
				Required:    true,
				ElementType: types.StringType,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the alert rule is enabled.",
				Optional:    true,
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

func (r *alertRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *alertRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan alertRuleResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var channelIDs []string
	diags = plan.NotificationChannelIDs.ElementsAs(ctx, &channelIDs, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert condition map to map[string]interface{}
	conditionMap := make(map[string]interface{})
	for key, value := range plan.Condition.Elements() {
		if strVal, ok := value.(types.String); ok {
			conditionMap[key] = strVal.ValueString()
		}
	}

	createReq := client.CreateAlertRuleRequest{
		Name:                   plan.Name.ValueString(),
		Condition:              conditionMap,
		NotificationChannelIDs: channelIDs,
		Enabled:                plan.Enabled.ValueBool(),
	}

	rule, err := r.client.CreateAlertRule(plan.CheckID.ValueString(), createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Alert Rule", err.Error())
		return
	}

	plan.ID = types.StringValue(rule.ID)
	plan.CreatedAt = types.StringValue(rule.CreatedAt)
	plan.UpdatedAt = types.StringValue(rule.UpdatedAt)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *alertRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state alertRuleResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	rule, err := r.client.GetAlertRule(state.CheckID.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Alert Rule", err.Error())
		return
	}

	state.Name = types.StringValue(rule.Name)
	state.Enabled = types.BoolValue(rule.Enabled)
	state.CreatedAt = types.StringValue(rule.CreatedAt)
	state.UpdatedAt = types.StringValue(rule.UpdatedAt)

	// Convert condition map to Terraform Map type (string values)
	conditionStrMap := make(map[string]string)
	for k, v := range rule.Condition {
		conditionStrMap[k] = fmt.Sprintf("%v", v)
	}
	conditionMap, diags := types.MapValueFrom(ctx, types.StringType, conditionStrMap)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		resp.Diagnostics.AddError("Error Converting Condition", "Failed to convert condition map")
		return
	}
	state.Condition = conditionMap

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *alertRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan alertRuleResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state alertRuleResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var channelIDs []string
	diags = plan.NotificationChannelIDs.ElementsAs(ctx, &channelIDs, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build update request with only changed fields
	updateReq := client.UpdateAlertRuleRequest{}

	if plan.Name.ValueString() != state.Name.ValueString() {
		name := plan.Name.ValueString()
		updateReq.Name = &name
	}

	// Check if condition changed
	if !plan.Condition.Equal(state.Condition) {
		conditionMap := make(map[string]interface{})
		for key, value := range plan.Condition.Elements() {
			if strVal, ok := value.(types.String); ok {
				conditionMap[key] = strVal.ValueString()
			}
		}
		updateReq.Condition = &conditionMap
	}

	// Check if notification channels changed
	if !plan.NotificationChannelIDs.Equal(state.NotificationChannelIDs) {
		updateReq.NotificationChannelIDs = &channelIDs
	}

	// Check if enabled changed
	if plan.Enabled.ValueBool() != state.Enabled.ValueBool() {
		enabled := plan.Enabled.ValueBool()
		updateReq.Enabled = &enabled
	}

	// Perform update
	rule, err := r.client.UpdateAlertRule(state.CheckID.ValueString(), state.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Updating Alert Rule", err.Error())
		return
	}

	// Update state with response
	plan.ID = types.StringValue(rule.ID)
	plan.CreatedAt = types.StringValue(rule.CreatedAt)
	plan.UpdatedAt = types.StringValue(rule.UpdatedAt)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *alertRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state alertRuleResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteAlertRule(state.CheckID.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Deleting Alert Rule", err.Error())
		return
	}
}

func (r *alertRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: check_id:rule_id
	parts := strings.Split(req.ID, ":")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			"Expected import ID in format: check_id:rule_id",
		)
		return
	}

	checkID := parts[0]
	ruleID := parts[1]

	// Set both check_id and id in state
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("check_id"), checkID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), ruleID)...)
}
