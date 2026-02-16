package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/quismon/terraform-provider-quismon/internal/client"
)

var _ datasource.DataSource = &notificationChannelDataSource{}

func NewNotificationChannelDataSource() datasource.DataSource {
	return &notificationChannelDataSource{}
}

type notificationChannelDataSource struct {
	client *client.Client
}

func (d *notificationChannelDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_notification_channel"
}

func (d *notificationChannelDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a Quismon notification channel by name.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "Channel name.",
				Required:    true,
			},
			"id": schema.StringAttribute{
				Description: "Channel ID.",
				Computed:    true,
			},
			"type": schema.StringAttribute{
				Description: "Channel type.",
				Computed:    true,
			},
		},
	}
}

func (d *notificationChannelDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*client.Client)
}

func (d *notificationChannelDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data struct {
		Name types.String `tfsdk:"name"`
		ID   types.String `tfsdk:"id"`
		Type types.String `tfsdk:"type"`
	}

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	channel, err := d.client.GetNotificationChannelByName(data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Notification Channel", err.Error())
		return
	}

	data.ID = types.StringValue(channel.ID)
	data.Type = types.StringValue(channel.Type)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}
