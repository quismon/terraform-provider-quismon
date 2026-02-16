package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/quismon/terraform-provider-quismon/internal/client"
)

var (
	_ resource.Resource                = &notificationChannelResource{}
	_ resource.ResourceWithConfigure   = &notificationChannelResource{}
	_ resource.ResourceWithImportState = &notificationChannelResource{}
)

func NewNotificationChannelResource() resource.Resource {
	return &notificationChannelResource{}
}

type notificationChannelResource struct {
	client *client.Client
}

type notificationChannelResourceModel struct {
	ID        types.String `tfsdk:"id"`
	OrgID     types.String `tfsdk:"org_id"`
	Name      types.String `tfsdk:"name"`
	Type      types.String `tfsdk:"type"`
	Config    types.Map    `tfsdk:"config"`
	Enabled   types.Bool   `tfsdk:"enabled"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

func (r *notificationChannelResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_notification_channel"
}

func (r *notificationChannelResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Quismon notification channel.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Channel ID.",
				Computed:    true,
			},
			"org_id": schema.StringAttribute{
				Description: "Organization ID.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Channel name.",
				Required:    true,
			},
			"type": schema.StringAttribute{
				Description: "Channel type: email, webhook, ntfy, slack, or pagerduty.",
				Required:    true,
			},
			"config": schema.MapAttribute{
				Description: "Channel-specific configuration.",
				Required:    true,
				ElementType: types.StringType,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the channel is enabled.",
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

func (r *notificationChannelResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *notificationChannelResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan notificationChannelResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert config map to map[string]interface{}
	configMap := make(map[string]interface{})
	for key, value := range plan.Config.Elements() {
		if strVal, ok := value.(types.String); ok {
			configMap[key] = strVal.ValueString()
		}
	}

	createReq := client.CreateNotificationChannelRequest{
		Name:    plan.Name.ValueString(),
		Type:    plan.Type.ValueString(),
		Config:  configMap,
		Enabled: plan.Enabled.ValueBool(),
	}

	channel, err := r.client.CreateNotificationChannel(createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Notification Channel", err.Error())
		return
	}

	plan.ID = types.StringValue(channel.ID)
	plan.OrgID = types.StringValue(channel.OrgID)
	plan.CreatedAt = types.StringValue(channel.CreatedAt)
	plan.UpdatedAt = types.StringValue(channel.UpdatedAt)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *notificationChannelResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state notificationChannelResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	channel, err := r.client.GetNotificationChannel(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Notification Channel", err.Error())
		return
	}

	state.OrgID = types.StringValue(channel.OrgID)
	state.Name = types.StringValue(channel.Name)
	state.Type = types.StringValue(channel.Type)
	state.Enabled = types.BoolValue(channel.Enabled)
	state.CreatedAt = types.StringValue(channel.CreatedAt)
	state.UpdatedAt = types.StringValue(channel.UpdatedAt)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *notificationChannelResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan notificationChannelResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert config map to map[string]interface{}
	configMap := make(map[string]interface{})
	for key, value := range plan.Config.Elements() {
		if strVal, ok := value.(types.String); ok {
			configMap[key] = strVal.ValueString()
		}
	}

	name := plan.Name.ValueString()
	enabled := plan.Enabled.ValueBool()

	updateReq := client.UpdateNotificationChannelRequest{
		Name:    &name,
		Config:  &configMap,
		Enabled: &enabled,
	}

	channel, err := r.client.UpdateNotificationChannel(plan.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Updating Notification Channel", err.Error())
		return
	}

	plan.UpdatedAt = types.StringValue(channel.UpdatedAt)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *notificationChannelResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state notificationChannelResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteNotificationChannel(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Deleting Notification Channel", err.Error())
		return
	}
}

func (r *notificationChannelResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
